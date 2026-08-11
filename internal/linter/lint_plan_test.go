package linter

import (
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
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

type lintPlanTestRuntime struct {
	rule.SourceRuntime
	owned     map[*ast.SourceFile]struct{}
	ownsCalls atomic.Int32
}

func (r *lintPlanTestRuntime) SameSourceRuntime(other rule.SourceRuntime) bool {
	otherRuntime, ok := other.(*lintPlanTestRuntime)
	return ok && r == otherRuntime
}

func (r *lintPlanTestRuntime) OwnsSourceFile(file *ast.SourceFile) bool {
	r.ownsCalls.Add(1)
	_, ok := r.owned[file]
	return ok
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
		Programs:         []*compiler.Program{program},
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
		Programs:         []*compiler.Program{program},
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

func TestLintPlanRunsStandaloneSourceWithoutProgram(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"gap.ts": "const gap = missing;",
	})
	file := program.GetSourceFile(paths["gap.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain gap.ts")
	}

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
					Standalone: []StandaloneLintSourceSet{{
						Files:   []*ast.SourceFile{file},
						Runtime: rule.SourceRuntimeForProgram(program),
					}},
					SingleThreaded:   true,
					SyntaxErrorFiles: map[string]struct{}{},
					TypeInfoFiles:    typeInfo.files,
					GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
						return []ConfiguredRule{
							{
								Name:     "standalone-native",
								Severity: rule.SeverityError,
								Run: func(ctx rule.RuleContext) rule.RuleListeners {
									nativeRuns++
									if ctx.Program != nil || ctx.TypeChecker != nil {
										t.Fatal("standalone source received Program or TypeChecker")
									}
									if !ctx.HasSourceRuntime() || ctx.Refs == nil || !ctx.SourceFile.IsBound() {
										t.Fatal("standalone source lost runtime or binder services")
									}
									return rule.RuleListeners{
										ast.KindIdentifier: func(node *ast.Node) {
											if node.Text() == "missing" {
												ctx.ReportNode(node, rule.RuleMessage{Description: "standalone"})
											}
										},
									}
								},
							},
							{
								Name:             "standalone-type-aware",
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
				// Standalone itself is the authoritative declaration that no type
				// information exists, even if a caller includes its path in the
				// Program-backed TypeInfoFiles set.
				if prepared {
					opts.PreparedPlan = mustPrepareLintPlan(t, opts)
					targets := opts.PreparedPlan.Targets()
					if len(targets) != 1 || targets[0].File != file || len(targets[0].Rules) != 1 ||
						targets[0].Rules[0].Name != "standalone-native" {
						t.Fatalf("standalone prepared targets = %+v", targets)
					}
				}

				result, err := RunLinter(opts)
				if err != nil {
					t.Fatalf("RunLinter: %v", err)
				}
				if result.LintedFileCount != 1 || nativeRuns != 1 || typeAwareRuns != 0 || reports.Load() != 1 {
					t.Fatalf(
						"standalone result differs: files=%d native=%d typeAware=%d reports=%d",
						result.LintedFileCount,
						nativeRuns,
						typeAwareRuns,
						reports.Load(),
					)
				}
				if _, ok := result.ExecutedRules["standalone-native"]; !ok {
					t.Fatal("standalone native rule missing from ExecutedRules")
				}
				if _, ok := result.ExecutedRules["standalone-type-aware"]; ok {
					t.Fatal("type-aware rule executed for standalone source")
				}
			})
		}
	}
}

