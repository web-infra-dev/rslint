package linter

import (
	"context"
	"os"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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
	Program          *compiler.Program
	Scope            FileScope
	ExcludePaths     []string
	FileFilter       FileFilter
	TargetFiles      []string
	HasTargetFiles   bool
	SyntaxErrorFiles map[string]struct{}
	GetRulesForFile  RuleHandler
	// CollectExecutedRules controls whether runLintRulesInProgram builds the
	// per-program rule-name set returned in programLintResult. LintSingleFile
	// leaves this disabled because it does not consume that result.
	CollectExecutedRules bool
	// SingleThreaded, when true, lints this program's file shards
	// sequentially on the calling goroutine instead of in parallel workers.
	SingleThreaded bool
	// TypeInfoFiles is the set of files with reliable project type information.
	// Rules that require type information are filtered for every other file,
	// and remaining rules receive a nil TypeChecker. nil = no gap distinction.
	TypeInfoFiles map[string]struct{}
	OnDiagnostic  DiagnosticHandler
}

type programLintResult struct {
	lintedFileCount int32
	executedRules   map[string]struct{}
}

// runLintRulesInProgram lints files in a single Program. Files are filtered
// through ExcludePaths, Scope (Files+Dirs), and FileFilter before rule
// execution. Pass FileFilter=nil to disable that layer.
//
// Unless SingleThreaded is set, files are linted in parallel shards — one
// worker per pool checker, each worker owning its checker exclusively and
// processing the files associated to it (see the sharding comment in the
// function body for the invariants this preserves).
//
// This is the post-refactor internal implementation behind both RunLinter and
// LintSingleFile. It does NOT run type-check — type-check is a program-level
// concern handled by RunLinter directly.
func runLintRulesInProgram(opts runProgramOptions) programLintResult {
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}
	getRulesForFile := opts.GetRulesForFile
	if getRulesForFile == nil {
		return programLintResult{}
	}

	// Collect files to lint (applying all filters). Shared with
	// CollectLintTargets so the eslint-plugin dispatch path observes the
	// exact same file set as native linting.
	filesToLint := collectFilesToLint(opts)

	result := programLintResult{lintedFileCount: int32(len(filesToLint))}

	// Early-out: if every file in this program was filtered, do not pay the
	// cost of acquiring a TypeChecker (which forces program binding and is
	// non-trivial when the checker hasn't been created yet).
	if result.lintedFileCount == 0 {
		return result
	}

	// lintFile lints one file with its already-resolved rules and checker. All per-file state
	// (listener map, comments, DisableManager, rule contexts) lives inside
	// this function, so concurrent calls for different files are independent;
	// the checker is the only shared resource and is owned exclusively by the
	// calling worker (see the sharding below).
	lintFile := func(file *ast.SourceFile, rules []ConfiguredRule, chk *checker.Checker) {
		registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)

		// Computed once per file and shared via ctx.Comments so
		// rules that need every comment in the file (directive scanning,
		// max-lines, etc.) don't each repeat this same token-tree walk.
		comments := make([]*ast.CommentRange, 0)
		utils.ForEachComment(&file.Node, func(comment *ast.CommentRange) { comments = append(comments, comment) }, file)
		// ForEachComment can surface the same physical comment twice (once as
		// a token's trailing range, once as the next token's leading range)
		// and isn't guaranteed strictly ordered. Sort and dedup once here so
		// every downstream consumer can rely on a clean, ordered list instead
		// of each re-deriving it.
		sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
		comments = slices.CompactFunc(comments, func(a, b *ast.CommentRange) bool {
			return a.Pos() == b.Pos() && a.End() == b.End()
		})

		// Create disable manager for this file
		disableManager := rule.NewDisableManager(file, comments)

		// Parse inline `/* global */` comments once per file, same as
		// DisableManager above. Rules receive both declaration metadata and the
		// merged result instead of parsing comments or config themselves.
		inlineGlobals, inlineGlobalDeclarations := rule.ParseInlineGlobals(file, comments)
		fileChecker := chk
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
				ConfigGlobals:  r.Globals,
				InlineGlobals:  inlineGlobalDeclarations,
				Globals:        rule.MergeGlobals(r.Globals, inlineGlobals),
				Comments:       comments,
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
						FilePath:   file.FileName(),
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
						FilePath:   file.FileName(),
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
						FilePath:    file.FileName(),
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
						FilePath:   file.FileName(),
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
						FilePath:   file.FileName(),
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
						FilePath:    file.FileName(),
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
						FilePath:    file.FileName(),
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
						FilePath:    file.FileName(),
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

	// Phase 1 parallelism is per-file within the program: files are grouped
	// by the checker the pool associated to them (for the compiler pool this
	// is the stable index%N mapping built in checkerpool.go), and each group
	// is linted serially by ONE worker holding that checker exclusively.
	// This keeps three invariants:
	//   - a checker is never used by two goroutines at once (pool contract:
	//     checkers must not be accessed concurrently);
	//   - every file's diagnostics are emitted by a single worker, so the
	//     file-internal diagnostic order stays deterministic — the fixer's
	//     tie-breaking and reporters rely on this;
	//   - Phase 2 type-check visits files through the same association,
	//     reusing the type caches warmed during lint.
	// The LSP project pool builds its file association dynamically on first
	// GetChecker instead of precomputing index%N — with this loop's
	// acquire/release probing, a fresh project pool associates every file
	// to the first checker, so the grouping collapses to a single group
	// (no intra-program parallelism on that path; today it is only reached
	// via LintSingleFile, where one file means one group anyway).
	// Correctness never depends on the grouping: each worker only uses the
	// checker it acquired exclusively for its own shard.
	ctx := context.Background()
	rulesByFile := make(map[*ast.SourceFile][]ConfiguredRule, len(filesToLint))
	checkerGroups := make(map[*checker.Checker][]*ast.SourceFile)
	for _, file := range filesToLint {
		if shouldSkipRulesForSyntax(opts, file, ctx) {
			continue
		}
		rules := filterRulesForTypeInfo(
			getRulesForFile(file),
			file.FileName(),
			opts.TypeInfoFiles,
		)
		if opts.CollectExecutedRules && len(rules) > 0 {
			if result.executedRules == nil {
				result.executedRules = make(map[string]struct{}, len(rules))
			}
			for _, configuredRule := range rules {
				result.executedRules[configuredRule.Name] = struct{}{}
			}
		}
		rules = filterNativeRules(rules)
		if len(rules) == 0 {
			continue
		}
		rulesByFile[file] = rules
		if opts.TypeInfoFiles != nil {
			if _, hasTypeInfo := opts.TypeInfoFiles[file.FileName()]; !hasTypeInfo {
				checkerGroups[nil] = append(checkerGroups[nil], file)
				continue
			}
		}
		chk, release := opts.Program.GetTypeCheckerForFile(ctx, file)
		release()
		checkerGroups[chk] = append(checkerGroups[chk], file)
	}

	wg := core.NewWorkGroup(opts.SingleThreaded)
	for chk, files := range checkerGroups {
		wg.Queue(func() {
			if chk != nil {
				var done func()
				chk, done = opts.Program.GetTypeCheckerForFileExclusive(ctx, files[0])
				defer done()
			}
			for _, file := range files {
				lintFile(file, rulesByFile[file], chk)
			}
		})
	}
	wg.RunAndWait()

	return result
}

