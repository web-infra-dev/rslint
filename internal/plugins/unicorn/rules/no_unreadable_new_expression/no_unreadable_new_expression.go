// Package no_unreadable_new_expression ports eslint-plugin-unicorn's
// `no-unreadable-new-expression` rule.
package no_unreadable_new_expression

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDMemberAccess       = "member-access"
	messageIDComplexConstructor = "complex-constructor"
)

var (
	messageMemberAccess = rule.RuleMessage{
		Id:          messageIDMemberAccess,
		Description: "Do not access members directly from a `new` expression.",
	}
	messageComplexConstructor = rule.RuleMessage{
		Id:          messageIDComplexConstructor,
		Description: "Do not use a complex expression as a constructor.",
	}
)

// isStaticMemberExpression mirrors upstream's recursive MemberExpression
// predicate. Source parentheses and JavaScript JSDoc casts are transparent in
// ESTree, while authored TypeScript assertions and optional chains remain
// visible and therefore make the constructor complex.
func isStaticMemberExpression(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || node.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(node) {
		return false
	}

	access := node.AsPropertyAccessExpression()
	if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
		return false
	}

	object := utils.ESTreeRuntimeExpression(access.Expression)
	return object != nil && (object.Kind == ast.KindIdentifier || isStaticMemberExpression(object))
}

func isSimpleConstructor(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && (node.Kind == ast.KindIdentifier || isStaticMemberExpression(node))
}

func checkMemberExpression(ctx rule.RuleContext, node *ast.Node) {
	var object, property *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		object = access.Expression
		property = access.Name()
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		object = access.Expression
		property = utils.ESTreeRuntimeExpression(access.ArgumentExpression)
	default:
		return
	}

	object = utils.ESTreeRuntimeExpression(object)
	if object == nil || object.Kind != ast.KindNewExpression || property == nil {
		return
	}

	ctx.ReportNode(property, messageMemberAccess)
}

// NoUnreadableNewExpressionRule disallows member access directly from a new
// expression and disallows complex expressions as constructors.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-unreadable-new-expression.js
var NoUnreadableNewExpressionRule = rule.Rule{
	Name:   "unicorn/no-unreadable-new-expression",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberExpression(ctx, node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkMemberExpression(ctx, node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				callee := utils.ESTreeRuntimeExpression(node.AsNewExpression().Expression)
				if isSimpleConstructor(callee) {
					return
				}
				if callee != nil {
					ctx.ReportNode(callee, messageComplexConstructor)
				}
			},
		}
	},
}
