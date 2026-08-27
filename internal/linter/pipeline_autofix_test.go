package linter

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestAutofixPipelineOwnsMemoryRoundsAndReobservesSnapshots(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "b", "b": "c"},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != 2 ||
		provider.acquisitions != 3 || strings.Join(provider.observed, ",") != "a,b,c" {
		t.Fatalf("result/provider = %+v / %+v", applied, provider)
	}
	if len(applied.FinalChanges) != 1 || applied.FinalChanges[0].Before != "a" || applied.FinalChanges[0].After != "c" {
		t.Fatalf("final in-memory delta = %+v", applied.FinalChanges)
	}
	if applied.Rounds[0].AppliedDiagnostics != 1 || applied.Rounds[1].AppliedDiagnostics != 1 {
		t.Fatalf("applied diagnostic counts = %+v", applied.Rounds)
	}
}

func TestAutofixSyntaxGateStopsBeforeAfterNativePluginDispatch(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "const value = ;",
		next:     map[string]string{},
		rules: func(string) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}
		},
	}
	dispatches := 0
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{
			Demand: ArtifactDemand{
				Native: rule.EditDemandAutofix,
				Plugin: rule.EditDemandAutofix,
			},
			Plugin:        PluginAfterNativeJoined,
			PluginFailure: PluginDiscardOnFailure,
		},
		AutofixPolicy{StopOnTargetSyntaxErrors: true},
		func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			dispatches++
			return &EslintPluginLintResult{}, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != 0 || !result.Observation.Native.HasTargetSyntaxErrors {
		t.Fatalf("syntax-gated applied result = %+v", applied)
	}
	if dispatches != 0 {
		t.Fatalf("plugin dispatches after target syntax error = %d, want 0", dispatches)
	}
}

func TestAutofixSyntaxGateStopsConcurrentPluginBeforeDispatch(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	brokenPath := tspath.ResolvePath(root, "broken.ts")
	pluginPath := tspath.ResolvePath(root, "plugin.ts")
	generation := Generation{
		Native: NativeGeneration{
			Programs: []*program.Program{
				pipelineTestProgram(t, root, brokenPath, "const value = ;"),
				pipelineTestProgram(t, root, pluginPath, "const value = 1;"),
			},
			TargetsByProgram: [][]string{{brokenPath}, {pluginPath}},
			SingleThreaded:   true,
			RulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
				if source.FileName() != pluginPath {
					return nil
				}
				return []rule.ConfiguredRule{{Name: "plugin/fix", IsEslintPluginRule: true}}
			},
		},
		Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
			return source.Text(), nil
		}},
		Plugin: &PluginGeneration{},
	}
	dispatches := 0
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{
			Demand: ArtifactDemand{Plugin: rule.EditDemandAutofix},
			Plugin: PluginConcurrentJoined,
		},
		autofixPolicyForTest(1, AutofixPolicy{StopOnTargetSyntaxErrors: true}),
		func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			dispatches++
			return &EslintPluginLintResult{}, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != 0 ||
		!result.Observation.Native.HasTargetSyntaxErrors || dispatches != 0 {
		t.Fatalf("syntax-gated result/dispatches = %+v/%d", applied, dispatches)
	}
}