func TestStandalonePlanSeparatesSourceUniverseFromExecutionProjection(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
		"b.ts": "export const b = 1;",
	})
	a := program.GetSourceFile(paths["a.ts"])
	b := program.GetSourceFile(paths["b.ts"])
	if a == nil || b == nil {
		t.Fatal("fixture Program did not contain both source files")
	}
	for _, prepared := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepared=%t", prepared), func(t *testing.T) {
			var resolved atomic.Int32
			var runs atomic.Int32
			opts := RunLinterOptions{
				Standalone: []StandaloneLintSourceSet{{
					Files:   []*ast.SourceFile{nil, a, a, a, b},
					Runtime: rule.SourceRuntimeForProgram(program),
				}},
				SingleThreaded:   true,
				ExcludePaths:     []string{string(b.Path())},
				SyntaxErrorFiles: map[string]struct{}{},
				GetRulesForFile: func(file *ast.SourceFile) []ConfiguredRule {
					resolved.Add(1)
					if file != a {
						t.Fatalf("resolved rules for %q, want only a.ts", file.FileName())
					}
					return []ConfiguredRule{{
						Name:     "standalone-projection",
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
				plan := opts.PreparedPlan.standalone[0]
				if !sameSourceFiles(plan.sourceFiles, []*ast.SourceFile{a, b}) {
					t.Fatalf("standalone source universe = %v, want [a.ts b.ts]", plan.sourceFiles)
				}
				if !sameSourceFiles(plan.files, []*ast.SourceFile{a}) {
					t.Fatalf("standalone execution projection = %v, want [a.ts]", plan.files)
				}
			}

			result, err := RunLinter(opts)
			if err != nil {
				t.Fatalf("RunLinter: %v", err)
			}
			if result.LintedFileCount != 1 || resolved.Load() != 1 || runs.Load() != 1 {
				t.Fatalf(
					"standalone projection result: files=%d resolved=%d runs=%d",
					result.LintedFileCount,
					resolved.Load(),
					runs.Load(),
				)
			}
		})
	}
}

func TestStandaloneRejectsDifferentASTsForSamePath(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
	})
	a := program.GetSourceFile(paths["a.ts"])
	if a == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	reparsed := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: a.FileName(),
		Path:     a.Path(),
	}, a.Text(), core.ScriptKindTS)
	opts := RunLinterOptions{
		Standalone: []StandaloneLintSourceSet{{
			// Put the Runtime's source second: silently keeping the first would
			// make module-resolution targets disagree with graph node identity.
			Files:   []*ast.SourceFile{reparsed, a},
			Runtime: rule.SourceRuntimeForProgram(program),
		}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			t.Fatal("conflicting standalone source set resolved rules")
			return nil
		},
	}
	want := fmt.Sprintf("linter: standalone source set contains different ASTs for path %q", a.Path())
	if _, err := PrepareLintPlan(opts); err == nil || err.Error() != want {
		t.Fatalf("PrepareLintPlan conflict error = %v, want %q", err, want)
	}
	if _, err := RunLinter(opts); err == nil || err.Error() != want {
		t.Fatalf("RunLinter conflict error = %v, want %q", err, want)
	}
}

func TestStandaloneEmptySourceSet(t *testing.T) {
	for _, files := range [][]*ast.SourceFile{nil, {}} {
		for _, prepared := range []bool{false, true} {
			t.Run(fmt.Sprintf("nil=%t/prepared=%t", files == nil, prepared), func(t *testing.T) {
				opts := RunLinterOptions{
					Standalone: []StandaloneLintSourceSet{{Files: files}},
					GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
						t.Fatal("empty standalone source set resolved rules")
						return nil
					},
				}
				if prepared {
					opts.PreparedPlan = mustPrepareLintPlan(t, opts)
				}
				result, err := RunLinter(opts)
				if err != nil {
					t.Fatalf("RunLinter: %v", err)
				}
				if result.LintedFileCount != 0 {
					t.Fatalf("LintedFileCount = %d, want 0", result.LintedFileCount)
				}
			})
		}
	}
}

