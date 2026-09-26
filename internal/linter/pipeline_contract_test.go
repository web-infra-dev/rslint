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
	t.Run("invalid deferred source set", func(t *testing.T) {
		var resolutions, releases int
		generation := pipelineDeferredTestGeneration(t, map[string]string{"valid.ts": "export {};"}, func(string) []rule.ConfiguredRule {
			resolutions++
			return nil
		})
		generation.Native.DeferredSources = append(generation.Native.DeferredSources, nil)
		_, err := RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, func() { releases++ }), ObservationPolicy{}, nil,
		))
		if err == nil || !strings.Contains(err.Error(), "invalid deferred source set") || resolutions != 0 || releases != 1 {
			t.Fatalf("error/resolutions/releases = %v/%d/%d", err, resolutions, releases)
		}
	})

	for _, projection := range []string{"empty", "duplicate"} {
		t.Run("deferred "+projection+" projection", func(t *testing.T) {
			var calls, releases atomic.Int32
			generation := pipelineDeferredTestGeneration(t, map[string]string{
				"first.ts": "export {};", "second.ts": "export {};",
			}, func(string) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "native/check", Run: func(rule.RuleContext) rule.RuleListeners {
					calls.Add(1)
					return nil
				}}}
			})
			generation.Target.Path = func(string) string {
				if projection == "empty" {
					return ""
				}
				return "same.ts"
			}
			_, err := RunPipeline(context.Background(), NewLintRequest(
				pipelineTestProvider(generation, func() { releases.Add(1) }), ObservationPolicy{}, nil,
			))
			if err == nil || !strings.Contains(err.Error(), "projected target") || calls.Load() != 0 || releases.Load() != 1 {
				t.Fatalf("error/calls/releases = %v/%d/%d", err, calls.Load(), releases.Load())
			}
		})
	}

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

func TestPipelineDetachesTypeCheckOnlyDiagnosticSources(t *testing.T) {
	compiled, paths := createTestProgramWithFiles(t, map[string]string{
		"source.ts": "const value: number = 'wrong';",
	})
	result, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(Generation{Native: NativeGeneration{
			Programs: wrapTestPrograms(compiled), TypeCheck: true, SingleThreaded: true,
		}}, nil),
		ObservationPolicy{}, nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := result.Observation.Native.Diagnostics
	if len(diagnostics) == 0 {
		t.Fatal("type-check-only observation lost its diagnostics")
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Origin != rule.DiagnosticOriginTypeScript || !diagnostic.PreFormatted || diagnostic.FilePath != paths["source.ts"] {
			t.Fatalf("type diagnostic metadata changed: %+v", diagnostic)
		}
		if _, retainedAST := diagnostic.SourceFile.(*ast.SourceFile); retainedAST {
			t.Fatal("type-check-only diagnostic retained its AST")
		}
		if diagnostic.SourceFile.Text() != "const value: number = 'wrong';" {
			t.Fatalf("type diagnostic source = %q", diagnostic.SourceFile.Text())
		}
	}
}

func TestPipelineCollectsLintedFilesOnlyWhenDemanded(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(t, root, fileName, "const value = 1;", nil, nil)

	withoutFiles, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	if withoutFiles.Observation.Native.Files != nil {
		t.Fatalf("unrequested linted files = %+v", withoutFiles.Observation.Native.Files)
	}

	withFiles, err := RunPipeline(context.Background(), NewLintRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{Demand: ArtifactDemand{LintedFiles: true}},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	files := withFiles.Observation.Native.Files
	if len(files) != 1 || files[0].Path != fileName || files[0].SourceFile == nil {
		t.Fatalf("requested linted files = %+v", files)
	}
}
