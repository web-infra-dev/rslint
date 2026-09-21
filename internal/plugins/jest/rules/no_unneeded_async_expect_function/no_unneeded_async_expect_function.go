package no_unneeded_async_expect_function

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unneeded_async_expect_function"
)

func hasPromiseExpectModifier(jestFnCall *jestUtils.ParsedJestFnCall) bool {
	return slices.Contains(jestFnCall.Modifiers, "resolves") ||
		slices.Contains(jestFnCall.Modifiers, "rejects")
}

var NoUnneededAsyncExpectFunctionRule = shared.NewRule(shared.Config{
	Name: "jest/no-unneeded-async-expect-function",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{ParseExpect: func(node *ast.Node) *ast.Node {
			jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
			if jestFnCall == nil ||
				jestFnCall.Kind != jestUtils.JestFnTypeExpect ||
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
	Report: func(ctx rule.RuleContext, match shared.Match) {
		replacement := rslintUtils.TrimmedNodeText(ctx.SourceFile, match.Awaited)
		ctx.ReportNodeWithFixes(
			match.Wrapper,
			shared.Message,
			rule.RuleFixReplace(ctx.SourceFile, match.Wrapper, replacement),
		)
	},
})
