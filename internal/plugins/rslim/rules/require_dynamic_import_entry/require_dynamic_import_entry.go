package require_dynamic_import_entry

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var RequireDynamicImportEntryRule = rule.Rule{
	Name:   "rslim/require-dynamic-import-entry",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if node.AsCallExpression().Expression.Kind != ast.KindImportKeyword {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "reviewDynamicImport",
					Description: "Review this dynamic import for Rslim: identify every module Rspack may load, including module side effects. Add source files whose runtime code may be removed to Rslim entries. Once all targets are safe, suppress this call with rslint-disable-next-line rslim/require-dynamic-import-entry and a reason.",
				})
			},
		}
	},
}
