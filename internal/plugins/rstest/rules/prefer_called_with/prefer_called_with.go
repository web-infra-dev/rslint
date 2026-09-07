package prefer_called_with

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_called_with"
)

var PreferCalledWithRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-called-with",
	Replacements: map[string]string{
		"toBeCalled":           "toBeCalledWith",
		"toHaveBeenCalled":     "toHaveBeenCalledWith",
		"toHaveBeenCalledOnce": "toHaveBeenCalledExactlyOnceWith",
	},
	Autofix: true,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.Head == nil || len(parsed.Matchers) == 0 ||
				parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
				return nil
			}
			return &shared.ExpectCall{
				Matcher:      parsed.Matcher,
				MatcherEntry: parsed.Matchers[0].Entry,
				Modifiers:    parsed.Modifiers,
			}
		}}
	},
})
