package no_disabled_tests

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_disabled_tests"
)

var NoDisabledTestsRule = shared.NewRule(shared.Config{
	Name: "rstest/no-disabled-tests",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &shared.ParsedCall{
					Call: &parsed.ParsedCall,
					// Semantic fields rather than Members, so aliases retain
					// modifiers consumed while resolving their declarations.
					HasSkip: parsed.Skipped,
					HasTodo: parsed.Todo,
				}
			},
		}
	},
	// `TestCall` accepts `(description, options, fn?)`.
	HasOptionsOverload: true,
})
