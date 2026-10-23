package require_dynamic_import_entry

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const reviewRequireDescription = "TypeScript does not collect this require() dependency for Rslim. Review every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason."

var RequireDynamicImportEntryRule = rule.Rule{
	Name:   "rslim/require-dynamic-import-entry",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				expression := call.Expression
				if expression.Kind == ast.KindImportKeyword {
					if len(call.Arguments.Nodes) > 0 && ast.IsStringLiteralLike(call.Arguments.Nodes[0]) {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "reviewDynamicImport",
						Description: "TypeScript does not collect this import() dependency because its path is not a string literal or a template literal without substitutions. Review every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason.",
					})
				} else if expression.Kind == ast.KindIdentifier && expression.Text() == "require" {
					// TypeScript collects literal require calls only in JavaScript files.
					if ast.IsInJSFile(node) && ast.IsRequireCall(node, true) {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "reviewRequire",
						Description: reviewRequireDescription,
					})
				}
			},
		}
	},
}
