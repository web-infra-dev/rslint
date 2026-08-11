package utils

import "github.com/microsoft/typescript-go/shim/ast"

// GetRequireCall returns node as a bare require(...) call with exactly one
// argument. Parentheses around the call, callee, or argument are transparent
// because ESTree does not expose them as separate nodes.
//
// When requireStringLiteralLikeArgument is true, the sole argument must be a
// string literal or no-substitution template literal. Callers that require the
// narrower ESTree Literal shape, or that reject optional calls, should apply
// those checks to the returned call.
func GetRequireCall(node *ast.Node, requireStringLiteralLikeArgument bool) *ast.CallExpression {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	// Keep tsgo's canonical require-call predicate as the fast path. The
	// fallback below only bridges parenthesized callee/argument nodes, which
	// tsgo preserves but ESTree (and therefore eslint-plugin-import) erases.
	if ast.IsRequireCall(node, requireStringLiteralLikeArgument) {
		return node.AsCallExpression()
	}
	if !ast.IsCallExpression(node) {
		return nil
	}

	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}

	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsIdentifier(callee) || callee.Text() != "require" {
		return nil
	}

	if requireStringLiteralLikeArgument {
		argument := ast.SkipParentheses(call.Arguments.Nodes[0])
		if !ast.IsStringLiteralLike(argument) {
			return nil
		}
	}
	return call
}
