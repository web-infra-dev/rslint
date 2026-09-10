// Ported from eslint-plugin-unicorn v74.0.0 (MIT); see LICENSE.
package no_unreadable_iife

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUnreadableIifeRule = rule.Rule{
	Name:   "unicorn/no-unreadable-iife",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
				if callee == nil || callee.Kind != ast.KindArrowFunction {
					return
				}
				body := callee.AsArrowFunction().Body
				if body.Kind != ast.KindParenthesizedExpression {
					return
				}
				bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
				ctx.ReportRangeWithDeferredSuggestions(bodyRange, rule.RuleMessage{
					Id:          "no-unreadable-iife",
					Description: "IIFE with parenthesized arrow function body is considered unreadable.",
				}, func() []rule.RuleSuggestion {
					expression := utils.ESTreeRuntimeExpression(body)
					expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, expression)
					for _, comment := range utils.CommentsInSpan(ctx.Comments.All(), bodyRange.Pos(), bodyRange.End()) {
						if comment.Pos() < expressionRange.Pos() || comment.End() > expressionRange.End() {
							return nil
						}
					}
					return []rule.RuleSuggestion{{
						Message:  rule.RuleMessage{Id: "suggestion", Description: "Use a block statement body."},
						FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(bodyRange, "{ return "+ctx.SourceFile.Text()[expressionRange.Pos():expressionRange.End()]+"; }")},
					}}
				})
			},
		}
	},
}
