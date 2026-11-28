package linter

// cspell:ignore synctest

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/web-infra-dev/rslint/internal/typeservice"
)

func typedPluginTestGeneration(t *testing.T, options string) Generation {
	t.Helper()
	raw, paths := createTestProgramWithFilesAndCompilerOptions(t,
		map[string]string{"value.ts": "export const value = { count: 1 };"}, options)
	return Generation{
		Native: NativeGeneration{Programs: wrapTestPrograms(raw), TargetsByProgram: [][]string{{paths["value.ts"]}}, Cwd: raw.GetCurrentDirectory(), RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}
		}},
		Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) { return source.Text(), nil }},
		Plugin: &PluginGeneration{ConfigForFile: func(string) EslintPluginFileConfig { return EslintPluginFileConfig{TypeServices: true} }},
	}
}

func TestEligiblePluginDoesNotSendUnusedTypeServices(t *testing.T) {
	generation := typedPluginTestGeneration(t, `{"noLib":true}`)
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		wire, err := json.Marshal(request)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(wire, []byte(`"sources":`)) || bytes.Contains(wire, []byte(`"origins":`)) {
			return nil, errors.New("plugin dispatch sent Program metadata or origins before a type service request")
		}
		_, frames := request.WithFrameAttachments()
		if len(frames) != len(request.Files) {
			return nil, errors.New("plugin dispatch sent extra compiler syntax")
		}
		result := &EslintPluginLintResult{}
		for _, file := range request.Files {
			if file.TypeServices == nil {
				return nil, errors.New("plugin cannot discover available type services")
			}
			result.Results = append(result.Results, EslintPluginFileResult{FilePath: file.Path})
		}
		return result, nil
	})
	result, err := RunPipeline(t.Context(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, dispatch))
	outcome, _ := result.Observation.JoinedPluginOutcome()
	if err := errors.Join(err, outcome.DispatchError); err != nil {
		t.Fatal(err)
	}
}

func TestProgressiveTypedPluginOwnsGenerationUntilRunOrDiscard(t *testing.T) {
	for _, discard := range []bool{false, true} {
		t.Run(fmt.Sprintf("discard=%v", discard), func(t *testing.T) {
			generation := typedPluginTestGeneration(t, `{"noLib":true}`)
			var released atomic.Int32
			presentation := &pipelineProgressiveDiagnostics{}
			_, err := RunPipeline(t.Context(), NewProgressiveLintRequest(pipelineTestProvider(generation, func() { released.Add(1) }), ArtifactDemand{}, presentation))
			if err != nil {
				t.Fatal(err)
			}
			if presentation.run == nil || released.Load() != 0 {
				t.Fatalf("typed generation released before enrichment: %d", released.Load())
			}
			if discard {
				presentation.run.Close()
				presentation.run.Close()
			} else {
				var request EslintPluginTypeRequest
				outcome, err := presentation.run.Run(t.Context(), EslintPluginDispatchFunc(func(ctx context.Context, got EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					if len(got.Files) == 0 {
						return &EslintPluginLintResult{}, nil
					}
					if released.Load() != 0 {
						t.Fatal("typed producer released during query")
					}
					service := got.Files[0].TypeServices
					if service == nil {
						t.Fatal("type service binding missing")
					}
					opened, openErr := QueryEslintPluginTypes(ctx, EslintPluginTypeRequest{Session: service.Session, Generation: service.Binding.Generation, Open: true, Projection: &service.Binding.SourceFileName})
					if openErr != nil || opened.Descriptor == nil || opened.Projection == nil || len(opened.Projection.Tree) == 0 || len(opened.Projection.Origins) == 0 {
						t.Fatalf("on-demand syntax projection missing: %v", openErr)
					}
					request = EslintPluginTypeRequest{Session: service.Session, Generation: service.Binding.Generation, Queries: []typeservice.Query{{Scope: service.Binding.Scope, Request: &typeservice.CheckerRequest{Op: "void"}}}}
					result, err := QueryEslintPluginTypes(ctx, request)
					if err != nil || len(result.Packets) != 1 {
						t.Fatalf("leased query failed: %v", err)
					}
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					if _, err := QueryEslintPluginTypes(cancelled, request); !errors.Is(err, context.Canceled) {
						t.Fatalf("cancelled query = %v", err)
					}
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: got.Files[0].Path}}}, nil
				}))
				if err != nil || outcome.DispatchError != nil {
					t.Fatalf("typed work = %v / %v", err, outcome.DispatchError)
				}
				if _, err := QueryEslintPluginTypes(t.Context(), request); err == nil {
					t.Fatal("retired session accepted a query")
				}
			}
			if released.Load() != 1 {
				t.Fatalf("release calls = %d", released.Load())
			}
		})
	}
}

