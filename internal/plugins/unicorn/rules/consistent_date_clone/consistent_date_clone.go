package consistent_date_clone

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "consistent-date-clone/error"

var unnecessaryGetTimeMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Unnecessary `.getTime()` call.",
}

func isSingleArgumentDateConstruction(node *ast.Node) bool {
	if node == nil || !ast.IsNewExpression(node) {
		return false
	}

	arguments := node.Arguments()
	if len(arguments) != 1 || ast.IsSpreadElement(arguments[0]) {
		return false
	}

	callee := ast.SkipParentheses(node.Expression())
	return callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Date"
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/consistent-date-clone.md
var ConsistentDateCloneRule = rule.Rule{
	Name:   "unicorn/consistent-date-clone",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		zeroArguments := 0
		return rule.RuleListeners{
			// Exit order reports a nested clone before its enclosing clone, matching
			// the source-ordered diagnostics exposed by ESLint.
			rule.ListenerOnExit(ast.KindNewExpression): func(node *ast.Node) {
				if !isSingleArgumentDateConstruction(node) {
					return
				}

				argument := ast.SkipParentheses(node.Arguments()[0])
				if argument == nil || ast.IsOptionalChain(argument) {
					return
				}
				getTimeCall, ok := unicornutil.MatchDotMethodCall(argument, unicornutil.DotMethodCallOptions{
					Method:          "getTime",
					ArgumentsLength: &zeroArguments,
				})
				if !ok {
					return
				}

				reportRange := utils.TrimNodeTextRange(ctx.SourceFile, getTimeCall.Property).
					WithEnd(getTimeCall.Call.End())
				ctx.ReportRangeWithDeferredFixes(reportRange, unnecessaryGetTimeMessage, func() []rule.RuleFix {
					return unicornutil.RemoveMethodCallFixes(getTimeCall)
				})
			},
		}
	},
}
