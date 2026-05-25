package no_throw_literal

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-throw-literal
var NoThrowLiteralRule = rule.Rule{
	Name: "no-throw-literal",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				throwStmt := node.AsThrowStatement()
				if throwStmt == nil || throwStmt.Expression == nil {
					return
				}
				expr := throwStmt.Expression

				if !utils.CouldBeError(expr) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "object",
						Description: "Expected an error object to be thrown.",
					})
					return
				}

				if utils.IsUndefinedIdentifier(expr) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "undef",
						Description: "Do not throw undefined.",
					})
				}
			},
		}
	},
}
