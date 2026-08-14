package linter

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func mustPrepareLintPlan(t *testing.T, opts RunLinterOptions) *LintPlan {
	t.Helper()
	plan, err := PrepareLintPlan(opts)
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	return plan
}

func TestLintProjectionReusesUnchangedProgramUniverse(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
		"b.ts": "export const b = 1;",
	})
	a := raw.GetSourceFile(paths["a.ts"])
	b := raw.GetSourceFile(paths["b.ts"])
	if a == nil || b == nil {
		t.Fatal("fixture Program did not contain both source files")
	}
	sourceProgram := mustSourceOnlyTestProgram(t, raw, []*ast.SourceFile{a, b})
	universe := sourceProgram.SourceFiles()

	scanned := collectFilesToLint(programPlanOptions{
		Program:      sourceProgram,
		ExcludePaths: []string{},
	})
	if len(scanned) != len(universe) || &scanned[0] != &universe[0] {
		t.Fatal("unchanged scan projection copied the Program source universe")
	}

	exact := collectFilesToLint(programPlanOptions{
		Program:        sourceProgram,
		ExcludePaths:   []string{},
		TargetFiles:    []string{a.FileName(), b.FileName()},
		HasTargetFiles: true,
	})
	if len(exact) != len(universe) || &exact[0] != &universe[0] {
		t.Fatal("unchanged exact projection copied the Program source universe")
	}

	excluded := collectFilesToLint(programPlanOptions{
		Program:      sourceProgram,
		ExcludePaths: []string{string(b.Path())},
	})
	if len(excluded) != 1 || excluded[0] != a || len(universe) != 2 || universe[1] != b {
		t.Fatalf("filtered projection=%v corrupted universe=%v", excluded, universe)
	}
}

func TestExactLintProjectionCallsFilterOncePerTarget(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
		"b.ts": "export const b = 1;",
	})
	a := raw.GetSourceFile(paths["a.ts"])
	b := raw.GetSourceFile(paths["b.ts"])
	if a == nil || b == nil {
		t.Fatal("fixture Program did not contain both source files")
	}
	sourceProgram := mustSourceOnlyTestProgram(t, raw, []*ast.SourceFile{a, b})
	calls := map[string]int{}
	files := collectFilesToLint(programPlanOptions{
		Program:        sourceProgram,
		ExcludePaths:   []string{},
		TargetFiles:    []string{a.FileName(), b.FileName()},
		HasTargetFiles: true,
		FileFilter: func(fileName string) bool {
			calls[fileName]++
			return fileName != b.FileName()
		},
	})
	if len(files) != 1 || files[0] != a || calls[a.FileName()] != 1 || calls[b.FileName()] != 1 {
		t.Fatalf("projection=%v filter calls=%v", files, calls)
	}
}

