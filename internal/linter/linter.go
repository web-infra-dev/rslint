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

// IsFileAllowed checks if fileName matches any path in allowFiles.
// It first tries fast string equality, then falls back to os.SameFile
// (using pre-computed FileInfo) to handle symlinks (e.g. /var vs /private/var on macOS).
func IsFileAllowed(fileName string, allowFiles []string, allowFileInfos []os.FileInfo) bool {
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

// PrecomputeAllowFileInfos collects os.FileInfo for each allowFile once,
// so that IsFileAllowed can use os.SameFile without repeated os.Stat calls.
// Files that do not exist are silently skipped.
func PrecomputeAllowFileInfos(allowFiles []string) []os.FileInfo {
	infos := make([]os.FileInfo, 0, len(allowFiles))
	for _, f := range allowFiles {
		if info, err := os.Stat(f); err == nil {
			infos = append(infos, info)
		}
	}
	return infos
}

// IsDirAllowed checks if fileName is inside any directory in allowDirs.
// Uses tspath.StartsWithDirectory to correctly handle src/ vs src-other/.
// caseSensitive is the filesystem's case sensitivity — callers pass
// fsys.UseCaseSensitiveFileNames(); the in-Program RunLinter path passes
// true to preserve its prior behavior.
func IsDirAllowed(fileName string, allowDirs []string, caseSensitive bool) bool {
	for _, dirPath := range allowDirs {
		if tspath.StartsWithDirectory(fileName, dirPath, caseSensitive) {
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
	OnDiagnostic  DiagnosticHandler

	// CompatRuleDispatcher is the eslint-plugin hook (see RunLinterOptions doc).
	// When non-nil, runLintRulesInProgram will collect compat rules per file,
	// invoke this dispatcher with one batch per program, and stream the
	// resulting diagnostics into OnDiagnostic alongside native ones.
	CompatRuleDispatcher CompatBatchHandler
	// SendCompatFileText ships each file's overlay text in the compat batch
	// (CompatLintFile.Text) instead of the worker re-reading disk — see
	// RunLinterOptions.SendCompatFileText (#3).
	SendCompatFileText bool
	// CollectFixes controls whether the runner materialises plugin
	// `descriptor.fix(fixer)` into per-diagnostic fixes. See
	// RunLinterOptions.CollectFixes for full semantics.
	CollectFixes bool
	// SuggestionsMode is "off" or "eager"; controls whether the runner
	// invokes plugin `descriptor.suggest[i].fix(fixer)` to materialize
	// suggestion text/fixes (off skips the work, eager runs it).
	SuggestionsMode string

	// Ctx propagated from RunLinterOptions.Ctx. nil = uncancellable.
	Ctx context.Context
}

// runLintRulesInProgram lints files in a single Program. Files are filtered
// through ExcludePaths, Scope (Files+Dirs), and FileFilter before rule
// execution. Pass FileFilter=nil to disable that layer.
//
// Returns the lintedFileCount and a "compat dispatch failed" flag. The
// flag is true when the program had compat rules to dispatch and the
// dispatcher returned an error — callers aggregate this to surface a
// runner-failure exit code.
//
// This is the post-refactor internal implementation behind both RunLinter and
// LintSingleFile. It does NOT run type-check — type-check is a program-level
// concern handled by RunLinter directly.
func runLintRulesInProgram(opts runProgramOptions) (lintedFileCount int32, compatDispatchFailed bool) {
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}
	getRulesForFile := opts.GetRulesForFile
	if getRulesForFile == nil {
		return 0, false
	}

	// ctxCancelled is consulted at the top of each file iteration below;
	// pulling it out here avoids the per-file Done() channel allocation
	// when no Ctx was supplied.
	ctxCancelled := func() bool { return false }
	if opts.Ctx != nil {
		ctxDone := opts.Ctx.Done()
		ctxCancelled = func() bool {
			select {
			case <-ctxDone:
				return true
			default:
				return false
			}
		}
	}

	// Pre-compute FileInfo for Scope.Files once to avoid N×M stat calls in the loop.
	var allowFileInfos []os.FileInfo
	if opts.Scope.Files != nil {
		allowFileInfos = PrecomputeAllowFileInfos(opts.Scope.Files)
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
			fileAllowed := opts.Scope.Files != nil && IsFileAllowed(file.FileName(), opts.Scope.Files, allowFileInfos)
			dirAllowed := opts.Scope.Dirs != nil && IsDirAllowed(file.FileName(), opts.Scope.Dirs, true)
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

	lintedFileCount = int32(len(filesToLint))

	// Early-out: if every file in this program was filtered, do not pay the
	// cost of acquiring a TypeChecker (which forces program binding and is
	// non-trivial when the checker hasn't been created yet).
	if lintedFileCount == 0 {
		return 0, false
	}

	// Run lint rules. Acquires a checker from the pool for type-aware rules.
	checker, done := opts.Program.GetTypeChecker(context.Background())

	// ── eslint-plugin compat-rule plumbing ─────────────────────────
	// Collected during the per-file native loop, dispatched at the end
	// of this program (one batch per rule-config-signature bucket — see
	// groupCompatByRuleSig).
	var compatPerFile []CompatFileEntry

	for _, file := range filesToLint {
		// Cancellation poll at the file boundary. The user-visible
		// promise is "SIGINT exits within one file's rule traversal";
		// we don't try to interrupt mid-AST-walk because that requires
		// threading the ctx through every rule's listener (huge surface).
		if ctxCancelled() {
			break
		}
		registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)
		filePath := file.FileName()

		rules := getRulesForFile(file)
		if len(rules) == 0 {
			continue
		}

		// Pre-scan rules: separate native vs compat. Compat rules have a
		// placeholder Run (returns nil); skipping their Run() avoids a
		// no-op call per rule per file. Their dispatch happens after the
		// native loop completes for the program.
		var nativeRules []ConfiguredRule
		var compatRules []ConfiguredRule
		for _, r := range rules {
			if r.IsEslintPluginRule {
				compatRules = append(compatRules, r)
			} else {
				nativeRules = append(nativeRules, r)
			}
		}

		// Build disable manager only when native rules need it. The
		// compat worker (packages/eslint-plugin-runner) runs its
		// own `applyDisableDirectives` (see apply-disable-directives.ts)
		// over the same source it parses, so Go side doesn't need to
		// re-derive comment positions for the compat path. Empirically
		// (cpuprof on the vscode bench) this saves ~9.6s wall on
		// compat-only configurations — the unconditional ForEachComment
		// + DisableManager build had 0 effect on output for those configs
		// (no `eslint-disable tse-js/*` directives in the source).
		var disableManager *rule.DisableManager
		if len(nativeRules) > 0 {
			comments := make([]*ast.CommentRange, 0)
			utils.ForEachComment(&file.Node, func(comment *ast.CommentRange) { comments = append(comments, comment) }, file)
			disableManager = rule.NewDisableManager(file, comments)
		}

		if len(compatRules) > 0 && opts.CompatRuleDispatcher != nil {
			// Project compat rules into the per-file maps (shared with the
			// CLI ingest path). ok is guaranteed by the enclosing
			// len(compatRules) > 0 guard.
			ruleMap, sevMap, langOpts, settings, configKey, _ :=
				CompatRuleMaps(compatRules)
			compatPerFile = append(compatPerFile, CompatFileEntry{
				Path:            filePath,
				Text:            file.Text(),
				SourceFile:      file,
				Rules:           ruleMap,
				Severity:        sevMap,
				LanguageOptions: langOpts,
				Settings:        settings,
				ConfigKey:       configKey,
			})
		}

		// Replace the rules slice with native-only for the listener registration
		// phase below.
		rules = nativeRules

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
						FilePath:   filePath,
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
						FilePath:   filePath,
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
						FilePath:    filePath,
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
						FilePath:   filePath,
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
						FilePath:   filePath,
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
						FilePath:    filePath,
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
						FilePath:    filePath,
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
						FilePath:    filePath,
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
		// Only walk the AST when native rules registered listeners.
		// Compat-only files have an empty listener set (compat rules
		// dispatch separately, after this loop), so the walk would visit
		// every node for zero effect — skip it.
		if len(nativeRules) > 0 {
			file.Node.ForEachChild(childVisitor)
		}
		clear(registeredListeners)
	}
	done()

	// ── eslint-plugin compat dispatch ─────────────────────────────
	// All native rules have completed for this program. Now hand off
	// to the compat dispatcher.
	//
	// Files in the same Program can have DIFFERENT effective rule
	// configurations: (1) multi-config monorepo where a shared tsconfig
	// includes files from multiple configs, (2) a single config that
	// uses ESLint's `overrides` mechanism to vary rules per file.
	// Pre-fix we built one batch with a single `unionRules` map keyed
	// by rule name — last-file-wins overwrote per-file options, so
	// every file in such a Program was linted with the LAST file's
	// options. Bucket by (rule-options + severity) signature so each
	// emitted batch is homogeneous; the typical single-config case
	// still emits one batch (one bucket).
	if len(compatPerFile) > 0 && opts.CompatRuleDispatcher != nil {
		// Delegate to the shared compat dispatch entry — same code path
		// the cmd-side compat-only fast path uses. See compat_runner.go
		// for bucketing / batching / diagnostic shaping.
		subRes, _ := DispatchCompat(DispatchCompatOptions{
			Files:           compatPerFile,
			Dispatcher:      opts.CompatRuleDispatcher,
			OnDiagnostic:    opts.OnDiagnostic,
			CollectFixes:    opts.CollectFixes,
			SuggestionsMode: opts.SuggestionsMode,
			Ctx:             opts.Ctx,
			IncludeFileText: opts.SendCompatFileText,
		})
		if subRes != nil && subRes.CompatDispatchErrors > 0 {
			compatDispatchFailed = true
		}
	}

	return lintedFileCount, compatDispatchFailed
}

// RunLinter runs all configured lint rules across the given programs in
// parallel, then optionally collects program-level type-check diagnostics
// aligned with `tsc --noEmit` semantics.
//
// Phase 1 — lint rules: each program is processed via
// runLintRulesInProgram, with files filtered through opts.ExcludePaths,
// opts.Scope, opts.PerProgramFilter and the program's own owned-file set.
// When opts.GetRulesForFile is nil, Phase 1 is skipped entirely — no work
// group is created, no per-program goroutines are spawned, and no
// owned-file sets are built. This is how callers run a pure type-check
// pass (--type-check-only) without paying lint-side setup cost.
//
// Phase 2 — type-check (skipped when opts.TypeCheck is false): each
// non-skipped program is handed to runTypeCheckAcrossPrograms, which
// aggregates diagnostics through collectNoEmitDiagnostics — a helper that
// mirrors compiler.GetDiagnosticsOfAnyProgram(file=nil) but enforces
// `tsc --noEmit` semantics regardless of whether the user's tsconfig
// sets noEmit. Type-check is NOT constrained by Scope / PerProgramFilter
// / ExcludePaths — it covers the full program just like tsc.
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
	var lintedFileCount atomic.Int32
	var compatDispatchErrors atomic.Int32

	// Phase 1: lint rules per program (parallel). Skipped when no rule
	// handler was supplied — see doc above.
	if opts.GetRulesForFile != nil {
		base := opts.GetRulesForFile
		trackedGetRules := func(sf *ast.SourceFile) []ConfiguredRule {
			rules := base(sf)
			for _, r := range rules {
				executedRules.Store(r.Name, struct{}{})
			}
			return rules
		}

		wg := core.NewWorkGroup(opts.SingleThreaded)
		for i, program := range opts.Programs {
			var perProgramFilter FileFilter
			if i < len(opts.PerProgramFilter) {
				perProgramFilter = opts.PerProgramFilter[i]
			}
			ownedFiles := buildOwnedFileSet(program)
			filter := composeOwnedFilter(perProgramFilter, ownedFiles)

			programOpts := runProgramOptions{
				Program:              program,
				Scope:                opts.Scope,
				ExcludePaths:         opts.ExcludePaths,
				FileFilter:           filter,
				GetRulesForFile:      trackedGetRules,
				TypeInfoFiles:        opts.TypeInfoFiles,
				OnDiagnostic:         opts.OnDiagnostic,
				CompatRuleDispatcher: opts.CompatRuleDispatcher,
				SendCompatFileText:   opts.SendCompatFileText,
				CollectFixes:         opts.CollectFixes,
				SuggestionsMode:      opts.SuggestionsMode,
				Ctx:                  opts.Ctx,
			}
			wg.Queue(func() {
				n, dispatchFailed := runLintRulesInProgram(programOpts)
				lintedFileCount.Add(n)
				if dispatchFailed {
					compatDispatchErrors.Add(1)
				}
			})
		}
		wg.RunAndWait()
	}

	// Phase 2: program-level type-check (tsc-aligned). Skip when the
	// lint-level ctx has been cancelled — type-check is expensive and
	// not currently ctx-aware itself, so without this gate a SIGINT
	// during Phase 1 leaves the user waiting for full type-check
	// completion (typically seconds, can be minutes on big projects).
	if opts.TypeCheck && (opts.Ctx == nil || opts.Ctx.Err() == nil) {
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       opts.Programs,
			Skip:           opts.SkipTypeCheckPrograms,
			SingleThreaded: opts.SingleThreaded,
			TypeInfoFiles:  opts.TypeInfoFiles,
			OnDiagnostic:   opts.OnDiagnostic,
		})
	}

	result := &LintResult{
		LintedFileCount:      lintedFileCount.Load(),
		ExecutedRules:        collectMapKeys(&executedRules),
		CompatDispatchErrors: compatDispatchErrors.Load(),
	}
	// Surface cancellation to the caller. Diagnostics that were already
	// streamed via OnDiagnostic stay; partial-result + ctx.Err() lets
	// callers distinguish "incomplete" from "complete with zero files".
	if opts.Ctx != nil && opts.Ctx.Err() != nil {
		return result, opts.Ctx.Err()
	}
	return result, nil
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
		Program:              opts.Program,
		Scope:                FileScope{Files: []string{opts.File}},
		ExcludePaths:         opts.ExcludePaths,
		GetRulesForFile:      opts.GetRulesForFile,
		OnDiagnostic:         opts.OnDiagnostic,
		CompatRuleDispatcher: opts.CompatRuleDispatcher,
		// LintSingleFile is the IDE/LSP single-file path — always an editor
		// buffer that may be unsaved, so ship its text (#3).
		SendCompatFileText: true,
		CollectFixes:       opts.CollectFixes,
		SuggestionsMode:    opts.SuggestionsMode,
		Ctx:                opts.Ctx,
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
