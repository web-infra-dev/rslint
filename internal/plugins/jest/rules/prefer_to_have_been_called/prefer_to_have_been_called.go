package prefer_to_have_been_called

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_have_been_called"
)

var PreferToHaveBeenCalledRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-to-have-been-called",
	Prepare: func(ctx rule.RuleContext) func(*ast.Node) []shared.Assertion {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return func(node *ast.Node) []shared.Assertion {
			parsed := analysis.ParseExpectCall(node)
			// An outer call on the matcher result parses to the same matcher.
			if parsed == nil || parsed.MatcherEntry == nil || testFramework.InvokedAccessorCall(parsed.MatcherEntry) != node {
				return nil
			}
			var not *testFramework.MemberEntry
			for i := range parsed.ModifierEntries {
				if parsed.ModifierEntries[i].Name == "not" {
					not = &parsed.ModifierEntries[i]
					break
				}
			}
			return []shared.Assertion{{
				Matcher: *parsed.MatcherEntry,
				Not:     not,
				CanFix:  true,
			}}
		}
	},
})
