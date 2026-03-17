package no_new_wrappers

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-new-wrappers
var NoNewWrappersRule = rule.Rule{
	Name: "no-new-wrappers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		wrapperObjects := map[string]bool{
			"String":  true,
			"Number":  true,
			"Boolean": true,
		}

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil {
					return
				}

				callee := newExpr.Expression
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}

				name := callee.Text()
				if wrapperObjects[name] {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noConstructor",
						Description: fmt.Sprintf("Do not use %s as a constructor.", name),
					})
				}
			},
		}
	},
}
