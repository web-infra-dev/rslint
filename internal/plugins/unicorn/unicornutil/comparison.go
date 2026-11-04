package unicornutil

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// IsComparisonOperator recognizes the strict equality and ordering operators
// used by Unicorn's comparison rules.
func IsComparisonOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return true
	}
	return false
}

// FlipComparisonOperator reverses an ordering comparison's operand order.
// Equality operators are symmetric and remain unchanged.
func FlipComparisonOperator(operator ast.Kind) ast.Kind {
	switch operator {
	case ast.KindLessThanToken:
		return ast.KindGreaterThanToken
	case ast.KindLessThanEqualsToken:
		return ast.KindGreaterThanEqualsToken
	case ast.KindGreaterThanToken:
		return ast.KindLessThanToken
	case ast.KindGreaterThanEqualsToken:
		return ast.KindLessThanEqualsToken
	}
	return operator
}

// ContainsOptionalChain includes chains nested inside a comparison's operands.
func ContainsOptionalChain(node *ast.Node) bool {
	return ast.IsOptionalChain(node) || node.ForEachChild(ContainsOptionalChain)
}
