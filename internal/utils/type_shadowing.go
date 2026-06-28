package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsTypeShadowed reports whether a type reference to `name` is shadowed by a
// local type-namespace declaration at the usage site. It is the type-namespace
// counterpart of IsShadowed: it walks from node up to the SourceFile and only
// considers declarations that introduce a type binding — class, interface, type
// alias, enum, namespace/module, type parameters, and imports — never plain
// value bindings such as `var`/`let`/`const`/function parameters.
//
// This must stay syntactic and must not reuse the checker/symbol: when a local
// declaration collides with the lib `Promise` (e.g. `type Promise = string`)
// the symbols merge and point back at the lib, so only a scope walk over the
// source declarations stays correct.
func IsTypeShadowed(node *ast.Node, name string) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				return hasTypeBindingInStatements(sf.Statements.Nodes, name)
			}
			return false

		case ast.KindModuleBlock:
			block := current.AsModuleBlock()
			if block != nil && block.Statements != nil && hasTypeBindingInStatements(block.Statements.Nodes, name) {
				return true
			}

		case ast.KindBlock:
			block := current.AsBlock()
			if block != nil && block.Statements != nil && hasTypeBindingInStatements(block.Statements.Nodes, name) {
				return true
			}
		}

		// Type parameters scope to their declaration (function, class,
		// interface, type alias, method, etc.).
		if hasTypeParameterNamed(current, name) {
			return true
		}

		current = current.Parent
	}
	return false
}

// hasTypeBindingInStatements checks a statement list for a declaration that
// binds `name` in the type namespace.
func hasTypeBindingInStatements(statements []*ast.Node, name string) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}

		case ast.KindModuleDeclaration:
			// `namespace X {}` / `module X {}` introduce a type binding when
			// named by an identifier. Ambient string-named modules don't.
			modDecl := stmt.AsModuleDeclaration()
			if modDecl != nil && modDecl.Name() != nil &&
				modDecl.Name().Kind == ast.KindIdentifier && modDecl.Name().Text() == name {
				return true
			}

		case ast.KindImportEqualsDeclaration:
			importEquals := stmt.AsImportEqualsDeclaration()
			if importEquals != nil && importEquals.Name() != nil && importEquals.Name().Text() == name {
				return true
			}

		case ast.KindImportDeclaration:
			if importBindsName(stmt, name) {
				return true
			}
		}
	}
	return false
}

// importBindsName reports whether an import declaration binds `name` (default,
// namespace, or named import). Imports can introduce a type binding, so they
// count for type-namespace shadowing.
func importBindsName(stmt *ast.Node, name string) bool {
	importDecl := stmt.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return false
	}
	importClause := importDecl.ImportClause.AsImportClause()
	if importClause == nil {
		return false
	}
	if importClause.Name() != nil && importClause.Name().Text() == name {
		return true
	}
	if importClause.NamedBindings == nil {
		return false
	}
	switch importClause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		nsImport := importClause.NamedBindings.AsNamespaceImport()
		return nsImport != nil && nsImport.Name() != nil && nsImport.Name().Text() == name
	case ast.KindNamedImports:
		namedImports := importClause.NamedBindings.AsNamedImports()
		if namedImports == nil || namedImports.Elements == nil {
			return false
		}
		for _, elem := range namedImports.Elements.Nodes {
			if elem == nil {
				continue
			}
			importSpec := elem.AsImportSpecifier()
			if importSpec != nil && importSpec.Name() != nil && importSpec.Name().Text() == name {
				return true
			}
		}
	}
	return false
}

// hasTypeParameterNamed reports whether `node` declares a type parameter named
// `name`. TypeParameters() is only valid for type-parameterized declarations,
// so guard the call accordingly.
func hasTypeParameterNamed(node *ast.Node, name string) bool {
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		// always type-parameterizable
	default:
		if !ast.IsFunctionLikeDeclaration(node) {
			return false
		}
	}
	for _, tp := range node.TypeParameters() {
		if tp == nil {
			continue
		}
		if n := tp.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
			return true
		}
	}
	return false
}
