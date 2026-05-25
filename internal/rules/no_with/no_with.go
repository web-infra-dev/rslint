package no_with

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-with
var NoWithRule = rule.Rule{
	Name: "no-with",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindWithStatement: func(node *ast.Node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedWith",
					Description: "Unexpected use of 'with' statement.",
				})
			},
		}
	},
}
