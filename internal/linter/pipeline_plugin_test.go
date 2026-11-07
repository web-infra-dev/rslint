package linter

// cspell:ignore synctest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func pluginTextTestPlan(count int) *LintPlan {
	files := make([]lintFilePlan, count)
	rules := &preparedRuleSet{plugins: []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}}
	for i := range files {
		files[i] = lintFilePlan{file: pluginSyntax(fmt.Sprintf("/src/%d.ts", i), "x;"), rules: rules}
	}
	return &LintPlan{programs: []programLintPlan{{files: files}}}
}

func TestPluginSourceReadsShareBudgetAndPreserveOrder(t *testing.T) {
	for _, singleThreaded := range []bool{false, true} {
		t.Run(fmt.Sprintf("singleThreaded=%v", singleThreaded), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				limit := runtime.GOMAXPROCS(0)
				if singleThreaded {
					limit = 1
				}
				gate := make(chan struct{})
				release := sync.OnceFunc(func() { close(gate) })
				defer release()
				var active, calls atomic.Int32
				texts := make(map[string]string)
				target := TargetProjection{
					Path: func(path string) string { return "target:" + path },
					ReadText: func(path string, _ ast.SourceFileLike) (string, error) {
						text, ok := texts[path]
						if !ok {
							return "", fmt.Errorf("unexpected plugin target %q", path)
						}
						calls.Add(1)
						if n := active.Add(1); n > int32(limit) {
							t.Errorf("%d plugin source reads overlapped, limit is %d", n, limit)
						}
						defer active.Add(-1)
						<-gate
						return text, nil
					},
				}
				plugin := &PluginGeneration{WirePath: func(path string) string { return "wire:" + path }}
				var paths []string
				var loads []GenerationLoad
				for _, count := range []int{max(3, limit+1), 1} {
					files := make(map[string]string, count)
					targets := make([]string, count)
					for i := range targets {
						index := len(paths) + i
						targets[i] = fmt.Sprintf("%d.ts", index)
						files[targets[i]] = fmt.Sprintf("const x = %d;", index)
					}
					cp, pathsByName := createTestProgramWithFilesAndCompilerOptions(t, files, `{"noLib":true}`)
					for i, name := range targets {
						targets[i] = pathsByName[name]
						texts[target.Path(targets[i])] = files[name]
					}
					paths = append(paths, targets...)
					native := NativeGeneration{
						Programs:         []*program.Program{program.NewFromCompiler(cp)},
						TargetsByProgram: [][]string{targets},
						Cwd:              cp.GetCurrentDirectory(),
						RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
							return []rule.ConfiguredRule{{Name: "plugin/check", Severity: rule.SeverityError, IsEslintPluginRule: true}}
						},
					}
					loads = append(loads, GenerationLoad{TargetCount: count, CheckerCount: 1, Acquire: func(context.Context) (Generation, ReleaseFunc, error) {
						return Generation{Native: native, Target: target, Plugin: plugin}, nil, nil
					}})
				}
				dispatch := EslintPluginDispatchFunc(func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					results := make([]EslintPluginFileResult, len(req.Files))
					for i, file := range req.Files {
						results[i] = EslintPluginFileResult{FilePath: file.Path, Diagnostics: []EslintPluginDiagnostic{{
							RuleName: "plugin/check", Message: file.Path,
						}}}
					}
					return &EslintPluginLintResult{Results: results}, nil
				})
				var result PipelineResult
				var err error
				done := make(chan struct{})
				go func() {
					defer close(done)
					result, err = RunPipeline(t.Context(), NewLintRequest(pipelinePlanProvider{GenerationPlan{Loads: loads, SingleThreaded: singleThreaded}}, ObservationPolicy{}, dispatch))
				}()
				synctest.Wait()
				if got := active.Load(); got != int32(limit) {
					t.Errorf("active plugin source reads = %d, want %d", got, limit)
				}
				release()
				<-done
				if err != nil {
					t.Fatal(err)
				}
				if got := int(calls.Load()); got != len(paths) {
					t.Fatalf("read %d plugin inputs, want %d", got, len(paths))
				}
				diagnostics, complete := result.Observation.CompleteDiagnostics()
				if !complete || len(diagnostics) != len(paths) {
					t.Fatalf("received %d plugin diagnostics, complete=%v", len(diagnostics), complete)
				}
				for i, diagnostic := range diagnostics {
					if diagnostic.FilePath != "target:"+paths[i] || diagnostic.Message.Description != "wire:"+paths[i] {
						t.Errorf("input %d produced %q at %q, want wire and target for %q", i, diagnostic.Message.Description, diagnostic.FilePath, paths[i])
					}
				}
			})
		})
	}
}

