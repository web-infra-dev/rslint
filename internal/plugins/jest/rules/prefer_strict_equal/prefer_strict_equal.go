package prefer_strict_equal

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Message Builders

func buildUseToStrictEqualErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToStrictEqual",
		Description: "Use `toStrictEqual()` instead",
	}
}

func buildSuggestReplaceWithStrictEqualErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestReplaceWithStrictEqual",
		Description: "Replace with `toStrictEqual()`",
	}
}

var PreferStrictEqualRule = rule.Rule{
	Name:   "jest/prefer-strict-equal",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
					return
				}

				MemberEntries := jestFnCall.MemberEntries
				if len(MemberEntries) == 0 {
					return
				}

				for _, memberEntry := range MemberEntries {
					if memberEntry.Name != "toEqual" {
						continue
					}

					fixRange, fixText, ok := testFramework.AccessorReplacement(ctx.SourceFile, memberEntry.Node, "toStrictEqual")
					if !ok {
						continue
					}

					ctx.ReportNodeWithSuggestions(
						memberEntry.Node,
						buildUseToStrictEqualErrorMessage(),
						rule.RuleSuggestion{
							Message: buildSuggestReplaceWithStrictEqualErrorMessage(),
							FixesArr: []rule.RuleFix{
								{
									Range: fixRange,
									Text:  fixText,
								},
							},
						},
					)
				}
			},
		}
	},
}
