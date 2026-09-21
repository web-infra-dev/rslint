package prefer_expect_resolves

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstest "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_expect_resolves"
)

var PreferExpectResolvesRule = shared.NewRule(shared.Config{
	Name:              "rstest/prefer-expect-resolves",
	Message:           rule.RuleMessage{Id: "expectResolves", Description: "Use `expect().resolves` instead"},
	RequirePromise:    true,
	ConservativeEdits: true,
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) *shared.ExpectCall {
		analysis := rstest.GetRstestCallAnalysis(ctx)
		return func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCallThroughTransparentExpressions(node)
			if parsed == nil || parsed.Entry != rstest.RstestExpectEntryCall ||
				parsed.Reason != rstest.RstestExpectParseReasonNone || len(parsed.Matchers) != 1 ||
				parsed.Matchers[0].Kind != rstest.RstestExpectMatcherCall || parsed.PromiseModifierEntry() != nil ||
				rstest.RSTEST_BROWSER_ELEMENT_MATCHERS[parsed.Matcher] || parsed.Matcher == "toHaveTitle" || parsed.Matcher == "toHaveURL" {
				return nil
			}
			matcher := rstest.MatcherCall(parsed.MatcherEntry)
			if matcher != node || parsed.MemberEntries[len(parsed.MemberEntries)-1].Node != parsed.MatcherEntry.Node {
				return nil
			}
			return &shared.ExpectCall{Head: parsed.Head, Matcher: matcher, AddResolves: true,
				Editable: shared.BuiltinValueMatcher(parsed.Matcher) && !analysis.IsExpectMatcherOverridden(parsed.Matcher)}
		}
	},
})