func TestPluginSourceReadsPreserveFirstInputError(t *testing.T) {
	for _, tc := range []struct {
		files          int
		singleThreaded bool
		inline         bool
	}{
		{files: 4},
		{files: 4, singleThreaded: true, inline: true},
		{files: 2, inline: true},
	} {
		var reads atomic.Int32
		generation := Generation{
			Native: NativeGeneration{SingleThreaded: tc.singleThreaded},
			Target: TargetProjection{ReadText: func(path string, source ast.SourceFileLike) (string, error) {
				reads.Add(1)
				if path == "/src/1.ts" {
					return "", errors.New("later read failed")
				}
				return source.Text(), nil
			}},
			Plugin: &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig {
				return EslintPluginFileConfig{Settings: map[string]any{"invalid": make(chan int)}}
			}},
		}
		_, err := materializePluginTask(t.Context(), pluginTextTestPlan(tc.files), generation, SourceSnapshot{}, ObservationPolicy{}, true, false, make(pluginSourceReadSlots, 8))
		if err == nil || !strings.Contains(err.Error(), `freeze plugin input "/src/0.ts": clone settings`) {
			t.Fatalf("%d files, single-threaded %v: first input error = %v", tc.files, tc.singleThreaded, err)
		}
		if tc.inline && reads.Load() != 1 {
			t.Fatal("inline preparation read past its first error")
		}
	}
}

func TestPluginSourceReadsStopAfterFirstFailure(t *testing.T) {
	for _, limit := range []int{2, 8} {
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				fail := make(chan struct{})
				finish := make(chan struct{})
				readErr := errors.New("source unavailable")
				var calls atomic.Int32
				generation := Generation{
					Target: TargetProjection{ReadText: func(path string, source ast.SourceFileLike) (string, error) {
						calls.Add(1)
						if path == "/src/0.ts" {
							<-fail
							return "", readErr
						}
						<-finish
						return source.Text(), nil
					}},
					Plugin: &PluginGeneration{},
				}
				done := make(chan struct{})
				var err error
				go func() {
					defer close(done)
					_, err = materializePluginTask(t.Context(), pluginTextTestPlan(80), generation, SourceSnapshot{}, ObservationPolicy{}, false, false, make(pluginSourceReadSlots, limit))
				}()
				synctest.Wait()
				if got := calls.Load(); got != int32(limit) {
					t.Errorf("started %d reads before failure, want %d", got, limit)
				}
				close(fail)
				synctest.Wait()
				select {
				case <-done:
					t.Error("returned before the other admitted readers finished")
				default:
				}
				close(finish)
				<-done
				if !errors.Is(err, readErr) {
					t.Errorf("materialization error = %v, want %v", err, readErr)
				}
				if got := calls.Load(); got != int32(limit) {
					t.Errorf("read %d inputs after admitting %d before failure", got, limit)
				}
			})
		})
	}
}

