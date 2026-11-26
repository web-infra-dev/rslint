package no_unused_modules

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// usage is what the analyzed importers use of one module: named exports,
// every export through a namespace (`import * as`, a literal `import()`), or
// every named export through `export * from`.
type usage struct {
	names     map[string]struct{}
	namespace bool
	star      bool
}

func (u *usage) addName(name string) {
	if u.names == nil {
		u.names = make(map[string]struct{})
	}
	u.names[name] = struct{}{}
}

func (u *usage) merge(other *usage) {
	for name := range other.names {
		u.addName(name)
	}
	u.namespace = u.namespace || other.namespace
	u.star = u.star || other.star
}

// uses reports whether name counts as used by another module. Upstream lets
// `export *` cover every name but default, and a namespace cover every name.
func (u *usage) uses(name string) bool {
	if u == nil {
		return false
	}
	if u.namespace || u.star && name != "default" {
		return true
	}
	_, ok := u.names[name]
	return ok
}

// moduleReference is one module request of an importer and what it uses of the
// requested module. It holds nodes of the importer only.
type moduleReference struct {
	source modules.Source
	usage  usage
}

// collectReferences summarizes what file imports, without resolving anything.
// Static declarations count at the top level only, as upstream visits only the
// Program body; `import()` and `require()` count anywhere.
func collectReferences(file *ast.SourceFile) []moduleReference {
	var references []moduleReference
	for _, statement := range file.Statements.Nodes {
		switch statement.Kind {
		case ast.KindImportDeclaration:
			declaration := statement.AsImportDeclaration()
			reference := moduleReference{source: modules.SourceFromSpecifier(declaration.ModuleSpecifier)}
			if clause := declaration.ImportClause; clause != nil {
				importClause := clause.AsImportClause()
				if importClause.Name() != nil {
					reference.usage.addName("default")
				}
				if bindings := importClause.NamedBindings; bindings != nil {
					if bindings.Kind == ast.KindNamespaceImport {
						reference.usage.namespace = true
					} else {
						for _, element := range bindings.AsNamedImports().Elements.Nodes {
							specifier := element.AsImportSpecifier()
							imported := specifier.PropertyName
							if imported == nil {
								imported = specifier.Name()
							}
							// Upstream drops `{ default as x }` because it keys
							// default imports separately; it imports default.
							reference.usage.addName(moduleExportNameText(imported))
						}
					}
				}
			}
			references = append(references, reference)
		case ast.KindImportEqualsDeclaration:
			// `import x = require('y')` binds the whole module.
			declaration := statement.AsImportEqualsDeclaration()
			if declaration.ModuleReference != nil && declaration.ModuleReference.Kind == ast.KindExternalModuleReference {
				specifier := declaration.ModuleReference.AsExternalModuleReference().Expression
				if specifier != nil && ast.IsStringLiteralLike(specifier) {
					references = append(references, moduleReference{
						source: modules.SourceFromSpecifier(specifier),
						usage:  usage{namespace: true},
					})
				}
			}
		case ast.KindExportDeclaration:
			declaration := statement.AsExportDeclaration()
			if declaration.ModuleSpecifier == nil {
				continue
			}
			reference := moduleReference{source: modules.SourceFromSpecifier(declaration.ModuleSpecifier)}
			clause := declaration.ExportClause
			if clause == nil || clause.Kind == ast.KindNamespaceExport {
				// Upstream records `export * as ns` as an `export *` dependency.
				reference.usage.star = true
			} else {
				for _, element := range clause.AsNamedExports().Elements.Nodes {
					specifier := element.AsExportSpecifier()
					local := specifier.PropertyName
					if local == nil {
						local = specifier.Name()
					}
					reference.usage.addName(moduleExportNameText(local))
				}
			}
			references = append(references, reference)
		}
	}
	if file.Flags&ast.NodeFlagsPossiblyContainsDynamicImport != 0 || mayCallRequire(file) {
		references = collectCallReferences(file, references)
	}
	return references
}

// mayCallRequire is a cheap text check that lets most ES modules skip the
// full-tree walk for require calls.
func mayCallRequire(file *ast.SourceFile) bool {
	return strings.Contains(file.Text(), "require")
}

func collectCallReferences(file *ast.SourceFile, references []moduleReference) []moduleReference {
	var calls []*ast.Node
	requireShadowed := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindCallExpression:
			calls = append(calls, node)
		case ast.KindIdentifier:
			if node.Text() == "require" && declaresName(node) {
				requireShadowed = true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	for _, node := range calls {
		call := node.AsCallExpression()
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			continue
		}
		argument := call.Arguments.Nodes[0]
		callee := ast.SkipParentheses(call.Expression)
		switch {
		case callee.Kind == ast.KindImportKeyword:
			// Only a string literal; upstream skips template literals too.
			if argument.Kind == ast.KindStringLiteral {
				references = append(references, moduleReference{
					source: modules.SourceFromCall(argument, node, modules.ModuleReferenceDynamicImport),
					usage:  usage{namespace: true},
				})
			}
		case callee.Kind == ast.KindIdentifier && callee.Text() == "require" && !requireShadowed &&
			len(call.Arguments.Nodes) == 1 && argument.Kind == ast.KindStringLiteral:
			references = append(references, moduleReference{
				source: modules.SourceFromCall(argument, node, modules.ModuleReferenceRequire),
				usage:  requireUsage(node),
			})
		}
	}
	return references
}

// requireUsage reads the names a static `require()` call uses from the
// syntax around it: destructured or accessed names, nothing for a bare
// statement, and the whole module for anything else.
func requireUsage(call *ast.Node) usage {
	node := call
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node = parent
		parent = node.Parent
	}
	var result usage
	switch {
	case parent == nil:
		result.namespace = true
	case parent.Kind == ast.KindExpressionStatement:
		// A side-effect require uses no export.
	case parent.Kind == ast.KindPropertyAccessExpression && parent.Expression() == node:
		name := parent.AsPropertyAccessExpression().Name()
		if name.Kind == ast.KindIdentifier {
			result.addName(name.Text())
		} else {
			result.namespace = true
		}
	case parent.Kind == ast.KindElementAccessExpression && parent.Expression() == node:
		argument := parent.AsElementAccessExpression().ArgumentExpression
		if argument != nil && ast.IsStringLiteralLike(argument) {
			result.addName(argument.Text())
		} else {
			result.namespace = true
		}
	case parent.Kind == ast.KindVariableDeclaration && parent.Initializer() == node &&
		parent.Name().Kind == ast.KindObjectBindingPattern:
		for _, element := range parent.Name().AsBindingPattern().Elements.Nodes {
			binding := element.AsBindingElement()
			if binding.DotDotDotToken != nil {
				result.namespace = true
				continue
			}
			key := binding.PropertyName
			if key == nil {
				key = binding.Name()
			}
			switch key.Kind {
			case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
				result.addName(key.Text())
			default:
				result.namespace = true
			}
		}
	default:
		result.namespace = true
	}
	return result
}

// declaresName reports whether identifier is the name a declaration binds,
// which makes a `require` reference in the same file possibly local.
func declaresName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression, ast.KindImportClause, ast.KindNamespaceImport,
		ast.KindImportEqualsDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
		return parent.Name() == identifier
	case ast.KindBindingElement:
		return parent.AsBindingElement().Name() == identifier
	case ast.KindImportSpecifier:
		return parent.AsImportSpecifier().Name() == identifier
	}
	return false
}
