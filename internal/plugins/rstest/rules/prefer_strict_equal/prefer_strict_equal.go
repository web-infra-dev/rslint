package prefer_strict_equal

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_strict_equal"
)

var PreferStrictEqualRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-strict-equal",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				len(parsed.Matchers) == 0 ||
				parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
				return nil
			}

			matcher := parsed.Matchers[0]
			return &shared.ExpectCall{
				Matcher:      matcher.Name,
				MatcherEntry: matcher.Entry,
			}
		}}
	},
})
