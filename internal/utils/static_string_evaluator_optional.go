package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func (staticEvaluator *StaticStringEvaluator) optionalChainShortCircuits(node *ast.Node) bool {
	if node == nil || !ast.IsOptionalChain(node) {
		return false
	}
	var input *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		input = AccessExpressionObject(node)
	case ast.KindCallExpression:
		input = node.AsCallExpression().Expression
	case ast.KindNonNullExpression:
		input = node.AsNonNullExpression().Expression
	default:
		return false
	}
	if staticEvaluator.optionalChainShortCircuits(input) {
		return true
	}
	if !ast.IsOptionalChainRoot(node) {
		return false
	}
	value := staticEvaluator.evalValue(input)
	return value.ok && staticValueNullish(value.value)
}
