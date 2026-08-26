package no_identical_title

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_identical_title"
)

var NoIdenticalTitleRule = shared.NewRule(shared.Config{
	Name: "rstest/no-identical-title",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &shared.ParsedCall{
					Call:          &parsed.ParsedCall,
					Parameterized: parsed.IsParameterized(),
				}
			},
		}
	},
})
