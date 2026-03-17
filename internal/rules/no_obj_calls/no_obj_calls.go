package no_obj_calls

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var nonCallableGlobals = map[string]bool{
	"Math": true, "JSON": true, "Reflect": true, "Atomics": true, "Intl": true,
}

// https://eslint.org/docs/latest/rules/no-obj-calls
var NoObjCallsRule = rule.Rule{
	Name: "no-obj-calls",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkCallee := func(node *ast.Node, calleeNode *ast.Node) {
			if calleeNode.Kind == ast.KindIdentifier {
				name := calleeNode.AsIdentifier().Text
				if nonCallableGlobals[name] {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedCall",
						Description: fmt.Sprintf("'%s' is not a function.", name),
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				checkCallee(node, callExpr.Expression)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkCallee(node, newExpr.Expression)
			},
		}
	},
}