// filterNativeRules removes Node-dispatched ESLint plugin placeholders from
// the native pass without mutating the resolver's shared cached slice. The
// original list remains available to CollectLintTargets for plugin dispatch.
func filterNativeRules(rules []ConfiguredRule) []ConfiguredRule {
	firstPlugin := -1
	for i, configuredRule := range rules {
		if configuredRule.IsEslintPluginRule {
			firstPlugin = i
			break
		}
	}
	if firstPlugin < 0 {
		return rules
	}

	nativeRules := make([]ConfiguredRule, 0, len(rules)-1)
	nativeRules = append(nativeRules, rules[:firstPlugin]...)
	for _, configuredRule := range rules[firstPlugin+1:] {
		if !configuredRule.IsEslintPluginRule {
			nativeRules = append(nativeRules, configuredRule)
		}
	}
	return nativeRules
}

func filterRulesForTypeInfo(rules []ConfiguredRule, fileName string, typeInfoFiles map[string]struct{}) []ConfiguredRule {
	if typeInfoFiles == nil {
		return rules
	}
	if _, hasTypeInfo := typeInfoFiles[fileName]; hasTypeInfo {
		return rules
	}
	return FilterNonTypeAwareRules(rules)
}

func shouldSkipRulesForSyntax(opts runProgramOptions, file *ast.SourceFile, ctx context.Context) bool {
	if opts.SyntaxErrorFiles != nil {
		_, invalid := opts.SyntaxErrorFiles[file.FileName()]
		return invalid
	}
	return len(opts.Program.GetSyntacticDiagnostics(ctx, file)) > 0
}

