package linter

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
)

var (
	errNilProgram     = errors.New("linter: Program must not be nil")
	errInvalidProgram = errors.New("linter: Program must be created by internal/program")
)

func validateProgram(sourceProgram *program.Program) error {
	if sourceProgram == nil {
		return errNilProgram
	}
	if !sourceProgram.IsValid() {
		return errInvalidProgram
	}
	return nil
}

func validatePrograms(sourcePrograms []*program.Program) error {
	for _, sourceProgram := range sourcePrograms {
		if err := validateProgram(sourceProgram); err != nil {
			return err
		}
	}
	return nil
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

// programPlanOptions contains only rule-plan construction policy. Execution
// never reads it after a programLintPlan has frozen file identity, rules, and
// checker eligibility.
type programPlanOptions struct {
	Program         *program.Program
	Scope           FileScope
	ExcludePaths    []string
	FileFilter      FileFilter
	TargetFiles     []string
	HasTargetFiles  bool
	SkipSyntaxCheck bool
	GetRulesForFile RuleHandler
}

// programRunOptions contains the run-scoped sinks and scheduling knobs that do
// not change a prepared plan's meaning.
type programRunOptions struct {
	Cwd                  string
	CollectExecutedRules bool
	SingleThreaded       bool
	Timing               *TimingCollector
}

// Keep checker-free lint shards large enough to amortize their goroutine,
// listener-registry, and shared-consumer overhead. Checker-capable Programs
// retain their checker-defined shards and do not use this heuristic.
const minCheckerFreeFilesPerLintWorker = 128

func checkerFreeLintWorkerCount(fileCount int, maxWorkers int) int {
	if fileCount <= 0 || maxWorkers <= 0 {
		return 0
	}
	return max(1, min(maxWorkers, fileCount/minCheckerFreeFilesPerLintWorker))
}

type programLintResult struct {
	lintedFileCount int32
	executedRules   map[string]struct{}
}

type listenerRegistry struct {
	byKind      map[ast.Kind][]func(node *ast.Node)
	activeKinds []ast.Kind
}

func newListenerRegistry() listenerRegistry {
	return listenerRegistry{
		byKind:      make(map[ast.Kind][]func(node *ast.Node), 20),
		activeKinds: make([]ast.Kind, 0, 20),
	}
}

func (r *listenerRegistry) add(kind ast.Kind, listener func(node *ast.Node)) {
	listeners := r.byKind[kind]
	if len(listeners) == 0 {
		r.activeKinds = append(r.activeKinds, kind)
	}
	r.byKind[kind] = append(listeners, listener)
}

func (r *listenerRegistry) listeners(kind ast.Kind) []func(node *ast.Node) {
	return r.byKind[kind]
}

// reset releases every listener closure from the completed file while
// retaining the sparse registry's backing storage for the next file in the
// same checker-shard task.
func (r *listenerRegistry) reset() {
	for _, kind := range r.activeKinds {
		listeners := r.byKind[kind]
		clear(listeners)
		r.byKind[kind] = listeners[:0]
	}
	r.activeKinds = r.activeKinds[:0]
}

// runLintRulesInProgram lints files in a single Program. Files are filtered
// through ExcludePaths, Scope (Files+Dirs), and FileFilter before rule
// execution. Pass FileFilter=nil to disable that layer.
//
// Unless SingleThreaded is set, files are linted in parallel shards — one
// task per pool checker, each task owning its checker exclusively and
// processing the files associated to it (see the sharding comment in the
// function body for the invariants this preserves).
//
// This is the post-refactor internal implementation behind both RunLinter and
// LintSingleFile. It does NOT run type-check — type-check is a program-level
// concern handled by RunLinter directly. consumer is passed separately because
// edit demand belongs to the reporting pass, not to the immutable plan or the
// Program itself.
func runLintRulesInProgram(plan *programLintPlan, opts programRunOptions, consumer rule.DiagnosticConsumer) programLintResult {
	if plan == nil || !plan.program.IsValid() {
		return programLintResult{}
	}
	filesToLint := plan.files
	sourceProgram := plan.program

	result := programLintResult{lintedFileCount: int32(len(filesToLint))}

	// Early-out: if every file in this program was filtered, do not pay the
	// cost of acquiring a TypeChecker (which forces program binding and is
	// non-trivial when the checker hasn't been created yet).
	if result.lintedFileCount == 0 {
		return result
	}

	// lintFile lints one file with its already-resolved rules and checker. Its
	// comments, DisableManager, and rule contexts are per-file. The listener
	// registry belongs to the calling checker-shard task and is empty on entry;
	// reset clears all captured per-file state before the next serial file.
	lintFile := func(filePlan *lintFilePlan, rules []rule.ConfiguredRule, chk *checker.Checker, registeredListeners *listenerRegistry) {
		file := filePlan.file

		// Per-rule durations for this file, parallel to rules. Listeners are
		// wrapped at registration time, so when timing is off the traversal
		// hot path pays nothing. All rules share one AST traversal; timing
		// each listener invocation is what attributes traversal time to the
		// rule that registered the listener.
		var ruleDurations []time.Duration
		if opts.Timing != nil {
			ruleDurations = make([]time.Duration, len(rules))
		}

		// One lazy store is shared by directives, inline globals, and every
		// comment-aware rule in this file. Most files never materialize it.
		comments := rule.NewCommentStore(file)

		// Directive parsing is itself lazy and only runs when a rule reports.
		disableManager := rule.NewDisableManager(file, comments)

		// A cheap source-text check inside ParseInlineGlobals avoids asking
		// the store for all comments unless an inline directive is possible.
		inlineGlobals, inlineGlobalDeclarations := rule.ParseInlineGlobals(file, comments)

		// Resolve immutable language initialization once per file. Globals and
		// RefStore receive their own concrete data and never inspect the current
		// selection input (the file extension) themselves.
		globalsInit, refsInit := rule.ResolveLanguageDefaults(file.FileName())

		fileChecker := chk

		// One lazy reference index shared by every rule in this file; most
		// files never materialize it. fileChecker is passed as a fallback for
		// identifiers the binder scope walk can't resolve (declared outside
		// this file); nil there just disables the fallback.
		refs := rule.NewRefStore(file, sourceProgram.Options(), fileChecker, refsInit)

		// One lazy byte-order-mark answer shared by every rule in this file.
		// The mark is gone from the text by the time the file is parsed, so
		// answering means going back to whatever produced that text; a file no
		// rule asks about never does.
		sourceBOM := rule.NewSourceBOM(sourceProgram.FS(), file.FileName())
		fileCache := rule.NewFileCacheWithProcessCurrentDirectory(opts.Cwd)
		var environment rule.RuleEnvironment
		if filePlan.environment != nil {
			environment = *filePlan.environment
		}
		baseContext := (rule.RuleContext{
			SourceFile:      file,
			Settings:        environment.Settings,
			LanguageOptions: environment.LanguageOptions,
			Globals:         rule.NewGlobals(environment.LanguageOptions, globalsInit, environment.Globals, inlineGlobals, inlineGlobalDeclarations),
			Comments:        comments,
			Refs:            refs,
			BOM:             sourceBOM,
			TypeChecker:     fileChecker,
			DisableManager:  disableManager,
		}).WithProgram(sourceProgram).WithFileCache(fileCache)

		for ruleIndex, r := range rules {
			ctx := baseContext
			ctx = ctx.WithDiagnosticConsumer(
				r.Name,
				r.Severity,
				consumer,
			)

			var runStart time.Time
			if ruleDurations != nil {
				runStart = time.Now()
			}
			ruleListeners := r.Run(ctx)
			if ruleDurations != nil {
				ruleDurations[ruleIndex] += time.Since(runStart)
			}

			for kind, listener := range ruleListeners {
				if ruleDurations != nil {
					inner := listener
					listener = func(node *ast.Node) {
						start := time.Now()
						inner(node)
						ruleDurations[ruleIndex] += time.Since(start)
					}
				}
				registeredListeners.add(kind, listener)
			}
		}

		runListeners := func(kind ast.Kind, node *ast.Node) {
			for _, listener := range registeredListeners.listeners(kind) {
				listener(node)
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
				// Only the value of a pattern property is an assignment
				// target; its key stays an ordinary expression (a computed one
				// is even evaluated as such). ESTree visits that key like any
				// other child, so visit it through the normal path before
				// propagating pattern context to the initializer.
				if name := node.Name(); name != nil {
					childVisitor(name)
				}
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
		if opts.Timing != nil {
			opts.Timing.addFile(file.FileName(), rules, ruleDurations)
		}
		registeredListeners.reset()
	}

	// Phase 1 parallelism is per-file within the program: files are grouped
	// by the checker the pool associated to them (for the compiler pool this
	// is the stable index%N mapping built in checkerpool.go), and each group
	// is linted serially by one checker-shard task holding that checker
	// exclusively.
	// This keeps three invariants:
	//   - a checker is never used by two goroutines at once (pool contract:
	//     checkers must not be accessed concurrently);
	//   - every file's diagnostics are emitted by a single task, so the
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
	// Correctness never depends on the grouping: each task only uses the
	// checker it acquired exclusively for its own shard.
	ctx := context.Background()
	type lintFileTask struct {
		plan  *lintFilePlan
		rules []rule.ConfiguredRule
	}
	checkerGroups := make(map[*checker.Checker][]lintFileTask)
	checkerFreeGeneration := true
	for fileIndex := range filesToLint {
		filePlan := &filesToLint[fileIndex]
		file := filePlan.file
		rules := filePlan.rules
		if filePlan.hasTypeChecker {
			checkerFreeGeneration = false
		}
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
		task := lintFileTask{plan: filePlan, rules: rules}
		if !filePlan.hasTypeChecker {
			checkerGroups[nil] = append(checkerGroups[nil], task)
			continue
		}
		chk, release := sourceProgram.TypeCheckerForFile(ctx, file)
		release()
		checkerGroups[chk] = append(checkerGroups[chk], task)
	}

	wg := core.NewWorkGroup(opts.SingleThreaded)
	queueFiles := func(chk *checker.Checker, tasks []lintFileTask) {
		wg.Queue(func() {
			registeredListeners := newListenerRegistry()
			if chk != nil {
				var done func()
				chk, done = sourceProgram.TypeCheckerForFileExclusive(ctx, tasks[0].plan.file)
				defer done()
			}
			for _, task := range tasks {
				lintFile(task.plan, task.rules, chk, &registeredListeners)
			}
		})
	}
	for chk, tasks := range checkerGroups {
		// A source generation with no checker capability has no checker-owned
		// shard topology. Its files are independent, so split the nil-checker
		// group.
		if chk == nil && checkerFreeGeneration && !opts.SingleThreaded && len(tasks) > 1 {
			workerCount := checkerFreeLintWorkerCount(len(tasks), runtime.GOMAXPROCS(0))
			if workerCount < 2 {
				queueFiles(nil, tasks)
				continue
			}
			chunkSize := (len(tasks) + workerCount - 1) / workerCount
			for worker := range workerCount {
				start := worker * chunkSize
				end := min(start+chunkSize, len(tasks))
				if start < end {
					queueFiles(nil, tasks[start:end])
				}
			}
			continue
		}
		queueFiles(chk, tasks)
	}
	wg.RunAndWait()

	return result
}

// filterNativeRules removes Node-dispatched ESLint plugin placeholders from
// the native pass without mutating the resolver's shared cached slice. The
// prepared plan retains the original list for host-side plugin dispatch.
func filterNativeRules(rules []rule.ConfiguredRule) []rule.ConfiguredRule {
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

	nativeRules := make([]rule.ConfiguredRule, 0, len(rules)-1)
	nativeRules = append(nativeRules, rules[:firstPlugin]...)
	for _, configuredRule := range rules[firstPlugin+1:] {
		if !configuredRule.IsEslintPluginRule {
			nativeRules = append(nativeRules, configuredRule)
		}
	}
	return nativeRules
}

func shouldSkipRulesForSyntax(opts programPlanOptions, file *ast.SourceFile, ctx context.Context) bool {
	if opts.SkipSyntaxCheck {
		return false
	}
	return len(opts.Program.SyntacticDiagnostics(ctx, file)) > 0
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
// When opts.PreparedPlan is present, Phase 1 consumes its stable per-Program
// files and resolved rules without repeating target collection or invoking
// GetRulesForFile. CLI/API hosts use this to share one plan with optional
// third-party plugin dispatch; other callers retain the direct path.
//
// Phase 2 — type-check (skipped when opts.TypeCheck is false): the linter asks
// each Program whether it can supply program-wide diagnostics and schedules
// only that capability. Capable Programs aggregate diagnostics through
// collectNoEmitDiagnostics, a helper that
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
	if !opts.Consumer.Demand.IsValid() {
		return nil, errors.New("linter: invalid native edit demand")
	}
	consumer := normalizeDiagnosticConsumer(opts.Consumer)
	// Reject an invalid source generation before either phase can report
	// diagnostics. In particular, type-check-only runs skip Phase 1 entirely,
	// so validation cannot live inside the lint-rule branch.
	if opts.GetRulesForFile != nil || opts.TypeCheck {
		if err := validatePrograms(opts.Programs); err != nil {
			return nil, err
		}
		if opts.TypeCheck && opts.TypeCheckPrograms != nil {
			if err := validatePrograms(opts.TypeCheckPrograms); err != nil {
				return nil, err
			}
		}
	}

	executedRules := make(map[string]struct{})
	var lintedFileCount int32

	// Phase 1: lint rules per program (parallel). Skipped when no rule
	// handler was supplied — see doc above.
	if opts.GetRulesForFile != nil {
		plan := opts.PreparedPlan
		if plan == nil {
			var err error
			plan, err = PrepareLintPlan(opts)
			if err != nil {
				return nil, err
			}
		} else {
			if len(plan.programs) != len(opts.Programs) {
				return nil, errors.New("linter: prepared lint plan does not match Programs")
			}
			for programIndex, programPlan := range plan.programs {
				if programPlan.program != opts.Programs[programIndex] {
					return nil, errors.New("linter: prepared lint plan does not match Programs")
				}
			}
		}
		runOpts := programRunOptions{
			Cwd:                  opts.Cwd,
			CollectExecutedRules: true,
			SingleThreaded:       opts.SingleThreaded,
			Timing:               opts.Timing,
		}
		programResults := make([]programLintResult, len(opts.Programs))
		wg := core.NewWorkGroup(opts.SingleThreaded)
		for i := range opts.Programs {
			programIndex := i
			wg.Queue(func() {
				programResults[programIndex] = runLintRulesInProgram(&plan.programs[programIndex], runOpts, consumer)
			})
		}
		wg.RunAndWait()
		mergeResult := func(programResult programLintResult) {
			lintedFileCount += programResult.lintedFileCount
			for name := range programResult.executedRules {
				executedRules[name] = struct{}{}
			}
		}
		for _, programResult := range programResults {
			mergeResult(programResult)
		}
	}

	// Phase 2: program-level type-check (tsc-aligned).
	if opts.TypeCheck {
		typeCheckPrograms := opts.TypeCheckPrograms
		if typeCheckPrograms == nil {
			typeCheckPrograms = opts.Programs
		}
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       typeCheckPrograms,
			SingleThreaded: opts.SingleThreaded,
			OnDiagnostic:   consumer.Report,
		})
	}

	return &LintResult{
		LintedFileCount: lintedFileCount,
		ExecutedRules:   executedRules,
	}, nil
}

// collectFilesToLint applies the ExcludePaths / Scope / FileFilter layers to a
// program's source files. PrepareLintPlan retains this exact result when native
// execution and a host-side consumer need to share the same file/rule plan.
func collectFilesToLint(opts programPlanOptions) []*ast.SourceFile {
	if opts.HasTargetFiles {
		return collectExactFilesToLint(opts)
	}
	var allowFileInfos []os.FileInfo
	if opts.Scope.Files != nil {
		allowFileInfos = precomputeAllowFileInfos(opts.Scope.Files)
	}
	files := opts.Program.SourceFiles()
	for fileIndex, file := range files {
		if filePassesLintProjection(opts, file, allowFileInfos) {
			continue
		}
		// Program owns an immutable source slice. Allocate only when selection
		// changes its execution projection; in-place compaction would corrupt the
		// full universe used by cross-file rules.
		filesToLint := make([]*ast.SourceFile, 0, len(files)-1)
		filesToLint = append(filesToLint, files[:fileIndex]...)
		for _, remaining := range files[fileIndex+1:] {
			if filePassesLintProjection(opts, remaining, allowFileInfos) {
				filesToLint = append(filesToLint, remaining)
			}
		}
		return filesToLint
	}
	return files
}

func filePassesLintProjection(opts programPlanOptions, file *ast.SourceFile, allowFileInfos []os.FileInfo) bool {
	p := string(file.Path())
	for _, skipPattern := range opts.ExcludePaths {
		if strings.Contains(p, skipPattern) {
			return false
		}
	}
	// Scope dimensions use OR semantics when either one is present.
	if opts.Scope.Files != nil || opts.Scope.Dirs != nil {
		fileAllowed := opts.Scope.Files != nil && isFileAllowed(file.FileName(), opts.Scope.Files, allowFileInfos)
		dirAllowed := opts.Scope.Dirs != nil && isDirAllowed(file.FileName(), opts.Scope.Dirs)
		if !fileAllowed && !dirAllowed {
			return false
		}
	}
	return opts.FileFilter == nil || opts.FileFilter(file.FileName())
}

func collectExactFilesToLint(opts programPlanOptions) []*ast.SourceFile {
	// Exact target plans commonly select a Program's complete universe in the
	// same stable order. Preserve the Program-owned slice when selection makes
	// no change, avoiding a map and pointer-slice allocation without inspecting
	// how the Program was constructed.
	files := opts.Program.SourceFiles()
	if opts.FileFilter == nil && len(opts.TargetFiles) == len(files) {
		exact := true
		for fileIndex, target := range opts.TargetFiles {
			file := opts.Program.GetSourceFile(target)
			if file != files[fileIndex] || !filePassesExactProjection(opts, file) {
				exact = false
				break
			}
		}
		if exact {
			return files
		}
	}

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
		if !filePassesExactProjection(opts, file) {
			continue
		}
		filesToLint = append(filesToLint, file)
	}
	return filesToLint
}