func TestPreparedLintPlanPreservesNativeSemanticsAndIsReused(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts":    "const a = 1;",
		"bad.ts":  "const bad = 1;",
		"gap.ts":  "const gap = 1;",
		"zero.ts": "const zero = 1;",
	})
	targets := [][]string{{paths["a.ts"], paths["bad.ts"], paths["gap.ts"], paths["zero.ts"]}}
	syntaxErrorFiles := map[string]struct{}{paths["bad.ts"]: {}}
	typeInfoFiles := map[string]struct{}{paths["a.ts"]: {}}

	newRuleHandler := func(calls map[string]int) RuleHandler {
		return func(file *ast.SourceFile) []ConfiguredRule {
			calls[file.FileName()]++
			switch file.FileName() {
			case paths["a.ts"]:
				rules := noopRule()
				return append(rules, ConfiguredRule{
					Name:               "community/plugin-rule",
					Severity:           rule.SeverityWarning,
					IsEslintPluginRule: true,
				})
			case paths["gap.ts"]:
				typeAwareRule := noopRule()[0]
				typeAwareRule.Name = "type-aware-rule"
				typeAwareRule.RequiresTypeInfo = true
				return []ConfiguredRule{typeAwareRule}
			default:
				return nil
			}
		}
	}

	directCalls := make(map[string]int)
	directResult, err := RunLinter(RunLinterOptions{
		Programs:         wrapTestPrograms(program),
		SingleThreaded:   true,
		TargetFiles:      targets,
		GetRulesForFile:  newRuleHandler(directCalls),
		TypeInfoFiles:    typeInfoFiles,
		SyntaxErrorFiles: syntaxErrorFiles,
	})
	if err != nil {
		t.Fatalf("direct RunLinter failed: %v", err)
	}

	preparedCalls := make(map[string]int)
	preparedOpts := RunLinterOptions{
		Programs:         wrapTestPrograms(program),
		SingleThreaded:   true,
		TargetFiles:      targets,
		GetRulesForFile:  newRuleHandler(preparedCalls),
		TypeInfoFiles:    typeInfoFiles,
		SyntaxErrorFiles: syntaxErrorFiles,
	}
	preparedOpts.PreparedPlan = mustPrepareLintPlan(t, preparedOpts)
	pluginTargets := preparedOpts.PreparedPlan.Targets()
	if len(pluginTargets) != 1 || pluginTargets[0].File.FileName() != paths["a.ts"] {
		t.Fatalf("prepared plugin projection = %+v, want only a.ts", pluginTargets)
	}
	if len(pluginTargets[0].Rules) != 2 {
		t.Fatalf("prepared a.ts rules = %d, want native and plugin rules", len(pluginTargets[0].Rules))
	}

	preparedResult, err := RunLinter(preparedOpts)
	if err != nil {
		t.Fatalf("prepared RunLinter failed: %v", err)
	}
	if !reflect.DeepEqual(preparedResult, directResult) {
		t.Fatalf("prepared result = %#v, direct result = %#v", preparedResult, directResult)
	}
	if preparedResult.LintedFileCount != 4 {
		t.Fatalf("prepared LintedFileCount = %d, want syntax and zero-rule files included", preparedResult.LintedFileCount)
	}
	if _, ok := preparedResult.ExecutedRules["community/plugin-rule"]; !ok {
		t.Fatal("prepared ExecutedRules omitted the configured plugin rule")
	}
	if _, ok := preparedResult.ExecutedRules["type-aware-rule"]; ok {
		t.Fatal("prepared ExecutedRules retained a type-aware rule for a gap file")
	}
	if !reflect.DeepEqual(preparedCalls, directCalls) {
		t.Fatalf("prepared callback calls = %v, direct callback calls = %v", preparedCalls, directCalls)
	}
	if preparedCalls[paths["a.ts"]] != 1 || preparedCalls[paths["gap.ts"]] != 1 || preparedCalls[paths["zero.ts"]] != 1 {
		t.Fatalf("prepared callback should run exactly once per eligible file, got %v", preparedCalls)
	}
	if preparedCalls[paths["bad.ts"]] != 0 {
		t.Fatalf("prepared callback ran for syntax-error file: %v", preparedCalls)
	}
}

