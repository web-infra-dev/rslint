package no_process_exit

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var message = rule.RuleMessage{
	Id:          "noProcessExit",
	Description: "Don't use process.exit(); throw an error instead.",
}

var NoProcessExitRule = rule.Rule{
	Name:   "node/no-process-exit",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				obj, name := utils.MemberExpressionParts(callee)
				obj = utils.ESTreeRuntimeExpression(obj)
				if obj == nil || obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != "process" {
					return
				}
				name = utils.ESTreeRuntimeExpression(name)
				if name == nil {
					return
				}
				// Upstream matches property.name without excluding computed or private names.
				if name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "exit" ||
					name.Kind == ast.KindPrivateIdentifier && name.AsPrivateIdentifier().Text == "#exit" {
					ctx.ReportNode(node, message)
				}
			},
		}
	},
}
