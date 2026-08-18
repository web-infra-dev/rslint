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
	if node == nil {
		return nil
	}
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

// GetRequireCallWithStringLiteralArgument returns a require call whose only
// argument is an ESTree string Literal. Unlike IsStringLiteralLike, this
// deliberately excludes no-substitution template literals. Optional calls are
// left to the caller because several upstream helpers accept them.
func GetRequireCallWithStringLiteralArgument(node *ast.Node) *ast.CallExpression {
	call := GetRequireCall(node, false)
	if call == nil {
		return nil
	}
	if !IsESTreeStringLiteral(call.Arguments.Nodes[0]) {
		return nil
	}
	return call
}

// GetStaticRequireCall returns the require-call shape import/order can follow
// into a top-level variable declaration. The shared upstream staticRequire
// predicate accepts optional calls, but ESTree wraps those in ChainExpression,
// which stops order's parent walk; tsgo needs the rejection explicitly.
func GetStaticRequireCall(node *ast.Node) *ast.CallExpression {
	call := GetRequireCallWithStringLiteralArgument(node)
	if call == nil || ast.IsOptionalChain(call.AsNode()) {
		return nil
	}
	return call
}

// GetRequireCallWithLiteralArgument returns the require-expression shape used
// by eslint-plugin-import/order for named destructuring and fixer barriers.
// Unlike GetStaticRequireCall, the argument may be any ESTree Literal (string,
// number, bigint, boolean, null, or regexp), but not a template literal,
// dynamic expression, or optional call.
func GetRequireCallWithLiteralArgument(node *ast.Node) *ast.CallExpression {
	call := GetRequireCall(node, false)
	if call == nil || ast.IsOptionalChain(call.AsNode()) {
		return nil
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[0])
	if argument == nil {
		return nil
	}
	switch argument.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return call
	default:
		return nil
	}
}

// FindStaticRequireCallInChain follows the callee/receiver side of ordinary
// call and member expressions until it reaches a static require call. It does
// not cross optional chains or TypeScript-only wrappers, matching the ESTree
// walk used by eslint-plugin-import's require-ordering rules.
func FindStaticRequireCallInChain(node *ast.Node) *ast.CallExpression {
	if node == nil {
		return nil
	}
	current := ast.SkipParentheses(node)
	for current != nil {
		// ESTree wraps the complete optional chain in ChainExpression, which
		// stops order's parent walk before it can reach a declarator.
		if ast.IsOptionalChain(current) {
			return nil
		}
		if call := GetStaticRequireCall(current); call != nil {
			return call
		}

		var receiver *ast.Node
		switch current.Kind {
		case ast.KindCallExpression:
			call := current.AsCallExpression()
			receiver = call.Expression
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			receiver = access.Expression
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			receiver = access.Expression
		default:
			return nil
		}
		current = ast.SkipParentheses(receiver)
	}
	return nil
}
