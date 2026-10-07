package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func (state *staticEvaluationState) optionalChainShortCircuits(node *ast.Node, staticEvaluator *StaticStringEvaluator) bool {
	if node == nil || !ast.IsOptionalChain(node) {
		return false
	}
	if result, found := state.optionalShortCircuits[node]; found {
		return result
	}
	result := state.computeOptionalChainShortCircuit(node, staticEvaluator)
	state.optionalShortCircuits[node] = result
	return result
}

func (state *staticEvaluationState) computeOptionalChainShortCircuit(node *ast.Node, staticEvaluator *StaticStringEvaluator) bool {
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
	if state.optionalChainShortCircuits(input, staticEvaluator) {
		return true
	}
	if !ast.IsOptionalChainRoot(node) {
		return false
	}
	value := staticEvaluator.evalValue(input)
	return value.ok && staticValueNullish(value.value)
}
