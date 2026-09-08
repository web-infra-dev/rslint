package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// FollowTypeAssertionChain unwraps parentheses plus `as` and angle-bracket
// type assertions. It deliberately stops at non-null and `satisfies`, matching
// the assertion-chain helpers in eslint-plugin-jest and eslint-plugin-vitest.
func FollowTypeAssertionChain(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
	return nil
}

// InvokedAccessorCall returns the call expression that directly invokes
// entry's accessor, or nil when the accessor is not called.
//
// MemberEntry.Call is not sufficient for nested calls: GetMemberEntries marks
// the last entry of a callee chain as invoked, so an outer call can overwrite
// the matcher entry's Call field. Walking from the accessor always finds the
// argument list owned by that matcher.
func InvokedAccessorCall(entry *MemberEntry) *ast.Node {
	_, accessor := AccessorReceiverAndParent(entry)
	if accessor == nil {
		return nil
	}

	child := accessor
	parent := accessor.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child = parent
		parent = parent.Parent
	}
	if parent == nil ||
		parent.Kind != ast.KindCallExpression ||
		parent.AsCallExpression().Expression != child {
		return nil
	}
	return parent
}

// CallArgumentListRange returns the span between a call's parentheses. The
// range contains every argument, comma, comment and whitespace, but not the
// parentheses themselves.
func CallArgumentListRange(sourceFile *ast.SourceFile, call *ast.Node) (core.TextRange, bool) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return core.TextRange{}, false
	}
	callExpression := call.AsCallExpression()
	callee := callExpression.Expression
	if callee == nil {
		return core.TextRange{}, false
	}

	start := callee.End()
	// A type argument list may contain parentheses of its own, so scanning has
	// to begin after it.
	if typeArguments := callExpression.TypeArguments; typeArguments != nil {
		start = max(start, typeArguments.End())
	}

	tokens := utils.TokensOfNode(sourceFile, call)
	if len(tokens) == 0 {
		return core.TextRange{}, false
	}
	closeParen := tokens[len(tokens)-1]
	if closeParen.Kind != ast.KindCloseParenToken {
		return core.TextRange{}, false
	}

	for _, token := range tokens {
		if token.Start < start {
			continue
		}
		if token.Kind == ast.KindOpenParenToken {
			if token.End > closeParen.Start {
				return core.TextRange{}, false
			}
			return core.NewTextRange(token.End, closeParen.Start), true
		}
	}
	return core.TextRange{}, false
}
