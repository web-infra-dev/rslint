package promiseutil

import "github.com/microsoft/typescript-go/shim/ast"

const skipTransparent = ast.OEKParentheses

// IsMemberCall reports whether node is a call expression whose callee is a
// non-computed dotted member with the given name, mirroring the simple
// MemberExpression checks in eslint-plugin-promise.
func IsMemberCall(node *ast.Node, memberName string) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return name != nil && ast.IsIdentifier(name) && name.AsIdentifier().Text == memberName
}

// IsPromiseLikeCall mirrors eslint-plugin-promise's lib/is-promise helper.
// It recognizes .then/.catch/.finally chains and Promise static calls, except
// Promise.withResolvers, which upstream intentionally excludes.
func IsPromiseLikeCall(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode == nil || !ast.IsIdentifier(nameNode) {
		return false
	}
	name := nameNode.AsIdentifier().Text
	if name == "then" || name == "catch" || name == "finally" {
		return true
	}
	object := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
	if IsPromiseLikeCall(object) {
		return true
	}
	return object != nil && ast.IsIdentifier(object) && object.AsIdentifier().Text == "Promise" && IsPromiseStatic(name)
}

func IsPromiseStatic(name string) bool {
	switch name {
	case "all", "allSettled", "any", "race", "reject", "resolve":
		return true
	default:
		return false
	}
}
