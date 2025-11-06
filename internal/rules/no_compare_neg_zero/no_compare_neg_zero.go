package no_compare_neg_zero

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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

// isNegativeZero checks if a node represents -0
// This matches: UnaryExpression with operator "-" and argument being Literal with value 0
func isNegativeZero(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}

	prefix := node.AsPrefixUnaryExpression()
	if prefix == nil || prefix.Operator != ast.KindMinusToken {
		return false
	}

	operand := prefix.Operand
	if operand == nil {
		return false
	}

	// Check if the operand is a numeric literal with value 0
	switch operand.Kind {
	case ast.KindNumericLiteral:
		numLiteral := operand.AsNumericLiteral()
		if numLiteral != nil && numLiteral.Text == "0" {
			return true
		}
	}

	return false
}

// NoCompareNegZeroRule disallows comparisons to negative zero
var NoCompareNegZeroRule = rule.CreateRule(rule.Rule{
	Name: "no-compare-neg-zero",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Define the operators we want to check
		operatorsToCheck := map[ast.Kind]bool{
			ast.KindGreaterThanToken:             true, // >
			ast.KindGreaterThanEqualsToken:       true, // >=
			ast.KindLessThanToken:                true, // <
			ast.KindLessThanEqualsToken:          true, // <=
			ast.KindEqualsEqualsToken:            true, // ==
			ast.KindEqualsEqualsEqualsToken:      true, // ===
			ast.KindExclamationEqualsToken:       true, // !=
			ast.KindExclamationEqualsEqualsToken: true, // !==
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				// Check if this is one of the operators we care about
				if !operatorsToCheck[binary.OperatorToken.Kind] {
					return
				}

				// Check if either side is -0
				if isNegativeZero(binary.Left) || isNegativeZero(binary.Right) {
					// Get the operator text for the error message
					operatorText := getOperatorText(binary.OperatorToken.Kind)
					ctx.ReportNode(node, buildCompareNegZeroMessage(operatorText))
				}
			},
		}
	},
})
