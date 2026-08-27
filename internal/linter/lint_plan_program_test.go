package linter

import (
	"fmt"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func mustPrepareLintPlan(t *testing.T, opts PrepareLintPlanOptions) *LintPlan {
	t.Helper()
	plan, err := PrepareLintPlan(opts)
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	return plan
}

func TestExactLintProjectionReusesUnchangedProgramUniverse(t *testing.T) {
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

	exact, err := resolveExactProgramFiles(sourceProgram, []string{a.FileName(), b.FileName()})
	if err != nil {
		t.Fatal(err)
	}
	if len(exact) != len(universe) || &exact[0] != &universe[0] {
		t.Fatal("unchanged exact projection copied the Program source universe")
	}

	subset, err := resolveExactProgramFiles(sourceProgram, []string{a.FileName()})
	if err != nil {
		t.Fatal(err)
	}
	if len(subset) != 1 || subset[0] != a || len(universe) != 2 || universe[1] != b {
		t.Fatalf("subset projection=%v corrupted universe=%v", subset, universe)
	}
}

func TestPreparedLintPlanPreservesNativeSemanticsAndIsReused(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts":    "const a = 1;",
		"gap.ts":  "const gap = 1;",
		"zero.ts": "const zero = 1;",
	})
	gapFile := program.GetSourceFile(paths["gap.ts"])
	if gapFile == nil {
		t.Fatal("fixture Program did not contain gap.ts")
	}
	typed := lintprogram.NewFromCompiler(program)
	gap := mustSourceOnlyTestProgram(t, program, []*ast.SourceFile{gapFile})
	programs := testPrograms(typed, gap)
	targets := [][]string{
		{paths["a.ts"], paths["zero.ts"]},
		{paths["gap.ts"]},
	}

	newRuleHandler := func(calls map[string]int) RuleHandler {
		return func(file *ast.SourceFile) []rule.ConfiguredRule {
			calls[file.FileName()]++
			switch file.FileName() {
			case paths["a.ts"]:
				rules := noopRule()
				return append(rules, rule.ConfiguredRule{
					Name:               "community/plugin-rule",
					Severity:           rule.SeverityWarning,
					IsEslintPluginRule: true,
				})
			case paths["gap.ts"]:
				typeAwareRule := noopRule()[0]
				typeAwareRule.Name = "type-aware-rule"
				typeAwareRule.RequiresTypeInfo = true
				return []rule.ConfiguredRule{typeAwareRule}
			default:
				return nil
			}
		}
	}

	preparedCalls := make(map[string]int)
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		SingleThreaded:   true,
		TargetsByProgram: targets,
		GetRulesForFile:  newRuleHandler(preparedCalls),
	})
	pluginTargets := plan.Targets()
	if len(pluginTargets) != 1 || pluginTargets[0].File.FileName() != paths["a.ts"] {
		t.Fatalf("prepared plugin projection = %+v, want only a.ts", pluginTargets)
	}
	if len(pluginTargets[0].Rules) != 2 {
		t.Fatalf("prepared a.ts rules = %d, want native and plugin rules", len(pluginTargets[0].Rules))
	}

	preparedResult, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       plan,
	})
	if err != nil {
		t.Fatalf("prepared RunLinter failed: %v", err)
	}
	if preparedResult.LintedFileCount != 3 {
		t.Fatalf("prepared LintedFileCount = %d, want zero-rule files included", preparedResult.LintedFileCount)
	}
	if _, ok := preparedResult.ExecutedRules["community/plugin-rule"]; !ok {
		t.Fatal("prepared ExecutedRules omitted the configured plugin rule")
	}
	if _, ok := preparedResult.ExecutedRules["type-aware-rule"]; ok {
		t.Fatal("prepared ExecutedRules retained a type-aware rule for a gap file")
	}
	if preparedCalls[paths["a.ts"]] != 1 || preparedCalls[paths["gap.ts"]] != 1 || preparedCalls[paths["zero.ts"]] != 1 {
		t.Fatalf("prepared callback should run exactly once per eligible file, got %v", preparedCalls)
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
	nativeRuns := 0
	typeAwareRuns := 0
	var reports atomic.Int32
	programs := testPrograms(sourceOnly)
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{file.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{
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
	})
	targets := plan.Targets()
	if len(targets) != 1 || targets[0].File != file || len(targets[0].Rules) != 1 ||
		targets[0].Rules[0].Name != "source-only-native" {
		t.Fatalf("source-only prepared targets = %+v", targets)
	}
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       plan,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(rule.RuleDiagnostic) {
				reports.Add(1)
			},
		},
	})
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
	var resolved atomic.Int32
	var runs atomic.Int32
	programs := testPrograms(sourceOnly)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{a.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(file *ast.SourceFile) []rule.ConfiguredRule {
			resolved.Add(1)
			if file != a {
				t.Fatalf("resolved rules for %q, want only a.ts", file.FileName())
			}
			return []rule.ConfiguredRule{{
				Name:     "source-only-projection",
				Severity: rule.SeverityError,
				Run: func(rule.RuleContext) rule.RuleListeners {
					runs.Add(1)
					return nil
				},
			}}
		},
	})
	plan := lintPlan.programs[0]
	if !slices.Equal(plan.program.SourceFiles(), []*ast.SourceFile{a, b}) {
		t.Fatalf("source universe = %v, want [a.ts b.ts]", plan.program.SourceFiles())
	}
	if len(plan.files) != 1 || plan.files[0].file != a {
		t.Fatalf("execution projection = %v, want [a.ts]", plan.files)
	}
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
	})
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
	programs := testPrograms(sourceOnly)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{files[0].FileName(), files[1].FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     "source-only-modules",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.Program() != sourceOnly || !ctx.Program().IsValid() || ctx.Program().CanProvideTypeChecker(ctx.SourceFile) {
						t.Fatal("source-only rule context lost its Program capabilities")
					}
					moduleGraph := ctx.Program().ModuleGraph()
					if got := len(moduleGraph.Files()); got != len(files) {
						t.Fatalf("source-only module graph has %d files, want %d", got, len(files))
					}
					references := moduleGraph.References(ctx.SourceFile, lintprogram.ESModuleReferences)
					if len(references) != 1 || references[0].Target == nil {
						t.Fatalf("source-only module references = %+v", references)
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
	})
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
	})
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
	programs := testPrograms(sourceOnly)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{files[0].FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_cycle.NoCycleRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if got := ctx.Program().ModuleGraph().Files(); !slices.Equal(got, files) {
						t.Fatalf("module graph files = %v, want complete source set", got)
					}
					return no_cycle.NoCycleRule.Run(ctx, nil)
				},
			}}
		},
	})
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				if diagnostic.RuleName != no_cycle.NoCycleRule.Name {
					t.Errorf("source-only diagnostic rule = %q", diagnostic.RuleName)
				}
				reports.Add(1)
			},
		},
	})
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
	programs := testPrograms(mustSourceOnlyTestProgram(t, program, files))
	targets := make([]string, len(files))
	for index, file := range files {
		targets[index] = file.FileName()
	}
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{targets},
		SingleThreaded:   singleThreaded,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     "checker-free-concurrency",
				Severity: rule.SeverityError,
				Run:      run,
			}}
		},
	})
	return RunLinterOptions{
		SingleThreaded: singleThreaded,
		LintPlan:       lintPlan,
	}
}

func mustSourceOnlyTestProgram(t testing.TB, typeScript *compiler.Program, files []*ast.SourceFile) *lintprogram.Program {
	t.Helper()
	sourceProgram, err := lintprogram.NewFromBoundSources(typeScript, files)
	if err != nil {
		t.Fatalf("NewFromBoundSources: %v", err)
	}
	return sourceProgram
}

func wrapTestPrograms(programs ...*compiler.Program) []*lintprogram.Program {
	return lintprogram.NewFromCompilers(programs)
}

func testPrograms(programs ...*lintprogram.Program) []*lintprogram.Program {
	return programs
}