func TestAutofixUsesProductRoundLimitThenVerifiesOnce(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	next := make(map[string]string, maxFixRounds)
	for round := range maxFixRounds {
		next[strconv.Itoa(round)] = strconv.Itoa(round + 1)
	}
	var fixArtifactBuilds atomic.Int32
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "0",
		next:     next,
		rules: func(string) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: "native/demand-probe",
				Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
					probeRange := core.NewTextRange(0, 0)
					ruleCtx.ReportRangeWithDeferredFixes(
						probeRange,
						rule.RuleMessage{Description: "probe"},
						func() []rule.RuleFix {
							fixArtifactBuilds.Add(1)
							return nil
						},
					)
					return nil
				},
			}}
		},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{
			VerifyAfterLastRound: true,
			VerificationDemand: ArtifactDemand{
				Native:      rule.EditDemandSuggestion,
				LintedFiles: true,
			},
		},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != maxFixRounds || applied.Last.Index != maxFixRounds {
		t.Fatalf("applied result = %+v", applied)
	}
	if applied.Initial.Native.Files != nil || len(applied.Last.Native.Files) != 1 {
		t.Fatalf(
			"per-observation linted-file demand = initial:%+v verification:%+v",
			applied.Initial.Native.Files,
			applied.Last.Native.Files,
		)
	}
	if provider.acquisitions != maxFixRounds+1 || provider.observed[len(provider.observed)-1] != strconv.Itoa(maxFixRounds) {
		t.Fatalf("provider acquisitions/observations = %d/%+v", provider.acquisitions, provider.observed)
	}
	if fixArtifactBuilds.Load() != maxFixRounds {
		t.Fatalf("autofix artifact builds = %d, want %d; final verification demand was not isolated", fixArtifactBuilds.Load(), maxFixRounds)
	}
}

func TestAutofixCanStopAtRoundLimitWithoutVerification(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "b", "b": "c", "c": "d"},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(2, AutofixPolicy{}),
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || applied.Verified || len(applied.Rounds) != 2 || applied.Last.Index != 1 {
		t.Fatalf("unverified applied result = %+v", applied)
	}
	if provider.acquisitions != 2 || strings.Join(provider.observed, ",") != "a,b" ||
		len(applied.FinalChanges) != 1 || applied.FinalChanges[0].After != "c" {
		t.Fatalf("provider/result = %+v / %+v", provider, applied)
	}
}

func TestAutofixRestoredInitialReusesInitialObservation(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "b", "b": "a"},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(3, AutofixPolicy{}),
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != 2 || !applied.Rounds[1].RestoredInitial ||
		applied.Last.Index != applied.Initial.Index || result.Observation.Index != 0 || len(applied.FinalChanges) != 0 ||
		len(applied.FinalSources) != 1 || applied.FinalSources[0].Path != provider.fileName || applied.FinalSources[0].Text != "a" {
		t.Fatalf("restored applied result/provider = %+v / %+v", applied, provider)
	}
	if provider.acquisitions != 2 || strings.Join(provider.observed, ",") != "a,b" {
		t.Fatalf("restored cycle provider = %+v", provider)
	}
}

func TestAutofixSameTextFixConsumesRound(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "a"},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(1, AutofixPolicy{}),
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Verified || len(applied.Rounds) != 1 || len(applied.FinalChanges) != 0 ||
		len(applied.FinalSources) != 1 || applied.FinalSources[0].Text != "a" ||
		len(applied.Rounds[0].ChangedPaths) != 1 || applied.Rounds[0].AppliedDiagnostics != 1 {
		t.Fatalf("same-text applied result/provider = %+v / %+v", applied, provider)
	}
}

