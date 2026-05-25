package forbid_foreign_prop_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const msgForbiddenPropType = "Using propTypes from another component is not safe because they may be removed in production builds"

type ruleOptions struct {
	allowInPropTypes bool
}

// parseOptions reads `allowInPropTypes` (boolean) from the rule options.
// Mirrors upstream's defaults — missing option / non-bool falls back to false.
func parseOptions(input any) ruleOptions {
	opts := ruleOptions{}
	m := utils.GetOptionsMap(input)
	if m == nil {
		return opts
	}
	if v, ok := m["allowInPropTypes"].(bool); ok {
		opts.allowInPropTypes = v
	}
	return opts
}

// effectiveParent walks up through ParenthesizedExpression wrappers to find
// the topmost paren-wrapped node and its non-paren ancestor. Mirrors
// ESTree's transparent-paren parent: in upstream `node.parent` already
// skips parentheses; in tsgo we must skip them explicitly to match.
func effectiveParent(node *ast.Node) (*ast.Node, *ast.Node) {
	current := node
	parent := current.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		current = parent
		parent = current.Parent
	}
	return current, parent
}

// isAssignmentLHS mirrors upstream's `ast.isAssignmentLHS`:
//
//	node.parent.type === 'AssignmentExpression' && node.parent.left === node
//
// Strict direct-LHS only — does NOT walk up through nested destructuring
// targets (e.g. `({a: Foo.propTypes} = X)`). Compound assignments
// (`+=`, `||=`, `??=`, …) all share the same ESTree
// `AssignmentExpression` type, so any assignment-style operator counts.
//
// We use this — not the broader `ast.IsAssignmentTarget` — because
// upstream only excludes the direct LHS form (`Foo.propTypes = X`).
// `ast.IsAssignmentTarget` walks up through nested patterns and would
// over-exclude.
func isAssignmentLHS(node *ast.Node) bool {
	current, parent := effectiveParent(node)
	if parent == nil || !ast.IsAssignmentExpression(parent, false) {
		return false
	}
	return parent.AsBinaryExpression().Left == current
}

// isAllowedAssignment reports whether `node` sits inside a `propTypes`
// declaration that the option `allowInPropTypes: true` excuses. Mirrors
// upstream's two-pronged check:
//  1. nearest enclosing AssignmentExpression has LHS `<expr>.propTypes`
//  2. nearest enclosing ClassProperty / PropertyDefinition is keyed by
//     the Identifier `propTypes`
//
// Both prongs require the dotted-identifier form — string-literal /
// computed key shapes don't satisfy upstream's `key.name === 'propTypes'`
// or `left.property.name === 'propTypes'` checks.
func isAllowedAssignment(node *ast.Node, allow bool) bool {
	if !allow {
		return false
	}

	// Closest enclosing assignment with LHS `<expr>.propTypes`.
	if assign := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return ast.IsAssignmentExpression(n, false)
	}); assign != nil {
		// Strip parens on the LHS — ESTree flattens parens, tsgo
		// preserves them. Matches upstream's `assignment.left.property`
		// access on a flat MemberExpression.
		left := ast.SkipParentheses(assign.AsBinaryExpression().Left)
		if left != nil && left.Kind == ast.KindPropertyAccessExpression {
			name := left.AsPropertyAccessExpression().Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "propTypes" {
				return true
			}
		}
	}

	// Closest enclosing class field whose key is the Identifier `propTypes`.
	// Upstream's `classProperty.key.name === 'propTypes'` only matches
	// Identifier keys (StringLiteral / ComputedPropertyName have no `.name`).
	if cp := ast.FindAncestorKind(node.Parent, ast.KindPropertyDeclaration); cp != nil {
		keyNode := cp.AsPropertyDeclaration().Name()
		if keyNode != nil && keyNode.Kind == ast.KindIdentifier && keyNode.AsIdentifier().Text == "propTypes" {
			return true
		}
	}

	return false
}

// isJsxTagName reports whether `node` is the TagName of a JSX element
// (or any link in a dotted JSX tag chain like `<Foo.Bar.Baz />`). tsgo
// represents dotted JSX tags as `PropertyAccessExpression`, but ESTree
// uses a distinct `JSXMemberExpression` node type — ESLint's
// `MemberExpression` listener does NOT fire on JSXMemberExpression.
// We exclude this case to align byte-for-byte.
func isJsxTagName(node *ast.Node) bool {
	// Walk up the PA chain (each inner PA is the `.Expression` of its
	// outer PA). If the outermost reaches a Jsx* element TagName, skip.
	outer := node
	for {
		p := outer.Parent
		if p == nil || p.Kind != ast.KindPropertyAccessExpression {
			break
		}
		if p.AsPropertyAccessExpression().Expression != outer {
			break
		}
		outer = p
	}
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxClosingElement:
		return true
	}
	return false
}

