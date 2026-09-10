package prefer_to_contain

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_contain"
)

func buildFixes(ctx rule.RuleContext, matched shared.Match) []rule.RuleFix {
	chain := make([]string, 0, len(matched.Expect.Modifiers)+1)
	for _, modifier := range matched.Expect.Modifiers {
		if modifier.Name != "not" {
			chain = append(chain, modifier.Name)
		}
	}
	if matched.ShouldHaveNot {
		chain = append(chain, "not")
	}
	chain = append(chain, "toContain")

	return []rule.RuleFix{
		rule.RuleFixReplace(ctx.SourceFile, matched.IncludesCall, utils.TrimmedNodeText(ctx.SourceFile, matched.Receiver)),
		rule.RuleFixReplaceRange(core.NewTextRange(matched.Expect.HeadCall.End(), matched.Expect.MatcherCall.AsCallExpression().Expression.End()), "."+strings.Join(chain, ".")),
		rule.RuleFixReplace(ctx.SourceFile, matched.MatcherValue, utils.TrimmedNodeText(ctx.SourceFile, matched.Item)),
	}
}

var PreferToContainRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-to-contain",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeExpect || parsed.MatcherEntry == nil ||
				testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
				return nil
			}
			matcherCall := testFramework.InvokedAccessorCall(parsed.MatcherEntry)
			if matcherCall == nil {
				return nil
			}
			headCall := parsed.Head.Local.Node.Parent
			if headCall == nil || headCall.Kind != ast.KindCallExpression {
				return nil
			}
			return &shared.ExpectCall{
				HeadCall: headCall, MatcherCall: matcherCall, MatcherEntry: *parsed.MatcherEntry,
				Matcher: parsed.Matcher, Modifiers: parsed.ModifierEntries,
			}
		}}
	},
	BuildFixes: buildFixes,
})