func TestAutofixFinalSourcesIncludePartiallyRestoredFiles(t *testing.T) {
	state := newAutofixState()
	if _, err := state.apply([]FileChange{
		{Path: "a.js", Before: "a", After: "b", AppliedDiagnostics: 1},
		{Path: "b.js", Before: "x", After: "y", AppliedDiagnostics: 1},
	}); err != nil {
		t.Fatal(err)
	}
	round, err := state.apply([]FileChange{
		{Path: "a.js", Before: "b", After: "a", AppliedDiagnostics: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if round.RestoredInitial {
		t.Fatal("one restored file must not hide another file's remaining change")
	}
	changes := state.finalChanges()
	if len(changes) != 1 || changes[0].Path != "b.js" || changes[0].After != "y" {
		t.Fatalf("net changes = %+v, want only b.js", changes)
	}
	sources := state.finalSources()
	if len(sources) != 2 ||
		sources[0] != (SourceFileSnapshot{Path: "a.js", Text: "a"}) ||
		sources[1] != (SourceFileSnapshot{Path: "b.js", Text: "y"}) {
		t.Fatalf("final fixed sources = %+v", sources)
	}
}

func TestAutofixRejectsRoundLimitAboveProductBound(t *testing.T) {
	provider := &pipelineAutofixProvider{}
	_, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(maxFixRounds+1, AutofixPolicy{}),
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "product safety bound") {
		t.Fatalf("round bound error = %v", err)
	}
	if provider.acquisitions != 0 {
		t.Fatalf("invalid request acquired %d generations", provider.acquisitions)
	}
}

func TestAutofixReobserveFailurePreservesLastSuccessfulObservation(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	nativeFinished := make(chan struct{})
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "b"},
		rules: func(content string) []rule.ConfiguredRule {
			rules := []rule.ConfiguredRule{{Name: "plugin/check", IsEslintPluginRule: true}}
			if content == "b" {
				rules = append(rules, rule.ConfiguredRule{
					Name: "native/second-only",
					Run: func(rule.RuleContext) rule.RuleListeners {
						close(nativeFinished)
						return nil
					},
				})
			}
			return rules
		},
	}
	committer := &pipelineFinalChangeRecorder{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var dispatches atomic.Int32
	result, err := RunPipeline(ctx, NewAutofixRequestWithCommitter(
		provider,
		committer,
		ObservationPolicy{
			Demand: ArtifactDemand{
				Native: rule.EditDemandAutofix,
				Plugin: rule.EditDemandAutofix,
			},
			Plugin:        PluginConcurrentJoined,
			PluginFailure: PluginDiscardOnFailure,
		},
		autofixPolicyForTest(1, AutofixPolicy{VerifyAfterLastRound: true}),
		func(dispatchCtx context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			if dispatches.Add(1) == 2 {
				<-nativeFinished
				cancel()
				return nil, dispatchCtx.Err()
			}
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
				FilePath: request.Files[0].Path,
			}}}, nil
		},
	))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pipeline error = %v, want context.Canceled", err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || applied.Verified || applied.Last.Index != 0 || result.Observation.Index != 0 ||
		len(applied.FinalChanges) != 1 || applied.FinalChanges[0].After != "b" || committer.commits != 0 {
		t.Fatalf("applied result = %+v, observation=%+v", applied, result.Observation)
	}
	records := result.PluginOutcomes()
	if len(records) != 2 || records[1].Observation != 1 || !errors.Is(records[1].DispatchError, context.Canceled) {
		t.Fatalf("plugin records = %+v, want failed observation 1 retained", records)
	}
	if _, leaked := result.ExecutedRules()["native/second-only"]; leaked {
		t.Fatal("failed re-observation polluted successful rule aggregation")
	}
}

func TestAutofixRejectsProviderThatIgnoresMemorySnapshot(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fileName := tspath.ResolvePath(root, "source.ts")
	generation := pipelineTestGeneration(t, root, fileName, "a", []rule.ConfiguredRule{{
		Name: "native/fix",
		Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
			textRange := core.NewTextRange(0, 1)
			ruleCtx.ReportRangeWithFixes(textRange, rule.RuleMessage{Description: "fix"}, rule.RuleFix{Range: textRange, Text: "b"})
			return nil
		},
	}}, nil)
	acquisitions := 0
	provider := GenerationProviderFunc(func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error) {
		acquisitions++
		return generation, nil, nil
	})
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		provider,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(2, AutofixPolicy{}),
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "did not materialize in-memory target") {
		t.Fatalf("stale generation error = %v", err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || acquisitions != 2 || len(applied.FinalChanges) != 1 || applied.FinalChanges[0].After != "b" {
		t.Fatalf("stale provider result/acquisitions = %+v/%d", applied, acquisitions)
	}
}
