package no_alert

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isProhibitedIdentifier checks if the name is alert, confirm, or prompt.
func isProhibitedIdentifier(name string) bool {
	return name == "alert" || name == "confirm" || name == "prompt"
}

// outerExpressionKinds covers parentheses, type assertions (as / angle-bracket),
// non-null assertions (!), and satisfies — all of which are transparent at runtime.
const outerExpressionKinds = ast.OEKParentheses | ast.OEKAssertions

// isGlobalThisOrWindow checks if the node is a reference to the global object:
// - `this` at global scope (not inside any function-like declaration or class)
// - non-shadowed `window`
// - non-shadowed `globalThis`
//
// Skips outer expression wrappers (parentheses, type assertions, non-null assertions)
// so that `(window).alert()` and `window!.alert()` are handled correctly.
func isGlobalThisOrWindow(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, outerExpressionKinds)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindThisKeyword {
		// `this` is at the global scope only when not enclosed by any
		// function-like declaration or class body. Class fields, static blocks,
		// and computed property names all live inside a class scope where `this`
		// refers to the class or instance, not globalThis.
		return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
			return ast.IsFunctionLikeDeclaration(n) || ast.IsClassLike(n)
		}) == nil
	}
	if node.Kind == ast.KindIdentifier {
		name := node.Text()
		if name == "window" || name == "globalThis" {
			return !utils.IsShadowed(node, name)
		}
	}
	return false
}

// https://eslint.org/docs/latest/rules/no-alert
var NoAlertRule = rule.Rule{
	Name: "no-alert",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipOuterExpressions(node.Expression(), outerExpressionKinds)
				if callee == nil {
					return
				}

				var name string

				switch callee.Kind {
				case ast.KindIdentifier:
					// Direct call: alert(), confirm(), prompt()
					name = callee.Text()
					if !isProhibitedIdentifier(name) {
						return
					}
					if utils.IsShadowed(callee, name) {
						return
					}

				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
					// Member access: window.alert(), window['alert'](), globalThis.alert(), this.alert()
					var ok bool
					name, ok = utils.AccessExpressionStaticName(callee)
					if !ok || !isProhibitedIdentifier(name) {
						return
					}
					if !isGlobalThisOrWindow(utils.AccessExpressionObject(callee)) {
						return
					}

				default:
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: fmt.Sprintf("Unexpected %s.", name),
				})
			},
		}
	},
}
