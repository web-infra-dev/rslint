package prefer_hooks_in_order

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_hooks_in_order"
)

var PreferHooksInOrderRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-hooks-in-order",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := jestUtils.ParseJestFnCall(node, ctx)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
		}
	},
})
