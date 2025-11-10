package no_debugger

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-debugger
var NoDebuggerRule = rule.Rule{
	Name: "no-debugger",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDebuggerStatement: func(node *ast.Node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "no-debugger",
					Description: "Unexpected 'debugger' statement.",
				})
			},
		}
	},
}