func TestPipelineConcurrentPluginStartsBeforeNativeAndPlansProjectedFix(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	projectedPath := "target:" + fileName
	pluginStarted := make(chan struct{})
	allowPluginFinish := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	nativeRan := false
	configuredRules := []rule.ConfiguredRule{
		{
			Name:     "native/order",
			Severity: rule.SeverityWarning,
			Run: func(rule.RuleContext) rule.RuleListeners {
				select {
				case <-pluginStarted:
				case <-ctx.Done():
					t.Fatal("native work started before plugin dispatch")
				}
				nativeRan = true
				close(allowPluginFinish)
				return nil
			},
		},
		{Name: "plugin/replace", Severity: rule.SeverityError, IsEslintPluginRule: true},
	}
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"a",
		configuredRules,
		&EslintPluginFileConfig{ConfigKey: "config"},
	)
	generation.Target.Path = func(string) string { return projectedPath }
	result, err := RunPipeline(ctx, NewAutofixRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{
			Demand: ArtifactDemand{Native: rule.EditDemandAutofix, Plugin: rule.EditDemandAutofix},
			Plugin: PluginConcurrentJoined,
		},
		autofixPolicyForTest(1, AutofixPolicy{}),
		EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			if request.Files[0].Path != fileName {
				t.Fatalf("plugin wire path = %q, want Program source path %q", request.Files[0].Path, fileName)
			}
			close(pluginStarted)
			<-allowPluginFinish
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
				FilePath: request.Files[0].Path,
				Diagnostics: []EslintPluginDiagnostic{{
					RuleName: "plugin/replace",
					Message:  "replace",
					StartPos: 0,
					EndPos:   1,
					Fixes:    []EslintPluginFix{{Range: [2]int{0, 1}, Text: "b"}},
				}},
			}}}, nil
		}),
	))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, complete := result.Observation.CompleteDiagnostics()
	applied, planned := result.AppliedFixes()
	changes := applied.FinalChanges
	if !nativeRan || !complete || len(diagnostics) != 1 || diagnostics[0].FilePath != projectedPath {
		t.Fatalf("observation = native:%v complete:%v diagnostics:%+v", nativeRan, complete, diagnostics)
	}
	if !planned || len(changes) != 1 || changes[0].Path != projectedPath || changes[0].Before != "a" || changes[0].After != "b" {
		t.Fatalf("planned changes = %+v, planned=%v", changes, planned)
	}
}

func TestPipelineJoinedModesPropagateCallerCancellation(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined} {
		t.Run(fmt.Sprintf("mode-%d", mode), func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			fileName := tspath.ResolvePath(root, "source.ts")
			nativeFinished := make(chan struct{})
			generation := pipelineTestGeneration(
				t,
				root,
				fileName,
				"a",
				[]rule.ConfiguredRule{
					{
						Name: "native/marker",
						Run: func(rule.RuleContext) rule.RuleListeners {
							close(nativeFinished)
							return nil
						},
					},
					{Name: "plugin/check", IsEslintPluginRule: true},
				},
				&EslintPluginFileConfig{},
			)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result, err := RunPipeline(ctx, NewLintRequest(
				pipelineTestProvider(generation, nil),
				ObservationPolicy{
					Demand:        ArtifactDemand{Plugin: rule.EditDemandAll},
					Plugin:        mode,
					PluginFailure: PluginDiscardOnFailure,
				},
				EslintPluginDispatchFunc(func(dispatchCtx context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					<-nativeFinished
					cancel()
					return nil, dispatchCtx.Err()
				}),
			))
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("pipeline error = %v, want context.Canceled", err)
			}
			records := result.PluginOutcomes()
			if len(records) != 1 || records[0].Observation != 0 || !errors.Is(records[0].DispatchError, context.Canceled) {
				t.Fatalf("plugin records = %+v, want canceled observation 0", records)
			}
			if result.Observation.Native.Lint != nil {
				t.Fatal("canceled observation was published as authoritative")
			}
		})
	}
}

func TestPipelineDoesNotPromoteIndependentPluginBudgetFailure(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"a",
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}},
		&EslintPluginFileConfig{},
	)
	result, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{
			Plugin:        PluginAfterNativeJoined,
			PluginFailure: PluginDiscardOnFailure,
		},
		EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return nil, context.DeadlineExceeded
		}),
	))
	if err != nil {
		t.Fatalf("pipeline promoted an independent plugin budget failure: %v", err)
	}
	outcome, ok := result.Observation.JoinedPluginOutcome()
	if !ok || !errors.Is(outcome.DispatchError, context.DeadlineExceeded) {
		t.Fatalf("joined plugin outcome = %+v, want deadline failure", outcome)
	}
}

func TestPipelineRejectsMissingJoinedPluginDispatcher(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(
		t,
		root,
		fileName,
		"a",
		[]rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}},
		&EslintPluginFileConfig{},
	)
	_, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{Plugin: PluginConcurrentJoined},
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "requires a dispatcher") {
		t.Fatalf("missing dispatcher error = %v", err)
	}
}

