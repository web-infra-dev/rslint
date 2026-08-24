package lsp

import (
	"context"
	"fmt"
	"log"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const codeActionKindSourceFixAllRslint = lsproto.CodeActionKind("source.fixAll.rslint")

// ruleFixToTextEdit converts a rule fix into an LSP TextEdit using the
// source file's line map for position encoding.
func ruleFixToTextEdit(sourceFile ast.SourceFileLike, fix rule.RuleFix) *lsproto.TextEdit {
	startLine, startChar := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, fix.Range.Pos())
	endLine, endChar := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, fix.Range.End())
	return &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: uint32(startLine), Character: uint32(startChar)},
			End:   lsproto.Position{Line: uint32(endLine), Character: uint32(endChar)},
		},
		NewText: fix.Text,
	}
}
func (s *Server) handleCodeAction(ctx context.Context, params *lsproto.CodeActionParams) (lsproto.CodeActionResponse, error) {
	log.Printf("Handling codeAction: %+v,%+v", params, ctx)
	uri := params.TextDocument.Uri

	// Handle source.fixAll requests (triggered by editor.codeActionsOnSave)
	if isFixAllRequest(params.Context) {
		return s.handleFixAllCodeAction(ctx, uri)
	}

	// Get stored diagnostics for this document
	ruleDiagnostics, exists := s.diagnostics[uri]
	if !exists {
		return lsproto.CodeActionResponse{
			CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{},
		}, nil
	}

	var codeActions []lsproto.CommandOrCodeAction

	// Find diagnostics that overlap with the requested range
	for _, ruleDiag := range ruleDiagnostics {
		// Check if diagnostic range overlaps with requested range
		diagStartLine, diagStartChar := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.Pos())
		diagEndLine, diagEndChar := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, ruleDiag.Range.End())

		diagRange := lsproto.Range{
			Start: lsproto.Position{Line: uint32(diagStartLine), Character: uint32(diagStartChar)},
			End:   lsproto.Position{Line: uint32(diagEndLine), Character: uint32(diagEndChar)},
		}

		if rangesOverlap(diagRange, params.Range) {
			// Add code action for fixes
			codeAction := createCodeActionFromRuleDiagnostic(ruleDiag, uri)
			if codeAction != nil {
				codeActions = append(codeActions, lsproto.CommandOrCodeAction{
					Command:    nil,
					CodeAction: codeAction,
				})
			}
			// add extract disable rule actions
			disableActions := createDisableRuleActions(ruleDiag, uri)
			codeActions = append(codeActions, disableActions...)

			// Add code actions for suggestions
			if ruleDiag.Suggestions != nil {
				for _, suggestion := range *ruleDiag.Suggestions {
					suggestionAction := createCodeActionFromSuggestion(ruleDiag, suggestion, uri)
					if suggestionAction != nil {
						codeActions = append(codeActions, lsproto.CommandOrCodeAction{
							Command:    nil,
							CodeAction: suggestionAction,
						})
					}
				}
			}
		}
	}

	return lsproto.CodeActionResponse{
		CommandOrCodeActionArray: &codeActions,
	}, nil
}

// isFixAllRequest returns true if the code action context requests source.fixAll actions.
func isFixAllRequest(ctx *lsproto.CodeActionContext) bool {
	if ctx == nil || ctx.Only == nil {
		return false
	}
	for _, kind := range *ctx.Only {
		if kind == lsproto.CodeActionKindSourceFixAll || kind == codeActionKindSourceFixAllRslint {
			return true
		}
	}
	return false
}

// maxFixPasses is the maximum number of lint-fix cycles to prevent infinite loops
// when two rules produce fixes that undo each other.
const maxFixPasses = 10

// handleFixAllCodeAction computes all auto-fixes for the given URI using
// multi-pass fixing: each pass lints and applies fixes to an isolated overlay,
// repeating until no more fixes are found or maxFixPasses is reached.
// This handles cascading fixes (e.g. no-wrapper-object-types fix triggers no-inferrable-types).
// It does NOT push diagnostics or update s.diagnostics — that is left to the
// subsequent didSave handler in the normal save flow.
func (s *Server) handleFixAllCodeAction(ctx context.Context, uri lsproto.DocumentUri) (lsproto.CodeActionResponse, error) {
	empty := lsproto.CodeActionResponse{CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{}}

	// Clear pending debounce for this URI — we are about to lint it fresh,
	// so any scheduled debounce lint for the same content is redundant.
	delete(s.pendingLintURIs, uri)

	if s.session == nil {
		return empty, nil
	}
	if !isLintableScriptFile(uri) {
		return empty, nil
	}
	snapshot := s.documentLintSnapshot(uri)
	if snapshot.unavailable {
		return empty, nil
	}
	originalContent := s.documents[uri]

	currentContent := s.computeFixAllContent(ctx, uri, originalContent, snapshot)

	if currentContent == originalContent {
		return empty, nil
	}

	// Produce a single TextEdit that replaces the entire document content.
	// Individual per-fix TextEdits can't be composed across passes (offsets shift),
	// so we replace the whole document with the final result.
	lastLine, lastChar := computeEndPosition(originalContent)

	codeAction := &lsproto.CodeAction{
		Title: "Fix all rslint auto-fixable problems",
		Kind:  ptrTo(codeActionKindSourceFixAllRslint),
		Edit: &lsproto.WorkspaceEdit{
			Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
				uri: {
					{
						Range: lsproto.Range{
							Start: lsproto.Position{Line: 0, Character: 0},
							End:   lsproto.Position{Line: uint32(lastLine), Character: uint32(lastChar)},
						},
						NewText: currentContent,
					},
				},
			},
		},
	}

	return lsproto.CodeActionResponse{
		CommandOrCodeActionArray: &[]lsproto.CommandOrCodeAction{
			{CodeAction: codeAction},
		},
	}, nil
}

