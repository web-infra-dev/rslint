package valid_expect_in_promise

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect_in_promise"
)

var ValidExpectInPromiseRule = shared.NewRule(shared.Config{
	Name:               "jest/valid-expect-in-promise",
	MessageDescription: "This promise should either be returned or awaited to ensure the expects in its chain are called",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		callbacks := analysis.Callbacks()
		return shared.Runtime{
			TestCallbackFunctions:    callbacks.Functions,
			IgnoredCallbackFunctions: callbacks.IgnoredDone,
			IsAssertionCall: func(node *ast.Node) bool {
				return analysis.ParseExpectCall(node) != nil
			},
			IsAsyncAssertionSink: func(callNode, value *ast.Node) bool {
				if callNode == nil || callNode.Kind != ast.KindCallExpression {
					return false
				}
				top := shared.TopMostCallExpressionOnCallee(callNode)
				parsed := analysis.ParseExpectCall(top)
				if parsed == nil ||
					(!slices.Contains(parsed.Modifiers, "resolves") &&
						!slices.Contains(parsed.Modifiers, "rejects")) {
					return false
				}
				head := shared.LeftMostCallExpression(top)
				if head == nil || head.AsCallExpression().Arguments == nil ||
					len(head.AsCallExpression().Arguments.Nodes) == 0 {
					return false
				}
				return ast.SkipParentheses(head.AsCallExpression().Arguments.Nodes[0]) ==
					ast.SkipParentheses(value)
			},
		}
	},
})
