package require_top_level_describe

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_top_level_describe"
)

var RequireTopLevelDescribeRule = shared.NewRule(shared.Config{
	Name: "jest/require-top-level-describe",
	Messages: shared.Messages{
		UnexpectedTestCase: "All test cases must be wrapped in a describe block.",
		UnexpectedHook:     "All hooks must be wrapped in a describe block.",
	},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := utils.ParseJestFnCall(node, ctx)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
		}
	},
})
