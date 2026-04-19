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

// GetJsxParentElement returns the JsxOpeningElement or JsxSelfClosingElement that
// owns the given JsxAttribute (or JsxSpreadAttribute), or nil if not applicable.
func GetJsxParentElement(attr *ast.Node) *ast.Node {
	if attr == nil || attr.Parent == nil {
		return nil
	}
	grandParent := attr.Parent.Parent
	if grandParent == nil {
		return nil
	}
	switch grandParent.Kind {
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return grandParent
	}
	return nil
}

// GetJsxTagName returns the tag-name node of a JsxOpeningElement or
// JsxSelfClosingElement, or nil for other kinds.
func GetJsxTagName(element *ast.Node) *ast.Node {
	if element == nil {
		return nil
	}
	switch element.Kind {
	case ast.KindJsxOpeningElement:
		return element.AsJsxOpeningElement().TagName
	case ast.KindJsxSelfClosingElement:
		return element.AsJsxSelfClosingElement().TagName
	}
	return nil
}

// IsDOMComponent reports whether a JSX opening/self-closing element refers to
// an intrinsic (DOM) element like <div> or <svg:path>, rather than a user
// component like <Foo> or <Foo.Bar>.
//
// Mirrors ESLint-plugin-react's `jsxUtil.isDOMComponent`: a tag name is
// intrinsic iff `elementType(node)` starts with a lowercase letter (regex
// `/^[a-z]/`). For member-expression tags (`<foo.bar>`, `<this.Foo>`) this
// means the classification is decided by the leftmost base identifier's
// first character — so `<foo.bar>` is DOM (matches ESLint, even though the
// runtime React behavior is "always user component"), while `<Foo.Bar>` is
// a user component.
func IsDOMComponent(element *ast.Node) bool {
	tagName := GetJsxTagName(element)
	if tagName == nil {
		return false
	}
	var text string
	switch tagName.Kind {
	case ast.KindIdentifier:
		text = tagName.AsIdentifier().Text
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns.Namespace == nil || ns.Name() == nil {
			return false
		}
		text = ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		// Walk to the leftmost base — its first character decides the
		// classification, matching `/^[a-z]/.test(elementType(node))`.
		base := tagName
		for base.Kind == ast.KindPropertyAccessExpression {
			base = base.AsPropertyAccessExpression().Expression
		}
		switch base.Kind {
		case ast.KindIdentifier:
			text = base.AsIdentifier().Text
		case ast.KindThisKeyword:
			// `<this.Foo>` — jsx-ast-utils' elementType returns "this.Foo",
			// first char is lowercase → DOM per ESLint's regex.
			text = "this"
		default:
			return false
		}
	default:
		return false
	}
	if text == "" {
		return false
	}
	first := text[0]
	return first >= 'a' && first <= 'z'
}