// RunLinter runs all configured lint rules across the given programs in
// parallel, then optionally collects program-level type-check diagnostics
// aligned with `tsc --noEmit` semantics.
//
// Phase 1 — lint rules: each program is processed via
// runLintRulesInProgram, with files filtered through opts.ExcludePaths,
// opts.Scope, opts.PerProgramFilter and, in legacy scan mode, the program's
// own owned-file set. When opts.TargetFiles is non-nil, Phase 1 uses that exact
// per-Program target plan instead of scanning Program roots.
// Within a program, files are linted in parallel shards (one per pool
// checker); diagnostics therefore arrive in nondeterministic cross-file
// order and callers that print them should impose an explicit order.
// When opts.GetRulesForFile is nil, Phase 1 is skipped entirely — no work
// group is created and no per-program goroutines are spawned. This is how
// callers run a pure type-check pass (--type-check-only) without paying
// lint-side setup cost.
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

	executedRules := make(map[string]struct{})
	var lintedFileCount int32

	// Phase 1: lint rules per program (parallel). Skipped when no rule
	// handler was supplied — see doc above.
	if opts.GetRulesForFile != nil {
		programResults := make([]programLintResult, len(opts.Programs))
		wg := core.NewWorkGroup(opts.SingleThreaded)
		for i, program := range opts.Programs {
			var perProgramFilter FileFilter
			if i < len(opts.PerProgramFilter) {
				perProgramFilter = opts.PerProgramFilter[i]
			}
			var targetFiles []string
			if opts.TargetFiles != nil && i < len(opts.TargetFiles) {
				targetFiles = opts.TargetFiles[i]
			}
			filter := perProgramFilter
			if opts.TargetFiles == nil {
				ownedFiles := buildOwnedFileSet(program)
				filter = composeOwnedFilter(perProgramFilter, ownedFiles)
			}

			programOpts := runProgramOptions{
				Program:              program,
				Scope:                opts.Scope,
				ExcludePaths:         opts.ExcludePaths,
				FileFilter:           filter,
				TargetFiles:          targetFiles,
				HasTargetFiles:       opts.TargetFiles != nil,
				GetRulesForFile:      opts.GetRulesForFile,
				CollectExecutedRules: true,
				SyntaxErrorFiles:     opts.SyntaxErrorFiles,
				SingleThreaded:       opts.SingleThreaded,
				TypeInfoFiles:        opts.TypeInfoFiles,
				OnDiagnostic:         opts.OnDiagnostic,
			}
			programIndex := i
			programOptions := programOpts
			wg.Queue(func() {
				programResults[programIndex] = runLintRulesInProgram(programOptions)
			})
		}
		wg.RunAndWait()
		for _, programResult := range programResults {
			lintedFileCount += programResult.lintedFileCount
			for name := range programResult.executedRules {
				executedRules[name] = struct{}{}
			}
		}
	}

	// Phase 2: program-level type-check (tsc-aligned).
	if opts.TypeCheck {
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       opts.Programs,
			Skip:           opts.SkipTypeCheckPrograms,
			SingleThreaded: opts.SingleThreaded,
			OnDiagnostic:   opts.OnDiagnostic,
		})
	}

	return &LintResult{
		LintedFileCount: lintedFileCount,
		ExecutedRules:   executedRules,
	}, nil
}

