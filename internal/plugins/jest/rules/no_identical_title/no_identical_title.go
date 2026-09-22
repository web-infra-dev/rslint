package no_identical_title

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_identical_title"
)

var NoIdenticalTitleRule = shared.NewRule(shared.Config{
	Name: "jest/no-identical-title",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ParsedCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return nil
				}
				return &shared.ParsedCall{
					Call:          &parsed.ParsedCall,
					Parameterized: slices.Contains(parsed.Members, "each"),
				}
			},
		}
	},
})
