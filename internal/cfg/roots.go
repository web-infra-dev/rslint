package cfg

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsRoot reports whether node owns a control-flow graph of its own.
func IsRoot(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Initializer != nil
	}
	return false
}

// RootOf returns the code path root node executes in.
func RootOf(node *ast.Node) *ast.Node {
	previous := node
	var decorator *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			decorator = current
		}
		if current.Kind == ast.KindSourceFile {
			return current
		}
		if IsRoot(current) && runsInsideRoot(current, previous, decorator) {
			return current
		}
		previous = current
	}
	return nil
}

// runsInsideRoot reports whether a direct child of a code path root runs in
// that root rather than in the surrounding one. A member's name runs where the
// member is declared, and so do the decorators on the member and on each of its
// parameters.
func runsInsideRoot(root *ast.Node, child *ast.Node, decorator *ast.Node) bool {
	switch root.Kind {
	case ast.KindPropertyDeclaration:
		return root.AsPropertyDeclaration().Initializer == child
	case ast.KindClassStaticBlockDeclaration:
		return root.AsClassStaticBlockDeclaration().Body == child
	}
	if root.Name() == child {
		return false
	}
	if decorator != nil && (decorator.Parent == root || decorator.Parent == child) {
		return false
	}
	return true
}