func TestLintPlanRunsSourceOnlyProgramWithoutChecker(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"gap.ts": "const gap = missing;",
	})
	file := program.GetSourceFile(paths["gap.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain gap.ts")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, program, []*ast.SourceFile{file})

	typeInfoCases := []struct {
		name  string
		files map[string]struct{}
	}{
		{name: "nil"},
		{name: "contains-file", files: map[string]struct{}{file.FileName(): {}}},
	}
	for _, typeInfo := range typeInfoCases {
		for _, prepared := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-info=%s/prepared=%t", typeInfo.name, prepared), func(t *testing.T) {
				nativeRuns := 0
				typeAwareRuns := 0
				var reports atomic.Int32
				opts := RunLinterOptions{
					Programs:         testPrograms(sourceOnly),
					SingleThreaded:   true,
					SyntaxErrorFiles: map[string]struct{}{},
					TypeInfoFiles:    typeInfo.files,
					GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
						return []ConfiguredRule{
							{
								Name:     "source-only-native",
								Severity: rule.SeverityError,
								Run: func(ctx rule.RuleContext) rule.RuleListeners {
									nativeRuns++
									if ctx.Program() != sourceOnly || ctx.Program().CanProvideTypeChecker(ctx.SourceFile) || ctx.TypeChecker != nil {
										t.Fatal("source-only Program received the wrong capabilities")
									}
									if !ctx.Program().IsValid() || ctx.Refs == nil || !ctx.SourceFile.IsBound() {
										t.Fatal("source-only Program lost source or binder services")
									}
									return rule.RuleListeners{
										ast.KindIdentifier: func(node *ast.Node) {
											if node.Text() == "missing" {
												ctx.ReportNode(node, rule.RuleMessage{Description: "source-only"})
											}
										},
									}
								},
							},
							{
								Name:             "source-only-type-aware",
								Severity:         rule.SeverityError,
								RequiresTypeInfo: true,
								Run: func(rule.RuleContext) rule.RuleListeners {
									typeAwareRuns++
									return nil
								},
							},
						}
					},
					Consumer: rule.DiagnosticConsumer{
						Demand: rule.EditDemandNone,
						Report: func(rule.RuleDiagnostic) {
							reports.Add(1)
						},
					},
				}
				// Program capability is authoritative even if run policy names this
				// file in TypeInfoFiles.
				if prepared {
					opts.PreparedPlan = mustPrepareLintPlan(t, opts)
					targets := opts.PreparedPlan.Targets()
					if len(targets) != 1 || targets[0].File != file || len(targets[0].Rules) != 1 ||
						targets[0].Rules[0].Name != "source-only-native" {
						t.Fatalf("source-only prepared targets = %+v", targets)
					}
				}

				result, err := RunLinter(opts)
				if err != nil {
					t.Fatalf("RunLinter: %v", err)
				}
				if result.LintedFileCount != 1 || nativeRuns != 1 || typeAwareRuns != 0 || reports.Load() != 1 {
					t.Fatalf(
						"source-only result differs: files=%d native=%d typeAware=%d reports=%d",
						result.LintedFileCount,
						nativeRuns,
						typeAwareRuns,
						reports.Load(),
					)
				}
				if _, ok := result.ExecutedRules["source-only-native"]; !ok {
					t.Fatal("source-only native rule missing from ExecutedRules")
				}
				if _, ok := result.ExecutedRules["source-only-type-aware"]; ok {
					t.Fatal("type-aware rule executed for source-only Program")
				}
			})
		}
	}
}

func TestSourceOnlyPlanSeparatesUniverseFromExecutionProjection(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
		"b.ts": "export const b = 1;",
	})
	a := program.GetSourceFile(paths["a.ts"])
	b := program.GetSourceFile(paths["b.ts"])
	if a == nil || b == nil {
		t.Fatal("fixture Program did not contain both source files")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, program, []*ast.SourceFile{nil, a, a, a, b})
	for _, prepared := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepared=%t", prepared), func(t *testing.T) {
			var resolved atomic.Int32
			var runs atomic.Int32
			opts := RunLinterOptions{
				Programs:         testPrograms(sourceOnly),
				SingleThreaded:   true,
				ExcludePaths:     []string{string(b.Path())},
				SyntaxErrorFiles: map[string]struct{}{},
				GetRulesForFile: func(file *ast.SourceFile) []ConfiguredRule {
					resolved.Add(1)
					if file != a {
						t.Fatalf("resolved rules for %q, want only a.ts", file.FileName())
					}
					return []ConfiguredRule{{
						Name:     "source-only-projection",
						Severity: rule.SeverityError,
						Run: func(rule.RuleContext) rule.RuleListeners {
							runs.Add(1)
							return nil
						},
					}}
				},
			}
			if prepared {
				opts.PreparedPlan = mustPrepareLintPlan(t, opts)
				plan := opts.PreparedPlan.programs[0]
				if !slices.Equal(plan.program.SourceFiles(), []*ast.SourceFile{a, b}) {
					t.Fatalf("source universe = %v, want [a.ts b.ts]", plan.program.SourceFiles())
				}
				if len(plan.files) != 1 || plan.files[0].file != a {
					t.Fatalf("execution projection = %v, want [a.ts]", plan.files)
				}
			}

			result, err := RunLinter(opts)
			if err != nil {
				t.Fatalf("RunLinter: %v", err)
			}
			if result.LintedFileCount != 1 || resolved.Load() != 1 || runs.Load() != 1 {
				t.Fatalf(
					"source-only projection result: files=%d resolved=%d runs=%d",
					result.LintedFileCount,
					resolved.Load(),
					runs.Load(),
				)
			}
		})
	}
}

