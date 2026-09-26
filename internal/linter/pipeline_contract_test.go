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
				RulesForPath: func(string) []rule.ConfiguredRule {
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
		generation.Native.RulesForPath = func(string) []rule.ConfiguredRule {
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
		recovered, releases := runPipelineWithParallelRuleResolver(t, func(source string) []rule.ConfiguredRule {
			if arrivals.Add(1) == 2 {
				close(releaseResolvers)
			}
			select {
			case <-releaseResolvers:
			case <-time.After(5 * time.Second):
				panic("timed out waiting for both parallel resolvers")
			}
			panic(source)
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
		recovered, releases := runPipelineWithParallelRuleResolver(t, func(string) []rule.ConfiguredRule {
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

func TestCanIsolateSourceFileKeepsUnknownAndExternalRulesConservative(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rules []rule.ConfiguredRule
		want  bool
	}{
		{"empty", nil, true},
		{"audited", []rule.ConfiguredRule{{SupportsFileIsolation: true}}, true},
		{"unknown", []rule.ConfiguredRule{{}}, false},
		{"type only", []rule.ConfiguredRule{{RequiresTypeInfo: true}}, true},
		{"mixed", []rule.ConfiguredRule{{SupportsFileIsolation: true}, {}}, false},
		{"external", []rule.ConfiguredRule{{SupportsFileIsolation: true, IsEslintPluginRule: true}}, false},
		{"external type", []rule.ConfiguredRule{{RequiresTypeInfo: true, IsEslintPluginRule: true}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := canIsolateSourceFile(tc.rules); got != tc.want {
				t.Fatalf("CanIsolateSourceFile = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRootGroupsMaterializeForSourceConsumers(t *testing.T) {
	for _, mode := range []string{"lint", "unknown rule", "external rule", "retained files", "edits", "plugins", "type check", "autofix", "progressive"} {
		t.Run(mode, func(t *testing.T) {
			generation := pipelineDeferredGeneration(t, 2)
			generation.Native.SingleThreaded = true
			var calls, runs int
			var sourceCounts []int
			generation.Native.RulesForPath = func(string) []rule.ConfiguredRule {
				calls++
				rules := []rule.ConfiguredRule{{Name: "native/check", SupportsFileIsolation: mode != "unknown rule", Run: func(ctx rule.RuleContext) rule.RuleListeners {
					runs++
					sourceCounts = append(sourceCounts, len(ctx.Program().SourceFiles()))
					ctx.ReportNode(ctx.SourceFile.AsNode(), rule.RuleMessage{Description: "source"})
					return nil
				}}}
				if mode == "external rule" {
					rules = append(rules, rule.ConfiguredRule{Name: "external/check", IsEslintPluginRule: true})
				}
				return rules
			}
			policy := ObservationPolicy{}
			switch mode {
			case "retained files":
				policy.Demand.LintedFiles = true
			case "edits":
				policy.Demand.Native = rule.EditDemandSuggestion
			case "plugins":
				generation.Plugin = &PluginGeneration{}
			case "type check":
				generation.Native.TypeCheck = true
			}
			provider := pipelineTestProvider(generation, nil)
			request := NewLintRequest(provider, policy, nil)
			if mode == "autofix" {
				policy.Demand.Native = rule.EditDemandAutofix
				request = NewAutofixRequest(provider, policy, AutofixPolicy{}, nil)
			}
			if mode == "progressive" {
				request = NewProgressiveLintRequest(provider, ArtifactDemand{}, &pipelineProgressiveDiagnostics{})
			}
			result, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 2 || runs != 2 || result.Observation.Native.Lint.LintedFileCount != 2 {
				t.Fatalf("resolver/rule/file counts = %d/%d/%d", calls, runs, result.Observation.Native.Lint.LintedFileCount)
			}
			streamed := mode == "lint" || mode == "edits" || mode == "progressive"
			for index, count := range sourceCounts {
				want := 2
				if streamed {
					want = 1
				}
				if count != want {
					t.Fatalf("source universe = %d, want %d", count, want)
				}
				_, retained := result.Observation.Native.Diagnostics[index].SourceFile.(*ast.SourceFile)
				if retained == streamed {
					t.Fatalf("AST retained = %v, streamed = %v", retained, streamed)
				}
			}
			if mode == "retained files" && len(result.Observation.Native.Files) != 2 {
				t.Fatal("requested source artifacts were lost")
			}
		})
	}
}

func TestRootGroupsRejectInvalidConstruction(t *testing.T) {
	for _, mode := range []string{"missing builder", "missing resolver", "empty name", "duplicate root", "duplicate projection", "nil program", "wrong source", "extra source", "checker"} {
		t.Run(mode, func(t *testing.T) {
			generation := pipelineDeferredGeneration(t, 2)
			generation.Native.SingleThreaded = true
			group := &generation.Native.RootGroups[0]
			path := group.FileNames[0]
			var ran bool
			generation.Native.RulesForPath = func(string) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{SupportsFileIsolation: true, Run: func(rule.RuleContext) rule.RuleListeners { ran = true; return nil }}}
			}
			switch mode {
			case "missing builder":
				group.Build = nil
			case "missing resolver":
				generation.Native.RulesForPath = nil
			case "empty name":
				group.FileNames = []string{""}
			case "duplicate root":
				group.FileNames = []string{path, path}
			case "duplicate projection":
				generation.Target.Path = func(string) string { return "same" }
			case "nil program":
				group.Build = func(context.Context, []string) (*program.Program, error) {
					return nil, nil //nolint:nilnil // Deliberately violate the construction contract.
				}
			case "wrong source":
				group.Build = func(context.Context, []string) (*program.Program, error) {
					return pipelineTestProgram(t, generation.Native.Cwd, path+"x.ts", "const value = 1;"), nil
				}
			case "extra source", "checker":
				compiled, paths := createTestProgramWithFiles(t, map[string]string{"source.ts": "export const value = 1;", "other.ts": "export const other = 1;"})
				group.FileNames = []string{paths["source.ts"]}
				p := program.NewFromCompiler(compiled)
				if mode == "extra source" {
					var err error
					p, err = program.NewFromBoundSources(compiled, compiled.GetSourceFiles())
					if err != nil {
						t.Fatal(err)
					}
				}
				group.Build = func(context.Context, []string) (*program.Program, error) { return p, nil }
			}
			var releases int
			_, err := RunPipeline(context.Background(), NewLintRequest(pipelineTestProvider(generation, func() { releases++ }), ObservationPolicy{}, nil))
			if err == nil || ran || releases != 1 {
				t.Fatalf("error/ran/releases = %v/%v/%d", err, ran, releases)
			}
		})
	}
}
