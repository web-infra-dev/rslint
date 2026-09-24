package no_hooks

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_hooks"
)

var NoHooksRule = shared.NewRule(shared.Config{
	Name: "jest/no-hooks",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &parsed.ParsedCall
			},
		}
	},
})
