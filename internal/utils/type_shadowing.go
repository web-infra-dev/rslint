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

		case ast.KindMappedType:
			mappedType := current.AsMappedTypeNode()
			if mappedType != nil &&
				typeParameterBindsName(mappedType.TypeParameter, name) &&
				mappedType.TypeParameter != node &&
				!mappedType.TypeParameter.Contains(node) {
				return true
			}

		case ast.KindInferType:
			inferType := current.AsInferTypeNode()
			if inferType != nil &&
				typeParameterBindsName(inferType.TypeParameter, name) &&
				inferType.TypeParameter != node &&
				!inferType.TypeParameter.Contains(node) {
				return true
			}

		case ast.KindConditionalType:
			conditionalType := current.AsConditionalTypeNode()
			if conditionalType != nil &&
				conditionalType.TrueType != nil &&
				conditionalType.TrueType.Contains(node) &&
				containsInferTypeBinding(conditionalType.ExtendsType, name) {
				return true
			}

		case ast.KindClassExpression:
			if classExpressionBindsName(current, name) {
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
			if moduleDeclarationBindsName(stmt, name) {
				return true
			}

		case ast.KindImportEqualsDeclaration:
			if importEqualsDeclarationBindsName(stmt, name) {
				return true
			}

		case ast.KindImportDeclaration:
			if importDeclarationBindsName(stmt, name) {
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
		if !ast.IsFunctionLike(node) {
			return false
		}
	}
	for _, tp := range node.TypeParameters() {
		if typeParameterBindsName(tp, name) {
			return true
		}
	}
	return false
}

func typeParameterBindsName(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	typeParameter := node.AsTypeParameterDeclaration()
	return typeParameter != nil &&
		typeParameter.Name() != nil &&
		typeParameter.Name().Kind == ast.KindIdentifier &&
		typeParameter.Name().Text() == name
}

func classExpressionBindsName(node *ast.Node, name string) bool {
	classExpression := node.AsClassExpression()
	return classExpression != nil &&
		classExpression.Name() != nil &&
		classExpression.Name().Kind == ast.KindIdentifier &&
		classExpression.Name().Text() == name
}

func containsInferTypeBinding(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionType, ast.KindConstructorType, ast.KindConditionalType:
		return false
	case ast.KindInferType:
		inferType := node.AsInferTypeNode()
		return inferType != nil && typeParameterBindsName(inferType.TypeParameter, name)
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsInferTypeBinding(child, name) {
			found = true
			return true
		}
		return false
	})
	return found
}
