package require_number_to_fixed_digits_argument

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "require-number-to-fixed-digits-argument"

var missingDigitsMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Missing the digits argument.",
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/require-number-to-fixed-digits-argument.md
var RequireNumberToFixedDigitsArgumentRule = rule.Rule{
	Name:   "unicorn/require-number-to-fixed-digits-argument",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		zeroArguments := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "toFixed",
					ArgumentsLength:     &zeroArguments,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				// Upstream intentionally exempts only a directly constructed
				// receiver. Calls on variables or factory results remain reportable,
				// even when they come from a Number-like library.
				object := ast.SkipParentheses(call.Object)
				if object != nil && object.Kind == ast.KindNewExpression {
					return
				}

				opening, closing, ok := unicornutil.CallExpressionParenthesesRange(ctx.SourceFile, node)
				if !ok {
					return
				}

				reportRange := core.NewTextRange(opening.Pos(), closing.End())
				insertRange := core.NewTextRange(closing.Pos(), closing.Pos())
				ctx.ReportRangeWithFixes(
					reportRange,
					missingDigitsMessage,
					rule.RuleFixReplaceRange(insertRange, "0"),
				)
			},
		}
	},
}