func TestStandaloneNonEmptySourceSetRequiresRuntime(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
	})
	file := program.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	const want = "linter: non-empty standalone source set requires a runtime"
	var typedNil *lintPlanTestRuntime
	for _, runtime := range []rule.SourceRuntime{nil, typedNil} {
		t.Run(fmt.Sprintf("typed-nil=%t", runtime != nil), func(t *testing.T) {
			opts := RunLinterOptions{
				Standalone: []StandaloneLintSourceSet{{
					Files:   []*ast.SourceFile{file},
					Runtime: runtime,
				}},
				SyntaxErrorFiles: map[string]struct{}{},
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					t.Fatal("runtime-less standalone source set resolved rules")
					return nil
				},
			}
			if _, err := PrepareLintPlan(opts); err == nil || err.Error() != want {
				t.Fatalf("PrepareLintPlan runtime error = %v, want %q", err, want)
			}
			if _, err := RunLinter(opts); err == nil || err.Error() != want {
				t.Fatalf("RunLinter runtime error = %v, want %q", err, want)
			}
		})
	}
}

func TestRunLinterValidatesStandaloneBeforeAnyLintSideEffects(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
	})
	file := program.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	var ruleCalls atomic.Int32
	var reports atomic.Int32
	_, err := RunLinter(RunLinterOptions{
		Programs:    []*compiler.Program{program},
		TargetFiles: [][]string{{file.FileName()}},
		Standalone:  []StandaloneLintSourceSet{{Files: []*ast.SourceFile{file}}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			ruleCalls.Add(1)
			return noopRule()
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) { reports.Add(1) },
		},
	})
	const want = "linter: non-empty standalone source set requires a runtime"
	if err == nil || err.Error() != want {
		t.Fatalf("RunLinter runtime error = %v, want %q", err, want)
	}
	if ruleCalls.Load() != 0 || reports.Load() != 0 {
		t.Fatalf("validation leaked lint side effects: rules=%d reports=%d", ruleCalls.Load(), reports.Load())
	}
}

