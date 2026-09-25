package prefer_ending_with_an_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_ending_with_an_expect"
)

// jestCallbackArgument returns the second argument when it is a function
// literal. Jest's only other overload puts a numeric timeout after the
// callback, so the callback never moves off this position.
func jestCallbackArgument(call *ast.CallExpression) *ast.Node {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return nil
	}
	callback := ast.SkipParentheses(call.Arguments.Nodes[1])
	if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		return nil
	}
	return callback
}

var PreferEndingWithAnExpectRule = shared.NewRule(shared.Config{
	Name:                       "jest/prefer-ending-with-an-expect",
	DefaultAssertFunctionNames: []string{"expect"},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			ClassifyTest: func(node *ast.Node) shared.TestBlock {
				return shared.TestBlock{
					IsTest: analysis.ParseTestCall(node) != nil,
				}
			},
			Callback: jestCallbackArgument,
			IsAssertion: func(node *ast.Node) bool {
				return analysis.ParseExpectCall(node) != nil
			},
		}
	},
})
