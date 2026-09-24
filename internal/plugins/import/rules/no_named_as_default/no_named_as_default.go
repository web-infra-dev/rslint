package no_named_as_default

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-as-default.js
var NoNamedAsDefaultRule = rule.Rule{
	Name:   "import/no-named-as-default",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkDefault := func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			if declaration.ImportClause == nil {
				return
			}
			name := declaration.ImportClause.Name()
			if name == nil || name.Text() == "default" {
				return
			}

			// The shared map supplies module resolution and export metadata.
			// Parser errors in dependencies are not reproduced as rule reports.
			exports, ok := import_utils.GetExportMap(ctx, declaration.ModuleSpecifier)
			if !ok || !exports.HasDefault() || exports.Get(name.Text()) == nil {
				return
			}
			// v2.32.0 compares the immediate re-export modules, even when the
			// original exported names differ. Local aliases and stars do not
			// qualify for this exemption.
			if source := exports.ReexportSource(name.Text()); source != "" && source == exports.ReexportSource("default") {
				return
			}

			ctx.ReportNode(name, rule.RuleMessage{
				Id:          "noNamedAsDefault",
				Description: fmt.Sprintf("Using exported name '%s' as identifier for default import.", name.Text()),
			})
		}

		// The upstream ExportDefaultSpecifier listener handles Babel's
		// `export name from "module"` proposal, which tsgo does not parse.
		return rule.RuleListeners{
			ast.KindImportDeclaration:   checkDefault,
			ast.KindJSImportDeclaration: checkDefault,
		}
	},
}