func TestStandaloneSourceMustBeBoundAndOwnedByRuntime(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
	})
	owned := program.GetSourceFile(paths["a.ts"])
	if owned == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	reparsed := func() *ast.SourceFile {
		return parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: owned.FileName(),
			Path:     owned.Path(),
		}, owned.Text(), core.ScriptKindTS)
	}

	tests := []struct {
		name string
		file *ast.SourceFile
		want string
	}{
		{
			name: "unbound",
			file: reparsed(),
			want: fmt.Sprintf("linter: standalone source %q is not bound", owned.FileName()),
		},
		{
			name: "different-bound-generation",
			file: func() *ast.SourceFile {
				file := reparsed()
				binder.BindSourceFile(file)
				return file
			}(),
			want: fmt.Sprintf("linter: standalone runtime does not own source %q", owned.FileName()),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := RunLinterOptions{
				Standalone: []StandaloneLintSourceSet{{
					Files:   []*ast.SourceFile{test.file},
					Runtime: rule.SourceRuntimeForProgram(program),
				}},
				SyntaxErrorFiles: map[string]struct{}{},
				GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
					t.Fatal("invalid standalone source resolved rules")
					return nil
				},
			}
			if _, err := PrepareLintPlan(opts); err == nil || err.Error() != test.want {
				t.Fatalf("PrepareLintPlan source error = %v, want %q", err, test.want)
			}
			if _, err := RunLinter(opts); err == nil || err.Error() != test.want {
				t.Fatalf("RunLinter source error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestPreparedStandalonePlanBindsSourceGenerationAndRuntime(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
	})
	file := program.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	reparsed := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: file.FileName(),
		Path:     file.Path(),
	}, file.Text(), core.ScriptKindTS)
	binder.BindSourceFile(reparsed)
	runtime := &lintPlanTestRuntime{
		SourceRuntime: rule.SourceRuntimeForProgram(program),
		owned: map[*ast.SourceFile]struct{}{
			file:     {},
			reparsed: {},
		},
	}
	opts := RunLinterOptions{
		Standalone: []StandaloneLintSourceSet{{
			Files:   []*ast.SourceFile{file},
			Runtime: runtime,
		}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return noopRule()
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	runtime.ownsCalls.Store(0)
	if _, err := RunLinter(opts); err != nil {
		t.Fatalf("exact prepared source set was rejected: %v", err)
	}
	if runtime.ownsCalls.Load() != 0 {
		t.Fatalf("exact prepared source set repeated %d owner checks", runtime.ownsCalls.Load())
	}
	opts.Standalone[0].Files = []*ast.SourceFile{reparsed}
	if _, err := RunLinter(opts); err == nil || err.Error() != "linter: prepared lint plan does not match standalone files" {
		t.Fatalf("reparsed source generation error = %v", err)
	}
	opts.Standalone[0].Files = []*ast.SourceFile{file}

	// The exact runtime used to prepare the plan remains valid.
	opts.Standalone[0].Runtime = runtime
	if _, err := RunLinter(opts); err != nil {
		t.Fatalf("same runtime identity was rejected: %v", err)
	}

	opts.Standalone[0].Runtime = &lintPlanTestRuntime{
		SourceRuntime: rule.SourceRuntimeForProgram(program),
		owned:         runtime.owned,
	}
	if _, err := RunLinter(opts); err == nil || err.Error() != "linter: prepared lint plan does not match standalone runtime" {
		t.Fatalf("different runtime error = %v", err)
	}
}

func TestStandaloneExecutionProjectionReusesUnchangedSourceUniverse(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "export const a = 1;",
		"b.ts": "export const b = 1;",
	})
	files := []*ast.SourceFile{
		program.GetSourceFile(paths["a.ts"]),
		program.GetSourceFile(paths["b.ts"]),
	}
	if files[0] == nil || files[1] == nil {
		t.Fatal("fixture Program did not contain both source files")
	}
	projection := collectStandaloneFilesToLint(files, []string{"/not-present/"})
	if !sameSourceFiles(projection, files) || &projection[0] != &files[0] {
		t.Fatal("unchanged standalone execution projection did not reuse source universe")
	}

	excluded := collectStandaloneFilesToLint(files, []string{string(files[1].Path())})
	if !sameSourceFiles(excluded, files[:1]) {
		t.Fatalf("excluded execution projection = %v, want only a.ts", excluded)
	}
	if &excluded[0] == &files[0] {
		t.Fatal("excluded execution projection reused source-universe backing array")
	}
	if !sameSourceFiles(files, []*ast.SourceFile{
		program.GetSourceFile(paths["a.ts"]),
		program.GetSourceFile(paths["b.ts"]),
	}) {
		t.Fatal("execution projection mutated complete source universe")
	}
}

type standaloneDerivedCacheTestKey struct{}

func TestStandaloneSourceSharesModuleGraphAndDerivedCache(t *testing.T) {
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
			t.Fatalf("fixture Program did not contain standalone file %d", i)
		}
	}

	var cacheBuilds atomic.Int32
	opts := RunLinterOptions{
		Standalone: []StandaloneLintSourceSet{{
			Files:   files,
			Runtime: rule.SourceRuntimeForProgram(program),
		}},
		SingleThreaded:   true,
		TypeInfoFiles:    map[string]struct{}{},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "standalone-modules",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.Program != nil || !ctx.HasSourceRuntime() || ctx.Modules == nil {
						t.Fatal("standalone rule context lost its runtime module graph")
					}
					if got := len(ctx.Modules.Files()); got != len(files) {
						t.Fatalf("standalone module graph has %d files, want %d", got, len(files))
					}
					edges := ctx.Modules.Edges(ctx.SourceFile, rule.ModuleSyntax{ESModule: true})
					if len(edges) != 1 || edges[0].Target == nil {
						t.Fatalf("standalone module edges = %+v", edges)
					}
					value := rule.CachedBySourceRuntime(ctx, standaloneDerivedCacheTestKey{}, func() int {
						cacheBuilds.Add(1)
						return 42
					})
					if value != 42 {
						t.Fatalf("standalone derived cache value = %d", value)
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
		t.Fatalf("standalone derived cache built %d times, want 1", got)
	}
}

