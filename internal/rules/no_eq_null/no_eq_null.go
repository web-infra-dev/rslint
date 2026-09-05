package no_eq_null

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoEqNullRule disallows null comparisons without type-checking operators.
// https://eslint.org/docs/latest/rules/no-eq-null
var NoEqNullRule = rule.Rule{
	Name:   "no-eq-null",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				if !utils.IsLooseEqualityOperator(binary.OperatorToken.Kind) {
					return
				}

				if utils.IsNullLiteral(binary.Left) || utils.IsNullLiteral(binary.Right) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Use '===' to compare with null.",
					})
				}
			},
		}
	},
}
