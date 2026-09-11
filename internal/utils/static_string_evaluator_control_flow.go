package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// EvalControlFlowTruthiness returns JavaScript truthiness under the same
// conservative policy as EvalControlFlowValue. It lets callers select a branch
// without evaluating that branch or depending on private static value types.
func (staticEvaluator *StaticStringEvaluator) EvalControlFlowTruthiness(node *ast.Node) (truthy bool, known bool) {
	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false, false
	}
	return staticValueTruthy(value)
}