func filePassesExactProjection(opts programPlanOptions, file *ast.SourceFile) bool {
	for _, skipPattern := range opts.ExcludePaths {
		if strings.Contains(string(file.Path()), skipPattern) {
			return false
		}
	}
	return opts.FileFilter == nil || opts.FileFilter(file.FileName())
}

// LintSingleFile runs lint rules against a single file in a single program.
// The caller owns syntactic diagnostics; this pass does not run type-check.
func LintSingleFile(opts LintSingleFileOptions) {
	if opts.ExcludePaths == nil {
		opts.ExcludePaths = utils.ExcludePaths
	}
	if !opts.Consumer.Demand.IsValid() {
		panic("linter: invalid native edit demand")
	}
	consumer := normalizeDiagnosticConsumer(opts.Consumer)
	getRulesForFile := opts.GetRulesForFile
	if getRulesForFile == nil {
		return
	}
	if !opts.HasTypeInfo {
		base := getRulesForFile
		getRulesForFile = func(file *ast.SourceFile) []rule.ConfiguredRule {
			return rule.FilterNonTypeAwareRules(base(file))
		}
	}
	plan, err := prepareProgramLintPlan(programPlanOptions{
		Program:         opts.Program,
		ExcludePaths:    opts.ExcludePaths,
		TargetFiles:     []string{opts.File},
		HasTargetFiles:  true,
		SkipSyntaxCheck: true,
		GetRulesForFile: getRulesForFile,
	})
	if err != nil {
		panic(err)
	}
	runLintRulesInProgram(&plan, programRunOptions{
		Cwd: opts.Cwd,
		// A single file is a single shard — run it on the calling goroutine
		// instead of scheduling a background task.
		SingleThreaded: true,
	}, consumer)
}

func normalizeDiagnosticConsumer(consumer rule.DiagnosticConsumer) rule.DiagnosticConsumer {
	if consumer.Report == nil {
		consumer.Demand = rule.EditDemandNone
		consumer.Report = discardDiagnostic
	}
	return consumer
}

func discardDiagnostic(rule.RuleDiagnostic) {}

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

// buildOwnedFileSet returns a set of file names that this Program directly owns
// through its root-file contract.
// Files in GetSourceFiles() but NOT in this set were pulled in through import
// resolution or project references — they belong to other programs.
// Returns nil for programs with no root files (should not happen in practice).
func buildOwnedFileSet(sourceProgram *program.Program) map[string]struct{} {
	fileNames := sourceProgram.RootFileNames()
	if len(fileNames) == 0 {
		return nil
	}
	owned := make(map[string]struct{}, len(fileNames))
	for _, fn := range fileNames {
		owned[fn] = struct{}{}
	}
	return owned
}
