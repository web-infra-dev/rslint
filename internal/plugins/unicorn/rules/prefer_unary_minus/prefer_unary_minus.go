// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_unary_minus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var (
	multiplicationMessage = rule.RuleMessage{
		Id:          "prefer-unary-minus",
		Description: "Prefer the unary minus operator over multiplying by `-1`.",
		Data:        map[string]string{"operation": "multiplying"},
	}
	divisionMessage = rule.RuleMessage{
		Id:          "prefer-unary-minus",
		Description: "Prefer the unary minus operator over dividing by `-1`.",
		Data:        map[string]string{"operation": "dividing"},
	}
)

var PreferUnaryMinusRule = rule.Rule{
	Name:   "unicorn/prefer-unary-minus",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind != ast.KindAsteriskToken && binary.OperatorToken.Kind != ast.KindSlashToken {
					return
				}
				left := utils.ESTreeRuntimeExpression(binary.Left)
				right := utils.ESTreeRuntimeExpression(binary.Right)
				leftIsNegativeOne, rightIsNegativeOne := isNegativeOne(left), isNegativeOne(right)
				var operand *ast.Node
				if rightIsNegativeOne && !leftIsNegativeOne {
					operand = left
				} else if binary.OperatorToken.Kind == ast.KindAsteriskToken && leftIsNegativeOne && !rightIsNegativeOne {
					operand = right
				} else {
					return
				}

				message := multiplicationMessage
				if binary.OperatorToken.Kind == ast.KindSlashToken {
					message = divisionMessage
				}
				ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
					nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
					operandRange := utils.TrimNodeTextRange(ctx.SourceFile, operand)
					comments := ctx.Comments.All()
					if utils.HasCommentInSpan(comments, nodeRange.Pos(), operandRange.Pos()) ||
						utils.HasCommentInSpan(comments, operandRange.End(), nodeRange.End()) {
						return nil
					}

					text := utils.TrimmedNodeText(ctx.SourceFile, operand)
					replacement := "-" + text
					if needsOperandParentheses(operand) {
						replacement = "-(" + text + ")"
					}
					if needsReplacementParentheses(node) {
						replacement = "(" + replacement + ")"
					}
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, replacement) {
						replacement = ";" + replacement
					} else if before, ok := utils.TokenBeforePosition(ctx.SourceFile, nodeRange.Pos()); ok &&
						before.Kind == ast.KindMinusToken && before.End == nodeRange.Pos() {
						replacement = " " + replacement
					}
					fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
					return append(fixes, rule.RuleFixReplace(ctx.SourceFile, node, replacement))
				})
			},
		}
	},
}

func isNegativeOne(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression || node.AsPrefixUnaryExpression().Operator != ast.KindMinusToken {
		return false
	}
	argument := utils.ESTreeRuntimeExpression(node.AsPrefixUnaryExpression().Operand)
	// The tsgo scanner already normalizes numeric values, including alternate
	// bases, separators, exponents, and floating-point rounding.
	return argument != nil && argument.Kind == ast.KindNumericLiteral && argument.Text() == "1"
}

func needsOperandParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNonNullExpression:
		// A trailing ! within an optional chain belongs to ESTree's
		// ChainExpression; an ordinary TSNonNullExpression needs parentheses.
		return utils.EslintLikePrecedence(node) < 0
	case ast.KindPostfixUnaryExpression, ast.KindTypeAssertionExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		return operator == ast.KindMinusToken || operator == ast.KindMinusMinusToken || operator == ast.KindPlusPlusToken
	default:
		return ast.GetExpressionPrecedence(node) < ast.OperatorPrecedenceUnary
	}
}

func needsReplacementParentheses(node *ast.Node) bool {
	outer := node
	for parent := utils.ESTreeParent(outer); parent != nil && ast.IsOuterExpression(parent, ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) && utils.ESTreeRuntimeExpression(parent.Expression()) == outer; parent = utils.ESTreeParent(outer) {
		outer = parent
	}
	parent := utils.ESTreeParent(outer)
	if parent == nil {
		return false
	}
	exponentiation := parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAsteriskAsteriskToken &&
		utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left) == outer
	for current := node; ; current = utils.ESTreeParent(current) {
		if current.Parent != nil && ast.IsParenthesizedExpression(current.Parent) {
			// Parentheses around erased assertions do not protect a unary left
			// operand of ** when TypeScript emits JavaScript.
			erasedAssertion := ast.IsAssertionExpression(current) || current.Kind == ast.KindSatisfiesExpression
			if !exponentiation || !erasedAssertion {
				return false
			}
		}
		if current == outer {
			break
		}
	}
	if exponentiation {
		return true
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression, ast.KindNewExpression:
		return utils.ESTreeRuntimeExpression(parent.Expression()) == outer
	case ast.KindTaggedTemplateExpression:
		return utils.ESTreeRuntimeExpression(parent.AsTaggedTemplateExpression().Tag) == outer
	default:
		return false
	}
}
