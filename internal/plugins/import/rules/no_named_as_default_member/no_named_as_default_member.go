package no_named_as_default_member

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-as-default-member.js
var NoNamedAsDefaultMemberRule = rule.Rule{
	Name:   "import/no-named-as-default-member",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		type defaultImport struct {
			source string
			names  map[string]bool
		}
		imports := make(map[string]defaultImport)
		// Seed imports before visiting expressions so uses preceding an import
		// are checked too. Upstream matches names, including shadowed bindings.
		for _, statement := range ctx.SourceFile.Statements.Nodes {
			if !ast.IsImportDeclarationOrJSImportDeclaration(statement) {
				continue
			}
			declaration := statement.AsImportDeclaration()
			if declaration.ImportClause == nil || declaration.ImportClause.Name() == nil {
				continue
			}
			names := import_utils.GetLocalExportNames(ctx, declaration.ModuleSpecifier)
			// The default property is always allowed, so a default-only module
			// cannot produce a diagnostic and needs no expression listeners.
			if len(names) == 0 || (len(names) == 1 && names[0] == "default") {
				continue
			}
			imported := defaultImport{source: declaration.ModuleSpecifier.Text(), names: make(map[string]bool, len(names))}
			for _, name := range names {
				imported.names[name] = true
			}
			imports[declaration.ImportClause.Name().Text()] = imported
		}
		if len(imports) == 0 {
			return nil
		}

		checkProperty := func(object, property, report *ast.Node) {
			object = utils.ESTreeRuntimeExpression(object)
			property = utils.ESTreeRuntimeExpression(property)
			if object == nil || object.Kind != ast.KindIdentifier || property == nil {
				return
			}
			// Match ESTree's .name, not a statically evaluated property value:
			// obj[key] is checked, whereas obj['key'] is not.
			if property.Kind != ast.KindIdentifier && property.Kind != ast.KindPrivateIdentifier {
				return
			}
			name := strings.TrimPrefix(property.Text(), "#")
			imported, ok := imports[object.Text()]
			if !ok || name == "default" || !imported.names[name] {
				return
			}
			ctx.ReportNode(report, rule.RuleMessage{
				Id: "noNamedAsDefaultMember",
				Description: fmt.Sprintf("Caution: `%s` also has a named export `%s`. Check if you meant to write `import {%s} from '%s'` instead.",
					object.Text(), name, name, imported.source),
			})
		}
		checkAccess := func(node *ast.Node) {
			if utils.IsInJsxTagName(node) {
				return
			}
			object, property := utils.MemberExpressionParts(node)
			checkProperty(object, property, node)
		}
		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkAccess,
			ast.KindElementAccessExpression:  checkAccess,
			ast.KindQualifiedName:            checkAccess,
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration.Name().Kind != ast.KindObjectBindingPattern || declaration.Initializer == nil {
					return
				}
				for _, element := range declaration.Name().AsBindingPattern().Elements.Nodes {
					binding := element.AsBindingElement()
					if binding.DotDotDotToken != nil {
						continue
					}
					key := ast.TryGetPropertyNameOfBindingOrAssignmentElement(element)
					if key != nil && key.Kind == ast.KindComputedPropertyName {
						key = key.AsComputedPropertyName().Expression
					}
					key = utils.ESTreeRuntimeExpression(key)
					checkProperty(declaration.Initializer, key, key)
				}
			},
		}
	},
}
