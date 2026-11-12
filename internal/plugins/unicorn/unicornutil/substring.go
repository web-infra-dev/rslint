package unicornutil

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsNonNegativeSafeInteger bounds an index to JavaScript's safe integer range.
func IsNonNegativeSafeInteger(value float64) bool {
	const maxSafeInteger = 1<<53 - 1
	return value >= 0 && value <= maxSafeInteger && math.Trunc(value) == value
}

// SafeIntegerLiteral reads a non-negative ESTree numeric literal, excluding
// unary expressions and TypeScript wrappers. Parentheses are transparent.
func SafeIntegerLiteral(node *ast.Node) (float64, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsNumericLiteral(node) {
		return 0, false
	}
	value := float64(jsnum.FromString(node.Text()))
	return value, IsNonNegativeSafeInteger(value)
}

// SingleCharacterSubstringIndex returns the original argument that represents
// the start of a one-character substring. Keeping the argument node preserves
// its source parentheses when prefer-at builds the suggestion.
func SingleCharacterSubstringIndex(args []*ast.Node) *ast.Node {
	if len(args) != 2 {
		return nil
	}
	first, second := args[0], args[1]
	start, startOK := SafeIntegerLiteral(first)
	end, endOK := SafeIntegerLiteral(second)
	if startOK && endOK && math.Abs(start-end) == 1 {
		if start < end {
			return first
		}
		return second
	}
	if isReferencePlusOne(second, first) {
		return first
	}
	if isReferencePlusOne(first, second) {
		return second
	}
	if expression := utils.ESTreeRuntimeExpression(first); ast.IsBinaryExpression(expression) {
		binary := expression.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindMinusToken && isOneLiteral(binary.Right) &&
			IsSameComparisonOperand(binary.Left, second) {
			return first
		}
	}
	return nil
}

func isReferencePlusOne(expression, reference *ast.Node) bool {
	expression = utils.ESTreeRuntimeExpression(expression)
	if !ast.IsBinaryExpression(expression) || expression.AsBinaryExpression().OperatorToken.Kind != ast.KindPlusToken {
		return false
	}
	binary := expression.AsBinaryExpression()
	return (isOneLiteral(binary.Right) && IsSameComparisonOperand(binary.Left, reference)) ||
		(isOneLiteral(binary.Left) && IsSameComparisonOperand(binary.Right, reference))
}

func isOneLiteral(node *ast.Node) bool {
	value, ok := SafeIntegerLiteral(node)
	return ok && value == 1
}
