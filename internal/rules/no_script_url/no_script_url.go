package no_script_url

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-script-url
var NoScriptUrlRule = rule.Rule{
	Name: "no-script-url",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		const jsScheme = "javascript:"

		check := func(node *ast.Node) {
			value := utils.GetStaticStringValue(node)
			if len(value) >= len(jsScheme) && strings.EqualFold(value[:len(jsScheme)], jsScheme) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedScriptURL",
					Description: "Script URL is a form of eval.",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				check(node)
			},
			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				// Skip tagged templates like foo`javascript:` — the tag function
				// controls interpretation, so the string is not used as a URL.
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}
				check(node)
			},
		}
	},
}