func TestPipelineNativeOnlySkipsPluginSourceMaterialization(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined} {
		for _, registered := range []bool{false, true} {
			t.Run(fmt.Sprintf("mode-%d/plugin-registered-%t", mode, registered), func(t *testing.T) {
				root := tspath.NormalizePath(t.TempDir())
				fileName := tspath.ResolvePath(root, "source.ts")
				var originalStatement *ast.Node
				nativeVisits := 0
				generation := pipelineTestGeneration(t, root, fileName, "debugger;", []rule.ConfiguredRule{{
					Name:     "native/check",
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{ast.KindDebuggerStatement: func(node *ast.Node) {
							nativeVisits++
							if node != originalStatement {
								t.Fatal("native traversal replaced the Program AST node")
							}
							ctx.ReportNode(node, rule.RuleMessage{Description: "native debugger"})
						}}
					},
				}}, nil)
				originalStatement = generation.Native.Programs[0].GetSourceFile(fileName).Statements.Nodes[0]
				if registered {
					generation.Plugin = &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig {
						t.Fatal("native-only target resolved plugin config")
						return EslintPluginFileConfig{}
					}}
				}
				generation.Target.ReadText = func(string, ast.SourceFileLike) (string, error) {
					return "", errors.New("native-only lint must not materialize plugin source")
				}
				result, err := RunPipeline(context.Background(), NewLintRequest(
					pipelineTestProvider(generation, nil),
					ObservationPolicy{Plugin: mode},
					nil,
				))
				if err != nil {
					t.Fatal(err)
				}
				diagnostics, complete := result.Observation.CompleteDiagnostics()
				if nativeVisits != 1 || !complete || len(diagnostics) != 1 || diagnostics[0].RuleName != "native/check" {
					t.Fatalf("native visits/complete/diagnostics = %d/%t/%+v", nativeVisits, complete, diagnostics)
				}
				if len(result.PluginOutcomes()) != 0 {
					t.Fatal("native-only lint dispatched plugin work")
				}
			})
		}
	}
}

func TestPipelineFreezesCompletePluginInputAfterMemoryChanges(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	paths := []string{
		tspath.ResolvePath(root, "changed.ts"),
		tspath.ResolvePath(root, "unchanged.ts"),
	}
	initial := map[string]string{paths[0]: "a", paths[1]: "stable"}
	acquisitions := 0
	provider := GenerationProviderFunc(func(_ context.Context, snapshot SourceSnapshot) (Generation, ReleaseFunc, error) {
		acquisitions++
		texts := map[string]string{paths[0]: initial[paths[0]], paths[1]: initial[paths[1]]}
		for _, file := range snapshot.Files() {
			texts[file.Path] = file.Text
		}
		programs := []*program.Program{
			pipelineTestProgram(t, root, paths[0], texts[paths[0]]),
			pipelineTestProgram(t, root, paths[1], texts[paths[1]]),
		}
		return Generation{
			Native: NativeGeneration{
				Programs:         programs,
				TargetsByProgram: [][]string{{paths[0]}, {paths[1]}},
				Cwd:              root,
				RulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
					rules := []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}
					if source.FileName() == paths[0] && source.Text() == "a" {
						rules = append(rules, rule.ConfiguredRule{
							Name: "native/fix",
							Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
								textRange := core.NewTextRange(0, 1)
								ruleCtx.ReportRangeWithFixes(
									textRange,
									rule.RuleMessage{Description: "fix"},
									rule.RuleFix{Range: textRange, Text: "b"},
								)
								return nil
							},
						})
					}
					return rules
				},
			},
			Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
				return source.Text(), nil
			}},
			Plugin: &PluginGeneration{},
		}, nil, nil
	})

	dispatches := 0
	var dispatchValidationErr error
	expectedSecond := map[string]string{paths[0]: "b", paths[1]: "stable"}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{
			Demand: ArtifactDemand{Native: rule.EditDemandAutofix},
			Plugin: PluginConcurrentJoined,
		},
		autofixPolicyForTest(1, AutofixPolicy{VerifyAfterLastRound: true}),
		EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			dispatches++
			if len(request.Files) != 2 {
				dispatchValidationErr = fmt.Errorf("plugin files = %+v", request.Files)
			}
			expected := initial
			if dispatches == 2 {
				expected = expectedSecond
			}
			for _, file := range request.Files {
				text, exists := expected[file.Path]
				if !exists || !bytes.HasSuffix(file.Frame, []byte(text)) {
					dispatchValidationErr = errors.New("generation frame mismatch")
				}
			}
			results := make([]EslintPluginFileResult, len(request.Files))
			for index, file := range request.Files {
				results[index].FilePath = file.Path
			}
			return &EslintPluginLintResult{Results: results}, nil
		}),
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || acquisitions != 2 || dispatches != 2 || len(applied.FinalChanges) != 1 || dispatchValidationErr != nil {
		t.Fatalf("result/acquisitions/dispatches/validation = %+v/%d/%d/%v", applied, acquisitions, dispatches, dispatchValidationErr)
	}
}

