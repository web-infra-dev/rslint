package prefer_equality_matcher

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_equality_matcher"
)

var PreferEqualityMatcherRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-equality-matcher",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{
			Parse: func(node *ast.Node) *shared.ExpectCall {
				parsed := jestUtils.ParseJestFnCall(node, ctx)
				if parsed == nil ||
					parsed.Kind != jestUtils.JestFnTypeExpect ||
					parsed.MatcherEntry == nil {
					return nil
				}

				headCall := parsed.Head.Local.Node.Parent
				if headCall == nil || headCall.Kind != ast.KindCallExpression {
					return nil
				}

				return &shared.ExpectCall{
					HeadCall:     headCall,
					MatcherCall:  node,
					MatcherEntry: *parsed.MatcherEntry,
					Matcher:      parsed.Matcher,
					Modifiers:    parsed.ModifierEntries,
				}
			},
		}
	},
	BuildFixes: func(ctx rule.RuleContext, match shared.Match, equalityMatcher string) []rule.RuleFix {
		_, matcherAccessor := testFramework.AccessorReceiverAndParent(&match.Expect.MatcherEntry)
		if matcherAccessor == nil {
			return nil
		}

		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, match.Comparison, match.LeftText),
			rule.RuleFixReplaceRange(
				core.NewTextRange(match.Expect.HeadCall.End(), matcherAccessor.End()),
				match.ModifierText+"."+equalityMatcher,
			),
			rule.RuleFixReplace(ctx.SourceFile, match.MatcherArgument, match.RightText),
		}
	},
})
