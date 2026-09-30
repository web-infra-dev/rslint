package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsStaticPrimitiveArgument reports whether an expression has a statically known
// primitive value whose numeric coercion cannot invoke user code.
func IsStaticPrimitiveArgument(evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	value, known := evaluator.EvalValue(node)
	if known {
		switch value.(type) {
		case string, bool, interface{ IsNaN() bool }:
			return true
		}
	}

	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindUndefinedKeyword, ast.KindVoidExpression:
		return true
	}
	return false
}
