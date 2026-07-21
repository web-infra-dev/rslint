package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// skipTransparentKinds matches parentheses + TS type assertions.
const skipTransparentKinds = ast.OEKParentheses | ast.OEKAssertions

// SkipAssertionsAndParens strips parentheses and all TS assertion wrappers
// (as, satisfies, !, <T>) from an expression, mirroring ESLint's
// unwrapTSAsExpression(uncast(node)). Returns nil when node is nil so callers
// can safely pass optional AST fields such as an absent initializer.
func SkipAssertionsAndParens(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, skipTransparentKinds)
}

// OutermostParenthesizedExpression returns node's outermost
// ParenthesizedExpression wrapper, or node itself when it is not wrapped.
// Unlike ast.WalkUpParenthesizedExpressions, this preserves the wrapper that
// the containing non-parenthesized node sees as its direct child.
func OutermostParenthesizedExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil &&
		ast.IsParenthesizedExpression(current.Parent) {
		parent := current.Parent.AsParenthesizedExpression()
		if parent == nil || parent.Expression != current {
			break
		}
		current = current.Parent
	}
	return current
}

// IsCallee checks if a node is the callee of a CallExpression or NewExpression,
// skipping parentheses and TS type assertions between the node and the call.
func IsCallee(node *ast.Node) bool {
	current := node
	parent := current.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparentKinds) {
		current = parent
		parent = current.Parent
	}
	if parent == nil {
		return false
	}
	if ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == current {
		return true
	}
	if parent.Kind == ast.KindNewExpression && parent.AsNewExpression().Expression == current {
		return true
	}
	return false
}

// GetStaticStringLiteralValue returns the string value and a presence flag if
// node is a string literal or a no-substitution template literal. It does not
// unwrap parentheses or TS assertions; callers choose which wrappers are
// transparent for their rule.
func GetStaticStringLiteralValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// GetStaticStringValue returns the string value if the node is a string literal
// or a no-substitution template literal. Returns "" if the value cannot be
// statically determined.
func GetStaticStringValue(node *ast.Node) string {
	value, _ := GetStaticStringLiteralValue(node)
	return value
}

// IsGlobalParseIntCallee reports whether callee references the built-in
// `parseInt` or `Number.parseInt` function. It mirrors ESLint's
// astUtils.isSpecificId / isSpecificMemberAccess shape: outer parentheses and
// optional chaining are transparent, TS-only assertion wrappers are not.
//
// globals is the config-declared `languageOptions.globals` / `/* global */`
// set (ctx.Globals); when it explicitly turns the referenced name `off`,
// the identifier no longer resolves to a known global and this returns
// false. Pass nil to skip that check (e.g. for callers whose upstream ESLint
// rule doesn't consult scope at all).
func IsGlobalParseIntCallee(callee *ast.Node, globals map[string]bool) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	if ast.IsIdentifier(callee) {
		if callee.AsIdentifier().Text != "parseInt" || IsShadowed(callee, "parseInt") {
			return false
		}
		return !isGlobalOff(globals, "parseInt")
	}

	if !IsSpecificMemberAccess(callee, "Number", "parseInt") {
		return false
	}

	obj := AccessExpressionObject(callee)
	obj = ast.SkipParentheses(obj)
	if obj == nil || !ast.IsIdentifier(obj) ||
		obj.AsIdentifier().Text != "Number" || IsShadowed(obj, "Number") {
		return false
	}
	return !isGlobalOff(globals, "Number")
}

// isGlobalOff reports whether globals explicitly un-declares name via an
// `off` setting (e.g. `/* global Foo: off */`). A name absent from globals
// is not considered off — it just falls back to the normal shadow check.
func isGlobalOff(globals map[string]bool, name string) bool {
	declared, ok := globals[name]
	return ok && !declared
}

// IsNonReferenceIdentifier checks if an identifier is NOT a value reference
// (i.e., it's a declaration name, property key, label, or module specifier name
// rather than a reference to a variable).
func IsNonReferenceIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Property access name: a.b — `b` is a property key, not a variable reference.
	if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == node {
		return true
	}

	// Qualified type name: A.B.C (used in types) — right-hand names are not refs.
	if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Right == node {
		return true
	}

	// Meta property: new.target, import.meta — `target`/`meta` are keywords.
	if parent.Kind == ast.KindMetaProperty {
		return true
	}

	// export { local as exported }: only `local` can read a runtime value.
	if parent.Kind == ast.KindExportSpecifier {
		if ast.IsTypeOnlyImportOrExportDeclaration(parent) || IsReExportSpecifier(parent) {
			return true
		}
		es := parent.AsExportSpecifier()
		if es == nil {
			return false
		}
		if es.PropertyName != nil {
			return es.PropertyName != node
		}
		return es.Name() != node
	}

	// ast.IsDeclarationName covers: variable, function, class, parameter,
	// property assignment, method, accessor, enum member, etc.
	if ast.IsDeclarationName(node) {
		// ShorthandPropertyAssignment { x } — x IS a reference to the variable.
		if parent.Kind == ast.KindShorthandPropertyAssignment {
			return false
		}
		return true
	}

	// Property name in destructuring: { x: y } — x is just a key.
	if parent.Kind == ast.KindBindingElement {
		be := parent.AsBindingElement()
		if be.PropertyName != nil && be.PropertyName == node {
			return true
		}
	}

	// Import source name: import { x as y } — x is the source module's export name.
	if parent.Kind == ast.KindImportSpecifier {
		importSpec := parent.AsImportSpecifier()
		if importSpec.PropertyName != nil && importSpec.PropertyName == node {
			return true
		}
	}

	// Label names: label: while(true) { break label; continue label; }
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	return false
}