type sourceOnlyDerivedCacheTestKey struct{}

func TestSourceOnlyProgramSharesModuleGraphAndDerivedCache(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `import "./b"; export const a = 1;`,
		"b.ts": `import "./a"; export const b = 1;`,
	})
	files := []*ast.SourceFile{
		program.GetSourceFile(paths["a.ts"]),
		program.GetSourceFile(paths["b.ts"]),
	}
	for i, file := range files {
		if file == nil {
			t.Fatalf("fixture Program did not contain source file %d", i)
		}
	}
	sourceOnly := mustSourceOnlyTestProgram(t, program, files)

	var cacheBuilds atomic.Int32
	opts := RunLinterOptions{
		Programs:         testPrograms(sourceOnly),
		SingleThreaded:   true,
		TypeInfoFiles:    map[string]struct{}{},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "source-only-modules",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.Program() != sourceOnly || !ctx.Program().IsValid() || ctx.Program().CanProvideTypeChecker(ctx.SourceFile) {
						t.Fatal("source-only rule context lost its Program capabilities")
					}
					moduleGraph := rule.ModuleGraphFor(ctx.Program())
					if got := len(moduleGraph.Files()); got != len(files) {
						t.Fatalf("source-only module graph has %d files, want %d", got, len(files))
					}
					edges := moduleGraph.Edges(ctx.SourceFile, rule.ModuleSyntax{ESModule: true})
					if len(edges) != 1 || edges[0].Target == nil {
						t.Fatalf("source-only module edges = %+v", edges)
					}
					value := rule.CachedByProgram(ctx, sourceOnlyDerivedCacheTestKey{}, func() int {
						cacheBuilds.Add(1)
						return 42
					})
					if value != 42 {
						t.Fatalf("source-only derived cache value = %d", value)
					}
					return nil
				},
			}}
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != int32(len(files)) {
		t.Fatalf("LintedFileCount = %d, want %d", result.LintedFileCount, len(files))
	}
	if got := cacheBuilds.Load(); got != 1 {
		t.Fatalf("source-only derived cache built %d times, want 1", got)
	}
}

func TestSourceOnlyProgramRunsProgramIndexedImportRule(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `import "./b"; export const a = 1;`,
		"b.ts": `import "./a"; export const b = 1;`,
	})
	files := []*ast.SourceFile{
		program.GetSourceFile(paths["a.ts"]),
		program.GetSourceFile(paths["b.ts"]),
	}
	for i, file := range files {
		if file == nil {
			t.Fatalf("fixture Program did not contain source file %d", i)
		}
	}
	sourceOnly := mustSourceOnlyTestProgram(t, program, files)

	var reports atomic.Int32
	opts := RunLinterOptions{
		Programs:         testPrograms(sourceOnly),
		SingleThreaded:   true,
		ExcludePaths:     []string{string(files[1].Path())},
		TypeInfoFiles:    map[string]struct{}{},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     no_cycle.NoCycleRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if got := rule.ModuleGraphFor(ctx.Program()).Files(); !slices.Equal(got, files) {
						t.Fatalf("module graph files = %v, want complete source set", got)
					}
					return no_cycle.NoCycleRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				if diagnostic.RuleName != no_cycle.NoCycleRule.Name {
					t.Errorf("source-only diagnostic rule = %q", diagnostic.RuleName)
				}
				reports.Add(1)
			},
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 1 || reports.Load() != 1 {
		t.Fatalf("source-only no-cycle result: files=%d reports=%d", result.LintedFileCount, reports.Load())
	}
}

