package default_rule

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DefaultRule ensures a default export is present when a module is imported
// through a default import.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/default.js
var DefaultRule = rule.Rule{
	Name:   "import/default",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkDefault := func(node *ast.Node) {
			importDecl := node.AsImportDeclaration()
			if importDecl.ImportClause == nil {
				return
			}

			defaultSpecifier := importDecl.ImportClause.AsImportClause().Name()
			if defaultSpecifier == nil {
				return
			}

			hasDefault, ok := import_utils.HasDefaultExport(ctx, importDecl.ModuleSpecifier)
			if !ok || hasDefault {
				return
			}

			ctx.ReportNode(defaultSpecifier, rule.RuleMessage{
				Id:          "noDefault",
				Description: fmt.Sprintf("No default export found in imported module %q.", importDecl.ModuleSpecifier.Text()),
			})
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration:   checkDefault,
			ast.KindJSImportDeclaration: checkDefault,
		}
	},
}
