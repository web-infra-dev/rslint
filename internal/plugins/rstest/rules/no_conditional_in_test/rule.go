package no_conditional_in

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_conditional_in_test"
)

var NoConditionalInTestRule = shared.NewRule(shared.Config{
	Name: "rstest/no-conditional-in-test",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			IsTestCall: func(node *ast.Node) bool {
				return analysis.ParseTestCall(node) != nil
			},
		}
	},
})