func TestPipelineRejectsDuplicatePluginWirePath(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	firstPath := tspath.ResolvePath(root, "first.ts")
	secondPath := tspath.ResolvePath(root, "second.ts")
	generation := Generation{
		Native: NativeGeneration{
			Programs: []*program.Program{
				pipelineTestProgram(t, root, firstPath, "a"),
				pipelineTestProgram(t, root, secondPath, "b"),
			},
			TargetsByProgram: [][]string{{firstPath}, {secondPath}},
			SingleThreaded:   true,
			Cwd:              root,
			RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}
			},
		},
		Plugin: &PluginGeneration{
			WirePath: func(string) string { return "shared.ts" },
		},
	}
	_, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{Plugin: PluginConcurrentJoined},
		EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return &EslintPluginLintResult{}, nil
		}),
	))
	if err == nil || !strings.Contains(err.Error(), "duplicate plugin wire path") {
		t.Fatalf("duplicate wire path error = %v", err)
	}
}

func TestPipelineRejectsDuplicateProjectedTargetBeforeExecution(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	firstPath := tspath.ResolvePath(root, "first.ts")
	secondPath := tspath.ResolvePath(root, "second.ts")
	ruleRan := false
	generation := Generation{
		Native: NativeGeneration{
			Programs: []*program.Program{
				pipelineTestProgram(t, root, firstPath, "a"),
				pipelineTestProgram(t, root, secondPath, "b"),
			},
			TargetsByProgram: [][]string{{firstPath}, {secondPath}},
			SingleThreaded:   true,
			Cwd:              root,
			RulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: "native/fix",
					Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
						ruleRan = true
						textRange := core.NewTextRange(0, len(source.Text()))
						ruleCtx.ReportRangeWithFixes(
							textRange,
							rule.RuleMessage{Description: "fix"},
							rule.RuleFix{Range: textRange, Text: "fixed"},
						)
						return nil
					},
				}}
			},
		},
		Target: TargetProjection{
			Path: func(string) string { return "shared.ts" },
			ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
				return source.Text(), nil
			},
		},
	}
	_, err := RunPipeline(context.Background(), NewAutofixRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(1, AutofixPolicy{}),
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "duplicate projected target") || ruleRan {
		t.Fatalf("duplicate projected target error/rule ran = %v/%v", err, ruleRan)
	}
}

