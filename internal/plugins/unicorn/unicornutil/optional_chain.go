package unicornutil

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// HasOptionalChainElement reports whether node's callee/object path contains
// an optional-chain link. Its transparent-node list is deliberately closed:
// tsgo's ParenthesizedExpression plus the four TypeScript wrappers that
// Unicorn unwraps (as, satisfies, non-null, and angle-bracket assertions).
// Call arguments and element-access keys are deliberately outside the path.
func HasOptionalChainElement(node *ast.Node) bool {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			node = node.Expression()
		case ast.KindCallExpression:
			if ast.IsOptionalChain(node) {
				return true
			}
			node = node.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression:
			if ast.IsOptionalChain(node) {
				return true
			}
			node = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			if ast.IsOptionalChain(node) {
				return true
			}
			node = node.AsElementAccessExpression().Expression
		default:
			return false
		}
	}
	return false
}
