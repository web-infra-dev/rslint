package utils

import "github.com/microsoft/typescript-go/shim/ast"

// IsInJsxTagName reports whether node is part of a dotted JSX tag name such as
// <Foo.Bar.Baz />. tsgo represents every dotted link as a
// PropertyAccessExpression, so inner links must climb to the outermost one
// before checking whether the JSX element owns it as its tag name.
func IsInJsxTagName(node *ast.Node) bool {
	if node == nil {
		return false
	}

	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindPropertyAccessExpression &&
		outer.Parent.AsPropertyAccessExpression().Expression == outer {
		outer = outer.Parent
	}
	return outer.Parent != nil && ast.IsJsxTagName(outer)
}
