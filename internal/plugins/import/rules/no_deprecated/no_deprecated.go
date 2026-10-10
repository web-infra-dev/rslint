package no_deprecated

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-deprecated.js
var NoDeprecatedRule = rule.Rule{
	Name:   "import/no-deprecated",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		deprecated := make(map[string]*import_utils.Deprecation)
		namespaces := make(map[string]*import_utils.ExportMap)
		type report struct {
			node        *ast.Node
			deprecation *import_utils.Deprecation
		}
		imports := make(map[*ast.Node][]report)
		// Seed before walking references, including those before the import.
		for _, stmt := range ctx.SourceFile.Statements.Nodes {
			if !ast.IsImportDeclarationOrJSImportDeclaration(stmt) {
				continue
			}
			decl := stmt.AsImportDeclaration()
			exports, ok := import_utils.GetExportMap(ctx, modules.SourceFromSpecifier(decl.ModuleSpecifier))
			if !ok {
				continue
			}
			if dep := exports.Deprecation(ctx); dep != nil {
				imports[stmt] = append(imports[stmt], report{stmt, dep})
			}
			if decl.ImportClause == nil {
				continue
			}
			clause := decl.ImportClause.AsImportClause()
			capture := func(imported string, local, specifier *ast.Node) {
				meta := exports.Get(imported)
				if meta == nil && imported == "default" {
					if dep := exports.DefaultDeprecation(ctx); dep != nil {
						deprecated[local.Text()] = dep
						imports[stmt] = append(imports[stmt], report{specifier, dep})
					}
					return
				}
				if meta == nil {
					return
				}
				if meta.Namespace != nil {
					namespaces[local.Text()] = meta.Namespace
				}
				if dep := meta.Deprecation(ctx); dep != nil {
					deprecated[local.Text()] = dep
					imports[stmt] = append(imports[stmt], report{specifier, dep})
				}
			}
			if clause.Name() != nil {
				capture("default", clause.Name(), clause.Name())
			}
			if clause.NamedBindings == nil {
				continue
			}
			switch clause.NamedBindings.Kind {
			case ast.KindNamespaceImport:
				if exports.Size() != 0 {
					namespaces[clause.NamedBindings.Name().Text()] = exports
				}
			case ast.KindNamedImports:
				for _, spec := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					imported := spec.AsImportSpecifier().PropertyName
					if imported == nil {
						imported = spec.Name()
					}
					// v2.32.0 uses ESTree imported.name, not a string-literal value.
					if imported.Kind == ast.KindIdentifier {
						capture(imported.Text(), spec.Name(), spec)
					}
				}
			}
		}
		names := make(map[string]struct{}, len(deprecated)+len(namespaces))
		for name := range deprecated {
			names[name] = struct{}{}
		}
		for name := range namespaces {
			names[name] = struct{}{}
		}
		if len(names) == 0 && len(imports) == 0 {
			return nil
		}
		scopes := scopeanalysis.References(ctx, names)
		moduleScope := func(node *ast.Node, name string) bool {
			// declaredScope inspects the first same-name reference in the active
			// scope, rather than resolving the current identifier. Preserve that
			// behavior for non-reference identifiers such as object property keys.
			current := scopes.Acquire(node)
			if current == nil {
				return false
			}
			for _, reference := range current.References {
				if reference.Identifier.Text() != name {
					continue
				}
				resolved := reference.Resolved()
				return resolved != nil && resolved.Scope == scopes.Global
			}
			return false
		}
		checkImport := func(node *ast.Node) {
			for _, item := range imports[node] {
				ctx.ReportNode(item.node, message(item.deprecation))
			}
		}
		// Member listeners precede receiver identifiers. Emit at the property
		// visit so a deprecated namespace receiver still reports before its member.
		memberDeprecations := make(map[*ast.Node][]*import_utils.Deprecation)
		checkMember := func(node *ast.Node) {
			if utils.IsInJsxTagName(node) {
				return
			}
			object, _ := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			if object == nil || object.Kind != ast.KindIdentifier {
				return
			}
			namespace := namespaces[object.Text()]
			if namespace == nil || !moduleScope(node, object.Text()) {
				return
			}
			for current := node; namespace != nil && current != nil; current = utils.ESTreeParent(current) {
				receiver, property := utils.MemberExpressionParts(current)
				if receiver == nil || property == nil || current.Kind == ast.KindElementAccessExpression {
					return
				}
				// PrivateIdentifier.name excludes the source # prefix.
				if property.Kind != ast.KindIdentifier && property.Kind != ast.KindPrivateIdentifier {
					return
				}
				meta := namespace.Get(strings.TrimPrefix(property.Text(), "#"))
				if meta == nil {
					return
				}
				if dep := meta.Deprecation(ctx); dep != nil {
					memberDeprecations[property] = append(memberDeprecations[property], dep)
				}
				namespace = meta.Namespace
			}
		}
		return rule.RuleListeners{
			ast.KindImportDeclaration:   checkImport,
			ast.KindJSImportDeclaration: checkImport,
			ast.KindIdentifier: func(node *ast.Node) {
				for _, dep := range memberDeprecations[node] {
					ctx.ReportNode(node, message(dep))
				}
				dep := deprecated[node.Text()]
				if dep == nil || utils.IsInJsxTagName(node) || utils.IsJSDocSyntaxNode(node) {
					return
				}
				parent := utils.ESTreeParent(node)
				if parent == nil {
					return
				}
				_, property := utils.MemberExpressionParts(parent)
				if utils.ESTreeRuntimeExpression(property) == node {
					return
				}
				switch parent.Kind {
				case ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportClause, ast.KindImportEqualsDeclaration, ast.KindJsxAttribute, ast.KindJsxNamespacedName:
					return
				}
				if moduleScope(node, node.Text()) {
					ctx.ReportNode(node, message(dep))
				}
			},
			ast.KindPrivateIdentifier: func(node *ast.Node) {
				for _, dep := range memberDeprecations[node] {
					ctx.ReportNode(node, message(dep))
				}
			},
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindElementAccessExpression:  checkMember,
			ast.KindQualifiedName:            checkMember,
		}
	},
}

func message(deprecation *import_utils.Deprecation) rule.RuleMessage {
	text := "Deprecated."
	if deprecation.Description != "" {
		text = "Deprecated: " + deprecation.Description
	}
	return rule.RuleMessage{Id: "deprecated", Description: text}
}
