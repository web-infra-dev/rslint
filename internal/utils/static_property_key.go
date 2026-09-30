package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func isStaticDeterministicMathMethod(name string) bool {
	switch name {
	case "abs", "acos", "acosh", "asin", "asinh",
		"atan", "atanh", "atan2", "cbrt", "ceil",
		"clz32", "cos", "cosh", "exp", "expm1",
		"f16round", "floor", "fround", "hypot", "imul",
		"log", "log10", "log1p", "log2", "max",
		"min", "pow", "round", "sign", "sin",
		"sinh", "sqrt", "tan", "tanh", "trunc":
		return true
	default:
		return false
	}
}

// HasStaticAccessExpressionKey reports whether a member has a statically known
// JavaScript property key. String/number keys use the evaluator's ordinary
// ToPropertyKey path; Symbol keys remain first-class static values so aliases,
// conditionals, and allowed Symbol calls compose through the same evaluator.
func (staticEvaluator *StaticStringEvaluator) HasStaticAccessExpressionKey(node *ast.Node) bool {
	if staticEvaluator == nil || node == nil {
		return false
	}
	if _, known := staticEvaluator.EvalAccessExpressionName(node); known {
		return true
	}
	if node.Kind != ast.KindElementAccessExpression {
		return false
	}
	expression := node.AsElementAccessExpression().ArgumentExpression
	if value := staticEvaluator.evalValue(expression); value.ok {
		if _, isSymbol := value.value.(staticSymbolValue); isSymbol {
			return true
		}
	}
	return staticEvaluator.isStaticDeterministicMathCall(expression)
}

func (staticEvaluator *StaticStringEvaluator) isStaticDeterministicMathCall(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	method, ok := staticEvaluator.builtinMethodName(
		node.AsCallExpression().Expression,
		"Math",
		map[*ast.Symbol]bool{},
	)
	if !ok {
		return false
	}
	if !isStaticDeterministicMathMethod(method) {
		return false
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return false
	}
	for _, argument := range arguments {
		if _, ok := staticValueToNumber(argument); !ok {
			return false
		}
	}
	return true
}
