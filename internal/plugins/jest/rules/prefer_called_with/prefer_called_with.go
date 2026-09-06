package prefer_called_with

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_called_with"
)

var PreferCalledWithRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-called-with",
	Replacements: map[string]string{
		"toBeCalled":       "toBeCalledWith",
		"toHaveBeenCalled": "toHaveBeenCalledWith",
	},
	Autofix: false,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeExpect {
				return nil
			}
			call := &shared.ExpectCall{Matcher: parsed.Matcher, Modifiers: parsed.Modifiers}
			if parsed.MatcherEntry != nil {
				call.MatcherEntry = *parsed.MatcherEntry
			}
			return call
		}}
	},
})
