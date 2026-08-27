package lsp

import (
	"context"
	"testing"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func configuredSpeculativePipelineResultForTest(
	s *Server,
	uri lsproto.DocumentUri,
	ctx context.Context,
	content string,
	entries config.RslintConfig,
	configDirectory string,
	usesJavaScriptConfig bool,
	typeScriptConfigPaths []string,
) (linter.PipelineResult, error) {
	return speculativePipelineResultForTest(
		s,
		ctx,
		uri,
		content,
		documentLintSnapshotForTest(
			s,
			uri,
			entries,
			configDirectory,
			usesJavaScriptConfig,
			typeScriptConfigPaths,
		),
	)
}

func configuredDocumentPipelineResultForTest(
	s *Server,
	ctx context.Context,
	uri lsproto.DocumentUri,
	entries config.RslintConfig,
	configDirectory string,
	usesJavaScriptConfig bool,
	typeScriptConfigPaths []string,
) (linter.PipelineResult, error) {
	snapshot := documentLintSnapshotForTest(
		s,
		uri,
		entries,
		configDirectory,
		usesJavaScriptConfig,
		typeScriptConfigPaths,
	)
	return linter.RunPipeline(ctx, linter.NewLintRequest(
		&documentGenerationProvider{server: s, uri: uri, snapshot: snapshot},
		linter.ObservationPolicy{
			Demand: linter.ArtifactDemand{
				Native: rule.EditDemandAll,
				Plugin: rule.EditDemandAll,
			},
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

func speculativePipelineResultForTest(
	s *Server,
	ctx context.Context,
	uri lsproto.DocumentUri,
	content string,
	snapshot documentLintSnapshot,
) (linter.PipelineResult, error) {
	environment := s.freezeSpeculativeLintEnvironment(uri, snapshot.target)
	if environment.baseFS == nil {
		environment.baseFS = osvfs.FS()
	}
	generation, release, err := acquireSpeculativeGeneration(ctx, content, snapshot, environment)
	if err != nil {
		return linter.PipelineResult{}, err
	}
	return runLSPGenerationForTest(
		ctx,
		generation,
		release,
		linter.ArtifactDemand{
			Native: rule.EditDemandAutofix,
			Plugin: rule.EditDemandAll,
		},
	)
}

func runSpeculativeFixAllForTest(
	t *testing.T,
	s *Server,
	ctx context.Context,
	uri lsproto.DocumentUri,
	original string,
	snapshot documentLintSnapshot,
) string {
	t.Helper()
	provider := s.newSpeculativeGenerationProvider(uri, original, snapshot)
	if provider.environment.baseFS == nil {
		provider.environment.baseFS = osvfs.FS()
	}
	dispatch, cancel := s.pluginDispatchWithinBudget(ctx, snapshot.pluginGeneration)
	defer cancel()
	result, err := linter.RunPipeline(ctx, linter.NewAutofixRequest(
		provider,
		linter.ObservationPolicy{
			Demand: linter.ArtifactDemand{
				Native: rule.EditDemandAutofix,
				Plugin: rule.EditDemandAutofix,
			},
			Plugin:        linter.PluginAfterNativeJoined,
			PluginFailure: linter.PluginDiscardOnFailure,
		},
		linter.AutofixPolicy{
			MaxRounds:                maxFixRounds,
			StopOnTargetSyntaxErrors: true,
		},
		dispatch,
	))
	if err != nil {
		t.Fatalf("run speculative fix-all: %v", err)
	}
	reportFixAllPluginOutcomes(uri, result.PluginOutcomes())
	content, err := speculativeContentFromResult(result, snapshot.target.Path, original)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
