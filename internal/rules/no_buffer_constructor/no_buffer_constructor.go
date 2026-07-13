package no_buffer_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-buffer-constructor
var NoBufferConstructorRule = rule.Rule{
	Name: "no-buffer-constructor",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.SkipAssertionsAndParens(call.Expression)
				if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Buffer" {
					if declared, ok := ctx.Globals["Buffer"]; ok && !declared {
						return
					}
					if !utils.IsShadowed(callee, "Buffer") {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "deprecated",
							Description: "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						})
					}
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				callee := utils.SkipAssertionsAndParens(newExpr.Expression)
				if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Buffer" {
					if declared, ok := ctx.Globals["Buffer"]; ok && !declared {
						return
					}
					if !utils.IsShadowed(callee, "Buffer") {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "deprecated",
							Description: "new Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						})
					}
				}
			},
		}
	},
}