// computeFixAllContent runs the multi-pass lint→fix loop and returns the final
// fixed content (== originalContent when nothing changed). Each pass folds
// eslint-plugin fixes into the native fixes so source.fixAll applies both, on
// the SAME content (so their byte offsets align with ApplyRuleFixes's input).
// The per-pass native lint goes through s.fixAllNativeLint, which tests override
// to drive the fold loop without a real TS session.
func (s *Server) computeFixAllContent(
	ctx context.Context,
	uri lsproto.DocumentUri,
	originalContent string,
	snapshot documentLintSnapshot,
) string {
	nativeLint := s.fixAllNativeLint
	if nativeLint == nil {
		nativeLint = s.defaultFixAllNativeLint
	}

	// Bound the eslint-plugin reverse requests across the WHOLE fixAll, not per
	// pass: source.fixAll runs inline on the dispatch loop, so a wedged or
	// mid-rebuild client that never answers rslint/pluginLint must not
	// freeze editor interaction — nor multiply the stall by maxFixPasses. Only
	// the plugin pass gets this deadline; the native pass keeps the original ctx
	// (it is in-process and does not depend on a client reply). Once the budget
	// expires lintPluginRulesSync returns nil and the remaining passes fold
	// native-only fixes.
	pluginTimeout := s.pluginReverseTimeout
	if pluginTimeout <= 0 {
		pluginTimeout = defaultPluginReverseTimeout
	}
	pluginCtx, cancelPlugin := context.WithTimeout(ctx, pluginTimeout)
	defer cancelPlugin()

	currentContent := originalContent
	for pass := range maxFixPasses {
		lintResult, err := nativeLint(ctx, uri, pass, currentContent, snapshot)
		if err != nil {
			log.Printf("Error running lint for fixAll pass %d: %v", pass, err)
			break
		}
		if lintResult.HasSyntaxErrors {
			break
		}
		ruleDiags := lintResult.Diagnostics

		// Fold in eslint-plugin fixes so source.fixAll applies plugin rule fixes
		// too, not just native. The plugin pass lints the SAME currentContent, so
		// its fix byte offsets align with ApplyRuleFixes's input; suggestionsMode
		// is "off" because fixAll applies only autofixes.
		// Skip the plugin pass once the budget is spent: lintPluginRulesSync on an
		// already-expired pluginCtx would still enqueue a (wasted) reverse request
		// to the client before returning nil.
		if pluginCtx.Err() == nil {
			if pluginDiags := s.lintPluginRulesSyncWithSnapshot(
				pluginCtx,
				uri,
				currentContent,
				true,
				linter.SuggestionsModeOff,
				snapshot,
			); len(pluginDiags) > 0 {
				ruleDiags = append(ruleDiags, pluginDiags...)
			}
		}

		fixedContent, _, wasFixed := linter.ApplyRuleFixes(currentContent, ruleDiags)
		if !wasFixed {
			break
		}
		currentContent = fixedContent
		if currentContent == originalContent {
			break // cycle detected — fixes reverted to original content
		}
	}
	return currentContent
}

// defaultFixAllNativeLint builds each pass from an isolated editor overlay.
// The pass number is intentionally unused: speculative content never enters
// the real TypeScript Session, regardless of how many fix cycles run.
func (s *Server) defaultFixAllNativeLint(
	ctx context.Context,
	uri lsproto.DocumentUri,
	_ int,
	content string,
	snapshot documentLintSnapshot,
) (lintPassResult, error) {
	return s.runConfiguredLintForContentWithSnapshot(uri, ctx, content, snapshot)
}

// computeEndPosition returns the line and UTF-16 character offset of the end
// of a text string, suitable for constructing an LSP Range that covers the
// entire document. Uses core.UTF16Len for correct UTF-16 code unit counting.
func computeEndPosition(text string) (int, int) {
	line := 0
	lastLineStart := 0
	for i := range len(text) {
		if text[i] == '\n' {
			line++
			lastLineStart = i + 1
		}
	}
	return line, int(core.UTF16Len(text[lastLineStart:]))
}

