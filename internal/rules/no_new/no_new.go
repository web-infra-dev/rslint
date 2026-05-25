package no_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-new
//
// ESLint's selector `ExpressionStatement > NewExpression` relies on ESTree
// stripping parentheses from the AST. tsgo keeps `ParenthesizedExpression`
// nodes, so walk through them with `ast.SkipParentheses` to match ESLint on
// forms like `(new Foo());`.
var NoNewRule = rule.Rule{
	Name: "no-new",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				stmt := node.AsExpressionStatement()
				if stmt == nil {
					return
				}
				expr := ast.SkipParentheses(stmt.Expression)
				if expr == nil || expr.Kind != ast.KindNewExpression {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noNewStatement",
					Description: "Do not use 'new' for side effects.",
				})
			},
		}
	},
}