func TestStandaloneSourceRunsProgramIndexedImportRule(t *testing.T) {
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
			t.Fatalf("fixture Program did not contain standalone file %d", i)
		}
	}

	var reports atomic.Int32
	opts := RunLinterOptions{
		Standalone: []StandaloneLintSourceSet{{
			Files:   files,
			Runtime: rule.SourceRuntimeForProgram(program),
		}},
		SingleThreaded:   true,
		ExcludePaths:     []string{string(files[1].Path())},
		TypeInfoFiles:    map[string]struct{}{},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     no_cycle.NoCycleRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if got := ctx.Modules.Files(); !sameSourceFiles(got, files) {
						t.Fatalf("standalone module graph files = %v, want complete source set", got)
					}
					return no_cycle.NoCycleRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				if diagnostic.RuleName != no_cycle.NoCycleRule.Name {
					t.Errorf("standalone diagnostic rule = %q", diagnostic.RuleName)
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
		t.Fatalf("standalone no-cycle result: files=%d reports=%d", result.LintedFileCount, reports.Load())
	}
}

func standaloneExecutionTestOptions(
	t *testing.T,
	singleThreaded bool,
	run func(rule.RuleContext) rule.RuleListeners,
) RunLinterOptions {
	t.Helper()
	fileCount := minStandaloneFilesPerLintWorker * 2
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
		Standalone: []StandaloneLintSourceSet{{
			Files:   files,
			Runtime: rule.SourceRuntimeForProgram(program),
		}},
		SingleThreaded:   singleThreaded,
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "standalone-concurrency",
				Severity: rule.SeverityError,
				Run:      run,
			}}
		},
	}
	opts.PreparedPlan = mustPrepareLintPlan(t, opts)
	return opts
}

func TestStandaloneLintWorkerCountKeepsSmallSetsSerial(t *testing.T) {
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
		if got := standaloneLintWorkerCount(test.files, test.procs); got != test.want {
			t.Fatalf("standaloneLintWorkerCount(%d, %d) = %d, want %d", test.files, test.procs, got, test.want)
		}
	}
}

func TestRunLinterParallelizesStandaloneFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	release := make(chan struct{})
	twoActive := make(chan struct{})
	var active atomic.Int32
	var signaled atomic.Bool
	opts := standaloneExecutionTestOptions(t, false, func(rule.RuleContext) rule.RuleListeners {
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
		t.Fatal("standalone files did not overlap across workers")
	}
	result := <-done
	if result.err != nil {
		t.Fatalf("RunLinter: %v", result.err)
	}
	wantFiles := int32(minStandaloneFilesPerLintWorker * 2)
	if result.result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.result.LintedFileCount, wantFiles)
	}
}

func TestRunLinterSingleThreadedSerializesStandaloneFiles(t *testing.T) {
	previousProcs := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previousProcs)

	var active atomic.Int32
	var maxActive atomic.Int32
	opts := standaloneExecutionTestOptions(t, true, func(rule.RuleContext) rule.RuleListeners {
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
	wantFiles := int32(minStandaloneFilesPerLintWorker * 2)
	if result.LintedFileCount != wantFiles {
		t.Fatalf("LintedFileCount = %d, want %d", result.LintedFileCount, wantFiles)
	}
	if maxActive.Load() != 1 {
		t.Fatalf("single-threaded standalone execution had %d active files", maxActive.Load())
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
		Programs:    []*compiler.Program{program},
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
		Programs:         []*compiler.Program{program},
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
		Programs:         []*compiler.Program{program, program},
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
		Programs:         []*compiler.Program{programA},
		SingleThreaded:   true,
		TargetFiles:      [][]string{{pathsA["a.ts"]}},
		SyntaxErrorFiles: map[string]struct{}{},
		GetRulesForFile:  func(*ast.SourceFile) []ConfiguredRule { return noopRule() },
	})
	_, err := RunLinter(RunLinterOptions{
		Programs:         []*compiler.Program{programB},
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
