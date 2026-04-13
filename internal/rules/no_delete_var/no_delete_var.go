package no_delete_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-delete-var
var NoDeleteVarRule = rule.Rule{
	Name: "no-delete-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				// SkipParentheses to match ESTree semantics: ESTree strips ParenthesizedExpression,
				// so ESLint sees `delete (x)` as deleting an Identifier directly.
				expr := ast.SkipParentheses(node.AsDeleteExpression().Expression)
				if expr != nil && expr.Kind == ast.KindIdentifier {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Variables should not be deleted.",
					})
				}
			},
		}
	},
}
