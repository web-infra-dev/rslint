package no_template_curly_in_string

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-template-curly-in-string
var NoTemplateCurlyInString = rule.Rule{
	Name: "no-template-curly-in-string",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		regex := regexp.MustCompile(`\$\{[^}]+\}`)
		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				expr := node.AsStringLiteral()
				if regex.MatchString(expr.Text) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedTemplateExpression",
						Description: "Template literal placeholder syntax in regular strings is not allowed.",
					})
				}
			},
		}
	},
}