func TestProgressiveTypedPluginReleasesDeferredSessionOnSiblingPanic(t *testing.T) {
	previous := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previous)
	generation := typedPluginTestGeneration(t, `{"noLib":true}`)
	var releases atomic.Int32
	panicked := errors.New("second load failed")
	provider := pipelinePlanProvider{plan: GenerationPlan{Loads: []GenerationLoad{
		{CheckerCount: 4, Acquire: func(context.Context) (Generation, ReleaseFunc, error) {
			return generation, func() { releases.Add(1) }, nil
		}},
		{CheckerCount: 4, Acquire: func(context.Context) (Generation, ReleaseFunc, error) {
			panic(panicked)
		}},
	}}}
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Errorf("panic = <nil>, want %v", panicked)
		} else if err, ok := recovered.(error); !ok || !errors.Is(err, panicked) {
			t.Errorf("panic = %v, want %v", recovered, panicked)
		}
		if releases.Load() != 1 {
			t.Errorf("completed unit released %d times, want 1", releases.Load())
		}
	}()
	_, _ = RunPipeline(t.Context(), NewProgressiveLintRequest(provider, ArtifactDemand{}, &pipelineProgressiveDiagnostics{}))
}

func TestTypedPluginWaitsForNativeLint(t *testing.T) {
	for _, tc := range []struct {
		name      string
		options   string
		typeCheck bool
	}{
		{name: "declaration-diagnostics", options: `{"declaration":true,"checkers":1}`, typeCheck: true},
		{name: "plain-typed", options: `{"noLib":true,"checkers":1}`, typeCheck: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generation := typedPluginTestGeneration(t, tc.options)
			generation.Native.TypeCheck = tc.typeCheck
			synctest.Test(t, func(t *testing.T) {
				nativeStarted := make(chan struct{})
				resume := make(chan struct{})
				unblock := sync.OnceFunc(func() { close(resume) })
				defer unblock()
				pluginRules := generation.Native.RulesForFile
				generation.Native.RulesForFile = func(file *ast.SourceFile) []rule.ConfiguredRule {
					return append(pluginRules(file), rule.ConfiguredRule{Name: "native/block", Run: func(rule.RuleContext) rule.RuleListeners {
						close(nativeStarted)
						<-resume
						return nil
					}})
				}
				var calls atomic.Int32
				dispatch := EslintPluginDispatchFunc(func(ctx context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					calls.Add(1)
					if len(request.Files) == 0 {
						return &EslintPluginLintResult{}, nil
					}
					file := request.Files[0]
					service := file.TypeServices
					if service == nil {
						return nil, errors.New("typed plugin lost its service")
					}
					_, err := QueryEslintPluginTypes(ctx, EslintPluginTypeRequest{Session: service.Session, Generation: service.Binding.Generation, Queries: []typeservice.Query{{Scope: service.Binding.Scope, Request: &typeservice.CheckerRequest{Op: "void"}}}})
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: file.Path}}}, err
				})
				done := make(chan error, 1)
				go func() {
					result, err := RunPipeline(t.Context(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, dispatch))
					outcome, _ := result.Observation.JoinedPluginOutcome()
					done <- errors.Join(err, outcome.DispatchError)
				}()
				select {
				case <-nativeStarted:
				case err := <-done:
					t.Fatalf("native lint did not start: %v", err)
				}
				synctest.Wait()
				if calls.Load() != 0 {
					t.Error("typed plugin dispatched before the binding's native lint pass")
				}
				unblock()
				if err := <-done; err != nil || calls.Load() != 2 {
					t.Fatalf("typed observation returned %v with %d requests", err, calls.Load())
				}
			})
		})
	}
}

