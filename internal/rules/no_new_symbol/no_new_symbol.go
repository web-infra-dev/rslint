package no_new_symbol

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-new-symbol
var NoNewSymbolRule = rule.Rule{
	Name: "no-new-symbol",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				if node == nil {
					return
				}

				expr := node.Expression()
				if expr == nil || expr.Kind != ast.KindIdentifier {
					return
				}

				if expr.Text() == "Symbol" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noNewSymbol",
						Description: "`Symbol` cannot be called as a constructor.",
					})
				}
			},
		}
	},
}
