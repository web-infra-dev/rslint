package unicornutil

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// IsDirectEvalCall matches a non-optional call to eval, including parentheses
// and erased TypeScript wrappers. A local binding can hold the intrinsic eval.
func IsDirectEvalCall(node *ast.Node) bool {
	if !ast.IsCallExpression(node) || node.AsCallExpression().QuestionDotToken != nil {
		return false
	}
	callee := ast.SkipOuterExpressions(node.Expression(), ast.OEKAll)
	return ast.IsIdentifier(callee) && callee.Text() == "eval"
}
