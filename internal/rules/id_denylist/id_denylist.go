package id_denylist

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed id_denylist.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/id-denylist

func parseOptions(options []any) map[string]struct{} {
	denyList := make(map[string]struct{}, len(options))
	for _, option := range options {
		if name, ok := option.(string); ok {
			denyList[name] = struct{}{}
		}
	}
	return denyList
}

var IdDenylistRule = rule.Rule{
	Name:   "id-denylist",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		denyList := parseOptions(options)
		if len(denyList) == 0 {
			return rule.RuleListeners{}
		}

		check := func(node *ast.Node) {
			name := node.Text()
			isPrivate := node.Kind == ast.KindPrivateIdentifier
			if isPrivate {
				// A tsgo PrivateIdentifier carries the `#` in its text, while
				// the denied name is written without it.
				name = strings.TrimPrefix(name, "#")
			}
			if _, restricted := denyList[name]; !restricted {
				return
			}
			if !shouldCheck(ctx, node) {
				return
			}
			if isPrivate {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "restrictedPrivate",
					Description: fmt.Sprintf("Identifier '#%s' is restricted.", name),
				})
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "restricted",
				Description: fmt.Sprintf("Identifier '%s' is restricted.", name),
			})
		}

		return rule.RuleListeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
		}
	},
}

// shouldCheck determines whether a restricted name in this position is
// reported. It mirrors the upstream rule's shouldCheck.
func shouldCheck(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Import attributes are defined by environments, so naming conventions
	// shouldn't apply to them.
	if isImportAttributeKey(node) {
		return false
	}

	// A JSX tag or attribute name is a JSXIdentifier upstream, a node type the
	// rule's Identifier listener never visits. tsgo spells both with plain
	// Identifier and PropertyAccessExpression nodes, so they are filtered here
	// instead.
	if isJsxName(node) {
		return false
	}

	// `import.meta`, `new.target`.
	if parent.Kind == ast.KindMetaProperty {
		return false
	}

	// Member access has special rules for checking property names. Read access
	// to a property with a restricted name is allowed, because it can be on an
	// object that user has no control over. Write access isn't allowed, because
	// it potentially creates a new property with a restricted name.
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return isAssignmentTarget(parent)
	}

	// A callee or an argument of a call or a `new` is never checked. tsgo
	// spells `import(specifier, options)` as a CallExpression too, where
	// upstream has a distinct ImportExpression that this exclusion misses.
	switch parent.Kind {
	case ast.KindCallExpression:
		if !ast.IsImportCall(parent) {
			return false
		}
	case ast.KindNewExpression:
		return false
	}

	return !isRenamedImport(node) &&
		!isPropertyNameInDestructuring(node) &&
		!isReferenceToGlobalVariable(ctx, node)
}

// isAssignmentTarget reports whether the given member access is written to,
// either on its own or as an element of a destructuring pattern. Upstream
// recognizes the pattern positions by the ArrayPattern, RestElement,
// ObjectPattern-property and AssignmentPattern node types; tsgo spells all of
// them with the expression syntax they share with a plain array or object
// literal, so each one additionally confirms that it sits in a destructuring
// target. Prefix/postfix `++`/`--` and a bare `for (x.y in z)` target are
// deliberately absent: upstream reports neither.
func isAssignmentTarget(node *ast.Node) bool {
	target := node
	for target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
		target = target.Parent
	}
	parent := target.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// `foo.bar = 1` and `foo.bar += 1`, plus the `obj.bar = baz` default of
		// `({ a: obj.bar = baz } = qux)` that upstream spells AssignmentPattern.
		binary := parent.AsBinaryExpression()
		return ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == target
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement, ast.KindSpreadAssignment:
		return isDestructuredFrom(target)
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == target &&
			isDestructuredFrom(target)
	}
	return false
}

// isDestructuredFrom reports whether node is an element of an array or object
// literal that an enclosing assignment reinterprets as a destructuring pattern.
// Upstream reads this off the node type alone, because ESTree gives a pattern
// its own kinds; tsgo reuses the expression syntax, so the enclosing assignment
// has to be found by walking out. The walk climbs only the nodes a pattern is
// built from and stops at the first parent that consumes the literal as a value
// — `[{ a: obj.b }.c] = d` destructures the result of a member access, not the
// literal the member access reads from. An update expression is not one of
// upstream's arms, so it does not count as an enclosing assignment here.
func isDestructuredFrom(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	assignment := ast.GetAssignmentTarget(node)
	return assignment != nil &&
		assignment.Kind != ast.KindPrefixUnaryExpression &&
		assignment.Kind != ast.KindPostfixUnaryExpression
}