// collectFilesToLint applies the ExcludePaths / Scope / FileFilter layers
// to a program's source files. Shared by runLintRulesInProgram (native
// lint) and CollectLintTargets (eslint-plugin dispatch) so both observe an
// identical file set.
func collectFilesToLint(opts runProgramOptions) []*ast.SourceFile {
	if opts.HasTargetFiles {
		return collectExactFilesToLint(opts)
	}

	var allowFileInfos []os.FileInfo
	if opts.Scope.Files != nil {
		allowFileInfos = precomputeAllowFileInfos(opts.Scope.Files)
	}
	var filesToLint []*ast.SourceFile
	for _, file := range opts.Program.GetSourceFiles() {
		p := string(file.Path())
		// skip lint node_modules and bundled files
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
	return filesToLint
}

func collectExactFilesToLint(opts runProgramOptions) []*ast.SourceFile {
	var filesToLint []*ast.SourceFile
	seen := make(map[string]struct{}, len(opts.TargetFiles))
	for _, target := range opts.TargetFiles {
		file := opts.Program.GetSourceFile(target)
		if file == nil {
			continue
		}
		fileName := file.FileName()
		if _, ok := seen[fileName]; ok {
			continue
		}
		seen[fileName] = struct{}{}
		p := string(file.Path())
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
		if opts.FileFilter != nil && !opts.FileFilter(fileName) {
			continue
		}
		filesToLint = append(filesToLint, file)
	}
	return filesToLint
}

// LintTarget is one file paired with the rules configured for it, as
// resolved by RunLinterOptions.GetRulesForFile.
type LintTarget struct {
	File  *ast.SourceFile
	Rules []ConfiguredRule
}

// CollectLintTargets resolves, for every file RunLinter would lint, the
// rules configured for it — WITHOUT running them. The CLI/LSP host uses it
// to split out eslint-plugin rules and dispatch them to the Node worker in
// parallel with native linting, reusing the exact same file-set filtering
// as RunLinter (exact TargetFiles when present, otherwise Scope / legacy
// owned-file filtering, plus ExcludePaths and per-program filters).
func CollectLintTargets(opts RunLinterOptions) []LintTarget {
	if opts.GetRulesForFile == nil {
		return nil
	}
	excludePaths := opts.ExcludePaths
	if excludePaths == nil {
		excludePaths = utils.ExcludePaths
	}
	var targets []LintTarget
	for i, program := range opts.Programs {
		var perProgramFilter FileFilter
		if i < len(opts.PerProgramFilter) {
			perProgramFilter = opts.PerProgramFilter[i]
		}
		var targetFiles []string
		if opts.TargetFiles != nil && i < len(opts.TargetFiles) {
			targetFiles = opts.TargetFiles[i]
		}
		filter := perProgramFilter
		if opts.TargetFiles == nil {
			filter = composeOwnedFilter(perProgramFilter, buildOwnedFileSet(program))
		}
		files := collectFilesToLint(runProgramOptions{
			Program:          program,
			Scope:            opts.Scope,
			ExcludePaths:     excludePaths,
			FileFilter:       filter,
			TargetFiles:      targetFiles,
			HasTargetFiles:   opts.TargetFiles != nil,
			SyntaxErrorFiles: opts.SyntaxErrorFiles,
			TypeInfoFiles:    opts.TypeInfoFiles,
		})
		targets = append(targets, collectLintTargetsForFiles(opts, program, files)...)
	}
	return targets
}

// collectLintTargetsForFiles resolves rules for one program's lint target
// files in parallel. Unlike the real lint pass, this has no type-checker
// affinity to preserve, so files are simply chunked evenly across goroutines.
// This is what turns the serial per-file config/glob resolution — the
// dominant cost on large repos with many ignore/files patterns — into a
// wall-clock win before the real, already-parallel lint pass even starts.
func collectLintTargetsForFiles(opts RunLinterOptions, program *compiler.Program, files []*ast.SourceFile) []LintTarget {
	if len(files) == 0 {
		return nil
	}
	shardCount := runtime.GOMAXPROCS(0)
	if opts.SingleThreaded {
		shardCount = 1
	} else if shardCount > len(files) {
		shardCount = len(files)
	}
	if shardCount < 1 {
		shardCount = 1
	}
	chunkSize := (len(files) + shardCount - 1) / shardCount
	shardResults := make([][]LintTarget, shardCount)

	wg := core.NewWorkGroup(opts.SingleThreaded)
	for shard := range shardCount {
		start := shard * chunkSize
		end := min(start+chunkSize, len(files))
		if start >= end {
			continue
		}
		shardIndex := shard
		shardFiles := files[start:end]
		wg.Queue(func() {
			ctx := context.Background()
			var result []LintTarget
			for _, file := range shardFiles {
				if shouldSkipRulesForSyntax(runProgramOptions{
					Program:          program,
					SyntaxErrorFiles: opts.SyntaxErrorFiles,
				}, file, ctx) {
					continue
				}
				rules := filterRulesForTypeInfo(opts.GetRulesForFile(file), file.FileName(), opts.TypeInfoFiles)
				if len(rules) == 0 {
					continue
				}
				result = append(result, LintTarget{File: file, Rules: rules})
			}
			shardResults[shardIndex] = result
		})
	}
	wg.RunAndWait()

	var targets []LintTarget
	for _, result := range shardResults {
		targets = append(targets, result...)
	}
	return targets
}

// LintSingleFile runs lint rules against a single file in a single program.
// The caller owns syntactic diagnostics; this pass does not run type-check.
func LintSingleFile(opts LintSingleFileOptions) {
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}
	if opts.OnDiagnostic == nil {
		opts.OnDiagnostic = func(rule.RuleDiagnostic) {}
	}
	getRulesForFile := opts.GetRulesForFile
	if !opts.HasTypeInfo && getRulesForFile != nil {
		base := getRulesForFile
		getRulesForFile = func(file *ast.SourceFile) []ConfiguredRule {
			return FilterNonTypeAwareRules(base(file))
		}
	}
	runLintRulesInProgram(runProgramOptions{
		Program:          opts.Program,
		ExcludePaths:     opts.ExcludePaths,
		TargetFiles:      []string{opts.File},
		HasTargetFiles:   true,
		GetRulesForFile:  getRulesForFile,
		SyntaxErrorFiles: map[string]struct{}{},
		// A single file is a single shard — run it on the calling goroutine
		// instead of spawning a worker.
		SingleThreaded: true,
		OnDiagnostic:   opts.OnDiagnostic,
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