func checkerFreeExecutionTestOptions(
	t *testing.T,
	singleThreaded bool,
	run func(rule.RuleContext) rule.RuleListeners,
) RunLinterOptions {
	t.Helper()
	fileCount := minCheckerFreeFilesPerLintWorker * 2
	sources := make(map[string]string, fileCount)
	names := make([]string, fileCount)
	for i := range fileCount {
		name := fmt.Sprintf("file-%03d.ts", i)
		names[i] = name
		sources[name] = "const value = 1;"
	}
	program, paths := createTestProgramWithFiles(t, sources)
	files := make([]*ast.SourceFile, 0, fileCount)
	for _, name := range names {
		file := program.GetSourceFile(paths[name])
		if file == nil {
			t.Fatalf("fixture Program did not contain %s", name)
		}
		files = append(files, file)
	}
	opts := RunLinterOptions{
		Programs:         testPrograms(mustSourceOnlyTestProgram(t, program, files)),
		SingleThreaded:   singleThreaded,
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "checker-free-concurrency",
				Severity: rule.SeverityError,
				Run:      run,
			}}
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	return opts
}

func TestCheckerFreeLintWorkerCountKeepsSmallSetsSerial(t *testing.T) {
	tests := []struct {
		files int
		procs int
		want  int
	}{
		{files: 0, procs: 10, want: 0},
		{files: 1, procs: 10, want: 1},
		{files: 255, procs: 10, want: 1},
		{files: 256, procs: 10, want: 2},
		{files: 1119, procs: 10, want: 8},
		{files: 10200, procs: 10, want: 10},
	}
	for _, test := range tests {
		if got := checkerFreeLintWorkerCount(test.files, test.procs); got != test.want {
			t.Fatalf("checkerFreeLintWorkerCount(%d, %d) = %d, want %d", test.files, test.procs, got, test.want)
		}
	}
}

func TestRunLinterParallelizesCheckerFreeFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	release := make(chan struct{})
	twoActive := make(chan struct{})
	var active atomic.Int32
	var signaled atomic.Bool
	opts := checkerFreeExecutionTestOptions(t, false, func(rule.RuleContext) rule.RuleListeners {
		if active.Add(1) >= 2 && signaled.CompareAndSwap(false, true) {
			close(twoActive)
		}
		<-release
		active.Add(-1)
		return nil
	})

	type outcome struct {
		result *LintResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := RunLinter(opts)
		done <- outcome{result: result, err: err}
	}()

	select {
	case <-twoActive:
		close(release)
	case <-time.After(5 * time.Second):
		close(release)
		result := <-done
		if result.err != nil {
			t.Fatalf("RunLinter: %v", result.err)
		}
		t.Fatal("checker-free files did not overlap across workers")
	}
	result := <-done
	if result.err != nil {
		t.Fatalf("RunLinter: %v", result.err)
	}
	wantFiles := int32(minCheckerFreeFilesPerLintWorker * 2)
	if result.result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.result.LintedFileCount, wantFiles)
	}
}

func TestRunLinterSingleThreadedSerializesCheckerFreeFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	var active atomic.Int32
	var maxActive atomic.Int32
	opts := checkerFreeExecutionTestOptions(t, true, func(rule.RuleContext) rule.RuleListeners {
		current := active.Add(1)
		for observed := maxActive.Load(); current > observed; observed = maxActive.Load() {
			if maxActive.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return nil
	})
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	wantFiles := int32(minCheckerFreeFilesPerLintWorker * 2)
	if result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.LintedFileCount, wantFiles)
	}
	if maxActive.Load() != 1 {
		t.Fatalf("single-threaded checker-free execution had %d active files", maxActive.Load())
	}
}

func TestPrepareLintPlanParallelizesRuleResolution(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 1;",
		"c.ts": "const c = 1;",
		"d.ts": "const d = 1;",
	})
	release := make(chan struct{})
	twoActive := make(chan struct{})
	type prepareResult struct {
		plan *LintPlan
		err  error
	}
	done := make(chan prepareResult, 1)
	var active atomic.Int32
	var signaled atomic.Bool

	opts := RunLinterOptions{
		Programs:    wrapTestPrograms(program),
		TargetFiles: [][]string{{paths["a.ts"], paths["b.ts"], paths["c.ts"], paths["d.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			if active.Add(1) >= 2 && signaled.CompareAndSwap(false, true) {
				close(twoActive)
			}
			<-release
			active.Add(-1)
			return noopRule()
		},
		SyntaxErrorFiles: map[string]struct{}{},
	}
	go func() {
		plan, err := PrepareLintPlan(opts)
		done <- prepareResult{plan: plan, err: err}
	}()

	select {
	case <-twoActive:
		close(release)
	case <-time.After(5 * time.Second):
		close(release)
		<-done
		t.Fatal("rule resolution did not overlap across workers")
	}
	prepared := <-done
	if prepared.err != nil {
		t.Fatalf("PrepareLintPlan: %v", prepared.err)
	}
	if len(prepared.plan.Targets()) != 4 {
		t.Fatalf("prepared targets = %d, want 4", len(prepared.plan.Targets()))
	}
}

