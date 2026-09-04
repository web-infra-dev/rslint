package no_process_exit

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var message = rule.RuleMessage{
	Id:          "noProcessExit",
	Description: "Don't use process.exit(); throw an error instead.",
}

var NoProcessExitRule = rule.Rule{
	Name:   "n/no-process-exit",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				obj := ast.SkipParentheses(prop.Expression)
				if obj == nil || obj.Kind != ast.KindIdentifier {
					return
				}
				if obj.AsIdentifier().Text != "process" {
					return
				}
				name := prop.Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "exit" {
					return
				}
				ctx.ReportNode(node, message)
			},
		}
	},
}
