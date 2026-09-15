package no_caller

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-caller
var NoCallerRule = rule.Rule{
	Name:   "no-caller",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			object, propName := utils.MemberExpressionParts(node)
			if object == nil || propName == nil {
				return
			}
			// Skip parentheses to handle (arguments).callee, ((arguments)).callee, etc.
			obj := ast.SkipParentheses(object)
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != "arguments" {
				return
			}

			name := propName.Text()
			if name == "callee" || name == "caller" {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: fmt.Sprintf("Avoid arguments.%s.", name),
				})
			}
		}
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: check,
			ast.KindQualifiedName:            check,
		}
	},
}
