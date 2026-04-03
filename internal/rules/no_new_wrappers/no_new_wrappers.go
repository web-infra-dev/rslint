package no_new_wrappers

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-new-wrappers
var wrapperObjects = map[string]bool{
	"String":  true,
	"Number":  true,
	"Boolean": true,
}

var NoNewWrappersRule = rule.Rule{
	Name: "no-new-wrappers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
				if !wrapperObjects[name] {
					return
				}

				// If the name is shadowed by a local declaration, it's not the
				// global built-in — skip reporting.
				if utils.IsShadowed(callee, name) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noConstructor",
					Description: fmt.Sprintf("Do not use %s as a constructor.", name),
				})
			},
		}
	},
}