func TestPipelinePluginFailureAppliesToAllLoadedUnits(t *testing.T) {
	for _, failure := range []PluginFailurePolicy{PluginKeepPartialWithSynthetic, PluginDiscardOnFailure} {
		t.Run(fmt.Sprint(failure), func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			loads := make([]GenerationLoad, 3)
			for i := range loads {
				path := tspath.ResolvePath(root, fmt.Sprintf("%d.ts", i))
				configKey := "first"
				if i == 1 {
					configKey = "second"
				}
				loads[i] = GenerationLoad{TargetCount: 1, Acquire: func(context.Context) (Generation, ReleaseFunc, error) {
					return pipelineTestGeneration(t, root, path, "a", []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true, Severity: rule.SeverityError}}, &EslintPluginFileConfig{ConfigKey: configKey}), nil, nil
				}}
			}
			wantErr := errors.New("transport failed")
			result, err := RunPipeline(context.Background(), NewLintRequest(pipelinePlanProvider{GenerationPlan{Loads: loads}}, ObservationPolicy{PluginFailure: failure},
				EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					path := request.Files[0].Path
					if strings.HasSuffix(path, "/1.ts") {
						return nil, errors.New("second config group failed")
					}
					if !strings.HasSuffix(path, "/0.ts") {
						return nil, wantErr
					}
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: path, Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "success"}}}}}, nil
				})))
			if err != nil {
				t.Fatal(err)
			}
			outcome, ok := result.Observation.JoinedPluginOutcome()
			if !ok || !errors.Is(outcome.DispatchError, wantErr) {
				t.Fatalf("lost dispatch failure: %+v", outcome)
			}
			want := 2
			if failure == PluginDiscardOnFailure {
				want = 0
			}
			if len(outcome.Diagnostics) != want {
				t.Fatalf("diagnostics = %+v, want %d", outcome.Diagnostics, want)
			}
			if want != 0 && outcome.Diagnostics[1].FilePath != tspath.ResolvePath(root, "0.ts") {
				t.Fatal("synthetic failure lost whole-observation target")
			}
		})
	}
}

func TestPipelineUnusedDirectives(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined} {
		for _, tc := range []struct {
			name, source         string
			native, plugin, fail bool
			unused               int
		}{
			{name: "zero rules", source: "/* eslint-disable */\nconst x = 1;", unused: 1},
			{name: "mixed used", source: "/* eslint-disable native/check, plugin/check */\ndebugger;", native: true, plugin: true},
			{name: "plugin only", source: "/* eslint-disable plugin/check */\ndebugger;", plugin: true},
			{name: "partial", source: "/* eslint-disable native/check, plugin/check, unused */\ndebugger;", native: true, plugin: true, unused: 1},
			{name: "plugin failure", source: "/* eslint-disable unused */\ndebugger;", plugin: true, fail: true},
		} {
			t.Run(fmt.Sprintf("%d/%s", mode, tc.name), func(t *testing.T) {
				root := tspath.NormalizePath(t.TempDir())
				path := tspath.ResolvePath(root, "source.ts")
				var rules []rule.ConfiguredRule
				if tc.native {
					rules = append(rules, rule.ConfiguredRule{Name: "native/check", Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{ast.KindDebuggerStatement: func(node *ast.Node) { ctx.ReportNode(node, rule.RuleMessage{Description: "native"}) }}
					}})
				}
				var config *EslintPluginFileConfig
				if tc.plugin {
					rules = append(rules, rule.ConfiguredRule{Name: "plugin/check", Severity: rule.SeverityError, IsEslintPluginRule: true})
					config = &EslintPluginFileConfig{}
				}
				generation := pipelineTestGeneration(t, root, path, tc.source, rules, config)
				generation.Native.UnusedDirectiveSeverity = func(string) rule.DiagnosticSeverity { return rule.SeverityWarning }
				dispatch := EslintPluginDispatchFunc(func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					if tc.fail {
						return nil, errors.New("plugin failed")
					}
					if !req.Files[0].CollectDirectiveUsage {
						t.Fatal("missing per-file usage request")
					}
					name := "plugin/check"
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: req.Files[0].Path, UsedDisableDirectives: []rule.DisableDirectiveUsage{{Line: 0, Column: 0, RuleID: &name}}}}}, nil
				})
				result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{Plugin: mode}, dispatch))
				if err != nil {
					t.Fatal(err)
				}
				diagnostics, _ := result.Observation.CompleteDiagnostics()
				unused := 0
				for _, d := range diagnostics {
					if strings.HasPrefix(d.Message.Description, "Unused eslint-") {
						unused++
						if d.Severity != rule.SeverityWarning {
							t.Fatal("wrong severity")
						}
					} else if !tc.fail {
						t.Fatalf("unsuppressed diagnostic: %+v", d)
					}
				}
				if unused != tc.unused {
					t.Fatalf("got %d unused, want %d; diagnostics: %+v", unused, tc.unused, diagnostics)
				}
			})
		}
	}
}

