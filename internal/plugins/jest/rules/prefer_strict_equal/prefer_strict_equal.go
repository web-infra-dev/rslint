package prefer_strict_equal

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_strict_equal"
)

var PreferStrictEqualRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-strict-equal",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := utils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != utils.JestFnTypeExpect || parsed.MatcherEntry == nil {
				return nil
			}
			return &shared.ExpectCall{
				Matcher:      parsed.Matcher,
				MatcherEntry: *parsed.MatcherEntry,
			}
		}}
	},
})