func TestTypedRetirementBypassesFileSlotsAndPreservesBatchOrder(t *testing.T) {
	resources := &pluginDispatchResources{slots: make(chan struct{}, 1)}
	firstGroupError := errors.New("first configuration failed in second unit")
	laterGroupError := errors.New("later configuration failed in first unit")
	closed := false
	lease := &pluginTypeLease{
		ids:     []string{"session"},
		release: func() { closed = true },
	}
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if len(request.Files) > 0 && len(resources.slots) != 1 {
			t.Error("file request did not use the observation's dispatch slot")
		}
		if len(request.RetiredTypeServices) > 0 {
			if len(resources.slots) != 0 {
				t.Error("retirement request contended for a file dispatch slot")
			}
			if !closed {
				t.Error("semantic lease retained while waiting for host retirement")
			}
		}
		if len(request.Files) > 0 {
			switch request.Files[0].Path {
			case "/repo/b.ts":
				return nil, laterGroupError
			case "/repo/a2.ts":
				return nil, firstGroupError
			}
		}
		return &EslintPluginLintResult{}, nil
	})
	first := pluginTask{resources: resources, types: []*pluginTypeLease{lease}, inputs: []EslintPluginFileInput{
		{Path: "/repo/a.ts", ConfigKey: "a", Frame: []byte{1}, TypeServices: &EslintPluginFileTypeServices{Session: "session"}},
		{Path: "/repo/b.ts", ConfigKey: "b", Frame: []byte{1}},
	}}
	second := pluginTask{resources: resources, inputs: []EslintPluginFileInput{{Path: "/repo/a2.ts", ConfigKey: "a", Frame: []byte{1}}}}
	outcome := mergePluginOutcomes([]EslintPluginDispatchOutcome{first.runRaw(t.Context(), dispatch), second.runRaw(t.Context(), dispatch)})
	if !errors.Is(outcome.DispatchError, firstGroupError) || len(resources.slots) != 0 {
		t.Fatalf("dispatch error = %v, remaining slots = %d", outcome.DispatchError, len(resources.slots))
	}
	t.Run("retirement failure preserves the completed diagnostic", func(t *testing.T) {
		cause := errors.New("host retirement failed")
		const name = "plugin/check"
		text := "const x = 1;"
		task := pluginTask{resources: resources, types: []*pluginTypeLease{{ids: lease.ids}}, inputs: []EslintPluginFileInput{
			{Path: "/repo/a.ts", Text: &text, Frame: []byte{1}, Rules: []rule.ConfiguredRule{{Name: name}}, TypeServices: &EslintPluginFileTypeServices{Session: "session"}},
		}}
		result := task.runRaw(t.Context(), EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			if len(request.RetiredTypeServices) != 0 {
				return nil, cause
			}
			if len(request.Files) == 0 {
				return &EslintPluginLintResult{}, nil
			}
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: request.Files[0].Path, Diagnostics: []EslintPluginDiagnostic{{RuleName: name, Message: "completed"}}}}}, nil
		}))
		outcome := applyPluginFailure(mergePluginOutcomes([]EslintPluginDispatchOutcome{result}), PluginDiscardOnFailure)
		if outcome.DispatchError != nil || len(outcome.Diagnostics) != 1 || outcome.Diagnostics[0].RuleName != name {
			t.Fatalf("retirement changed the completed lint outcome: %+v", outcome)
		}
		if len(outcome.Notices) != 1 || outcome.Notices[0].Kind != EslintPluginTypeServiceRetirementFailed || !errors.Is(outcome.Notices[0].Err, cause) {
			t.Fatalf("retirement notice = %+v", outcome.Notices)
		}
		fileFailure := pluginTask{resources: resources, inputs: []EslintPluginFileInput{{Path: "/repo/b.ts", ConfigKey: "b", Frame: []byte{1}}}}
		fileResult := fileFailure.runRaw(t.Context(), dispatch)
		merged := mergePluginOutcomes([]EslintPluginDispatchOutcome{result, fileResult})
		if !errors.Is(merged.DispatchError, laterGroupError) || len(merged.Notices) != 1 {
			t.Fatalf("cleanup notice changed a later dispatch error: %+v", merged)
		}
	})
	t.Run("failed retirement releases partial output", func(t *testing.T) {
		cause := errors.New("retirement open failed")
		released := 0
		provider := testPluginBatchProvider(func(context.Context) (*EslintPluginBatch, error) {
			return &EslintPluginBatch{Release: func() { released++ }}, cause
		})
		err := dispatchTypeRetirement(t.Context(), provider, lease.ids)
		if !errors.Is(err, cause) || released != 1 || len(resources.slots) != 0 {
			t.Fatalf("dispatch error = %v, released = %d, remaining slots = %d", err, released, len(resources.slots))
		}
	})
	t.Run("panic during file dispatch still retires the session", func(t *testing.T) {
		cause := errors.New("file output release failed")
		opened, retired := 0, 0
		provider := testPluginBatchProvider(func(context.Context) (*EslintPluginBatch, error) {
			opened++
			fileBatch := opened == 1
			return &EslintPluginBatch{
				Dispatch: func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					retired += len(request.RetiredTypeServices)
					return &EslintPluginLintResult{}, nil
				},
				Release: func() {
					if fileBatch {
						panic(cause)
					}
				},
			}, nil
		})
		task := pluginTask{resources: resources, types: []*pluginTypeLease{{ids: lease.ids}}, inputs: []EslintPluginFileInput{{Path: "/repo/a.ts", Frame: []byte{1}, TypeServices: &EslintPluginFileTypeServices{Session: "session"}}}}
		result := task.runRaw(t.Context(), provider)
		if !errors.Is(result.DispatchError, cause) || retired != 1 || len(resources.slots) != 0 {
			t.Fatalf("dispatch error = %v, retired = %d, remaining slots = %d", result.DispatchError, retired, len(resources.slots))
		}
	})
	t.Run("undispatched session skips retirement RPC", func(t *testing.T) {
		canceled, cancel := context.WithCancel(t.Context())
		cancel()
		var called atomic.Int32
		task := pluginTask{resources: resources, types: []*pluginTypeLease{{ids: lease.ids}}, inputs: []EslintPluginFileInput{{Path: "/repo/a.ts", Frame: []byte{1}, TypeServices: &EslintPluginFileTypeServices{Session: "session"}}}}
		result := task.runRaw(canceled, EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			called.Add(1)
			return &EslintPluginLintResult{}, nil
		}))
		if !errors.Is(result.DispatchError, context.Canceled) || called.Load() != 0 {
			t.Fatalf("dispatch error = %v, calls = %d", result.DispatchError, called.Load())
		}
	})
}

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
