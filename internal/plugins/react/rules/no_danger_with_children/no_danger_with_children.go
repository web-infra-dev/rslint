package no_danger_with_children

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isCreateElementCallee reports whether a CallExpression callee is
// `<anything>.createElement` — mirroring upstream's check:
//
//	node.callee.type === 'MemberExpression'
//	  && 'name' in node.callee.property
//	  && node.callee.property.name === 'createElement'
//
// Upstream deliberately does NOT restrict the object to `React` or the
// configured pragma; any `x.createElement(...)` is inspected. We therefore
// do NOT reuse `reactutil.IsCreateElementCall`, which enforces a pragma
// match. Computed access (`x['createElement']`, modeled as
// ElementAccessExpression) is excluded because upstream's
// `'name' in property` guard skips computed keys.
func isCreateElementCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	callee = ast.SkipParentheses(callee)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return name.AsIdentifier().Text == "createElement"
}

const dangerWithChildrenMessage = "Only set one of `children` or `props.dangerouslySetInnerHTML`"

var NoDangerWithChildrenRule = rule.Rule{
	Name: "react/no-danger-with-children",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// resolveObjectLiteralInit walks an Identifier back to its
		// VariableDeclaration via the TypeChecker and returns the initializer
		// when it is an object literal. Mirrors upstream's
		// `variable.defs[0].node.init` lookup. Parentheses are transparent
		// on the initializer (tsgo preserves them; ESTree flattens them).
		resolveObjectLiteralInit := func(ident *ast.Node) *ast.ObjectLiteralExpression {
			if ident == nil || ident.Kind != ast.KindIdentifier {
				return nil
			}
			decl := utils.GetDeclaration(ctx.TypeChecker, ident)
			if decl == nil || decl.Kind != ast.KindVariableDeclaration {
				return nil
			}
			init := decl.AsVariableDeclaration().Initializer
			if init == nil {
				return nil
			}
			init = ast.SkipParentheses(init)
			if init.Kind != ast.KindObjectLiteralExpression {
				return nil
			}
			return init.AsObjectLiteralExpression()
		}

		// findObjectPropByName reports whether the object literal has a
		// property with the given name — either directly (Identifier key, to
		// match upstream's `prop.key.name === propName`) or via a spread whose
		// initializer resolves to another object literal.
		//
		// `seen` tracks Identifier names already followed through a spread,
		// so self-referencing initializers like `const props = { ...props }`
		// terminate instead of looping.
		//
		// Upstream intentionally matches ONLY Identifier keys — a
		// StringLiteral key (`"children": "x"`) has `.value` but no `.name`,
		// so upstream's `prop.key.name === propName` is always false on it.
		// We keep parity by not using `utils.GetStaticPropertyName` here.
		var findObjectPropByName func(obj *ast.ObjectLiteralExpression, name string, seen map[string]bool) bool
		findObjectPropByName = func(obj *ast.ObjectLiteralExpression, name string, seen map[string]bool) bool {
			if obj == nil || obj.Properties == nil {
				return false
			}
			for _, prop := range obj.Properties.Nodes {
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					keyNode := prop.AsPropertyAssignment().Name()
					if keyNode != nil && keyNode.Kind == ast.KindIdentifier &&
						keyNode.AsIdentifier().Text == name {
						return true
					}
				case ast.KindShorthandPropertyAssignment:
					keyNode := prop.AsShorthandPropertyAssignment().Name()
					if keyNode != nil && keyNode.Kind == ast.KindIdentifier &&
						keyNode.AsIdentifier().Text == name {
						return true
					}
				case ast.KindSpreadAssignment:
					expr := ast.SkipParentheses(prop.AsSpreadAssignment().Expression)
					if expr.Kind != ast.KindIdentifier {
						continue
					}
					spreadName := expr.AsIdentifier().Text
					if seen[spreadName] {
						continue
					}
					inner := resolveObjectLiteralInit(expr)
					if inner == nil {
						continue
					}
					nextSeen := make(map[string]bool, len(seen)+1)
					for k, v := range seen {
						nextSeen[k] = v
					}
					nextSeen[spreadName] = true
					if findObjectPropByName(inner, name, nextSeen) {
						return true
					}
				}
			}
			return false
		}

		// asPropsObject returns the ObjectLiteralExpression to inspect for a
		// given props node, together with a `seen` set pre-populated with the
		// identifier name (if any) that was just resolved. Callers should use
		// the returned `seen` for their first recursive lookup so a
		// self-referencing initializer (`const props = { ...props, ... }`)
		// terminates on the first recursion.
		asPropsObject := func(node *ast.Node) (*ast.ObjectLiteralExpression, map[string]bool) {
			node = ast.SkipParentheses(node)
			if node.Kind == ast.KindObjectLiteralExpression {
				return node.AsObjectLiteralExpression(), map[string]bool{}
			}
			if node.Kind == ast.KindIdentifier {
				if obj := resolveObjectLiteralInit(node); obj != nil {
					return obj, map[string]bool{node.AsIdentifier().Text: true}
				}
			}
			return nil, nil
		}

		// findJsxAttr reports whether a JSX opening / self-closing element
		// carries an attribute with the given name, either as a plain
		// JsxAttribute (Identifier name) or via a JsxSpreadAttribute whose
		// *Identifier* argument resolves to an object literal containing that
		// name (transitively through further spreads).
		//
		// NOTE: Upstream's `findSpreadVariable` reads `attribute.argument.name`,
		// which only exists when the argument is an Identifier. Inline object
		// spreads (`{...{ ... }}`) are therefore opaque to upstream — we keep
		// parity here by requiring Identifier too, rather than inspecting the
		// inline object directly.
		findJsxAttr := func(element *ast.Node, name string) bool {
			for _, attr := range reactutil.GetJsxElementAttributes(element) {
				switch attr.Kind {
				case ast.KindJsxAttribute:
					nameNode := attr.AsJsxAttribute().Name()
					if nameNode != nil && nameNode.Kind == ast.KindIdentifier &&
						nameNode.AsIdentifier().Text == name {
						return true
					}
				case ast.KindJsxSpreadAttribute:
					expr := ast.SkipParentheses(attr.AsJsxSpreadAttribute().Expression)
					if expr.Kind != ast.KindIdentifier {
						continue
					}
					inner := resolveObjectLiteralInit(expr)
					if inner == nil {
						continue
					}
					seen := map[string]bool{expr.AsIdentifier().Text: true}
					if findObjectPropByName(inner, name, seen) {
						return true
					}
				}
			}
			return false
		}

		// isLineBreak mirrors ESLint's isLineBreak helper: a JsxText that
		// spans more than one line AND contains only whitespace. A single-line
		// whitespace JsxText (e.g. `<div> </div>`) counts as meaningful
		// children and is NOT a line break — matches upstream.
		isLineBreak := func(child *ast.Node) bool {
			if child == nil || child.Kind != ast.KindJsxText {
				return false
			}
			text := child.AsJsxText().Text
			if strings.TrimSpace(text) != "" {
				return false
			}
			return strings.Contains(text, "\n")
		}

		// checkJsx is shared between JsxElement and JsxSelfClosingElement.
		// elementForAttrs carries the attributes (opening element or the
		// self-closing element itself); reportNode is the node the diagnostic
		// is attached to (the full JsxElement / JsxSelfClosingElement).
		checkJsx := func(elementForAttrs, reportNode *ast.Node, firstChild *ast.Node) {
			hasChildren := false
			if firstChild != nil && !isLineBreak(firstChild) {
				hasChildren = true
			} else if findJsxAttr(elementForAttrs, "children") {
				hasChildren = true
			}
			if !hasChildren {
				return
			}
			if findJsxAttr(elementForAttrs, "dangerouslySetInnerHTML") {
				ctx.ReportNode(reportNode, rule.RuleMessage{
					Id:          "dangerWithChildren",
					Description: dangerWithChildrenMessage,
				})
			}
		}

		return rule.RuleListeners{
			ast.KindJsxElement: func(node *ast.Node) {
				jsx := node.AsJsxElement()
				var firstChild *ast.Node
				if jsx.Children != nil && len(jsx.Children.Nodes) > 0 {
					firstChild = jsx.Children.Nodes[0]
				}
				checkJsx(jsx.OpeningElement, node, firstChild)
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				checkJsx(node, node, nil)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !isCreateElementCallee(call.Expression) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
					return
				}
				obj, seen := asPropsObject(call.Arguments.Nodes[1])
				if obj == nil {
					return
				}
				if !findObjectPropByName(obj, "dangerouslySetInnerHTML", seen) {
					return
				}
				hasChildren := len(call.Arguments.Nodes) > 2
				if !hasChildren {
					hasChildren = findObjectPropByName(obj, "children", seen)
				}
				if hasChildren {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "dangerWithChildren",
						Description: dangerWithChildrenMessage,
					})
				}
			},
		}
	},
}
