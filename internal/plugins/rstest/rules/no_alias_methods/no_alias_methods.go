package no_alias_methods

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_alias_methods"
)

var NoAliasMethodsRule = shared.NewRule(shared.Config{
	Name: "rstest/no-alias-methods",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *testFramework.MemberEntry {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.MatcherEntry == nil {
				return nil
			}
			return parsed.MatcherEntry
		}}
	},
})
