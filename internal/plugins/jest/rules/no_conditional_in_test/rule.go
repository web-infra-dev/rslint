package no_conditional_in

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_in_test"
)

var NoConditionalInTestRule = shared.NewRule(shared.Config{
	Name: "jest/no-conditional-in-test",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			IsTestCall: func(node *ast.Node) bool {
				return analysis.ParseTestCall(node) != nil
			},
		}
	},
})
