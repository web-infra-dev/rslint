package no_nesting

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/promiseutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

// isThenOrCatchCall reports whether node is a call whose callee is a non-computed
// .then or .catch member access. Mirrors eslint-plugin-promise's has-promise-callback.
func isThenOrCatchCall(node *ast.Node) bool {
	return promiseutil.IsMemberCall(node, "then") || promiseutil.IsMemberCall(node, "catch")
}

// isPromiseCallback reports whether node is a FunctionExpression or ArrowFunction
// directly passed as argument to a .then() or .catch() call.
// Mirrors eslint-plugin-promise's lib/is-inside-promise helper.
func isPromiseCallback(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	// In tsgo, an argument node's parent is the containing CallExpression.
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	return isThenOrCatchCall(parent)
}

// walkIdentifiers calls fn for each identifier name found anywhere in node's
// subtree, stopping early across the whole traversal if fn returns true.
func walkIdentifiers(node *ast.Node, fn func(name string) bool) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return fn(node.AsIdentifier().Text)
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if walkIdentifiers(child, fn) {
			found = true
			return true
		}
		return false
	})
	return found
}

// argsContainRef reports whether any identifier anywhere in the call's argument
// list has a name that is bound in fn's parameter list or local variable
// declarations. Uses utils.HasShadowingParameter for params and
// utils.HasShadowingDeclaration / utils.HasHoistedVarDeclaration for body
// declarations, mirroring the upstream's scope-reference check.
func argsContainRef(callNode *ast.Node, fn *ast.Node) bool {
	args := callNode.AsCallExpression().Arguments
	if args == nil {
		return false
	}
	body := fn.Body()
	for _, arg := range args.Nodes {
		found := false
		walkIdentifiers(arg, func(name string) bool {
			if utils.HasShadowingParameter(fn, name) {
				found = true
				return true
			}
			if body != nil && (utils.HasShadowingDeclaration(body, name) || utils.HasHoistedVarDeclaration(body, name)) {
				found = true
				return true
			}
			return false
		})
		if found {
			return true
		}
	}
	return false
}

func buildAvoidNestingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "avoidNesting",
		Description: "Avoid nesting promises.",
	}
}

var NoNestingRule = rule.Rule{
	Name: "promise/no-nesting",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Stack of promise-callback function nodes, closest last.
		callbackStack := []*ast.Node{}

		onEnter := func(node *ast.Node) {
			if isPromiseCallback(node) {
				callbackStack = append(callbackStack, node)
			}
		}
		onExit := func(node *ast.Node) {
			if isPromiseCallback(node) && len(callbackStack) > 0 {
				callbackStack = callbackStack[:len(callbackStack)-1]
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionExpression:                      onEnter,
			ast.KindArrowFunction:                           onEnter,
			rule.ListenerOnExit(ast.KindFunctionExpression): onExit,
			rule.ListenerOnExit(ast.KindArrowFunction):      onExit,

			ast.KindCallExpression: func(node *ast.Node) {
				if !isThenOrCatchCall(node) || len(callbackStack) == 0 {
					return
				}
				closestCallback := callbackStack[len(callbackStack)-1]
				if argsContainRef(node, closestCallback) {
					return
				}
				callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
				if callee == nil || !ast.IsPropertyAccessExpression(callee) {
					return
				}
				nameNode := callee.AsPropertyAccessExpression().Name()
				if nameNode != nil {
					ctx.ReportNode(nameNode, buildAvoidNestingMessage())
				}
			},
		}
	},
}