func TestUnusedDirectiveCompletion(t *testing.T) {
	for _, nativeFirst := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("nativeFirst=%v/failure=%v", nativeFirst, fail), func(t *testing.T) {
				source := pluginSyntax("/test.ts", "/* eslint-disable unused */\ndebugger;")
				manager := rule.NewDisableManager(source, rule.NewCommentStore(source))
				manager.Freeze(newTextSourceFile(source.Text()))
				state := &fileDirectives{manager: manager, path: "/test.ts", severity: rule.SeverityWarning, demand: rule.EditDemandAutofix}
				if got := state.complete(nativeFirst, false, "/test.ts"); len(got) != 0 {
					t.Fatal("premature unused report")
				}
				got := state.complete(!nativeFirst, fail, "/test.ts")
				want := 1
				if fail {
					want = 0
				}
				if len(got) != want {
					t.Fatalf("got %d reports, want %d", len(got), want)
				}
				if got := state.complete(nativeFirst, false, "/test.ts"); len(got) != 0 {
					t.Fatal("duplicate finalization")
				}
			})
		}
	}
}

func TestUnusedDirectiveMultiplePluginBatches(t *testing.T) {
	source := pluginSyntax("/test.ts", "/* eslint-disable unused */\nx();")
	manager := rule.NewDisableManager(source, rule.NewCommentStore(source))
	manager.Freeze(newTextSourceFile(source.Text()))
	state := &fileDirectives{manager: manager, path: "/test.ts", severity: rule.SeverityWarning, pendingPlugins: 2}
	if got := state.complete(true, false, "/test.ts"); len(got) != 0 {
		t.Fatal("native work finalized pending plugins")
	}
	if got := state.complete(false, false, "/test.ts"); len(got) != 0 {
		t.Fatal("first batch finalized file")
	}
	if got := state.complete(false, false, "/test.ts"); len(got) != 1 {
		t.Fatalf("last batch produced %d reports", len(got))
	}
}

func TestPipelineUnusedDirectiveUsageValidation(t *testing.T) {
	name := "plugin/check"
	for _, tc := range []struct {
		name   string
		usage  []rule.DisableDirectiveUsage
		failed bool
	}{
		{"missing", nil, true},
		{"empty", []rule.DisableDirectiveUsage{}, false},
		{"unknown", []rule.DisableDirectiveUsage{{Line: 9, RuleID: &name}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			path := tspath.ResolvePath(root, "source.ts")
			generation := pipelineTestGeneration(t, root, path, "/* eslint-disable plugin/check */\nx();", []rule.ConfiguredRule{{Name: name, Severity: rule.SeverityError, IsEslintPluginRule: true}}, &EslintPluginFileConfig{})
			generation.Native.UnusedDirectiveSeverity = func(string) rule.DiagnosticSeverity { return rule.SeverityWarning }
			result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}}, EslintPluginDispatchFunc(func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
				return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: req.Files[0].Path, UsedDisableDirectives: tc.usage}}}, nil
			})))
			if err != nil {
				t.Fatal(err)
			}
			diagnostics, _ := result.Observation.CompleteDiagnostics()
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics: %+v", diagnostics)
			}
			if tc.failed {
				if diagnostics[0].RuleName != "rslint/plugin-lint-error" || len(diagnostics[0].Fixes()) != 0 {
					t.Fatalf("unsafe failure result: %+v", diagnostics)
				}
			} else if diagnostics[0].RuleName != "" || len(diagnostics[0].Fixes()) != 1 {
				t.Fatalf("missing unused fix: %+v", diagnostics)
			}
		})
	}
}

func TestPipelineUnusedReportingOffSkipsUsageState(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	path := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(t, root, path, "/* eslint-disable plugin/check */\nx();", []rule.ConfiguredRule{{Name: "plugin/check", Severity: rule.SeverityError, IsEslintPluginRule: true}}, &EslintPluginFileConfig{})
	generation.Native.DirectiveRuleExists = func(string, string) bool { t.Fatal("off reporting parsed directive state"); return false }
	result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, EslintPluginDispatchFunc(func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if req.Files[0].CollectDirectiveUsage {
			t.Fatal("off reporting requested usage")
		}
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: req.Files[0].Path}}}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, _ := result.Observation.CompleteDiagnostics()
	if len(diagnostics) != 0 {
		t.Fatalf("off reporting emitted diagnostics: %+v", diagnostics)
	}
}

