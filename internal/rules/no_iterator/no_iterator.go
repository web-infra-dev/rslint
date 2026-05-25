package no_iterator

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-iterator
var NoIteratorRule = rule.Rule{
	Name: "no-iterator",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noIterator",
				Description: "Reserved name '__iterator__'.",
			})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				name := node.Name()
				if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "__iterator__" {
					report(node)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				argExpr := node.AsElementAccessExpression().ArgumentExpression
				if argExpr != nil {
					val, ok := utils.GetStaticExpressionValue(argExpr)
					if ok && val == "__iterator__" {
						report(node)
					}
				}
			},
		}
	},
}
