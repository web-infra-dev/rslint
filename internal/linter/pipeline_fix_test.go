package linter

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"

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
