package linter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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
		func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
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
		},
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
				func(dispatchCtx context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					<-nativeFinished
					cancel()
					return nil, dispatchCtx.Err()
				},
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
		func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return nil, context.DeadlineExceeded
		},
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
				SingleThreaded:   true,
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
			Plugin: &PluginGeneration{HostReadsInitialText: true},
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
		func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			dispatches++
			if len(request.Files) != 2 {
				dispatchValidationErr = fmt.Errorf("plugin files = %+v", request.Files)
			}
			for _, file := range request.Files {
				if dispatches == 1 && file.Text != nil {
					dispatchValidationErr = fmt.Errorf("initial disk-backed input %q was inlined", file.Path)
				}
				if dispatches == 2 {
					if file.Text == nil || *file.Text != expectedSecond[file.Path] {
						dispatchValidationErr = fmt.Errorf("memory-generation input = %+v", file)
					}
				}
			}
			results := make([]EslintPluginFileResult, len(request.Files))
			for index, file := range request.Files {
				results[index].FilePath = file.Path
			}
			return &EslintPluginLintResult{Results: results}, nil
		},
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
		func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return &EslintPluginLintResult{}, nil
		},
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