// Helper function to check if two ranges overlap
func rangesOverlap(a, b lsproto.Range) bool {
	// Ranges overlap if a starts before or at b's end AND b starts before or at a's end
	aStartsBefore := a.Start.Line < b.End.Line ||
		(a.Start.Line == b.End.Line && a.Start.Character <= b.End.Character)
	bStartsBefore := b.Start.Line < a.End.Line ||
		(b.Start.Line == a.End.Line && b.Start.Character <= a.End.Character)

	return aStartsBefore && bStartsBefore
}

// Helper function to create a code action from a rule diagnostic
func createCodeActionFromRuleDiagnostic(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri) *lsproto.CodeAction {
	fixes := ruleDiag.Fixes()
	if len(fixes) == 0 {
		return nil
	}

	// Convert rule fixes to LSP text edits
	var textEdits []*lsproto.TextEdit
	for _, fix := range fixes {
		textEdits = append(textEdits, ruleFixToTextEdit(ruleDiag.SourceFile, fix))
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	return &lsproto.CodeAction{
		Title:       "Fix: " + ruleDiag.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{convertRuleDiagnosticToLSP(ruleDiag)},
		IsPreferred: ptrTo(true), // Mark auto-fixes as preferred
	}
}

// Helper function to create a code action from a rule suggestion
func createCodeActionFromSuggestion(ruleDiag rule.RuleDiagnostic, suggestion rule.RuleSuggestion, uri lsproto.DocumentUri) *lsproto.CodeAction {
	fixes := suggestion.Fixes()
	if len(fixes) == 0 {
		return nil
	}

	// Convert rule fixes to LSP text edits
	var textEdits []*lsproto.TextEdit
	for _, fix := range fixes {
		textEdits = append(textEdits, ruleFixToTextEdit(ruleDiag.SourceFile, fix))
	}

	// Create workspace edit
	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: textEdits,
		},
	}

	return &lsproto.CodeAction{
		Title:       "Suggestion: " + suggestion.Message.Description,
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{convertRuleDiagnosticToLSP(ruleDiag)},
		IsPreferred: ptrTo(false), // Mark suggestions as not preferred
	}
}

// Helper function to create disable rule actions for diagnostics without fixes
func createDisableRuleActions(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri) []lsproto.CommandOrCodeAction {
	if ruleDiag.Origin == rule.DiagnosticOriginTypeScript {
		return nil
	}
	var actions []lsproto.CommandOrCodeAction

	lspDiagnostic := convertRuleDiagnosticToLSP(ruleDiag)

	// Action 1: Disable rule for this line
	disableLineAction := createDisableRuleForLineAction(ruleDiag, uri, lspDiagnostic)
	if disableLineAction != nil {
		actions = append(actions, lsproto.CommandOrCodeAction{
			Command:    nil,
			CodeAction: disableLineAction,
		})
	}

	// Action 2: Disable rule for entire file
	disableFileAction := createDisableRuleForFileAction(ruleDiag, uri, lspDiagnostic)
	if disableFileAction != nil {
		actions = append(actions, lsproto.CommandOrCodeAction{
			Command:    nil,
			CodeAction: disableFileAction,
		})
	}

	return actions
}

// Helper function to create a "disable rule for this line" action
func createDisableRuleForLineAction(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri, lspDiagnostic *lsproto.Diagnostic) *lsproto.CodeAction {
	// Get the line where the diagnostic occurs
	lineStart := lspDiagnostic.Range.Start.Line

	// Create text edit to add rslint-disable-next-line comment
	disableComment := fmt.Sprintf("// rslint-disable-next-line %s\n", ruleDiag.RuleName)

	// Find the start of the line to insert the comment
	lineStartPos := lsproto.Position{Line: lineStart, Character: 0}

	textEdit := &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: lineStartPos,
			End:   lineStartPos,
		},
		NewText: disableComment,
	}

	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: {textEdit},
		},
	}

	return &lsproto.CodeAction{
		Title:       fmt.Sprintf("Disable %s for this line", ruleDiag.RuleName),
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
		IsPreferred: ptrTo(false),
	}
}

// Helper function to create a "disable rule for entire file" action
func createDisableRuleForFileAction(ruleDiag rule.RuleDiagnostic, uri lsproto.DocumentUri, lspDiagnostic *lsproto.Diagnostic) *lsproto.CodeAction {
	// Create text edit to add rslint-disable comment at the top of the file
	disableComment := fmt.Sprintf("/* rslint-disable %s */\n", ruleDiag.RuleName)

	// Insert at the very beginning of the file
	fileStartPos := lsproto.Position{Line: 0, Character: 0}

	textEdit := &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: fileStartPos,
			End:   fileStartPos,
		},
		NewText: disableComment,
	}

	workspaceEdit := &lsproto.WorkspaceEdit{
		Changes: &map[lsproto.DocumentUri][]*lsproto.TextEdit{
			uri: {textEdit},
		},
	}

	return &lsproto.CodeAction{
		Title:       fmt.Sprintf("Disable %s for entire file", ruleDiag.RuleName),
		Kind:        ptrTo(lsproto.CodeActionKind("quickfix")),
		Edit:        workspaceEdit,
		Diagnostics: &[]*lsproto.Diagnostic{lspDiagnostic},
		IsPreferred: ptrTo(false),
	}
}
