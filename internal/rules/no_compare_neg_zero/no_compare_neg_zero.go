package no_compare_neg_zero

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildCompareNegZeroMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not use the '" + operator + "' operator to compare against -0.",
	}
}

// getOperatorText converts an operator Kind to its string representation
func getOperatorText(kind ast.Kind) string {
	switch kind {
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindGreaterThanEqualsToken:
		return ">="
	case ast.KindLessThanToken:
		return "<"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	default:
		return ""
	}
}

func isZeroNumericLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindNumericLiteral {
		return false
	}
	numeric := node.AsNumericLiteral()
	return numeric != nil && numeric.Text == "0"
}

// isNegativeZero checks whether node represents the ESTree shape `-0`.
// Parentheses are transparent around both the unary expression and its
// numeric operand.
func isNegativeZero(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	prefix := node.AsPrefixUnaryExpression()
	return prefix != nil &&
		prefix.Operator == ast.KindMinusToken &&
		isZeroNumericLiteral(prefix.Operand)
}

// NoCompareNegZeroRule disallows comparisons to negative zero
var NoCompareNegZeroRule = rule.Rule{
	Name: "no-compare-neg-zero",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil || prefix.Operator != ast.KindMinusToken {
					return
				}

				operand := utils.OutermostParenthesizedExpression(node)
				comparisonNode := operand.Parent
				if comparisonNode == nil || comparisonNode.Kind != ast.KindBinaryExpression {
					return
				}

				comparison := comparisonNode.AsBinaryExpression()
				if comparison == nil || comparison.OperatorToken == nil ||
					comparison.Left != operand && comparison.Right != operand {
					return
				}

				operatorText := getOperatorText(comparison.OperatorToken.Kind)
				if operatorText == "" {
					return
				}
				if !isZeroNumericLiteral(prefix.Operand) {
					return
				}

				// When both operands are -0, the left listener reports the
				// comparison and the right listener skips the duplicate.
				if comparison.Right == operand && isNegativeZero(comparison.Left) {
					return
				}

				ctx.ReportNode(comparisonNode, buildCompareNegZeroMessage(operatorText))
			},
		}
	},
}
