package require_to_throw_message

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_to_throw_message"
)

var RequireToThrowMessageRule = shared.NewRule(shared.Config{
	Name: "jest/require-to-throw-message",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := utils.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			ParseExpectCall: func(node *ast.Node) *shared.ExpectCall {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil || parsed.MatcherEntry == nil {
					return nil
				}

				return &shared.ExpectCall{
					Matcher:      parsed.Matcher,
					MatcherEntry: parsed.MatcherEntry,
					Modifiers:    parsed.Modifiers,
					MatcherArgs:  node.Arguments(),
				}
			},
		}
	},
})