func TestPipelineUnusedDirectivesIgnoreSkippedTypeRules(t *testing.T) {
	for _, name := range []string{"only-skipped", "mixed"} {
		mixed := name == "mixed"
		t.Run(name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			path := tspath.ResolvePath(root, "source.ts")
			rules := []rule.ConfiguredRule{{Name: "needs-types", Severity: rule.SeverityError, RequiresTypeInfo: true}}
			names := "needs-types"
			if mixed {
				names += ", plain"
				rules = append(rules, rule.ConfiguredRule{Name: "plain", Severity: rule.SeverityError, Run: func(rule.RuleContext) rule.RuleListeners { return rule.RuleListeners{} }})
			}
			generation := pipelineTestGeneration(t, root, path, "/* eslint-disable "+names+" */\nx();", rules, nil)
			generation.Native.UnusedDirectiveSeverity = func(string) rule.DiagnosticSeverity { return rule.SeverityWarning }
			result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
			if err != nil {
				t.Fatal(err)
			}
			diagnostics, _ := result.Observation.CompleteDiagnostics()
			if !mixed {
				if len(diagnostics) != 0 {
					t.Fatalf("skipped type rule reported as unused: %+v", diagnostics)
				}
			} else if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message.Description, "'plain'") || strings.Contains(diagnostics[0].Message.Description, "'needs-types'") {
				t.Fatalf("wrong unused entries for mixed rules: %+v", diagnostics)
			}
		})
	}
}

func TestPipelineDirectiveUsageSelectionIsPerFile(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	onPath := tspath.ResolvePath(root, "on.ts")
	offPath := tspath.ResolvePath(root, "off.ts")
	source := "/* eslint-disable plugin/check */\nx();"
	rules := []rule.ConfiguredRule{{Name: "plugin/check", Severity: rule.SeverityError, IsEslintPluginRule: true}}
	generation := pipelineTestGeneration(t, root, onPath, source, rules, &EslintPluginFileConfig{})
	generation.Native.Programs = append(generation.Native.Programs, pipelineTestProgram(t, root, offPath, source))
	generation.Native.TargetsByProgram = append(generation.Native.TargetsByProgram, []string{offPath})
	generation.Target.Path = nil
	generation.Native.UnusedDirectiveSeverity = func(path string) rule.DiagnosticSeverity {
		if path == onPath {
			return rule.SeverityWarning
		}
		return rule.SeverityOff
	}
	result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, EslintPluginDispatchFunc(func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		result := &EslintPluginLintResult{}
		for _, file := range req.Files {
			if file.CollectDirectiveUsage != (file.Path == onPath) {
				t.Fatalf("wrong per-file collection: %+v", file)
			}
			row := EslintPluginFileResult{FilePath: file.Path}
			if file.CollectDirectiveUsage {
				row.UsedDisableDirectives = []rule.DisableDirectiveUsage{}
			}
			result.Results = append(result.Results, row)
		}
		return result, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, _ := result.Observation.CompleteDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].FilePath != onPath || diagnostics[0].RuleName != "" {
		t.Fatalf("wrong per-file unused reports: %+v", diagnostics)
	}
}

func TestPipelineUnusedDirectivePathProjectedOnce(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	sourcePath := tspath.ResolvePath(root, "source.ts")
	targetPath := tspath.ResolvePath(root, "target.ts")
	generation := pipelineTestGeneration(t, root, sourcePath, "/* eslint-disable */\nx();", nil, nil)
	generation.Native.UnusedDirectiveSeverity = func(string) rule.DiagnosticSeverity { return rule.SeverityWarning }
	generation.Target.Path = func(path string) string {
		if path == sourcePath {
			return targetPath
		}
		if path == targetPath {
			return tspath.ResolvePath(root, "incorrect.ts")
		}
		return path
	}
	result, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, _ := result.Observation.CompleteDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].FilePath != targetPath {
		t.Fatalf("unexpected projected diagnostics: %+v", diagnostics)
	}
}
