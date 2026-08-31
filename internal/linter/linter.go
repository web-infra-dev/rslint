package linter

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
)

var (
	errNilProgram                    = errors.New("linter: Program must not be nil")
	errInvalidProgram                = errors.New("linter: Program must be created by internal/program")
	errTypeCheckOnlyProgramsWithPlan = errors.New("linter: TypeCheckOnlyPrograms must be nil when LintPlan is provided")
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

// runLintRulesInProgram executes the files and rules already frozen for one
// Program. It does not perform target discovery, filtering, or rule resolution.
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

	// Early-out: if this Program has no selected lint targets, do not pay the
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

		// A cheap source-text check inside each parser avoids asking the store
		// for all comments unless that inline directive is possible.
		inlineGlobals, inlineGlobalDeclarations := rule.ParseInlineGlobals(file, comments)
		inlineExported, inlineExportedDeclarations := rule.ParseInlineExported(file, comments)
		var environment rule.RuleEnvironment
		if filePlan.environment != nil {
			environment = *filePlan.environment
		}

		// Resolve immutable language initialization once per file. Globals and
		// RefStore receive their own concrete data and never inspect the current
		// selection input (the file extension) themselves.
		globalsInit, refsInit, languageOptions := rule.ResolveLanguageDefaults(file.FileName(), environment.LanguageOptions)

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
		baseContext := (rule.RuleContext{
			SourceFile:      file,
			Settings:        environment.Settings,
			LanguageOptions: languageOptions,
			Globals:         rule.NewGlobals(languageOptions, globalsInit, environment.Globals, inlineGlobals, inlineGlobalDeclarations),
			Exported:        rule.NewExported(inlineExported, inlineExportedDeclarations),
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
			if expression := utils.JSDocTypeCastExpression(node); expression != nil {
				patternVisitor(expression)
				return
			}
			if utils.IsJSDocSyntaxNode(node) {
				return
			}
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
			if expression := utils.JSDocTypeCastExpression(node); expression != nil {
				childVisitor(expression)
				return false
			}
			if utils.IsJSDocSyntaxNode(node) {
				return false
			}
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

// RunLinter runs all configured lint rules across the given programs in
// parallel, then optionally collects program-level type-check diagnostics
// aligned with `tsc --noEmit` semantics.
//
// Phase 1 — lint rules: each Program is processed through the exact immutable
// file/rule projection in opts.LintPlan.
// Within a program, files are linted in parallel shards (one per pool
// checker); diagnostics therefore arrive in nondeterministic cross-file
// order and callers that print them should impose an explicit order.
// When opts.LintPlan is nil, Phase 1 is skipped entirely — no work group is
// created and no per-Program goroutines are spawned. This is how callers run a
// pure type-check pass (--type-check-only) without paying lint-side setup cost.
//
// Phase 2 — type-check (skipped when opts.TypeCheck is false): the linter asks
// each Program whether it can supply program-wide diagnostics and schedules
// only that capability. Capable Programs aggregate diagnostics through
// collectNoEmitDiagnostics, a helper that
// mirrors compiler.GetDiagnosticsOfAnyProgram(file=nil) but enforces
// `tsc --noEmit` semantics regardless of whether the user's tsconfig sets
// noEmit. Type-check is not constrained by the lint plan; it covers the full
// Program just like tsc.
//
// See RunLinterOptions for each field's zero-value semantics.
func RunLinter(opts RunLinterOptions) (*LintResult, error) {
	if !opts.Consumer.Demand.IsValid() {
		return nil, errors.New("linter: invalid native edit demand")
	}
	if opts.LintPlan != nil && opts.TypeCheckOnlyPrograms != nil {
		return nil, errTypeCheckOnlyProgramsWithPlan
	}
	consumer := normalizeDiagnosticConsumer(opts.Consumer)
	var sourcePrograms []*program.Program
	if opts.LintPlan != nil {
		for _, programPlan := range opts.LintPlan.programs {
			if err := validateProgram(programPlan.program); err != nil {
				return nil, err
			}
		}
		if opts.TypeCheck {
			sourcePrograms = opts.LintPlan.sourcePrograms()
		}
	} else if opts.TypeCheck {
		sourcePrograms = opts.TypeCheckOnlyPrograms
		if err := validatePrograms(sourcePrograms); err != nil {
			return nil, err
		}
	}

	executedRules := make(map[string]struct{})
	var lintedFileCount int32

	// Phase 1: lint rules per Program (parallel). Skipped when no plan was
	// supplied — see doc above.
	if opts.LintPlan != nil {
		plan := opts.LintPlan
		runOpts := programRunOptions{
			Cwd:                  opts.Cwd,
			CollectExecutedRules: true,
			SingleThreaded:       opts.SingleThreaded,
			Timing:               opts.Timing,
		}
		programResults := make([]programLintResult, len(plan.programs))
		wg := core.NewWorkGroup(opts.SingleThreaded)
		for i := range plan.programs {
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
		runTypeCheckAcrossPrograms(typeCheckRequest{
			Programs:       sourcePrograms,
			SingleThreaded: opts.SingleThreaded,
			OnDiagnostic:   consumer.Report,
		})
	}

	return &LintResult{
		LintedFileCount: lintedFileCount,
		ExecutedRules:   executedRules,
	}, nil
}

// LintSingleFile runs lint rules against one already selected file in one
// Program. The caller owns target selection and syntactic diagnostics; this
// pass does not apply path exclusions or run type-check. A non-empty request
// whose File is absent from Program violates the exact-file contract and
// panics.
func LintSingleFile(opts LintSingleFileOptions) {
	if !opts.Consumer.Demand.IsValid() {
		panic("linter: invalid native edit demand")
	}
	consumer := normalizeDiagnosticConsumer(opts.Consumer)
	getRulesForFile := opts.GetRulesForFile
	if getRulesForFile == nil {
		return
	}
	if err := validateProgram(opts.Program); err != nil {
		panic(err)
	}
	file := opts.Program.GetSourceFile(opts.File)
	if file == nil {
		panic(fmt.Errorf("%w: %q", errTargetNotInProgram, opts.File))
	}
	if !opts.HasTypeInfo {
		base := getRulesForFile
		getRulesForFile = func(file *ast.SourceFile) []rule.ConfiguredRule {
			return rule.FilterNonTypeAwareRules(base(file))
		}
	}
	plan, err := prepareProgramLintPlanForFiles(programRulePlanOptions{
		Program:         opts.Program,
		SkipSyntaxCheck: true,
		GetRulesForFile: getRulesForFile,
	}, []*ast.SourceFile{file})
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
