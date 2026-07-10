package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func unusedExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedExpression",
		Description: "Expected an assignment or function call and instead saw an expression.",
	}
}

// https://typescript-eslint.io/rules/no-unused-expressions
var NoUnusedExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := utils.ParseNoUnusedExpressionOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				exprStmt := node.AsExpressionStatement()
				if exprStmt == nil || exprStmt.Expression == nil {
					return
				}
				expr := exprStmt.Expression

				if !utils.IsDisallowedUnusedExpression(expr, opts) {
					return
				}

				if utils.IsDirectivePrologueStatementIncludingClassStaticBlocks(node) {
					return
				}

				ctx.ReportNode(node, unusedExpressionMessage())
			},
		}
	},
})
