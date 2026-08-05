package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// Generic AST primitives for upcoming expect-consuming rules (valid-expect,
// no-standalone-expect, valid-expect-in-promise). They are framework-agnostic;
// the jest counterparts live in internal/plugins/jest/utils and are copied
// here rather than shared, following the "don't modify jest utils" porting
// rule. If a third plugin ever needs them, extract to
// internal/utils/test_framework then.

// IsFunction reports nodes that declare a callable body. Mirrors jest's
// IsFunction (jest.go:481); the nil guard is an addition, jest's callers all
// pre-check.
func IsFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsFunctionDeclaration(node) ||
		ast.IsFunctionExpressionOrArrowFunction(node) ||
		node.Kind == ast.KindMethodDeclaration ||
		node.Kind == ast.KindConstructor ||
		node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor
}

// CalleeChainName returns a dotted name for a call callee expression,
// mirroring eslint-plugin-jest's getNodeName for CallExpression callees and
// jest's CalleeChainName (jest.go:195).
//
// It differs from GetRstestFnMemberEntries: bracket notation contributes a
// segment only when the index matches eslint-plugin-jest's supported accessor
// names (identifier, string literal, or no-substitution template). Unsupported
// keys break the chain entirely. NewExpression is peeled so e.g.
// new (require('x')).y becomes a chain.
func CalleeChainName(expr *ast.Node) string {
	if expr == nil {
		return ""
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil {
		return ""
	}

	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		property := expr.AsPropertyAccessExpression()
		left := CalleeChainName(property.Expression)
		name := property.Name()
		if name == nil {
			return left
		}
		propertyName := calleeChainPropertyName(name)
		if left == "" || propertyName == "" {
			return left
		}
		return left + "." + propertyName
	case ast.KindElementAccessExpression:
		element := expr.AsElementAccessExpression()
		left := CalleeChainName(element.Expression)
		key := calleeChainLiteralElementKey(ast.SkipParentheses(element.ArgumentExpression))
		if left == "" || key == "" {
			return ""
		}
		return left + "." + key
	case ast.KindCallExpression:
		return CalleeChainName(expr.AsCallExpression().Expression)
	case ast.KindNewExpression:
		newExpression := expr.AsNewExpression()
		if newExpression == nil {
			return ""
		}
		return CalleeChainName(newExpression.Expression)
	case ast.KindTaggedTemplateExpression:
		return CalleeChainName(expr.AsTaggedTemplateExpression().Tag)
	default:
		return ""
	}
}

// calleeChainPropertyName mirrors jest's getPropertyName (jest.go:149).
func calleeChainPropertyName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text
	}
	return ""
}

// calleeChainLiteralElementKey matches eslint-plugin-jest segments for
// MemberExpression computed with a supported accessor name only. Mirrors
// jest's calleeChainLiteralElementKey (jest.go:241).
func calleeChainLiteralElementKey(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	default:
		return ""
	}
}

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
