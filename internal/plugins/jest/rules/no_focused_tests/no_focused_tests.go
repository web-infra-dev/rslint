package no_focused_tests

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builders

func buildErrorFocusedTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "focusedTest",
		Description: "Unexpected focused test",
	}
}

func buildErrorSuggestRemoveFocusMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemoveFocus",
		Description: "Suggest removing focus from test",
	}
}

func buildFix() {}

var NoFocusedTestsRule = rule.Rule{
	Name: "jest/no-focused-tests",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					(jestFnCall.Kind != utils.JestFnTypeDescribe &&
						jestFnCall.Kind != utils.JestFnTypeTest) {
					return
				}

				if strings.HasPrefix(jestFnCall.Name, "f") {
					ctx.ReportNodeWithSuggestions(
						node,
						buildErrorFocusedTestMessage(),
						rule.RuleSuggestion{
							Message: buildErrorSuggestRemoveFocusMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, node, string(jestFnCall.Name[1:])),
							},
						},
					)
				} else if slices.Contains(jestFnCall.Members, "only") {
					ctx.ReportNodeWithSuggestions(
						node,
						buildErrorFocusedTestMessage(),
						rule.RuleSuggestion{
							Message: buildErrorSuggestRemoveFocusMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, node, jestFnCall.Name),
							},
						},
					)
				}
			},
		}
	},
}
