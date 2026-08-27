package linter

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestCommittedAutofixCommitsOnlyFinalDeltaOnce(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t:        t,
		root:     root,
		fileName: tspath.ResolvePath(root, "source.ts"),
		initial:  "a",
		next:     map[string]string{"a": "b", "b": "c"},
	}
	committer := &pipelineFinalChangeRecorder{}
	committer.commit = func(_ context.Context, changes []FileChange) (CommitResult, error) {
		if provider.acquisitions != 3 || strings.Join(provider.observed, ",") != "a,b,c" {
			t.Fatalf("terminal commit ran before all observations: %+v", provider)
		}
		return CommitResult{ConfirmedPaths: []string{changes[0].Path}}, nil
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequestWithCommitter(
		provider,
		committer,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{MaxRounds: MaxFixRounds},
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || !applied.Committed || committer.commits != 1 || len(committer.committed) != 1 || len(committer.committed[0]) != 1 {
		t.Fatalf("committed result/committer = %+v / %+v", applied, committer)
	}
	change := committer.committed[0][0]
	if change.Before != "a" || change.After != "c" || change.AppliedDiagnostics != 2 {
		t.Fatalf("terminal delta = %+v", change)
	}
}

func TestAutofixTerminalCommitErrorReturnsConfirmedExternalState(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	paths := [2]string{tspath.ResolvePath(root, "first.ts"), tspath.ResolvePath(root, "second.ts")}
	initial := map[string]string{paths[0]: "a", paths[1]: "b"}
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
				SingleThreaded:   true,
				Cwd:              root,
				RulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{Name: "native/fix", Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
						textRange := core.NewTextRange(0, len(source.Text()))
						ruleCtx.ReportRangeWithFixes(textRange, rule.RuleMessage{Description: "fix"}, rule.RuleFix{Range: textRange, Text: "fixed"})
						return nil
					}}}
				},
			},
			Target: TargetProjection{ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
				return source.Text(), nil
			}},
		}, nil, nil
	})
	committer := &pipelineFinalChangeRecorder{
		commit: func(_ context.Context, changes []FileChange) (CommitResult, error) {
			return CommitResult{ConfirmedPaths: []string{changes[0].Path}}, errors.New("partial commit")
		},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequestWithCommitter(
		provider,
		committer,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{MaxRounds: 1},
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "partial commit") {
		t.Fatalf("apply error = %v", err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || applied.Verified || applied.Committed || len(applied.Rounds) != 1 ||
		len(applied.CommittedPaths) != 1 || applied.CommittedPaths[0] != paths[0] ||
		applied.Rounds[0].AppliedDiagnostics != 2 {
		t.Fatalf("partial applied result = %+v", applied)
	}
	if acquisitions != 1 || committer.commits != 1 {
		t.Fatalf("acquisitions/commits = %d/%d, want 1/1", acquisitions, committer.commits)
	}
}

func TestAutofixPropagatesCancellationFromTerminalCommitter(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	commitErr := errors.New("commit failed")
	provider := &pipelineAutofixProvider{
		t: t, root: root, fileName: tspath.ResolvePath(root, "source.ts"), initial: "a", next: map[string]string{"a": "b"},
	}
	committer := &pipelineFinalChangeRecorder{}
	committer.commit = func(context.Context, []FileChange) (CommitResult, error) {
		cancel()
		return CommitResult{}, commitErr
	}
	result, err := RunPipeline(ctx, NewAutofixRequestWithCommitter(
		provider,
		committer,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{MaxRounds: 1},
		nil,
	))
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), commitErr.Error()) {
		t.Fatalf("pipeline error = %v, want joined commit and cancellation errors", err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || applied.Verified || applied.Committed || len(applied.Rounds) != 1 || len(applied.FinalChanges) != 1 {
		t.Fatalf("applied result = %+v", applied)
	}
}

func TestCommittedAutofixRejectsNilCommitterBeforeAcquisition(t *testing.T) {
	provider := &pipelineAutofixProvider{}
	_, err := RunPipeline(context.Background(), NewAutofixRequestWithCommitter(
		provider,
		nil,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{MaxRounds: 1},
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "committer must not be nil") {
		t.Fatalf("nil committer error = %v", err)
	}
	if provider.acquisitions != 0 {
		t.Fatalf("invalid request acquired %d generations", provider.acquisitions)
	}
}

func TestAutofixPipelineRejectsFalseTerminalCommitConfirmation(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	provider := &pipelineAutofixProvider{
		t: t, root: root, fileName: tspath.ResolvePath(root, "source.ts"), initial: "a", next: map[string]string{"a": "b"},
	}
	committer := &pipelineFinalChangeRecorder{
		commit: func(context.Context, []FileChange) (CommitResult, error) { return CommitResult{}, nil },
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequestWithCommitter(
		provider,
		committer,
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		AutofixPolicy{MaxRounds: 1},
		nil,
	))
	if err == nil || !strings.Contains(err.Error(), "confirmed 0 of 1") {
		t.Fatalf("commit contract error = %v", err)
	}
	applied, ok := result.AppliedFixes()
	if !ok || applied.Verified || len(applied.Rounds) != 1 {
		t.Fatalf("partial contract result = %+v", applied)
	}
}

func TestCommitResultRejectsInvalidExternalConfirmations(t *testing.T) {
	planned := []FileChange{{Path: "a.ts", Before: "a", After: "b"}}
	tests := []struct {
		name   string
		result CommitResult
		err    error
	}{
		{name: "partial without error", result: CommitResult{}},
		{name: "complete with error", result: CommitResult{ConfirmedPaths: []string{"a.ts"}}, err: errors.New("commit failed")},
		{name: "extra confirmation", result: CommitResult{ConfirmedPaths: []string{"a.ts", "extra.ts"}}},
		{name: "duplicate confirmation", result: CommitResult{ConfirmedPaths: []string{"a.ts", "a.ts"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confirmed, err := validateCommitResult(planned, test.result, test.err)
			if err == nil {
				t.Fatal("invalid terminal confirmation was accepted")
			}
			if test.name == "partial without error" && len(confirmed) != 0 {
				t.Fatalf("unexpected confirmed paths: %+v", confirmed)
			}
		})
	}
}
