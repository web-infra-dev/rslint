package scope

import "github.com/microsoft/typescript-go/shim/ast"

// Acquire returns the innermost scope associated with node's ESTree position.
// This is a syntactic query like SourceCode.getScope, not reference resolution:
// it does not skip type-only definitions or function bodies for parameter
// defaults. The index is built lazily for consumers that need this query.
func (m *Manager) Acquire(node *ast.Node) *Scope {
	if m.byBlock == nil {
		m.byBlock = make(map[*ast.Node]*Scope, len(m.Scopes))
		for _, current := range m.Scopes {
			if current.Block != nil {
				// A named function expression owns both a name scope and a
				// function scope. The later, inner scope is acquired first.
				m.byBlock[current.Block] = current
			}
		}
	}
	var child *ast.Node
	for current := node; current != nil; current = current.Parent {
		if acquired := m.byBlock[current]; acquired != nil && !outsideRuntimeMethodScope(current, child) {
			return acquired
		}
		child = current
	}
	return m.Global
}

// ESTree methods wrap a FunctionExpression. Their computed keys and member
// decorators sit outside that function, while parameters and their decorators
// sit inside. tsgo combines both parts into one runtime method node.
func outsideRuntimeMethodScope(method, child *ast.Node) bool {
	if child == nil {
		return false
	}
	switch method.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return child == method.Name() || child.Kind == ast.KindDecorator
	default:
		return false
	}
}
