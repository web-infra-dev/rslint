package linter

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
)

type ConfiguredRule struct {
	Name             string
	Settings         map[string]interface{}
	Severity         rule.DiagnosticSeverity
	RequiresTypeInfo bool
	Run              func(ctx rule.RuleContext) rule.RuleListeners
}

// FilterNonTypeAwareRules returns only rules that do not require type information.
func FilterNonTypeAwareRules(rules []ConfiguredRule) []ConfiguredRule {
	filtered := make([]ConfiguredRule, 0, len(rules))
	for _, r := range rules {
		if !r.RequiresTypeInfo {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// isFileAllowed checks if fileName matches any path in allowFiles.
// It first tries fast string equality, then falls back to os.SameFile
// (using pre-computed FileInfo) to handle symlinks (e.g. /var vs /private/var on macOS).
func isFileAllowed(fileName string, allowFiles []string, allowFileInfos []os.FileInfo) bool {
	for _, filePath := range allowFiles {
		if filePath == fileName {
			return true
		}
	}
	// Fallback: compare by inode to handle directory symlinks
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	for _, info := range allowFileInfos {
		if os.SameFile(fileInfo, info) {
			return true
		}
	}
	return false
}

// precomputeAllowFileInfos collects os.FileInfo for each allowFile once,
// so that isFileAllowed can use os.SameFile without repeated os.Stat calls.
// Files that do not exist are silently skipped.
func precomputeAllowFileInfos(allowFiles []string) []os.FileInfo {
	infos := make([]os.FileInfo, 0, len(allowFiles))
	for _, f := range allowFiles {
		if info, err := os.Stat(f); err == nil {
			infos = append(infos, info)
		}
	}
	return infos
}

// isDirAllowed checks if fileName is inside any directory in allowDirs.
// Uses tspath.StartsWithDirectory to correctly handle src/ vs src-other/.
func isDirAllowed(fileName string, allowDirs []string) bool {
	for _, dirPath := range allowDirs {
		if tspath.StartsWithDirectory(fileName, dirPath, true) {
			return true
		}
	}
	return false
}

// flattenDiagnosticMessage builds a human-readable message from a TypeScript
// diagnostic, including its MessageChain and RelatedInformation.
// The format follows tsc's output style.
func flattenDiagnosticMessage(d *ast.Diagnostic) string {
	var b strings.Builder
	b.WriteString(d.String())
	for _, chain := range d.MessageChain() {
		flattenMessageChain(&b, chain, 1)
	}
	for _, related := range d.RelatedInformation() {
		if related.File() != nil {
			line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(related.File(), related.Pos())
			fmt.Fprintf(&b, "\n  %s:%d: %s", related.File().FileName(), line+1, related.String())
		}
	}
	return b.String()
}

func flattenMessageChain(b *strings.Builder, chain *ast.Diagnostic, level int) {
	b.WriteByte('\n')
	for range level {
		b.WriteString("  ")
	}
	b.WriteString(chain.String())
	for _, child := range chain.MessageChain() {
		flattenMessageChain(b, child, level+1)
	}
}

// RunLinterInProgram lints files in a single Program. Files are filtered through
// skipFiles, allowFiles/allowDirs, and the optional fileFilter before rule execution.
// fileFilter is used in multi-config mode for ownership-based deduplication: only files
// owned by this program's config pass the filter. Pass nil to lint all files.
func RunLinterInProgram(program *compiler.Program, allowFiles []string, allowDirs []string, skipFiles []string, getRulesForFile RuleHandler, typeCheck bool, onDiagnostic DiagnosticHandler, typeInfoFiles map[string]struct{}, fileFilter func(string) bool) int32 {
	// Pre-compute FileInfo for allowFiles once to avoid N×M stat calls in the loop.
	var allowFileInfos []os.FileInfo
	if allowFiles != nil {
		allowFileInfos = precomputeAllowFileInfos(allowFiles)
	}

	// Collect files to lint (applying all filters).
	var filesToLint []*ast.SourceFile
	for _, file := range program.GetSourceFiles() {
		p := string(file.Path())
		// skip lint node_modules and bundled files
		// FIXME: we may have better api to tell whether a file is a bundled file or not
		skipFile := false
		for _, skipPattern := range skipFiles {
			if strings.Contains(p, skipPattern) {
				skipFile = true
				break
			}
		}
		if skipFile {
			continue
		}
		// Filter by allowFiles / allowDirs (OR logic: match either one)
		if allowFiles != nil || allowDirs != nil {
			fileAllowed := allowFiles != nil && isFileAllowed(file.FileName(), allowFiles, allowFileInfos)
			dirAllowed := allowDirs != nil && isDirAllowed(file.FileName(), allowDirs)
			if !fileAllowed && !dirAllowed {
				continue
			}
		}
		// Ownership filter: in multi-config mode, only lint files owned by
		// the current program's config (nearest config == program's config).
		if fileFilter != nil && !fileFilter(file.FileName()) {
			continue
		}
		filesToLint = append(filesToLint, file)
	}

	lintedFileCount := int32(len(filesToLint))

	// Phase 1: Run lint rules. Acquires a checker from the pool for type-aware rules.
	{
		checker, done := program.GetTypeChecker(context.Background())
		for _, file := range filesToLint {
			registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)

			rules := getRulesForFile(file)
			if len(rules) == 0 {
				continue
			}

			comments := make([]*ast.CommentRange, 0)
			utils.ForEachComment(&file.Node, func(comment *ast.CommentRange) { comments = append(comments, comment) }, file)

			// Create disable manager for this file
			disableManager := rule.NewDisableManager(file, comments)

			// For gap files (not in typeInfoFiles), pass nil TypeChecker
			// as defense-in-depth. Type-aware rules are already filtered
			// out by getRulesForFile, but this ensures rules with optional
			// TypeChecker usage degrade gracefully.
			fileChecker := checker
			if typeInfoFiles != nil {
				if _, hasTypeInfo := typeInfoFiles[file.FileName()]; !hasTypeInfo {
					fileChecker = nil
				}
			}

			for _, r := range rules {
				ctx := rule.RuleContext{
					SourceFile:     file,
					Program:        program,
					Settings:       r.Settings,
					TypeChecker:    fileChecker,
					DisableManager: disableManager,
					ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      textRange,
							Message:    msg,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportRangeWithFixes: func(textRange core.TextRange, msg rule.RuleMessage, fixes ...rule.RuleFix) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      textRange,
							Message:    msg,
							FixesPtr:   &fixes,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
						// Check if rule is disabled at this position
						if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:    r.Name,
							Range:       textRange,
							Message:     msg,
							Suggestions: &suggestions,
							SourceFile:  file,
							Severity:    r.Severity,
						})
					},
					ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
						// Trim leading trivia (comments/whitespace) so the line number
						// matches the actual code, not a preceding disable comment.
						trimmedRange := utils.TrimNodeTextRange(file, node)
						if disableManager.IsRuleDisabled(r.Name, trimmedRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      trimmedRange,
							Message:    msg,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},
					ReportNodeWithFixes: func(node *ast.Node, msg rule.RuleMessage, fixes ...rule.RuleFix) {
						trimmedRange := utils.TrimNodeTextRange(file, node)
						if disableManager.IsRuleDisabled(r.Name, trimmedRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:   r.Name,
							Range:      trimmedRange,
							Message:    msg,
							FixesPtr:   &fixes,
							SourceFile: file,
							Severity:   r.Severity,
						})
					},

					ReportNodeWithSuggestions: func(node *ast.Node, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
						trimmedRange := utils.TrimNodeTextRange(file, node)
						if disableManager.IsRuleDisabled(r.Name, trimmedRange.Pos()) {
							return
						}
						onDiagnostic(rule.RuleDiagnostic{
							RuleName:    r.Name,
							Range:       trimmedRange,
							Message:     msg,
							Suggestions: &suggestions,
							SourceFile:  file,
							Severity:    r.Severity,
						})
					},
				}

				for kind, listener := range r.Run(ctx) {
					listeners, ok := registeredListeners[kind]
					if !ok {
						listeners = make([](func(node *ast.Node)), 0, len(rules))
					}
					registeredListeners[kind] = append(listeners, listener)
				}
			}

			runListeners := func(kind ast.Kind, node *ast.Node) {
				if listeners, ok := registeredListeners[kind]; ok {
					for _, listener := range listeners {
						listener(node)
					}
				}
			}

			/* convert.ts -> allowPattern:
			catch name
			variabledeclaration name
			forinstatement initializer
			forofstatement initializer
			(propagation) allowPattern > arrayliteralexpression elements
			(propagation) allowPattern > objectliteralexpression properties
			(propagation) allowPattern > spreadassignment,spreadelement expression
			(propagation) allowPattern > propertyassignment value
			arraybindingpattern elements
			objectbindingpattern elements
			(init) binaryexpression(with '=' operator') left
			*/

			var childVisitor ast.Visitor
			var patternVisitor func(node *ast.Node)
			patternVisitor = func(node *ast.Node) {
				runListeners(node.Kind, node)
				kind := rule.ListenerOnAllowPattern(node.Kind)
				runListeners(kind, node)

				switch node.Kind {
				case ast.KindArrayLiteralExpression:
					for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
						patternVisitor(element)
					}
				case ast.KindObjectLiteralExpression:
					for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
						patternVisitor(property)
					}
				case ast.KindSpreadElement, ast.KindSpreadAssignment:
					patternVisitor(node.Expression())
				case ast.KindPropertyAssignment:
					patternVisitor(node.Initializer())
				default:
					node.ForEachChild(childVisitor)
				}

				runListeners(rule.ListenerOnExit(kind), node)
				runListeners(rule.ListenerOnExit(node.Kind), node)
			}
			childVisitor = func(node *ast.Node) bool {
				runListeners(node.Kind, node)

				switch node.Kind {
				case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
					kind := rule.ListenerOnNotAllowPattern(node.Kind)
					runListeners(kind, node)
					node.ForEachChild(childVisitor)
					runListeners(rule.ListenerOnExit(kind), node)
				default:
					if ast.IsAssignmentExpression(node, true) {
						expr := node.AsBinaryExpression()
						patternVisitor(expr.Left)
						childVisitor(expr.OperatorToken)
						childVisitor(expr.Right)
					} else {
						node.ForEachChild(childVisitor)
					}
				}

				runListeners(rule.ListenerOnExit(node.Kind), node)

				return false
			}
			file.Node.ForEachChild(childVisitor)
			clear(registeredListeners)
		}
		done()
	}

	// Phase 2: Collect TypeScript semantic diagnostics when type-check is enabled.
	// This runs after releasing the checker from Phase 1, because GetSemanticDiagnostics
	// internally acquires its own checker from the pool.
	if typeCheck {
		ctx := context.Background()
		for _, file := range filesToLint {
			// Skip semantic diagnostics for gap files (no reliable type info).
			if typeInfoFiles != nil {
				if _, hasTypeInfo := typeInfoFiles[file.FileName()]; !hasTypeInfo {
					continue
				}
			}
			for _, d := range program.GetSemanticDiagnostics(ctx, file) {
				onDiagnostic(rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", d.Code()),
					Range:        d.Loc(),
					Message:      rule.RuleMessage{Description: flattenDiagnosticMessage(d)},
					SourceFile:   file,
					Severity:     rule.SeverityError,
					PreFormatted: true,
				})
			}
		}
	}

	return lintedFileCount
}

