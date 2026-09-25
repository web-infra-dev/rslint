package no_unneeded_async_expect_function

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unneeded_async_expect_function"
)

func hasPromiseExpectModifier(jestFnCall *jestUtils.ParsedJestFnCall) bool {
	return slices.Contains(jestFnCall.Modifiers, "resolves") ||
		slices.Contains(jestFnCall.Modifiers, "rejects")
}

var NoUnneededAsyncExpectFunctionRule = shared.NewRule(shared.Config{
	Name: "jest/no-unneeded-async-expect-function",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{ParseExpect: func(node *ast.Node) *ast.Node {
			jestFnCall := analysis.ParseExpectCall(node)
			if jestFnCall == nil ||
				!hasPromiseExpectModifier(jestFnCall) {
				return nil
			}

			headCall := jestFnCall.Head.Local.Node.Parent
			if headCall == nil || headCall.Kind != ast.KindCallExpression {
				return nil
			}
			return headCall
		}}
	},
	// The fix is withheld when the unwrapped call would not mean the same
	// thing; the wrapper is still reported.
	Report: func(ctx rule.RuleContext, match shared.Match) {
		ctx.ReportNodeWithDeferredFixes(
			match.Wrapper,
			shared.Message,
			func() []rule.RuleFix {
				if fix := shared.UnwrapFix(ctx, match); fix != nil {
					return []rule.RuleFix{*fix}
				}
				return nil
			},
		)
	},
})
