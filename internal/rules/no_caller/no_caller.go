package no_caller

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-caller
var NoCallerRule = rule.Rule{
	Name: "no-caller",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propAccess := node.AsPropertyAccessExpression()
				if propAccess == nil {
					return
				}

				// Skip parentheses to handle (arguments).callee, ((arguments)).callee, etc.
				obj := ast.SkipParentheses(propAccess.Expression)
				if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != "arguments" {
					return
				}

				propName := propAccess.Name()
				if propName == nil {
					return
				}
				name := propName.Text()
				if name == "callee" || name == "caller" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: fmt.Sprintf("Avoid arguments.%s.", name),
					})
				}
			},
		}
	},
}