func TestPrepareLintPlanHonorsSingleThreadedOrder(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 1;",
		"c.ts": "const c = 1;",
	})
	wantOrder := []string{paths["c.ts"], paths["a.ts"], paths["b.ts"]}
	var gotOrder []string
	plan := mustPrepareLintPlan(t, RunLinterOptions{
		Programs:         wrapTestPrograms(program),
		SingleThreaded:   true,
		TargetFiles:      [][]string{wantOrder},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(file *ast.SourceFile) []ConfiguredRule {
			gotOrder = append(gotOrder, file.FileName())
			return noopRule()
		},
	})
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("single-threaded resolution order = %v, want %v", gotOrder, wantOrder)
	}
	if gotTargets := plan.Targets(); len(gotTargets) != len(wantOrder) {
		t.Fatalf("single-threaded prepared targets = %d, want %d", len(gotTargets), len(wantOrder))
	}
}

func TestPreparedLintPlanPreservesSameFileAcrossPrograms(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"shared.ts": "const shared = 1;",
	})
	calls := 0
	opts := RunLinterOptions{
		Programs:         wrapTestPrograms(program, program),
		SingleThreaded:   true,
		TargetFiles:      [][]string{{paths["shared.ts"]}, {paths["shared.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			calls++
			return noopRule()
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	if targets := opts.PreparedPlan.Targets(); len(targets) != 2 {
		t.Fatalf("prepared targets = %d, want one entry per Program", len(targets))
	}
	result, err := RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter failed: %v", err)
	}
	if result.LintedFileCount != 2 {
		t.Fatalf("LintedFileCount = %d, want one count per Program", result.LintedFileCount)
	}
	if calls != 2 {
		t.Fatalf("GetRulesForFile calls = %d, want one per Program and no execution-time repeats", calls)
	}
}

func TestRunLinterRejectsPreparedPlanForDifferentProgram(t *testing.T) {
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programB, pathsB := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const b = 1;",
	})
	plan := mustPrepareLintPlan(t, RunLinterOptions{
		Programs:         wrapTestPrograms(programA),
		SingleThreaded:   true,
		TargetFiles:      [][]string{{pathsA["a.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
	})
	_, err := RunLinter(RunLinterOptions{
		Programs:         wrapTestPrograms(programB),
		SingleThreaded:   true,
		TargetFiles:      [][]string{{pathsB["b.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
		PreparedPlan:     plan,
	})
	if err == nil {
		t.Fatal("RunLinter accepted a prepared plan bound to a different Program")
	}
}

func TestRunLinterRejectsNilProgramBeforeLintSideEffects(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	var ruleCalls atomic.Int32
	var filterCalls atomic.Int32
	opts := RunLinterOptions{
		Programs:         append(wrapTestPrograms(raw), nil),
		TargetFiles:      [][]string{{paths["a.ts"]}, nil},
		PerProgramFilter: []FileFilter{func(string) bool { filterCalls.Add(1); return true }},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
	}
	if _, err := PrepareLintPlan(opts); !errors.Is(err, errNilProgram) {
		t.Fatalf("PrepareLintPlan nil Program error = %v", err)
	}
	if ruleCalls.Load() != 0 || filterCalls.Load() != 0 {
		t.Fatalf("PrepareLintPlan produced side effects before rejecting nil Program: rules=%d filters=%d", ruleCalls.Load(), filterCalls.Load())
	}
	if _, err := RunLinter(opts); !errors.Is(err, errNilProgram) {
		t.Fatalf("RunLinter nil Program error = %v", err)
	}
	if ruleCalls.Load() != 0 || filterCalls.Load() != 0 {
		t.Fatalf("RunLinter produced side effects before rejecting nil Program: rules=%d filters=%d", ruleCalls.Load(), filterCalls.Load())
	}
	typeCheckOnly := RunLinterOptions{
		Programs:  wrapTestPrograms(raw),
		TypeCheck: true,
	}
	typeCheckOnly.Programs = append(typeCheckOnly.Programs, nil)
	if _, err := RunLinter(typeCheckOnly); !errors.Is(err, errNilProgram) {
		t.Fatalf("type-check-only RunLinter nil Program error = %v", err)
	}
	noOp, err := RunLinter(RunLinterOptions{Programs: []*lintprogram.Program{nil}})
	if err != nil || noOp.LintedFileCount != 0 || len(noOp.ExecutedRules) != 0 {
		t.Fatalf("no-op RunLinter result = %+v, error = %v", noOp, err)
	}

	invalid := &lintprogram.Program{}
	invalidOpts := RunLinterOptions{
		Programs:        []*lintprogram.Program{invalid},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
	}
	if _, err := PrepareLintPlan(invalidOpts); !errors.Is(err, errInvalidProgram) {
		t.Fatalf("PrepareLintPlan zero Program error = %v", err)
	}
	if _, err := RunLinter(invalidOpts); !errors.Is(err, errInvalidProgram) {
		t.Fatalf("RunLinter zero Program error = %v", err)
	}
	invalidSkipMask := RunLinterOptions{
		Programs:              wrapTestPrograms(raw),
		TypeCheck:             true,
		SkipTypeCheckPrograms: []bool{},
	}
	if _, err := RunLinter(invalidSkipMask); !errors.Is(err, errInvalidTypeCheckSkipMask) {
		t.Fatalf("RunLinter invalid type-check skip mask error = %v", err)
	}
}

func TestPreparedLintPlanFreezesPerFileTypeInfoEligibility(t *testing.T) {
	for _, initiallyHasTypeInfo := range []bool{false, true} {
		t.Run(fmt.Sprintf("initially-has-type-info=%t", initiallyHasTypeInfo), func(t *testing.T) {
			raw, paths := createTestProgramWithFiles(t, map[string]string{
				"a.ts": "const a = 1;",
			})
			typeInfoFiles := map[string]struct{}{}
			if initiallyHasTypeInfo {
				typeInfoFiles[paths["a.ts"]] = struct{}{}
			}
			var runs atomic.Int32
			sourceProgram := wrapTestPrograms(raw)[0]
			opts := RunLinterOptions{
				Programs:         testPrograms(sourceProgram),
				TargetFiles:      [][]string{{paths["a.ts"]}},
				TypeInfoFiles:    typeInfoFiles,
				SyntaxErrorFiles: map[string]struct{}{},
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					return []ConfiguredRule{{
						Name:             "type-aware-probe",
						RequiresTypeInfo: true,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							if ctx.TypeChecker == nil {
								t.Fatal("prepared type-aware rule received nil checker")
							}
							if ctx.Program() != sourceProgram {
								t.Fatal("prepared type-aware rule lost its Program generation")
							}
							runs.Add(1)
							return nil
						},
					}}
				},
			}
			opts.PreparedPlan = mustPrepareLintPlan(t, opts)
			if initiallyHasTypeInfo {
				delete(typeInfoFiles, paths["a.ts"])
			} else {
				typeInfoFiles[paths["a.ts"]] = struct{}{}
			}
			if _, err := RunLinter(opts); err != nil {
				t.Fatalf("RunLinter: %v", err)
			}
			wantRuns := int32(0)
			if initiallyHasTypeInfo {
				wantRuns = 1
			}
			if got := runs.Load(); got != wantRuns {
				t.Fatalf("type-aware runs = %d, want %d", got, wantRuns)
			}
		})
	}
}

func TestLintSingleFileWithoutRuleHandlerIsNoOp(t *testing.T) {
	LintSingleFile(LintSingleFileOptions{})
}
