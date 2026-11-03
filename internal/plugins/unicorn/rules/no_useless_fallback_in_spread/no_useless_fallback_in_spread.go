// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_useless_fallback_in_spread

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUselessFallbackInSpreadRule = rule.Rule{
	Name:   "unicorn/no-useless-fallback-in-spread",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindObjectLiteralExpression: func(node *ast.Node) {
			if !ast.IsEmptyObjectLiteral(node) {
				return
			}
			logical := utils.ESTreeParent(node)
			if logical == nil || logical.Kind != ast.KindBinaryExpression {
				return
			}
			binary := logical.AsBinaryExpression()
			if (binary.OperatorToken.Kind != ast.KindBarBarToken && binary.OperatorToken.Kind != ast.KindQuestionQuestionToken) ||
				utils.ESTreeRuntimeExpression(binary.Right) != node {
				return
			}
			spread := utils.ESTreeParent(logical)
			if spread == nil || spread.Kind != ast.KindSpreadAssignment || spread.Parent.Kind != ast.KindObjectLiteralExpression {
				return
			}

			ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
				Id: "no-useless-fallback-in-spread", Description: "The empty object is useless.",
			}, func() []rule.RuleFix {
				// The raw left operand retains its parentheses and their comments.
				start, end := binary.Left.End(), logical.End()
				if utils.HasCommentInSpan(ctx.Comments.All(), start, end) {
					return nil
				}
				fixes := []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, end))}
				// Delete only the surrounding parentheses, preserving comments and
				// the left operand's grouping (in particular, sequence expressions).
				// The validated spread boundary lets us walk this wrapper chain once.
				for outer := logical.Parent; outer != spread; outer = outer.Parent {
					if ast.IsParenthesizedExpression(outer) {
						start := scanner.GetTokenPosOfNode(outer, ctx.SourceFile, false)
						fixes = append(fixes,
							rule.RuleFixRemoveRange(core.NewTextRange(start, start+1)),
							rule.RuleFixRemoveRange(core.NewTextRange(outer.End()-1, outer.End())))
					}
				}
				return fixes
			})
		}}
	},
}
