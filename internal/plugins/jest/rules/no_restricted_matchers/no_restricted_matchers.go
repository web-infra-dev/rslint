package no_restricted_matchers

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_restricted_matchers"
)

var NoRestrictedMatchersRule = shared.NewRule(shared.Config{
	Name:      "jest/no-restricted-matchers",
	Modifiers: jestUtils.EXPECT_MODIFIER_NAMES,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ExpectChain {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return nil
				}
				return &shared.ExpectChain{Names: parsed.Members, Entries: parsed.MemberEntries}
			},
		}
	},
})
