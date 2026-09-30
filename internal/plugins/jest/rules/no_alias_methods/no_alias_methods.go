package no_alias_methods

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_alias_methods"
)

var NoAliasMethodsRule = shared.NewRule(shared.Config{
	Name: "jest/no-alias-methods",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *testFramework.MemberEntry {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.MatcherEntry == nil {
				return nil
			}
			// Every call further along the chain, as in
			// `expect(a).toBeCalled().then()`, parses to the same matcher.
			// Only the call that invokes the matcher owns it.
			if testFramework.InvokedAccessorCall(parsed.MatcherEntry) != node {
				return nil
			}
			return parsed.MatcherEntry
		}}
	},
})
