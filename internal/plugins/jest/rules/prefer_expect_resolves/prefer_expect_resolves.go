package prefer_expect_resolves

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jest "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	framework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_expect_resolves"
)

var PreferExpectResolvesRule = shared.NewRule(shared.Config{
	Name:    "jest/prefer-expect-resolves",
	Message: rule.RuleMessage{Id: "expectResolves", Description: "Use `await expect(...).resolves instead"},
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) *shared.ExpectCall {
		return func(node *ast.Node) *shared.ExpectCall {
			parsed := jest.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jest.JestFnTypeExpect || parsed.MatcherEntry == nil {
				return nil
			}
			head := parsed.Head.Local.Node.Parent
			if head == nil || head.Kind != ast.KindCallExpression {
				return nil
			}
			matcher := framework.InvokedAccessorCall(parsed.MatcherEntry)
			if matcher == nil || matcher != node {
				return nil
			}
			return &shared.ExpectCall{
				Head: head, Matcher: matcher, Editable: true,
				AddResolves: !slices.Contains(parsed.Modifiers, "resolves") && !slices.Contains(parsed.Modifiers, "rejects"),
			}
		}
	},
})
