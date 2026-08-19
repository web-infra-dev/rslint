package no_continue

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-continue
var NoContinueRule = rule.Rule{
	Name:   "no-continue",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindContinueStatement: func(node *ast.Node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected use of continue statement.",
				})
			},
		}
	},
}
