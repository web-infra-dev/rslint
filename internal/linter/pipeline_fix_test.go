package linter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestFixSourcesRejectDistinctSourcesForSamePath(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	firstPath := tspath.ResolvePath(root, "first.ts")
	secondPath := tspath.ResolvePath(root, "second.ts")
	first := pipelineTestProgram(t, root, firstPath, "a").SourceFiles()[0]
	second := pipelineTestProgram(t, root, secondPath, "b").SourceFiles()[0]
	firstFixes := []rule.RuleFix{{Range: core.NewTextRange(0, 1), Text: "x"}}
	secondFixes := []rule.RuleFix{{Range: core.NewTextRange(0, 1), Text: "y"}}
	_, err := fixSourcesFromDiagnostics([]rule.RuleDiagnostic{
		{FilePath: "shared.ts", SourceFile: first, FixesPtr: &firstFixes},
		{FilePath: "shared.ts", SourceFile: second, FixesPtr: &secondFixes},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate fix target") {
		t.Fatalf("duplicate fix source error = %v", err)
	}
}

func TestPipelineDetachesSourcesAfterFreezingNativeAndPluginFixes(t *testing.T) {
	for _, mode := range []PluginExecution{PluginConcurrentJoined, PluginAfterNativeJoined} {
		t.Run(fmt.Sprintf("mode-%d", mode), func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			path := tspath.ResolvePath(root, "source.ts")
			generation := pipelineTestGeneration(t, root, path, "ab", []rule.ConfiguredRule{
				{
					Name: "native/fix", Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						span := core.NewTextRange(0, 1)
						ctx.ReportRangeWithFixes(span, rule.RuleMessage{Description: "native"}, rule.RuleFix{Range: span, Text: "A"})
						return nil
					},
				},
				{Name: "plugin/fix", Severity: rule.SeverityError, IsEslintPluginRule: true},
			}, &EslintPluginFileConfig{})
			original := generation.Native.Programs[0].SourceFiles()[0]
			generation.Target.ReadText = func(_ string, source ast.SourceFileLike) (string, error) {
				if source != original {
					return "", errors.New("source identity changed before fix text was frozen")
				}
				return source.Text(), nil
			}
			result, err := RunPipeline(context.Background(), NewAutofixRequest(
				pipelineTestProvider(generation, nil),
				ObservationPolicy{Plugin: mode, Demand: ArtifactDemand{Native: rule.EditDemandAutofix, Plugin: rule.EditDemandAutofix}},
				autofixPolicyForTest(1, AutofixPolicy{}),
				func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
					return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
						FilePath: request.Files[0].Path,
						Diagnostics: []EslintPluginDiagnostic{{
							RuleName: "plugin/fix", Message: "plugin", StartPos: 1, EndPos: 2,
							Fixes: []EslintPluginFix{{Range: [2]int{1, 2}, Text: "B"}},
						}},
					}}}, nil
				},
			))
			if err != nil {
				t.Fatal(err)
			}
			applied, ok := result.AppliedFixes()
			if !ok || len(applied.FinalChanges) != 1 || applied.FinalChanges[0].After != "AB" {
				t.Fatalf("combined fixes = %+v", applied.FinalChanges)
			}
			diagnostics, _ := applied.Initial.CompleteDiagnostics()
			for _, diagnostic := range diagnostics {
				if _, retained := diagnostic.SourceFile.(*ast.SourceFile); retained {
					t.Fatal("fix result retained a compiler AST")
				}
			}
		})
	}
}

func TestPipelineFreezesTextOnlyForFixableTargets(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	fixablePath := tspath.ResolvePath(root, "fixable.ts")
	nonFixablePath := tspath.ResolvePath(root, "clean.ts")
	fixableProgram := pipelineTestProgram(t, root, fixablePath, "a")
	nonFixableProgram := pipelineTestProgram(t, root, nonFixablePath, "clean")
	generation := Generation{
		Native: NativeGeneration{
			Programs:         []*program.Program{fixableProgram, nonFixableProgram},
			TargetsByProgram: [][]string{{fixablePath}, {nonFixablePath}},
			SingleThreaded:   true,
			Cwd:              root,
			RulesForFile: func(source *ast.SourceFile) []rule.ConfiguredRule {
				if source.FileName() != fixablePath {
					return nil
				}
				return []rule.ConfiguredRule{{
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
				}}
			},
		},
		Target: TargetProjection{
			ReadText: func(path string, source ast.SourceFileLike) (string, error) {
				if path != fixablePath {
					return "", fmt.Errorf("unexpected fix text read for %q", path)
				}
				return source.Text(), nil
			},
		},
	}
	result, err := RunPipeline(context.Background(), NewAutofixRequest(
		pipelineTestProvider(generation, nil),
		ObservationPolicy{Demand: ArtifactDemand{Native: rule.EditDemandAutofix}},
		autofixPolicyForTest(1, AutofixPolicy{}),
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	applied, ok := result.AppliedFixes()
	changes := applied.FinalChanges
	if !ok || len(changes) != 1 || changes[0].Path != fixablePath {
		t.Fatalf("planned changes = %+v, planned=%v", changes, ok)
	}
}

func TestPlanFixesIsPureAndDeterministic(t *testing.T) {
	unsortedFixes := []rule.RuleFix{
		{Range: core.NewTextRange(1, 2), Text: "Z"},
		{Range: core.NewTextRange(0, 1), Text: "A"},
	}
	diagnostics := []rule.RuleDiagnostic{
		{
			FilePath: "z.ts",
			Range:    core.NewTextRange(0, 1),
			FixesPtr: &[]rule.RuleFix{{Range: core.NewTextRange(0, 1), Text: "Z"}},
		},
		{
			FilePath: "a.ts",
			Range:    core.NewTextRange(0, 2),
			FixesPtr: &unsortedFixes,
		},
	}
	changes, err := planFixes(diagnostics, fixTextSnapshot{"z.ts": "z", "a.ts": "ab"})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0].Path != "a.ts" || changes[0].After != "AZ" || changes[1].Path != "z.ts" {
		t.Fatalf("changes = %+v", changes)
	}
	if unsortedFixes[0].Range.Pos() != 1 || unsortedFixes[1].Range.Pos() != 0 {
		t.Fatalf("planner mutated caller-owned fixes: %+v", unsortedFixes)
	}
}
