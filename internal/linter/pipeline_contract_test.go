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

func TestDeferredRootsRejectInvalidInputsBeforeBuilding(t *testing.T) {
	for _, name := range []string{"missing builder", "missing resolver", "duplicate deferred", "duplicate eager", "empty name", "empty projection", "retained files", "edits", "plugins", "type check", "autofix", "progressive"} {
		t.Run(name, func(t *testing.T) {
			generation := pipelineDeferredGeneration(t, 1)
			path := generation.Native.DeferredRoots.FileNames[0]
			var builds, releases int
			generation.Native.DeferredRoots.Build = func(context.Context, string) (*program.Program, error) {
				builds++
				return nil, errors.New("unexpected build")
			}
			policy := ObservationPolicy{}
			switch name {
			case "missing builder":
				generation.Native.DeferredRoots.Build = nil
			case "missing resolver":
				generation.Native.RulesForFile = nil
			case "duplicate deferred":
				generation.Native.DeferredRoots.FileNames = []string{path, path}
			case "duplicate eager":
				generation.Native.Programs = []*program.Program{pipelineTestProgram(t, generation.Native.Cwd, path, "const value = 1;")}
				generation.Native.TargetsByProgram = [][]string{{path}}
			case "empty name":
				generation.Native.DeferredRoots.FileNames = []string{""}
			case "empty projection":
				generation.Target.Path = func(string) string { return "" }
			case "retained files":
				policy.Demand.LintedFiles = true
			case "edits":
				policy.Demand.Native = rule.EditDemandSuggestion
			case "plugins":
				generation.Plugin = &PluginGeneration{}
			case "type check":
				generation.Native.TypeCheck = true
			}
			provider := pipelineTestProvider(generation, func() { releases++ })
			request := NewLintRequest(provider, policy, nil)
			if name == "autofix" {
				policy.Demand.Native = rule.EditDemandAutofix
				request = NewAutofixRequest(provider, policy, AutofixPolicy{}, nil)
			}
			if name == "progressive" {
				request = NewProgressiveLintRequest(provider, ArtifactDemand{}, &pipelineProgressiveDiagnostics{})
			}
			result, err := RunPipeline(context.Background(), request)
			if err == nil || builds != 0 || releases != 1 || result.Observation.Native.Lint != nil {
				t.Fatalf("error/builds/releases = %v/%d/%d", err, builds, releases)
			}
		})
	}
}

func TestDeferredRootsRejectInvalidProgramsAndRuleScopes(t *testing.T) {
	for _, name := range []string{"nil program", "wrong source", "extra source", "checker", "unknown rule", "external rule"} {
		t.Run(name, func(t *testing.T) {
			generation := pipelineDeferredGeneration(t, 1)
			path := generation.Native.DeferredRoots.FileNames[0]
			var ran bool
			generation.Native.RulesForFile = func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: "test", SupportsFileIsolation: name != "unknown rule", IsEslintPluginRule: name == "external rule", Run: func(rule.RuleContext) rule.RuleListeners { ran = true; return nil }}}
			}
			switch name {
			case "nil program":
				generation.Native.DeferredRoots.Build = func(context.Context, string) (*program.Program, error) {
					return nil, nil //nolint:nilnil // Exercise a builder that violates its contract.
				}
			case "wrong source":
				generation.Native.DeferredRoots.Build = func(context.Context, string) (*program.Program, error) {
					return pipelineTestProgram(t, generation.Native.Cwd, path+"x.ts", "const value = 1;"), nil
				}
			case "extra source", "checker":
				compiled, paths := createTestProgramWithFiles(t, map[string]string{"source.ts": "export const value = 1;", "other.ts": "export const other = 1;"})
				generation.Native.DeferredRoots.FileNames = []string{paths["source.ts"]}
				p := program.NewFromCompiler(compiled)
				if name == "extra source" {
					var err error
					p, err = program.NewFromBoundSources(compiled, compiled.GetSourceFiles())
					if err != nil {
						t.Fatal(err)
					}
				}
				generation.Native.DeferredRoots.Build = func(context.Context, string) (*program.Program, error) { return p, nil }
			}
			_, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, nil), ObservationPolicy{}, nil))
			if err == nil || ran {
				t.Fatalf("error/rule ran = %v/%v", err, ran)
			}
		})
	}
}
