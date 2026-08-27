package lsp

import (
	"context"
	"errors"
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

// handleFixAllCodeAction computes all auto-fixes for the given URI using
// bounded rounds: each observation lints an isolated overlay and each non-empty
// planned change set advances it by one round, until stable or the linter-owned
// product limit.
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
	provider := s.newSpeculativeGenerationProvider(uri, originalContent, snapshot)

	// Bound eslint-plugin reverse requests across the whole fix-all operation.
	// Native work remains on the caller context because it is in-process; once
	// this plugin budget expires, later observations continue native-only.
	budgetedDispatch, cancelPlugin := s.pluginDispatchWithinBudget(ctx, snapshot.pluginGeneration)
	defer cancelPlugin()
	pipelineResult, err := linter.RunPipeline(ctx, linter.NewAutofixRequest(
		provider,
		linter.ObservationPolicy{
			Demand: linter.ArtifactDemand{
				Native: rule.EditDemandAutofix,
				Plugin: rule.EditDemandAutofix,
			},
			Plugin:        linter.PluginAfterNativeJoined,
			PluginFailure: linter.PluginDiscardOnFailure,
		},
		linter.AutofixPolicy{
			StopOnTargetSyntaxErrors: true,
		},
		budgetedDispatch,
	))
	reportFixAllPluginOutcomes(uri, pipelineResult.PluginOutcomes())
	if err != nil {
		// Preserve the established LSP policy: a later observation failure still
		// returns the pipeline's earlier, valid request-local memory result.
		log.Printf("Error running lint for fixAll: %v", err)
	}
	currentContent, contentErr := speculativeContentFromResult(
		pipelineResult,
		snapshot.target.Path,
		originalContent,
	)
	if contentErr != nil {
		return empty, contentErr
	}

	if currentContent == originalContent {
		return empty, nil
	}

	// Produce a single TextEdit that replaces the entire document content.
	// Individual per-fix TextEdits cannot be composed across rounds (offsets shift),
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

func speculativeContentFromResult(
	result linter.PipelineResult,
	targetPath string,
	originalContent string,
) (string, error) {
	applied, ok := result.AppliedFixes()
	if !ok || len(applied.FinalChanges) == 0 {
		return originalContent, nil
	}
	if len(applied.FinalChanges) != 1 || applied.FinalChanges[0].Path != targetPath {
		return "", fmt.Errorf("invalid LSP fix result for %q", targetPath)
	}
	return applied.FinalChanges[0].After, nil
}

func reportFixAllPluginOutcomes(uri lsproto.DocumentUri, records []linter.PluginDispatchRecord) {
	// A failed observation must not hide protocol notices produced by later
	// observations. Consume every structured notice before coalescing repeated
	// whole-operation transport failures to one log entry.
	for _, record := range records {
		reportLSPPluginProtocolNotices(record.Notices)
	}
	for _, record := range records {
		err := record.DispatchError
		if err == nil {
			continue
		}
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			log.Printf("[rslint] eslint-plugin fixAll for %s timed out (client unresponsive); applying native-only fixes", uri)
		case errors.Is(err, context.Canceled):
		default:
			log.Printf("[rslint] eslint-plugin fixAll lint error for %s: %v", uri, err)
		}
		// One whole-operation budget can make every later observation report the
		// same terminal error; one log entry is sufficient.
		return
	}
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
