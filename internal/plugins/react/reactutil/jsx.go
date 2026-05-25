package reactutil

import "github.com/microsoft/typescript-go/shim/ast"

// GetJsxTagBaseIdentifier returns the leftmost Identifier of a JSX tag-name
// node — i.e. the symbol a rule must resolve to classify the tag. Pass the
// tag-name node obtained from `GetJsxTagName` (or directly from
// `JsxOpeningElement.TagName` / `JsxSelfClosingElement.TagName`). Returns nil
// when the tag does not terminate in an Identifier (ThisKeyword base,
// JsxNamespacedName, unknown shape).
//
// Shapes handled:
//
//   - `<Foo />`                 → Identifier("Foo")
//   - `<Foo.Bar />`             → Identifier("Foo")
//   - `<Foo.Bar.Baz />`         → Identifier("Foo")
//   - `<this />` / `<this.X />` → nil (ThisKeyword base)
//   - `<a:b />`                 → nil (JsxNamespacedName — not an identifier
//     reference in any scope)
//   - `<foo-bar />`             → Identifier("foo-bar") (tsgo preserves the
//     hyphenated text verbatim; callers decide whether that's DOM).
func GetJsxTagBaseIdentifier(tagName *ast.Node) *ast.Node {
	if tagName == nil {
		return nil
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName
	case ast.KindPropertyAccessExpression:
		base := tagName
		for base.Kind == ast.KindPropertyAccessExpression {
			base = base.AsPropertyAccessExpression().Expression
		}
		if base.Kind == ast.KindIdentifier {
			return base
		}
	}
	return nil
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

// GetJsxChildren returns the child-node list of a JsxElement or JsxFragment,
// or nil for other kinds (JsxSelfClosingElement has no children) and when the
// container's child list is absent.
func GetJsxChildren(parent *ast.Node) []*ast.Node {
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindJsxElement:
		if parent.AsJsxElement().Children == nil {
			return nil
		}
		return parent.AsJsxElement().Children.Nodes
	case ast.KindJsxFragment:
		if parent.AsJsxFragment().Children == nil {
			return nil
		}
		return parent.AsJsxFragment().Children.Nodes
	}
	return nil
}

// GetJsxElementAttributes returns the attribute nodes of a JsxOpeningElement or
// JsxSelfClosingElement, or nil for other kinds or when the element has no
// attributes. Each returned node is either a JsxAttribute or a JsxSpreadAttribute.
func GetJsxElementAttributes(element *ast.Node) []*ast.Node {
	if element == nil {
		return nil
	}
	var attrs *ast.Node
	switch element.Kind {
	case ast.KindJsxOpeningElement:
		attrs = element.AsJsxOpeningElement().Attributes
	case ast.KindJsxSelfClosingElement:
		attrs = element.AsJsxSelfClosingElement().Attributes
	default:
		return nil
	}
	if attrs == nil {
		return nil
	}
	list := attrs.AsJsxAttributes()
	if list == nil || list.Properties == nil {
		return nil
	}
	return list.Properties.Nodes
}

// GetJsxElementTypeString returns the jsx-ast-utils `elementType(node)`
// equivalent — the dotted / namespaced display string of a JSX tag name as
// an ESTree-compatible source caller would see it. `node` may be either a
// JsxOpeningElement / JsxSelfClosingElement, or a raw tag-name node. Returns
// "" for shapes that don't correspond to a legal React/JSX element type
// (e.g. a computed member access), so callers can treat "" as "not a user
// component".
//
// Supported tag shapes:
//
//   - `<Foo>` / `<foo>`       → "Foo" / "foo"
//   - `<Foo.Bar.Baz>`         → "Foo.Bar.Baz" (PropertyAccessExpression chain)
//   - `<this.Foo>`            → "this.Foo" (ThisKeyword base)
//   - `<ns:Name>`             → "ns:Name" (JsxNamespacedName)
//
// This is AST-driven — interior whitespace or comments in unusual forms
// (e.g. `<Foo . Bar />`) are normalized away, matching jsx-ast-utils.
func GetJsxElementTypeString(node *ast.Node) string {
	tagName := node
	if node != nil {
		if t := GetJsxTagName(node); t != nil {
			tagName = t
		}
	}
	return tagNameString(tagName)
}

func tagNameString(tagName *ast.Node) string {
	if tagName == nil {
		return ""
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.AsIdentifier().Text
	case ast.KindThisKeyword:
		return "this"
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns.Namespace == nil || ns.Name() == nil {
			return ""
		}
		if ns.Namespace.Kind != ast.KindIdentifier || ns.Name().Kind != ast.KindIdentifier {
			return ""
		}
		return ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		pa := tagName.AsPropertyAccessExpression()
		base := tagNameString(pa.Expression)
		if base == "" {
			return ""
		}
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return ""
		}
		return base + "." + nameNode.AsIdentifier().Text
	}
	return ""
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

// IsJsxElementLike reports whether node is a JsxElement or
// JsxSelfClosingElement — the two tsgo kinds that correspond to ESTree's
// single `JSXElement` type.
func IsJsxElementLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsJsxElement(node) || ast.IsJsxSelfClosingElement(node)
}

// IsJsxLike mirrors eslint-plugin-react's `jsxUtil.isJSX` — true for a JSX
// element (either tag form) or a JSX fragment.
func IsJsxLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return IsJsxElementLike(node) || ast.IsJsxFragment(node)
}

// JSXRootTagName returns the tag-name of a JsxElement / JsxSelfClosingElement
// (peeling paren / TS wrappers) when it's a plain Identifier, or "" otherwise.
// Member-expression tag-names (`<Foo.Bar />`) and namespaced names
// (`<svg:circle/>`) intentionally return "" — upstream's
// `getComponentNameFromJSXElement` only matches plain identifiers via the
// detected-components list keyed by the binding's local name.
func JSXRootTagName(expr *ast.Node) string {
	expr = SkipExpressionWrappers(expr)
	if expr == nil {
		return ""
	}
	var tag *ast.Node
	switch expr.Kind {
	case ast.KindJsxElement:
		opening := expr.AsJsxElement().OpeningElement
		if opening != nil {
			tag = opening.AsJsxOpeningElement().TagName
		}
	case ast.KindJsxSelfClosingElement:
		tag = expr.AsJsxSelfClosingElement().TagName
	default:
		return ""
	}
	if tag == nil || tag.Kind != ast.KindIdentifier {
		return ""
	}
	return tag.AsIdentifier().Text
}

// ReturnedJSXRootTagName extracts the root JSX tag name from a function's
// body — covers both the concise-body case (`() => <Foo/>`) and the
// block-body case where the FIRST top-level ReturnStatement is inspected.
// Returns empty string when the body doesn't return a JSX element directly.
func ReturnedJSXRootTagName(fn *ast.Node) string {
	if fn == nil {
		return ""
	}
	var body *ast.Node
	switch fn.Kind {
	case ast.KindArrowFunction:
		body = fn.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	default:
		return ""
	}
	if body == nil {
		return ""
	}
	if body.Kind == ast.KindBlock {
		var ret *ast.Node
		body.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindReturnStatement {
				ret = child
				return true
			}
			return false
		})
		if ret == nil {
			return ""
		}
		return JSXRootTagName(ret.AsReturnStatement().Expression)
	}
	return JSXRootTagName(body)
}
