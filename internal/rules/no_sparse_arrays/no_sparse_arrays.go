package no_sparse_arrays

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-sparse-arrays
var NoSparseArraysRule = rule.Rule{
	Name: "no-sparse-arrays",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				for _, v := range node.AsArrayLiteralExpression().Elements.Nodes {
					if v.Kind == ast.KindOmittedExpression {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "unexpectedSparseArray",
							Description: "Unexpected comma in middle of array.",
						})
					}
				}
			},
		}
	},
}
