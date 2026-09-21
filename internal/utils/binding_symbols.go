package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// BindingNameSymbol returns the binder symbol of a declared binding name. The
// binder attaches the symbol to the declaration node owning the name — the
// declaration itself (VariableDeclaration, Parameter, ImportSpecifier, ...)
// for a plain identifier, the enclosing BindingElement for names nested in a
// destructuring pattern. Binder symbols are the currency of ctx.Refs, so use
// them on the declaration side too. Returns nil when ident is not its
// parent's declared name.
func BindingNameSymbol(ident *ast.Node) *ast.Symbol {
	if ident == nil || ident.Parent == nil || ident.Parent.Name() != ident {
		return nil
	}
	declaration := ident.Parent
	if declaration.Parent != nil && ast.IsParameterPropertyDeclaration(declaration, declaration.Parent) && ast.IsIdentifier(ident) {
		// A parameter property stores its class member symbol on the declaration.
		// References to the parameter use the constructor's local symbol instead.
		return declaration.Parent.Locals()[ident.Text()]
	}
	return declaration.Symbol()
}

// BindingSymbols returns the binder symbol of every name bound by a
// declaration's name node: the lone symbol of a plain identifier, or one per
// name nested in a destructuring pattern.
func BindingSymbols(nameNode *ast.Node) []*ast.Symbol {
	var symbols []*ast.Symbol
	CollectBindingNames(nameNode, func(ident *ast.Node, _ string) {
		if sym := BindingNameSymbol(ident); sym != nil {
			symbols = append(symbols, sym)
		}
	})
	return symbols
}
