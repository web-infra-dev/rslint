package no_conditional_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_expect"
)

var NoConditionalExpectRule = shared.NewRule(shared.Config{
	Name: "jest/no-conditional-expect",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		callbacks := jestUtils.CollectJestTestCallbacks(ctx)
		return shared.Runtime{
			TestCallbackFunctions: callbacks.Functions,
			IsTestCall: func(node *ast.Node) bool {
				parsed := callbacks.ParseFnCall(node)
				return parsed != nil && parsed.Kind == jestUtils.JestFnTypeTest
			},
			IsExpectCall: func(node *ast.Node) bool {
				parsed := callbacks.ParseFnCall(node)
				return parsed != nil && parsed.Kind == jestUtils.JestFnTypeExpect
			},
		}
	},
})
