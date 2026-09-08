package prefer_equality_matcher

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestShouldAddNotTruthTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		comparisonNegated bool
		matcherValue      bool
		hasNot            bool
		want              bool
	}{
		{comparisonNegated: false, matcherValue: true, hasNot: false, want: false},
		{comparisonNegated: false, matcherValue: false, hasNot: false, want: true},
		{comparisonNegated: true, matcherValue: true, hasNot: false, want: true},
		{comparisonNegated: true, matcherValue: false, hasNot: false, want: false},
		{comparisonNegated: false, matcherValue: true, hasNot: true, want: true},
		{comparisonNegated: false, matcherValue: false, hasNot: true, want: false},
		{comparisonNegated: true, matcherValue: true, hasNot: true, want: false},
		{comparisonNegated: true, matcherValue: false, hasNot: true, want: true},
	}

	for _, test := range tests {
		if got := shouldAddNot(test.comparisonNegated, test.matcherValue, test.hasNot); got != test.want {
			t.Errorf("shouldAddNot(%t, %t, %t) = %t, want %t", test.comparisonNegated, test.matcherValue, test.hasNot, got, test.want)
		}
	}
}

func TestEqualityMatcherSuggestionOrder(t *testing.T) {
	t.Parallel()

	want := []string{"toBe", "toEqual", "toStrictEqual"}
	if len(equalityMatchers) != len(want) {
		t.Fatalf("equalityMatchers has %d entries, want %d", len(equalityMatchers), len(want))
	}
	for index := range want {
		if equalityMatchers[index] != want[index] {
			t.Errorf("equalityMatchers[%d] = %q, want %q", index, equalityMatchers[index], want[index])
		}
	}
}

func TestNewRuleBuildsThreeNonOverlappingSuggestions(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, `expect(a === b).not.toBe(false);`, core.ScriptKindTS)

	var calls []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			calls = append(calls, node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if len(calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(calls))
	}

	outer, head := calls[0], calls[1]
	entries := testFramework.GetMemberEntries(outer)
	if len(entries) != 3 {
		t.Fatalf("member count = %d, want 3", len(entries))
	}

	buildCount := 0
	r := NewRule(Config{
		Name: "test/prefer-equality-matcher",
		Prepare: func(rule.RuleContext) Runtime {
			return Runtime{Parse: func(node *ast.Node) *ExpectCall {
				if node != outer {
					return nil
				}
				return &ExpectCall{
					HeadCall:     head,
					MatcherCall:  outer,
					MatcherEntry: entries[2],
					Matcher:      entries[2].Name,
					Modifiers:    entries[1:2],
				}
			}}
		},
		BuildFixes: func(_ rule.RuleContext, match Match, equalityMatcher string) []rule.RuleFix {
			buildCount++
			if match.ShouldHaveNot {
				t.Fatal("existing not should be removed for !== false")
			}
			if match.ModifierText != "" {
				t.Fatalf("modifier text = %q, want empty", match.ModifierText)
			}
			return []rule.RuleFix{
				rule.RuleFixReplaceRange(core.NewTextRange(0, 1), "left"),
				rule.RuleFixReplaceRange(core.NewTextRange(2, 3), equalityMatcher),
				rule.RuleFixReplaceRange(core.NewTextRange(4, 5), "right"),
			}
		},
	})

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(
			r.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)
		r.Run(ctx, nil)[ast.KindCallExpression](outer)
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	if len(diagnosticsOnly) != 1 || diagnosticsOnly[0].Suggestions != nil || buildCount != 0 {
		t.Fatalf("diagnostics-only run = %#v, build count = %d", diagnosticsOnly, buildCount)
	}
	autofixOnly := run(rule.EditDemandAutofix)
	if len(autofixOnly) != 1 || autofixOnly[0].Suggestions != nil || autofixOnly[0].FixesPtr != nil || buildCount != 0 {
		t.Fatalf("autofix-only run = %#v, build count = %d", autofixOnly, buildCount)
	}
	diagnostics := run(rule.EditDemandSuggestion)

	if len(diagnostics) != 1 || diagnostics[0].Suggestions == nil {
		t.Fatalf("diagnostics = %#v, want one diagnostic with suggestions", diagnostics)
	}
	suggestions := *diagnostics[0].Suggestions
	if len(suggestions) != 3 {
		t.Fatalf("suggestions = %d, want 3", len(suggestions))
	}
	for index, matcher := range []string{"toBe", "toEqual", "toStrictEqual"} {
		if suggestions[index].Message.Description != "Use `"+matcher+"`" {
			t.Errorf("suggestion %d description = %q", index, suggestions[index].Message.Description)
		}
		fixes := suggestions[index].FixesArr
		if len(fixes) != 3 || fixes[1].Text != matcher {
			t.Errorf("suggestion %d fixes = %#v", index, fixes)
		}
		for fixIndex := 1; fixIndex < len(fixes); fixIndex++ {
			if fixes[fixIndex-1].Range.End() > fixes[fixIndex].Range.Pos() {
				t.Errorf("suggestion %d fixes %d and %d overlap: %#v", index, fixIndex-1, fixIndex, fixes)
			}
		}
	}
	if buildCount != 3 {
		t.Errorf("suggestion build count = %d, want 3", buildCount)
	}
	allEdits := run(rule.EditDemandAll)
	if len(allEdits) != 1 || allEdits[0].Suggestions == nil || len(*allEdits[0].Suggestions) != 3 || allEdits[0].FixesPtr != nil {
		t.Fatalf("all-edits run = %#v, want suggestions only", allEdits)
	}
	if buildCount != 6 {
		t.Errorf("all-edits build count = %d, want 6 total", buildCount)
	}
}
