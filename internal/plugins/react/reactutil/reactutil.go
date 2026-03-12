package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsCreateElementCall checks if the callee is React.createElement.
func IsCreateElementCall(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "createElement" {
		return false
	}
	if prop.Expression.Kind != ast.KindIdentifier || prop.Expression.AsIdentifier().Text != "React" {
		return false
	}
	return true
}

// GetJsxPropName returns the display name of a JSX node.
// For JsxAttribute: returns the attribute name (including namespaced names like "foo:bar").
// For JsxSpreadAttribute: returns "spread".
// For Identifier nodes (e.g. tag names): returns the identifier text.
// For unknown nodes: returns "".
func GetJsxPropName(node *ast.Node) string {
	if ast.IsJsxAttribute(node) {
		nameNode := node.AsJsxAttribute().Name()
		if nameNode.Kind == ast.KindIdentifier {
			return nameNode.AsIdentifier().Text
		}
		if nameNode.Kind == ast.KindJsxNamespacedName {
			ns := nameNode.AsJsxNamespacedName()
			return ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
		}
	}
	if ast.IsJsxSpreadAttribute(node) {
		return "spread"
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	return ""
}
