package no_conditional_expect

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "jest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		callbacks := analysis.Callbacks()
		return shared.Runtime{
			TestCallbackFunctions: callbacks.Functions,
			IsTestCall: func(node *ast.Node) bool {
				return analysis.ParseTestCall(node) != nil
			},
			IsExpectCall: func(node *ast.Node) bool {
				return analysis.ParseExpectCall(node) != nil
			},
		}
	},
})
