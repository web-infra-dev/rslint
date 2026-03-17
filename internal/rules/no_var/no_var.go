package no_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-var
var NoVarRule = rule.Rule{
	Name: "no-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt == nil || varStmt.DeclarationList == nil {
					return
				}

				// BlockScoped = Let | Const | Using
				// If none of those flags are set, it's a var declaration.
				if varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0 {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedVar",
					Description: "Unexpected var, use let or const instead.",
				})
			},
		}
	},
}