// isRenamedImport reports whether node is an imported or re-exported name that
// the same specifier renames: the `a` of `import { a as b } from 'mod'` or of
// `export { a as b } from 'mod'`.
func isRenamedImport(node *ast.Node) bool {
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportSpecifier:
		return parent.AsImportSpecifier().PropertyName == node
	case ast.KindExportSpecifier:
		return parent.AsExportSpecifier().PropertyName == node && utils.IsReExportSpecifier(parent)
	}
	return false
}

// isPropertyNameInDestructuring reports whether node is the source key of an
// object destructuring pattern: the `a` of `const { a: b } = foo` or of
// `({ a: obj.b } = foo)`. A shorthand key is deliberately absent — upstream
// skips its key half but still checks the value half, which is the same node
// here (see isVariableReference).
func isPropertyNameInDestructuring(node *ast.Node) bool {
	parent := node.Parent
	switch parent.Kind {
	case ast.KindBindingElement:
		return parent.AsBindingElement().PropertyName == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() == node &&
			isDestructuredFrom(parent.Parent)
	}
	return false
}

// isReferenceToGlobalVariable reports whether node references a global variable
// that the source code does not declare. Those identifiers are allowed, as it
// is assumed that user has no control over the names of external global
// variables.
func isReferenceToGlobalVariable(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Refs == nil || !isVariableReference(node) {
		return false
	}
	name := node.Text()
	if !ctx.Globals.Access(name).IsDeclared() {
		return false
	}
	// A name this file declares in the reference's own declaration space
	// resolves to that declaration instead of to the global.
	if ctx.Refs.ResolveInFile(node) != nil {
		return false
	}
	// In a script every top-level declaration shares the one global scope, so a
	// single one of them takes the name away from the global for the whole
	// file: `interface Number {}` at the top of a script also claims a later
	// `Number` used as a value, and a `Number` read from inside a function. A
	// module keeps its top level to itself, and a declaration nested in a scope
	// of its own reaches only what that scope encloses; the resolution above
	// already settles both.
	if ctx.SourceFile != nil && ast.IsGlobalSourceFile(ctx.SourceFile.AsNode()) &&
		ctx.SourceFile.Locals[name] != nil {
		return false
	}
	return true
}

// isVariableReference reports whether node occupies a position that upstream's
// scope analysis records as a variable reference, the only kind of identifier
// that can resolve to a global.
func isVariableReference(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	// `{ foo }` is one node here and two upstream: a property key and a value
	// reference. Only inside a destructuring pattern, where the key half is
	// skipped as a destructuring key, does the reference half decide the
	// outcome; in an object literal the key half is checked and resolves to
	// nothing.
	if node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		return isDestructuredFrom(node.Parent.Parent)
	}
	return !utils.IsNonReferenceIdentifier(node)
}

// isImportAttributeKey reports whether node names an import attribute: the key
// of a static `with { type: 'json' }` clause, or any key of the options object
// of a dynamic `import(specifier, { with: { type: 'json' } })`, including keys
// nested under such a key.
func isImportAttributeKey(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Static import / re-export.
	if parent.Kind == ast.KindImportAttribute && parent.AsImportAttribute().Name() == node {
		return true
	}

	// Dynamic import. A computed key is evaluated rather than named, so it is
	// checked like any other expression.
	if node.Kind == ast.KindComputedPropertyName {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Name() != node {
			return false
		}
	default:
		return false
	}
	object := parent.Parent
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	// Upstream sees the object literal directly, because ESTree has no node for
	// a parenthesis: `import('x', ({ type: 'json' }))` passes the same options
	// object as the unparenthesized form.
	for object.Parent != nil && object.Parent.Kind == ast.KindParenthesizedExpression {
		object = object.Parent
	}
	owner := object.Parent
	if owner == nil {
		return false
	}
	if ast.IsImportCall(owner) {
		arguments := owner.AsCallExpression().Arguments
		if arguments != nil && len(arguments.Nodes) > 1 && arguments.Nodes[1] == object {
			return true
		}
	}
	// Nested key: `type` in `{ with: { type: 'json' } }` is reached through
	// `with`.
	if owner.Kind == ast.KindPropertyAssignment && owner.AsPropertyAssignment().Initializer == object {
		return isImportAttributeKey(owner.Name())
	}
	return false
}

// isJsxName reports whether node is part of a JSX tag name or attribute name.
func isJsxName(node *ast.Node) bool {
	current := node
	for parent := current.Parent; parent != nil; parent = current.Parent {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindQualifiedName, ast.KindJsxNamespacedName:
			current = parent
		case ast.KindJsxOpeningElement:
			return parent.AsJsxOpeningElement().TagName == current
		case ast.KindJsxSelfClosingElement:
			return parent.AsJsxSelfClosingElement().TagName == current
		case ast.KindJsxClosingElement:
			return parent.AsJsxClosingElement().TagName == current
		case ast.KindJsxAttribute:
			return parent.AsJsxAttribute().Name() == current
		default:
			return false
		}
	}
	return false
}
