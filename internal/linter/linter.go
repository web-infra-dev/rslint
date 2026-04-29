package linter

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
)

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

// runProgramOptions is the internal per-program input to runLintRulesInProgram.
type runProgramOptions struct {
	Program         *compiler.Program
	Scope           FileScope
	ExcludePaths    []string
	FileFilter      FileFilter
	GetRulesForFile RuleHandler
	// TypeInfoFiles is the set of files with reliable type information.
	// Gap files (not in this set) get a nil TypeChecker passed to rule
	// contexts as defense-in-depth — type-aware rules are already filtered
	// out by GetRulesForFile, this just guards rules with optional
	// TypeChecker usage. nil = no gap-file distinction.
	TypeInfoFiles map[string]struct{}
	OnDiagnostic DiagnosticHandler
}

// runLintRulesInProgram lints files in a single Program. Files are filtered
// through ExcludePaths, Scope (Files+Dirs), and FileFilter before rule
// execution. Pass FileFilter=nil to disable that layer.
//
// This is the post-refactor internal implementation behind both RunLinter and
// LintSingleFile. It does NOT run type-check — type-check is a program-level
// concern handled by RunLinter directly.
func runLintRulesInProgram(opts runProgramOptions) int32 {
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}
	getRulesForFile := opts.GetRulesForFile
	if getRulesForFile == nil {
		return 0
	}

	// Pre-compute FileInfo for Scope.Files once to avoid N×M stat calls in the loop.
	var allowFileInfos []os.FileInfo
	if opts.Scope.Files != nil {
		allowFileInfos = precomputeAllowFileInfos(opts.Scope.Files)
	}

	// Collect files to lint (applying all filters).
	var filesToLint []*ast.SourceFile
	for _, file := range opts.Program.GetSourceFiles() {
		p := string(file.Path())
		// skip lint node_modules and bundled files
		// FIXME: we may have better api to tell whether a file is a bundled file or not
		skipFile := false
		for _, skipPattern := range opts.ExcludePaths {
			if strings.Contains(p, skipPattern) {
				skipFile = true
				break
			}
		}
		if skipFile {
			continue
		}
		// Filter by Scope.Files / Scope.Dirs (OR logic: match either one).
		if opts.Scope.Files != nil || opts.Scope.Dirs != nil {
			fileAllowed := opts.Scope.Files != nil && isFileAllowed(file.FileName(), opts.Scope.Files, allowFileInfos)
			dirAllowed := opts.Scope.Dirs != nil && isDirAllowed(file.FileName(), opts.Scope.Dirs)
			if !fileAllowed && !dirAllowed {
				continue
			}
		}
		// Caller-supplied filter (multi-config ownership / config `ignores`).
		if opts.FileFilter != nil && !opts.FileFilter(file.FileName()) {
			continue
		}
		filesToLint = append(filesToLint, file)
	}

	lintedFileCount := int32(len(filesToLint))

	// Early-out: if every file in this program was filtered, do not pay the
	// cost of acquiring a TypeChecker (which forces program binding and is
	// non-trivial when the checker hasn't been created yet).
	if lintedFileCount == 0 {
		return 0
	}

	// Run lint rules. Acquires a checker from the pool for type-aware rules.
	checker, done := opts.Program.GetTypeChecker(context.Background())
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

		// For gap files (not in caller's TypeInfoFiles), pass nil TypeChecker
		// as defense-in-depth. Type-aware rules are already filtered out by
		// getRulesForFile, but this ensures rules with optional TypeChecker
		// usage degrade gracefully.
		fileChecker := checker
		if opts.TypeInfoFiles != nil {
			if _, hasTypeInfo := opts.TypeInfoFiles[file.FileName()]; !hasTypeInfo {
				fileChecker = nil
			}
		}

		for _, r := range rules {
			ctx := rule.RuleContext{
				SourceFile:     file,
				Program:        opts.Program,
				Settings:       r.Settings,
				TypeChecker:    fileChecker,
				DisableManager: disableManager,
				ReportRange: func(textRange core.TextRange, msg rule.RuleMessage) {
					if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:   r.Name,
						Range:      textRange,
						Message:    msg,
						SourceFile: file,
						Severity:   r.Severity,
					})
				},
				ReportRangeWithFixes: func(textRange core.TextRange, msg rule.RuleMessage, fixes ...rule.RuleFix) {
					if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:   r.Name,
						Range:      textRange,
						Message:    msg,
						FixesPtr:   &fixes,
						SourceFile: file,
						Severity:   r.Severity,
					})
				},
				ReportRangeWithSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
					if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:    r.Name,
						Range:       textRange,
						Message:     msg,
						Suggestions: &suggestions,
						SourceFile:  file,
						Severity:    r.Severity,
					})
				},
				ReportNode: func(node *ast.Node, msg rule.RuleMessage) {
					trimmedRange := utils.TrimNodeTextRange(file, node)
					if disableManager.IsRuleDisabled(r.Name, trimmedRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
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
					opts.OnDiagnostic(rule.RuleDiagnostic{
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
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:    r.Name,
						Range:       trimmedRange,
						Message:     msg,
						Suggestions: &suggestions,
						SourceFile:  file,
						Severity:    r.Severity,
					})
				},
				ReportNodeWithFixesAndSuggestions: func(node *ast.Node, msg rule.RuleMessage, fixes []rule.RuleFix, suggestions []rule.RuleSuggestion) {
					trimmedRange := utils.TrimNodeTextRange(file, node)
					if disableManager.IsRuleDisabled(r.Name, trimmedRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:    r.Name,
						Range:       trimmedRange,
						Message:     msg,
						FixesPtr:    &fixes,
						Suggestions: &suggestions,
						SourceFile:  file,
						Severity:    r.Severity,
					})
				},
				ReportRangeWithFixesAndSuggestions: func(textRange core.TextRange, msg rule.RuleMessage, fixes []rule.RuleFix, suggestions []rule.RuleSuggestion) {
					if disableManager.IsRuleDisabled(r.Name, textRange.Pos()) {
						return
					}
					opts.OnDiagnostic(rule.RuleDiagnostic{
						RuleName:    r.Name,
						Range:       textRange,
						Message:     msg,
						FixesPtr:    &fixes,
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

	return lintedFileCount
}

// RunLinter runs all configured lint rules across the given programs in
// parallel, then optionally collects program-level type-check diagnostics
// aligned with `tsc --noEmit` semantics.
//
// Phase 1 — lint rules: each program is processed via
// runLintRulesInProgram, with files filtered through opts.ExcludePaths,
// opts.Scope, opts.PerProgramFilter and the program's own owned-file set.
//
// Phase 2 — type-check (skipped when opts.TypeCheck is false): each
// non-skipped program is handed to runTypeCheckAcrossPrograms, which calls
// compiler.GetDiagnosticsOfAnyProgram(file=nil). Type-check is NOT
// constrained by Scope / PerProgramFilter / ExcludePaths — it covers the
// full program just like tsc.
//
// See RunLinterOptions for each field's zero-value semantics.
func RunLinter(opts RunLinterOptions) (*LintResult, error) {
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}

	var executedRules sync.Map
	trackedGetRules := opts.GetRulesForFile
	if trackedGetRules != nil {
		base := opts.GetRulesForFile
		trackedGetRules = func(sf *ast.SourceFile) []ConfiguredRule {
			rules := base(sf)
			for _, r := range rules {
				executedRules.Store(r.Name, struct{}{})
			}
			return rules
		}
	}

	// Phase 1: lint rules per program (parallel).
	wg := core.NewWorkGroup(opts.SingleThreaded)
	var lintedFileCount atomic.Int32
	for i, program := range opts.Programs {
		var perProgramFilter FileFilter
		if i < len(opts.PerProgramFilter) {
			perProgramFilter = opts.PerProgramFilter[i]
		}
		ownedFiles := buildOwnedFileSet(program)
		filter := composeOwnedFilter(perProgramFilter, ownedFiles)

		programOpts := runProgramOptions{
			Program:         program,
			Scope:           opts.Scope,
			ExcludePaths:    opts.ExcludePaths,
			FileFilter:      filter,
			GetRulesForFile: trackedGetRules,
			TypeInfoFiles:   opts.TypeInfoFiles,
			OnDiagnostic:    opts.OnDiagnostic,
		}
		wg.Queue(func() {
			n := runLintRulesInProgram(programOpts)
			lintedFileCount.Add(n)
		})
	}
	wg.RunAndWait()

	// Phase 2: program-level type-check (tsc-aligned).
	if opts.TypeCheck {
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       opts.Programs,
			Skip:           opts.SkipTypeCheckPrograms,
			SingleThreaded: opts.SingleThreaded,
			TypeInfoFiles:  opts.TypeInfoFiles,
			OnDiagnostic:   opts.OnDiagnostic,
		})
	}

	return &LintResult{
		LintedFileCount: lintedFileCount.Load(),
		ExecutedRules:   collectMapKeys(&executedRules),
	}, nil
}

// LintSingleFile runs lint rules against a single file in a single program.
// Designed for IDE / LSP per-keystroke usage. Does not run type-check.
func LintSingleFile(opts LintSingleFileOptions) {
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}
	runLintRulesInProgram(runProgramOptions{
		Program:         opts.Program,
		Scope:           FileScope{Files: []string{opts.File}},
		ExcludePaths:    opts.ExcludePaths,
		GetRulesForFile: opts.GetRulesForFile,
		OnDiagnostic:    opts.OnDiagnostic,
	})
}

// composeOwnedFilter combines a caller-supplied filter with the program's
// owned-file restriction. Either component may be nil.
func composeOwnedFilter(extra FileFilter, owned map[string]struct{}) FileFilter {
	if extra == nil && owned == nil {
		return nil
	}
	return func(name string) bool {
		if extra != nil && !extra(name) {
			return false
		}
		if owned != nil {
			if _, ok := owned[name]; !ok {
				return false
			}
		}
		return true
	}
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

func collectMapKeys(m *sync.Map) map[string]struct{} {
	out := make(map[string]struct{})
	m.Range(func(k, _ any) bool {
		if s, ok := k.(string); ok {
			out[s] = struct{}{}
		}
		return true
	})
	return out
}