type RuleHandler = func(sourceFile *ast.SourceFile) []ConfiguredRule
type DiagnosticHandler = func(diagnostic rule.RuleDiagnostic)

// RunLinter runs all configured rules across the given programs in parallel.
//   - allowFiles: if non-nil, only lint files in this list; nil = all files
//   - allowDirs: if non-nil, also lint files under these dirs (OR with allowFiles)
//   - typeInfoFiles: files with type info; gap files not in this set skip type-aware rules
//   - fileFilters: optional per-program ownership filters (parallel to programs).
//     In multi-config mode, each filter ensures a program only lints files owned by
//     its nearest config. nil or missing entries = no filter (process all).
func RunLinter(programs []*compiler.Program, singleThreaded bool, allowFiles []string, allowDirs []string, excludedPaths []string, getRulesForFile RuleHandler, typeCheck bool, onDiagnostic DiagnosticHandler, typeInfoFiles map[string]struct{}, fileFilters []func(string) bool) (int32, error) {

	wg := core.NewWorkGroup(singleThreaded)

	var lintedFileCount atomic.Int32
	for i, program := range programs {
		var baseFilter func(string) bool
		if i < len(fileFilters) {
			baseFilter = fileFilters[i]
		}

		// Each program only lints its own root files (from tsconfig include/files
		// patterns or gap file list). Files pulled in through import resolution or
		// project references belong to other programs — linting them here would
		// cause duplicate diagnostics.
		ownedFiles := buildOwnedFileSet(program)
		filter := func(fileName string) bool {
			if baseFilter != nil && !baseFilter(fileName) {
				return false
			}
			if ownedFiles != nil {
				_, isOwned := ownedFiles[fileName]
				return isOwned
			}
			return true
		}

		wg.Queue(func() {
			fileCount := RunLinterInProgram(program, allowFiles, allowDirs, excludedPaths, getRulesForFile, typeCheck, onDiagnostic, typeInfoFiles, filter)
			lintedFileCount.Add(fileCount)
		})
	}
	wg.RunAndWait()
	return lintedFileCount.Load(), nil
}

// buildOwnedFileSet returns a set of file names that this program directly owns
// (listed in its tsconfig include/files patterns, or as gap file root files).
// Files in GetSourceFiles() but NOT in this set were pulled in through import
// resolution or project references — they belong to other programs.
// Returns nil for programs with no root files (should not happen in practice).
func buildOwnedFileSet(program *compiler.Program) map[string]struct{} {
	fileNames := program.CommandLine().FileNames()
	if len(fileNames) == 0 {
		return nil
	}
	owned := make(map[string]struct{}, len(fileNames))
	for _, fn := range fileNames {
		owned[fn] = struct{}{}
	}
	return owned
}
