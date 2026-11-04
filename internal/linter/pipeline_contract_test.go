package linter

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestPipelineReleasesGenerationOnPreparationFailureAndPanic(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		var releases atomic.Int32
		_, err := RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(Generation{Native: NativeGeneration{
				Programs: []*program.Program{nil},
				RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return nil
				},
			}}, func() {
				releases.Add(1)
			}),
			ObservationPolicy{},
			nil,
		))
		if err == nil || releases.Load() != 1 {
			t.Fatalf("error/releases = %v/%d, want error/1", err, releases.Load())
		}
	})

	t.Run("panic", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		fileName := tspath.ResolvePath(root, "source.ts")
		generation := pipelineTestGeneration(t, root, fileName, "const value = 1;", nil, nil)
		generation.Native.RulesForFile = func(*ast.SourceFile) []rule.ConfiguredRule {
			panic("resolver failed")
		}
		var releases atomic.Int32
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_, _ = RunPipeline(context.Background(), NewLintRequest(
				pipelineTestProvider(generation, func() { releases.Add(1) }),
				ObservationPolicy{},
				nil,
			))
		}()
		if recovered == nil || releases.Load() != 1 {
			t.Fatalf("panic/releases = %v/%d, want panic/1", recovered, releases.Load())
		}
	})

	t.Run("parallel panics", func(t *testing.T) {
		previousProcs := runtime.GOMAXPROCS(2)
		defer runtime.GOMAXPROCS(previousProcs)

		var arrivals atomic.Int32
		releaseResolvers := make(chan struct{})
		recovered, releases := runPipelineWithParallelRuleResolver(t, func(source *ast.SourceFile) []rule.ConfiguredRule {
			if arrivals.Add(1) == 2 {
				close(releaseResolvers)
			}
			select {
			case <-releaseResolvers:
			case <-time.After(5 * time.Second):
				panic("timed out waiting for both parallel resolvers")
			}
			panic(source.FileName())
		})
		firstPanic, firstOK := recovered.(string)
		if !firstOK ||
			(!strings.HasSuffix(firstPanic, "/first.ts") && !strings.HasSuffix(firstPanic, "/second.ts")) ||
			releases != 1 {
			t.Fatalf("parallel panic/releases = %v/%d, want either resolver panic/1", recovered, releases)
		}
	})

	t.Run("parallel abnormal worker exit", func(t *testing.T) {
		previousProcs := runtime.GOMAXPROCS(2)
		defer runtime.GOMAXPROCS(previousProcs)

		var calls atomic.Int32
		recovered, releases := runPipelineWithParallelRuleResolver(t, func(*ast.SourceFile) []rule.ConfiguredRule {
			if calls.Add(1) == 1 {
				runtime.Goexit()
			}
			return nil
		})
		recoveredErr, ok := recovered.(error)
		if !ok || !errors.Is(recoveredErr, errPlanWorkerAborted) || releases != 1 {
			t.Fatalf("abnormal worker exit/releases = %v/%d, want %v/1", recovered, releases, errPlanWorkerAborted)
		}
	})
}

func TestPipelineRejectsInvalidGenerationPortsBeforeExecution(t *testing.T) {
	t.Run("complete Programs require an ordinary lint request", func(t *testing.T) {
		acquired := false
		provider := GenerationProviderFunc(func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error) {
			acquired = true
			return Generation{}, nil, nil
		})
		request := NewAutofixRequest(provider, ObservationPolicy{
			ProgramExecution: ProgramComplete,
			Demand:           ArtifactDemand{Native: rule.EditDemandAutofix},
		}, AutofixPolicy{}, nil)
		if _, err := RunPipeline(context.Background(), request); err == nil || acquired {
			t.Fatalf("autofix accepted interleaved computation: error=%v acquired=%t", err, acquired)
		}
	})

	t.Run("nil provider function", func(t *testing.T) {
		_, err := RunPipeline(context.Background(), NewLintRequest(
			GenerationProviderFunc(nil),
			ObservationPolicy{},
			nil,
		))
		if err == nil || !strings.Contains(err.Error(), "provider function must not be nil") {
			t.Fatalf("nil provider error = %v", err)
		}
	})

	t.Run("empty target projection", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		fileName := tspath.ResolvePath(root, "source.ts")
		ruleRan := false
		generation := pipelineTestGeneration(t, root, fileName, "const value = 1;", []rule.ConfiguredRule{{
			Name: "native/must-not-run",
			Run: func(rule.RuleContext) rule.RuleListeners {
				ruleRan = true
				return nil
			},
		}}, nil)
		generation.Target.Path = func(string) string { return "" }
		_, err := RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, nil),
			ObservationPolicy{},
			nil,
		))
		if err == nil || !strings.Contains(err.Error(), "projected target path must not be empty") || ruleRan {
			t.Fatalf("projection error/rule ran = %v/%v", err, ruleRan)
		}
	})
}

func TestPipelineAcceptsGenerationWithoutLintPlan(t *testing.T) {
	result, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(Generation{Native: NativeGeneration{SingleThreaded: true}}, nil),
		ObservationPolicy{},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	if result.Observation.Native.Lint == nil ||
		result.Observation.Native.Lint.LintedFileCount != 0 ||
		len(result.Observation.Native.Diagnostics) != 0 ||
		len(result.Observation.Native.Files) != 0 ||
		result.Observation.Native.HasTargetSyntaxErrors {
		t.Fatalf("empty generation observation = %+v", result.Observation.Native)
	}
}