// findPropTypesKey returns the first element whose effective key is the
// Identifier `propTypes`, or nil. Shared between the
// ObjectBindingPattern listener (declaration form) and the
// ObjectLiteralExpression listener (destructuring-assignment form).
// Mirrors upstream's
// `'name' in property.key && property.key.name === 'propTypes'`:
// rest / spread elements, computed keys, string-literal keys, and
// methods / accessors are silently skipped.
func findPropTypesKey(elements []*ast.Node) *ast.Node {
	for _, prop := range elements {
		var key *ast.Node
		switch prop.Kind {
		case ast.KindBindingElement:
			be := prop.AsBindingElement()
			if be == nil || be.DotDotDotToken != nil {
				continue
			}
			if be.PropertyName != nil {
				key = be.PropertyName
			} else {
				key = be.Name()
			}
		case ast.KindPropertyAssignment:
			key = prop.AsPropertyAssignment().Name()
		case ast.KindShorthandPropertyAssignment:
			key = prop.AsShorthandPropertyAssignment().Name()
		default:
			// SpreadAssignment, methods, accessors, etc. — upstream's
			// `property.type === 'Property'` filter excludes these.
			continue
		}
		if key == nil || key.Kind != ast.KindIdentifier {
			continue
		}
		if key.AsIdentifier().Text != "propTypes" {
			continue
		}
		return prop
	}
	return nil
}

var ForbidForeignPropTypesRule = rule.Rule{
	Name: "react/forbid-foreign-prop-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		msg := rule.RuleMessage{
			Id:          "forbiddenPropType",
			Description: msgForbiddenPropType,
		}

		return rule.RuleListeners{
			// Dotted access `<expr>.propTypes`. ESTree's `MemberExpression`
			// non-computed branch with `property.type === 'Identifier'`.
			// PrivateIdentifier (`#propTypes`) is intentionally ignored —
			// upstream's `property.type === 'Identifier'` check excludes
			// PrivateIdentifier (its ESTree type is `PrivateIdentifier`).
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				name := node.AsPropertyAccessExpression().Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				if name.AsIdentifier().Text != "propTypes" {
					return
				}
				if isAssignmentLHS(node) {
					return
				}
				if isJsxTagName(node) {
					return
				}
				if isAllowedAssignment(node, opts.allowInPropTypes) {
					return
				}
				ctx.ReportNode(name, msg)
			},

			// Bracketed access `<expr>['propTypes']`. ESTree's
			// `MemberExpression` computed branch with property a Literal
			// whose value is `'propTypes'`. Upstream also accepts JSXText
			// here, but JSXText cannot syntactically appear inside a
			// MemberExpression's brackets, so it's a dead branch upstream
			// and we omit it.
			//
			// Template literals (no-substitution form) are NOT matched —
			// upstream's `property.type === 'Literal'` excludes
			// TemplateLiteral. Locked in by the `Foo[\`propTypes\`]` valid
			// test case.
			ast.KindElementAccessExpression: func(node *ast.Node) {
				arg := node.AsElementAccessExpression().ArgumentExpression
				if arg == nil || arg.Kind != ast.KindStringLiteral {
					return
				}
				if arg.AsStringLiteral().Text != "propTypes" {
					return
				}
				if isAssignmentLHS(node) {
					return
				}
				if isAllowedAssignment(node, opts.allowInPropTypes) {
					return
				}
				ctx.ReportNode(arg, msg)
			},

			// Object destructuring in declaration form: `var { propTypes }
			// = X`, `const { propTypes: alias } = X`, function parameters
			// `function f({ propTypes }) {}`, etc. tsgo represents these
			// as `ObjectBindingPattern` containing `BindingElement`s.
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				bp := node.AsBindingPattern()
				if bp == nil || bp.Elements == nil {
					return
				}
				if match := findPropTypesKey(bp.Elements.Nodes); match != nil {
					ctx.ReportNode(match, msg)
				}
			},

			// Object destructuring in expression form: `({ propTypes } =
			// X)`, `[{ propTypes }] = X`, nested patterns inside a larger
			// destructuring target. tsgo represents the LHS of such
			// assignments as `ObjectLiteralExpression` (only the
			// assignment context reclassifies it as a pattern), so we
			// gate on `ast.IsAssignmentTarget`, which walks up through
			// parens / arrays / nested object literals to confirm the
			// node is in an assignment-target position.
			//
			// Upstream covers this via its single `ObjectPattern`
			// listener — ESLint emits ObjectPattern for both var-decl
			// and assignment forms; tsgo splits them into two distinct
			// kinds, so the rule needs two listeners.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if !ast.IsAssignmentTarget(node) {
					return
				}
				obj := node.AsObjectLiteralExpression()
				if obj == nil || obj.Properties == nil {
					return
				}
				if match := findPropTypesKey(obj.Properties.Nodes); match != nil {
					ctx.ReportNode(match, msg)
				}
			},
		}
	},
}
