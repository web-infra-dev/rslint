package named

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed named.schema.json
var optionsSchema []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/named.js
var NamedRule = rule.Rule{
	Name:   "import/named",
	Schema: rule.NewSchema(optionsSchema),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		commonjs := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				commonjs, _ = option["commonjs"].(bool)
			}
		}
		check := func(source, name *ast.Node) {
			found, path := import_utils.FindExport(ctx, source, name.Text())
			if found || len(path) == 0 {
				return
			}
			message := fmt.Sprintf("%s not found in '%s'", name.Text(), source.Text())
			if len(path) > 1 {
				for i, file := range path {
					if relative, err := filepath.Rel(filepath.Dir(ctx.SourceFile.FileName()), file); err == nil {
						path[i] = relative
					}
				}
				message = fmt.Sprintf("%s not found via %s", name.Text(), strings.Join(path, " -> "))
			}
			// Upstream uses literal messages and has no message IDs.
			ctx.ReportNode(name, rule.RuleMessage{Description: message})
		}
		checkImport := func(node *ast.Node) {
			decl := node.AsImportDeclaration()
			if decl.ImportClause == nil {
				return
			}
			clause := decl.ImportClause.AsImportClause()
			if clause.IsTypeOnly() || clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
				return
			}
			for _, node := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
				if !node.AsImportSpecifier().IsTypeOnly {
					check(decl.ModuleSpecifier, node.PropertyNameOrName())
				}
			}
		}
		listeners := rule.RuleListeners{
			ast.KindImportDeclaration:   checkImport,
			ast.KindJSImportDeclaration: checkImport,
			ast.KindExportDeclaration: func(node *ast.Node) {
				decl := node.AsExportDeclaration()
				if decl.IsTypeOnly || decl.ModuleSpecifier == nil || decl.ExportClause == nil || decl.ExportClause.Kind != ast.KindNamedExports {
					return
				}
				for _, specifier := range decl.ExportClause.AsNamedExports().Elements.Nodes {
					// Upstream skips declaration-level type exports, but still
					// checks inline `export { type Name } from` specifiers.
					check(decl.ModuleSpecifier, specifier.PropertyNameOrName())
				}
			},
		}
		if commonjs {
			listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
				decl := node.AsVariableDeclaration()
				if decl.Name().Kind != ast.KindObjectBindingPattern {
					return
				}
				init := utils.ESTreeRuntimeExpression(decl.Initializer)
				if init == nil || init.Kind != ast.KindCallExpression || init.Flags&ast.NodeFlagsOptionalChain != 0 {
					return
				}
				source := import_utils.CommonJSRequireSource(init.AsCallExpression())
				if source == nil {
					return
				}
				for _, node := range decl.Name().AsBindingPattern().Elements.Nodes {
					binding := node.AsBindingElement()
					if binding.DotDotDotToken != nil {
						continue
					}
					key := node.PropertyNameOrName()
					if key.Kind == ast.KindComputedPropertyName {
						key = utils.ESTreeRuntimeExpression(key.AsComputedPropertyName().Expression)
					}
					if key.Kind == ast.KindIdentifier {
						check(source, key)
					}
				}
			}
		}
		return listeners
	},
}
