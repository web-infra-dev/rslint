package require_to_throw_message

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_to_throw_message"
)

var RequireToThrowMessageRule = shared.NewRule(shared.Config{
	Name: "rstest/require-to-throw-message",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			ParseExpectCall: func(node *ast.Node) *shared.ExpectCall {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil ||
					parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
					parsed.Head == nil ||
					parsed.MatcherEntry == nil ||
					len(parsed.Matchers) == 0 ||
					parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
					return nil
				}
				matcherCall := rstestUtils.MatcherCall(parsed.MatcherEntry)
				if matcherCall == nil {
					return nil
				}

				return &shared.ExpectCall{
					Matcher:      parsed.Matcher,
					MatcherEntry: parsed.MatcherEntry,
					Modifiers:    parsed.Modifiers,
					MatcherArgs:  matcherCall.Arguments(),
				}
			},
		}
	},
})
