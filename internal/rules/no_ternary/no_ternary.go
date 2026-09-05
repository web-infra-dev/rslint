package no_ternary

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-ternary
var NoTernaryRule = rule.Rule{
	Name:   "no-ternary",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noTernaryOperator",
					Description: "Ternary operator used.",
				})
			},
		}
	},
}