func TestPipelineCollectsLintedFilesOnlyWhenDemanded(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	paths := []string{tspath.ResolvePath(root, "zero-rules.ts"), tspath.ResolvePath(root, "syntax-error.ts"), tspath.ResolvePath(root, "unselected.ts")}
	targets := []string{tspath.ResolvePath(root, "selected/zero-rules.ts"), tspath.ResolvePath(root, "selected/syntax-error.ts")}
	generation := Generation{Native: NativeGeneration{
		Programs: []*program.Program{
			pipelineTestProgram(t, root, paths[0], "const value = 1;"),
			pipelineTestProgram(t, root, paths[1], "const broken = ;"),
			pipelineTestProgram(t, root, paths[2], ""),
		},
		TargetsByProgram: [][]string{{paths[0]}, {paths[1]}, nil},
		RulesForFile:     func(*ast.SourceFile) []rule.ConfiguredRule { return nil },
	}, Target: TargetProjection{Path: func(path string) string {
		for i, source := range paths[:len(targets)] {
			if path == source {
				return targets[i]
			}
		}
		return path
	}}}
	for _, demand := range []ArtifactDemand{{}, {LintedFiles: true}, {SourceFiles: true}} {
		result, err := RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, nil), ObservationPolicy{Demand: demand}, nil,
		))
		if err != nil {
			t.Fatal(err)
		}
		observation := result.Observation.Native
		if observation.Lint.LintedFileCount != 2 || !observation.HasTargetSyntaxErrors || len(observation.Diagnostics) == 0 {
			t.Fatalf("empty, zero-rule or syntax-error Program changed selection: %+v", observation)
		}
		for _, diagnostic := range observation.Diagnostics {
			if diagnostic.FilePath != targets[1] {
				t.Fatalf("syntax diagnostic lost its stable target projection: %+v", diagnostic)
			}
		}
		if !demand.LintedFiles && !demand.SourceFiles {
			if observation.Files != nil {
				t.Fatalf("unrequested linted files = %+v", observation.Files)
			}
			continue
		}
		if len(observation.Files) != 2 {
			t.Fatalf("selected files = %+v", observation.Files)
		}
		for i, file := range observation.Files {
			if file.Path != targets[i] || (file.SourceFile != nil) != demand.SourceFiles {
				t.Fatalf("requested target or AST = %+v", file)
			}
			if demand.SourceFiles && file.SourceFile != generation.Native.Programs[i].GetSourceFile(paths[i]) {
				t.Fatal("requested AST lost its original source generation")
			}
		}
	}
}

func TestPipelinePreparesCompleteBindingBeforeRuleExecution(t *testing.T) {
	for _, failure := range []string{"missing target", "extra targets", "duplicate target", "duplicate wire path", "plugin source"} {
		t.Run(failure, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			first := tspath.ResolvePath(root, "first.ts")
			second := tspath.ResolvePath(root, "second.ts")
			var nativeRuns, pluginRuns, releases atomic.Int32
			generation := Generation{
				Native: NativeGeneration{
					Programs: []*program.Program{
						pipelineTestProgram(t, root, first, "const first = 1;"),
						pipelineTestProgram(t, root, second, "const second = 2;"),
					},
					TargetsByProgram: [][]string{{first}, {second}},
					RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{
							{Name: "native/check", Run: func(rule.RuleContext) rule.RuleListeners { nativeRuns.Add(1); return nil }},
							{Name: "plugin/check", IsEslintPluginRule: true},
						}
					},
				},
				Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) { return source.Text(), nil }},
				Plugin: &PluginGeneration{},
			}
			var want string
			switch failure {
			case "missing target":
				generation.Native.TargetsByProgram[1] = []string{second + ".missing"}
				want = "target is not present"
			case "extra targets":
				generation.Native.TargetsByProgram = append(generation.Native.TargetsByProgram, nil)
				want = "one entry per Program"
			case "duplicate target":
				generation.Target.Path = func(string) string { return "same.ts" }
				want = "duplicate projected target"
			case "duplicate wire path":
				generation.Plugin.WirePath = func(string) string { return "same.ts" }
				want = "duplicate plugin wire path"
			case "plugin source":
				generation.Target.ReadText = func(path string, source ast.SourceFileLike) (string, error) {
					if path == second {
						return "", errors.New("source unavailable")
					}
					return source.Text(), nil
				}
				want = "source unavailable"
			}
			_, err := RunPipeline(context.Background(), NewLintRequest(
				pipelineTestProvider(generation, func() { releases.Add(1) }), ObservationPolicy{},
				EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					pluginRuns.Add(1)
					return &EslintPluginLintResult{}, nil
				})))
			if err == nil || !strings.Contains(err.Error(), want) || nativeRuns.Load() != 0 || pluginRuns.Load() != 0 || releases.Load() != 1 {
				t.Fatalf("error/native/plugin/releases = %v/%d/%d/%d, want %q/0/0/1", err, nativeRuns.Load(), pluginRuns.Load(), releases.Load(), want)
			}
		})
	}
}
