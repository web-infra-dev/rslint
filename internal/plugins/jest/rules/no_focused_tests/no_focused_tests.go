package no_focused_tests

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
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
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					(jestFnCall.Kind != jestUtils.JestFnTypeDescribe &&
						jestFnCall.Kind != jestUtils.JestFnTypeTest) {
					return
				}

				if strings.HasPrefix(jestFnCall.Name, "f") {
					callExpr := node.AsCallExpression()
					if callExpr == nil {
						return
					}

					callee := ast.SkipParentheses(callExpr.Expression)
					if callee == nil {
						return
					}

					calleeRange := rslintUtils.TrimNodeTextRange(ctx.SourceFile, callee)
					if jestFnCall.Head.Type == jestUtils.JEST_IMPORT_MODE && jestFnCall.Name != jestFnCall.Head.Local.Value {
						ctx.ReportNode(callee, buildErrorFocusedTestMessage())
					} else {
						ctx.ReportNodeWithSuggestions(
							callee,
							buildErrorFocusedTestMessage(),
							rule.RuleSuggestion{
								Message: buildErrorSuggestRemoveFocusMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixRemoveRange(core.NewTextRange(calleeRange.Pos(), calleeRange.Pos()+1)),
								},
							},
						)
					}
				} else {
					idx := slices.IndexFunc(jestFnCall.MemberEntries, func(entry jestUtils.ParsedJestFnMemberEntry) bool {
						return entry.Name == "only"
					})
					if idx >= 0 {
						entry := jestFnCall.MemberEntries[idx]
						startRange := entry.Node.Loc.Pos() - 1
						endRange := entry.Node.Loc.End()
						if entry.Node.Kind != ast.KindIdentifier {
							endRange = entry.Node.End() + 1
						}

						ctx.ReportNodeWithSuggestions(
							entry.Node,
							buildErrorFocusedTestMessage(),
							rule.RuleSuggestion{
								Message: buildErrorSuggestRemoveFocusMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixRemoveRange(core.NewTextRange(startRange, endRange)),
								},
							},
						)
					}
				}
			},
		}
	},
}
