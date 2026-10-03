package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// HasStaticAccessExpressionKey reports whether a member has a statically known
// JavaScript property key. String/number keys use the evaluator's ordinary
// ToPropertyKey path; Symbol keys remain first-class static values so aliases,
// conditionals, and allowed Symbol calls compose through the same evaluator.
func (staticEvaluator *StaticStringEvaluator) HasStaticAccessExpressionKey(node *ast.Node) bool {
	if staticEvaluator == nil || node == nil {
		return false
	}
	if node.Kind != ast.KindElementAccessExpression {
		_, known := AccessExpressionStaticName(node)
		return known
	}
	expression := node.AsElementAccessExpression().ArgumentExpression
	if value := staticEvaluator.evalValue(expression); value.ok {
		if _, isSymbol := value.value.(staticSymbolValue); isSymbol {
			return true
		}
		_, known := staticValueToString(value.value)
		return known
	}
	return false
}
