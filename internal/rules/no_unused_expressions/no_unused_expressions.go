package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func messageUnusedExpression() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedExpression",
		Description: "Expected an assignment or function call and instead saw an expression.",
	}
}

// https://eslint.org/docs/latest/rules/no-unused-expressions
var NoUnusedExpressionsRule = rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := utils.ParseNoUnusedExpressionOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				stmt := node.AsExpressionStatement()
				if stmt == nil || stmt.Expression == nil {
					return
				}

				if utils.IsDirectivePrologueStatement(node) {
					return
				}
				if utils.IsDisallowedUnusedExpression(stmt.Expression, opts) {
					ctx.ReportNode(node, messageUnusedExpression())
				}
			},
		}
	},
}
