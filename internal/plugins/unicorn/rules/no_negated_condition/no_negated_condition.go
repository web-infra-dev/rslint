// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
package no_negated_condition

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoNegatedConditionRule = rule.Rule{
	Name:   "unicorn/no-negated-condition",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			var condition, consequent, alternate *ast.Node
			isIf := node.Kind == ast.KindIfStatement
			if isIf {
				statement := node.AsIfStatement()
				if statement.ElseStatement == nil || statement.ElseStatement.Kind == ast.KindIfStatement {
					return
				}
				condition, consequent, alternate = statement.Expression, statement.ThenStatement, statement.ElseStatement
			} else {
				expression := node.AsConditionalExpression()
				condition, consequent, alternate = expression.Condition, expression.WhenTrue, expression.WhenFalse
			}
			test := utils.ESTreeRuntimeExpression(condition)
			switch test.Kind {
			case ast.KindPrefixUnaryExpression:
				if test.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
					return
				}
			case ast.KindBinaryExpression:
				operator := test.AsBinaryExpression().OperatorToken.Kind
				if operator != ast.KindExclamationEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
					return
				}
			default:
				return
			}
			ctx.ReportNodeWithDeferredFixes(test, rule.RuleMessage{
				Id: "no-negated-condition", Description: "Unexpected negated condition.",
			}, func() []rule.RuleFix {
				consequentRange, consequentText, ok := branchText(ctx, consequent, isIf)
				if !ok {
					return nil
				}
				alternateRange, alternateText, ok := branchText(ctx, alternate, isIf)
				if !ok {
					return nil
				}
				var fixes []rule.RuleFix
				if consequentText != alternateText {
					fixes = make([]rule.RuleFix, 0, 3)
					fixes = append(fixes,
						rule.RuleFixReplaceRange(consequentRange, alternateText),
						rule.RuleFixReplaceRange(alternateRange, consequentText),
					)
				}
				if test.Kind == ast.KindBinaryExpression {
					operator := test.AsBinaryExpression().OperatorToken
					replacement := "=="
					if operator.Kind == ast.KindExclamationEqualsEqualsToken {
						replacement = "==="
					}
					return append(fixes, rule.RuleFixReplace(ctx.SourceFile, operator, replacement))
				}

				bang, _ := utils.TokenAtOrAfter(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, test).Pos())
				fixes = append(fixes, rule.RuleFixRemoveRange(bang.Range()))
				if isIf {
					// The if statement already supplies the parentheses around its test.
					operand := test.AsPrefixUnaryExpression().Operand
					inner := utils.ESTreeRuntimeExpression(operand)
					for operand != inner {
						if ast.IsParenthesizedExpression(operand) {
							opening, _ := utils.TokenAtOrAfter(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, operand).Pos())
							closing, _ := utils.TokenBeforePosition(ctx.SourceFile, operand.End())
							fixes = append(fixes, rule.RuleFixRemoveRange(opening.Range()), rule.RuleFixRemoveRange(closing.Range()))
						}
						operand = operand.Expression()
					}
					return fixes
				}

				fixes = append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
				afterBang, _ := utils.TokenAtOrAfter(ctx.SourceFile, bang.End)
				needsGrouping := afterBang.Kind == ast.KindOpenBraceToken || afterBang.Kind == ast.KindFunctionKeyword || afterBang.Kind == ast.KindClassKeyword
				if afterBang.Kind == ast.KindAsyncKeyword {
					next, _ := utils.TokenAtOrAfter(ctx.SourceFile, afterBang.End)
					needsGrouping = next.Kind == ast.KindFunctionKeyword
				}
				if afterBang.Kind == ast.KindLetKeyword {
					next, _ := utils.TokenAtOrAfter(ctx.SourceFile, afterBang.End)
					needsGrouping = next.Kind == ast.KindOpenBracketToken
				}
				// Removing ! must not turn an expression into a block or declaration.
				isArrowBody := afterBang.Kind == ast.KindOpenBraceToken && test == condition &&
					node.Parent != nil && node.Parent.Kind == ast.KindArrowFunction && node.Parent.Body() == node
				if isArrowBody || needsGrouping && utils.IsStartOfExpressionStatement(ctx.SourceFile, test) {
					prefix := "("
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, prefix) {
						prefix = ";("
					}
					return append(fixes,
						rule.RuleFixReplaceRange(core.NewTextRange(bang.Start, bang.Start), prefix),
						rule.RuleFixReplaceRange(core.NewTextRange(test.End(), test.End()), ")"),
					)
				}
				if test == condition {
					parentheses := unicornutil.ReturnOrThrowParenthesesFixes(ctx.SourceFile, node, afterBang.Start)
					if len(parentheses) > 0 {
						return append(fixes, parentheses...)
					}
				}
				if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, afterBang.Text) {
					start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
					fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, start), ";"))
				}
				return fixes
			})
		}
		return rule.RuleListeners{ast.KindIfStatement: check, ast.KindConditionalExpression: check}
	},
}

// Leading comments travel with their branch. Trailing comments have ambiguous
// ownership, so upstream reports the condition without offering a fix.
func branchText(ctx rule.RuleContext, branch *ast.Node, isIf bool) (core.TextRange, string, bool) {
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, branch)
	if comments := ctx.Comments.All(); len(comments) > 0 {
		end := len(ctx.SourceFile.Text())
		if next, ok := utils.TokenAtOrAfter(ctx.SourceFile, textRange.End()); ok {
			end = next.Start
		}
		if utils.HasCommentInSpan(comments, textRange.End(), end) {
			return textRange, "", false
		}
		previous, _ := utils.TokenBeforePosition(ctx.SourceFile, textRange.Pos())
		if before := utils.CommentsInSpan(comments, previous.End, textRange.Pos()); len(before) > 0 {
			textRange = core.NewTextRange(before[0].Pos(), textRange.End())
		}
	}
	text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	if isIf && branch.Kind != ast.KindBlock {
		text = "{" + text + "}"
	}
	return textRange, text, true
}
