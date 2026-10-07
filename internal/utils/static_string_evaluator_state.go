package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

type staticEvaluationState struct {
	values                map[*ast.Node]staticEvalResult
	optionalShortCircuits map[*ast.Node]bool
}

func (staticEvaluator *StaticStringEvaluator) evalValue(node *ast.Node) staticEvalResult {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return staticEvalResult{}
	}
	if staticEvaluator.evaluation == nil {
		staticEvaluator.evaluation = &staticEvaluationState{
			values:                map[*ast.Node]staticEvalResult{},
			optionalShortCircuits: map[*ast.Node]bool{},
		}
		defer func() { staticEvaluator.evaluation = nil }()
	}
	if result, found := staticEvaluator.evaluation.values[node]; found {
		return result
	}
	result := staticEvaluator.evalUncachedValue(node)
	staticEvaluator.evaluation.values[node] = result
	return result
}
