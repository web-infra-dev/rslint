package lsp

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func runLSPGenerationForTest(
	ctx context.Context,
	generation linter.Generation,
	release linter.ReleaseFunc,
	demand linter.ArtifactDemand,
) (linter.PipelineResult, error) {
	return linter.RunPipeline(ctx, linter.NewLintRequest(
		linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
			return generation, release, nil
		}),
		linter.ObservationPolicy{
			Demand:        demand,
			Plugin:        linter.PluginAfterNativeJoined,
			PluginFailure: linter.PluginDiscardOnFailure,
		},
		func(_ context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			results := make([]linter.EslintPluginFileResult, len(request.Files))
			for index, file := range request.Files {
				results[index].FilePath = file.Path
			}
			return &linter.EslintPluginLintResult{Results: results}, nil
		},
	))
}

func TestLSPGenerationReleaseLifecycle(t *testing.T) {
	t.Run("normal execution", func(t *testing.T) {
		directory, file, program, fsys := programOver(t, "const value = 1;\n")
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		ruleRan := false
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name:     "test/release",
				Severity: rule.SeverityWarning,
				Run: func(rule.RuleContext) rule.RuleListeners {
					if releases != 0 {
						t.Fatal("generation finalized before rule execution completed")
					}
					ruleRan = true
					return nil
				},
			}},
			nil,
			nil,
		)
		_, err := runLSPGenerationForTest(
			context.Background(), generation, func() { releases++ }, linter.ArtifactDemand{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !ruleRan || releases != 1 {
			t.Fatalf("rule ran/release calls = %v/%d, want true/1", ruleRan, releases)
		}
	})

	t.Run("syntax error", func(t *testing.T) {
		directory, file, program, fsys := programOver(t, "const value = ;\n")
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		ruleRan := false
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name: "test/must-not-run",
				Run: func(rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					return nil
				},
			}},
			nil,
			nil,
		)
		result, err := runLSPGenerationForTest(
			context.Background(), generation, func() { releases++ }, linter.ArtifactDemand{Native: rule.EditDemandAll},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Observation.Native.HasTargetSyntaxErrors || ruleRan || releases != 1 {
			t.Fatalf("syntax/rule/release = %v/%v/%d, want true/false/1", result.Observation.Native.HasTargetSyntaxErrors, ruleRan, releases)
		}
	})

	t.Run("plugin work detached before release", func(t *testing.T) {
		const content = "const value = 1;\n"
		directory, file, program, fsys := programOver(t, content)
		sourceFile := sourceFileForPath(program, file, fsys)
		releases := 0
		generation := newLintGeneration(
			program,
			sourceFile,
			lspConfigTarget(file, directory, fsys),
			directory,
			true,
			[]rule.ConfiguredRule{{
				Name:               "test/plugin",
				Severity:           rule.SeverityWarning,
				IsEslintPluginRule: true,
			}},
			&linter.EslintPluginFileConfig{ConfigKey: "frozen-config"},
			nil,
		)
		presentation := &capturedProgressiveDiagnostics{}
		_, err := linter.RunPipeline(context.Background(), linter.NewProgressiveLintRequest(
			linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
				return generation, func() { releases++ }, nil
			}),
			linter.ArtifactDemand{Plugin: rule.EditDemandAll},
			presentation,
		))
		if err != nil {
			t.Fatal(err)
		}
		if releases != 1 || presentation.run == nil {
			t.Fatalf("release/plugin run = %d/%v, want 1/non-nil", releases, presentation.run != nil)
		}
		var request linter.EslintPluginLintRequest
		outcome, err := presentation.run(context.Background(), func(_ context.Context, got linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			request = got
			return &linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: got.Files[0].Path}}}, nil
		})
		if err != nil || outcome.DispatchError != nil {
			t.Fatalf("plugin work errors = %v/%v", err, outcome.DispatchError)
		}
		if len(request.Files) != 1 || request.Files[0].Text == nil || *request.Files[0].Text != content {
			t.Fatalf("detached plugin request = %+v", request.Files)
		}
	})

	t.Run("canceled before acquisition", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		acquisitions := 0
		result, err := linter.RunPipeline(ctx, linter.NewLintRequest(
			linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
				acquisitions++
				return linter.Generation{}, nil, nil
			}),
			linter.ObservationPolicy{},
			nil,
		))
		_ = result
		if !errors.Is(err, context.Canceled) || acquisitions != 0 {
			t.Fatalf("error/acquisitions = %v/%d, want context.Canceled/0", err, acquisitions)
		}
	})
}

func TestDocumentGenerationProviderFinalizesBeforePublication(t *testing.T) {
	fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
	server := fixture.server
	server.backgroundCtx = context.Background()
	if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
		t.Fatal(err)
	}
	server.lintPrograms = fixture.store
	snapshot := documentLintSnapshotForTest(
		server,
		fixture.sourceURI,
		config.RslintConfig{},
		filepath.Dir(fixture.configPath),
		false,
		[]string{fixture.configPath},
	)
	releases := 0
	provider := &documentGenerationProvider{
		server:   server,
		uri:      fixture.sourceURI,
		snapshot: snapshot,
		requestPrograms: func(
			context.Context,
			lsproto.DocumentUri,
			target.File,
		) (lintProjectLoaders, linter.ReleaseFunc) {
			return lintProjectLoaders{}, func() { releases++ }
		},
		buildGeneration: func(
			*compiler.Program,
			*ast.SourceFile,
			target.File,
			string,
			bool,
			documentLintSnapshot,
		) linter.Generation {
			panic("generation assembly failed")
		},
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _, _ = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
	}()
	if recovered == nil {
		t.Fatal("expected generation assembly panic")
	}
	if releases != 1 {
		t.Fatalf("resident Program finalizer calls = %d, want 1", releases)
	}
}
