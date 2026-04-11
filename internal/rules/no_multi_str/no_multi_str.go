package no_multi_str

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-multi-str
var NoMultiStrRule = rule.Rule{
	Name: "no-multi-str",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				// Skip string literals inside JSX contexts.
				// Possible JSX parents for a StringLiteral:
				//   - JsxAttribute:  <div attr="value">
				//   - JsxExpression: <div>{'value'}</div> or <div attr={'value'}>
				if node.Parent != nil &&
					(ast.IsJsxAttributeLike(node.Parent) || ast.IsJsxExpression(node.Parent)) {
					return
				}

				raw := utils.TrimmedNodeText(ctx.SourceFile, node)

				// A line break in the raw source of a string literal means it uses
				// the backslash-newline continuation syntax (e.g. 'line1 \<LF> line2').
				if strings.ContainsAny(raw, "\n\r\u2028\u2029") {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "multilineString",
						Description: "Multiline support is limited to browsers supporting ES5 only.",
					})
				}
			},
		}
	},
}
