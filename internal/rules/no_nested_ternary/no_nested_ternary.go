package no_nested_ternary

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-nested-ternary
var NoNestedTernaryRule = rule.Rule{
	Name: "no-nested-ternary",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				cond := node.AsConditionalExpression()
				// ESTree strips parentheses; unwrap to match ESLint's AST view.
				if ast.IsConditionalExpression(ast.SkipParentheses(cond.WhenTrue)) ||
					ast.IsConditionalExpression(ast.SkipParentheses(cond.WhenFalse)) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noNestedTernary",
						Description: "Do not nest ternary expressions.",
					})
				}
			},
		}
	},
}
