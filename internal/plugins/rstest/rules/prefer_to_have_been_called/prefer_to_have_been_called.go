package prefer_to_have_been_called

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_have_been_called"
)

var PreferToHaveBeenCalledRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-to-have-been-called",
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) []shared.Assertion {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return func(node *ast.Node) []shared.Assertion {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.Reason != rstestUtils.RstestExpectParseReasonNone || parsed.Head == nil || parsed.Entry == rstestUtils.RstestExpectEntryElement {
				return nil
			}
			var not *testFramework.MemberEntry
			for i := range parsed.ModifierEntries {
				if parsed.ModifierEntries[i].Name == "not" {
					not = &parsed.ModifierEntries[i]
				}
			}
			var assertions []shared.Assertion
			for _, matcher := range parsed.Matchers {
				if matcher.Kind != rstestUtils.RstestExpectMatcherCall || (matcher.Name != "toHaveBeenCalledTimes" && matcher.Name != "toBeCalledTimes") {
					continue
				}
				// Negation persists across Chai assertions, so changing it can affect sibling matchers.
				assertions = append(assertions, shared.Assertion{
					Matcher: matcher.Entry,
					Not:     not,
					CanFix: parsed.Entry != rstestUtils.RstestExpectEntryPoll &&
						len(parsed.Matchers) == 1 &&
						parsed.Expression == testFramework.InvokedAccessorCall(&matcher.Entry) &&
						isDiscardedAssertion(parsed.Expression),
				})
			}
			return assertions
		}
	},
})

func isDiscardedAssertion(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAwaitExpression:
			continue
		case ast.KindExpressionStatement:
			return true
		default:
			return false
		}
	}
	return false
}