// IsInAmbientContext reports whether node was parsed inside an ambient
// context. TypeScript-Go propagates this through declaration files and
// `declare` contexts via NodeFlagsAmbient.
func IsInAmbientContext(node *ast.Node) bool {
	return node != nil && node.Flags&ast.NodeFlagsAmbient != 0
}

// CouldBeError reports whether a node could plausibly evaluate to an Error
// object at runtime. Mirrors ESLint's `astUtils.couldBeError`, adapted to the
// tsgo AST where AssignmentExpression / LogicalExpression / SequenceExpression
// are all flattened into BinaryExpression and ChainExpression has no analog.
//
// Only parentheses are unwrapped — TS-only assertion wrappers (`x as T`,
// `<T>x`, `x satisfies T`, `x!`) are NOT unwrapped, because ESLint's
// `astUtils.couldBeError` does not list them and falls through to `false`.
// Verified against ESLint core run on a `.ts` file via `@typescript-eslint/parser`:
// `throw foo as Error;` and `throw foo!;` are both reported as "object".
//
// Used by rules whose ESLint counterparts call `astUtils.couldBeError`:
// `no-throw-literal`, `prefer-promise-reject-errors`, etc.
func CouldBeError(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindYieldExpression,
		ast.KindAwaitExpression:
		return true

	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		// `a, b, c` parses left-associatively in tsgo, so the rightmost
		// expression is `bin.Right` of the outer BinaryExpression.
		case ast.KindCommaToken:
			return CouldBeError(bin.Right)
		// `a = b` / `a &&= b` evaluate to the right operand.
		case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return CouldBeError(bin.Right)
		// `a ||= b` / `a ??= b` evaluate to either `a` or `b`.
		case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return CouldBeError(bin.Left) || CouldBeError(bin.Right)
		// `a && b` short-circuits to a falsy `a` (cannot be Error) or to `b`.
		case ast.KindAmpersandAmpersandToken:
			return CouldBeError(bin.Right)
		// `a || b` / `a ?? b` evaluate to either operand.
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return CouldBeError(bin.Left) || CouldBeError(bin.Right)
		default:
			// Arithmetic / bitwise / comparison / compound-assign other than
			// `=`, `&&=`, `||=`, `??=`: result is a primitive (or throws).
			return false
		}

	case ast.KindConditionalExpression:
		ce := node.AsConditionalExpression()
		if ce == nil {
			return false
		}
		return CouldBeError(ce.WhenTrue) || CouldBeError(ce.WhenFalse)
	}

	return false
}

// IsUndefinedIdentifier reports whether the node, after unwrapping parens, is
// the literal identifier `undefined`. Purely lexical — does not detect `void 0`,
// `undefined as any`, or a shadowed `undefined` binding, matching ESLint's
// `node.argument.name === "undefined"` check (which only sees an Identifier
// after parens are dropped at parse time, not after TS assertions).
func IsUndefinedIdentifier(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == "undefined"
}

// IsReExportSpecifier checks if an ExportSpecifier is part of a re-export
// declaration (export { ... } from 'mod').
func IsReExportSpecifier(exportSpec *ast.Node) bool {
	// ExportSpecifier → NamedExports → ExportDeclaration
	namedExports := exportSpec.Parent
	if namedExports == nil {
		return false
	}
	exportDecl := namedExports.Parent
	if exportDecl == nil || exportDecl.Kind != ast.KindExportDeclaration {
		return false
	}
	return exportDecl.AsExportDeclaration().ModuleSpecifier != nil
}

// IsClassExtendsHeritageClause reports whether an ExpressionWithTypeArguments
// node sits inside a class (not interface) `extends` clause — a value
// context, since the superclass expression is actually evaluated at runtime.
// Every other heritage use (interface extends, class implements) is a pure
// type position.
func IsClassExtendsHeritageClause(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindHeritageClause {
		return false
	}
	clause := parent.AsHeritageClause()
	if clause.Token != ast.KindExtendsKeyword {
		return false
	}
	grandparent := parent.Parent
	return grandparent != nil &&
		(grandparent.Kind == ast.KindClassDeclaration || grandparent.Kind == ast.KindClassExpression)
}
