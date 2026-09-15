package prefer_comparison_matcher

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_comparison_matcher"
)

var PreferComparisonMatcherRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-comparison-matcher",
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) *shared.ExpectCall {
		return func(node *ast.Node) *shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeExpect || parsed.MatcherEntry == nil || parsed.MatcherEntry.Call != node {
				return nil
			}
			head := ast.WalkUpParenthesizedExpressions(parsed.Head.Local.Node.Parent)
			if head == nil || head.Kind != ast.KindCallExpression {
				return nil
			}
			return &shared.ExpectCall{
				HeadCall: head, MatcherCall: node, MatcherEntry: *parsed.MatcherEntry,
				Matcher: parsed.Matcher, Modifiers: parsed.ModifierEntries,
			}
		}
	},
	BuildFixes: func(ctx rule.RuleContext, match shared.Match) []rule.RuleFix {
		_, accessor := testFramework.AccessorReceiverAndParent(&match.Expect.MatcherEntry)
		if accessor == nil {
			return nil
		}
		// Replacing across receiver parentheses would leave unmatched opening parentheses.
		for receiver := accessor.Expression(); receiver != match.Expect.HeadCall; receiver = receiver.Expression() {
			if receiver == nil || receiver.Kind == ast.KindParenthesizedExpression {
				return nil
			}
		}
		modifier := ""
		if entries := match.Expect.Modifiers; len(entries) > 0 && entries[0].Name != "not" {
			modifier = "." + entries[0].Name
		}
		return append(shared.OperandFixes(ctx, match), rule.RuleFixReplaceRange(
			core.NewTextRange(match.Expect.HeadCall.End(), accessor.End()), modifier+"."+match.Matcher,
		))
	},
})
