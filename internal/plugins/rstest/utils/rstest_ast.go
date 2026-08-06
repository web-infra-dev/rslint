package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// Rstest-local AST primitive. The other two primitives this file used to hold
// both had jest counterparts and now live in internal/utils/test_framework
// (IsFunction, CalleeChainName), which every rule imports directly.
// IsPromiseChainCall has no counterpart in any other plugin, so it stays here
// until a second consumer exists.

// IsPromiseChainCall reports promise-chain method calls: the callee is a
// member access of then / catch / finally with a statically known name, and
// the argument count is within the method's arity — at least 1, at most 2 for
// then and at most 1 for catch / finally. Mirrors eslint-plugin-jest's
// isPromiseChainCall (valid-expect-in-promise.ts:24-47); jest's Go port has no
// counterpart yet, so the upstream JS is the reference. Parentheses are
// skipped where ESTree would not materialize them.
func IsPromiseChainCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	name := promiseChainMethodName(call.Expression)
	if name == "" {
		return false
	}

	argumentCount := 0
	if call.Arguments != nil {
		argumentCount = len(call.Arguments.Nodes)
	}
	if argumentCount == 0 {
		// Promise methods need at least one argument to do anything.
		return false
	}

	switch name {
	case "then":
		return argumentCount < 3
	case "catch", "finally":
		return argumentCount < 2
	default:
		return false
	}
}

// promiseChainMethodName resolves the statically known member name of a
// promise-chain callee, applying eslint-plugin-jest's isSupportedAccessor
// rules: a plain identifier property, or a computed key that is a string
// literal or a no-substitution template. Dynamic keys yield "".
func promiseChainMethodName(callee *ast.Node) string {
	if callee == nil {
		return ""
	}
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return ""
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	case ast.KindElementAccessExpression:
		key := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if key == nil {
			return ""
		}
		switch key.Kind {
		case ast.KindStringLiteral:
			return key.AsStringLiteral().Text
		case ast.KindNoSubstitutionTemplateLiteral:
			return key.AsNoSubstitutionTemplateLiteral().Text
		}
	}
	return ""
}
