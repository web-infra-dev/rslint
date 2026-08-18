package scope

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsInsideOwnInitializer reports whether `location` falls inside the value a
// binding is initialized from — eslint-scope's `isInInitializer` walk, which
// answers "is this reference evaluated while its own binding is still being
// set up":
//
//	var a = a
//	var [a = a] = list
//	var {a = a} = obj
//	function f(a = a) {}
//	for (var a in a) {}
//	for (var a of a) {}
//
// `declarationID` is the identifier that binds the name. The walk climbs from
// it looking for the construct that supplies its value, and stops at the node
// kinds ESLint calls SENTINEL_TYPE — the boundaries a value cannot be
// initialized across.
//
// Callers are responsible for the scope-level preconditions their rule needs
// (for example "the reference and the declaration share a scope"): this
// function is purely positional.
func IsInsideOwnInitializer(declarationID *ast.Node, location int) bool {
	if declarationID == nil {
		return false
	}
	for node := declarationID.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindVariableDeclaration:
			declarator := node.AsVariableDeclaration()
			if declarator != nil && rangeContains(declarator.Initializer, location) {
				return true
			}
			// `for (var a in a)` / `for (var a of a)`: the iterated expression
			// is evaluated before the binding is initialized.
			if node.Parent != nil && node.Parent.Parent != nil {
				loop := node.Parent.Parent
				if loop.Kind == ast.KindForInStatement || loop.Kind == ast.KindForOfStatement {
					if forInOrOf := loop.AsForInOrOfStatement(); forInOrOf != nil &&
						rangeContains(forInOrOf.Expression, location) {
						return true
					}
				}
			}
			return false

		case ast.KindBindingElement:
			// ESTree's AssignmentPattern inside a destructuring pattern.
			if element := node.AsBindingElement(); element != nil && rangeContains(element.Initializer, location) {
				return true
			}

		case ast.KindParameter:
			// ESTree's AssignmentPattern in a parameter list.
			if rangeContains(node.Initializer(), location) {
				return true
			}

		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindArrowFunction, ast.KindCatchClause,
			ast.KindImportDeclaration, ast.KindExportDeclaration,
			// ESTree wraps every shorthand method and accessor body in a
			// FunctionExpression, which the walk stops at. tsgo has no such
			// wrapper — the member node *is* the function — so these kinds
			// stand in for it. Without them the walk escapes the method and
			// finds the enclosing `const x = { m(node) { node; } }`
			// initializer, treating every parameter read inside as a
			// self-reference.
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
	}
	return false
}

// rangeContains reports whether `location` falls within `node`'s source range.
// A nil node contains nothing.
func rangeContains(node *ast.Node, location int) bool {
	return node != nil && node.Pos() <= location && location <= node.End()
}
