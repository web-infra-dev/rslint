package utils

import (
	"slices"

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

// IsPlainClassMember reports whether a tsgo class member maps to ESTree's
// MethodDefinition or PropertyDefinition. Abstract members and auto-accessors
// use TypeScript-specific ESTree node kinds instead.
func IsPlainClassMember(member *ast.Node) bool {
	return member != nil &&
		!ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract|ast.ModifierFlagsAccessor)
}

// IsImportAttributeKey reports whether node is an import attribute key in a
// static import/re-export or in a dynamic import options object. These keys
// are fixed by the import-attributes protocol rather than freely chosen
// identifier or property names.
func IsImportAttributeKey(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent

	// import data from "./data.json" with { type: "json" }
	if parent.Kind == ast.KindImportAttribute && parent.AsImportAttribute().Name() == node {
		return true
	}

	// import("./data.json", { with: { type: "json" } })
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return false
	}
	if parent.Name() != node {
		return false
	}
	objectExpression := parent.Parent
	if objectExpression == nil || objectExpression.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	outer := OutermostParenthesizedExpression(objectExpression)
	container := outer.Parent
	if container == nil {
		return false
	}
	if container.Kind == ast.KindCallExpression {
		call := container.AsCallExpression()
		return call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword &&
			call.Arguments != nil && len(call.Arguments.Nodes) > 1 && call.Arguments.Nodes[1] == outer
	}
	if container.Kind == ast.KindPropertyAssignment && container.AsPropertyAssignment().Initializer == outer &&
		!ast.IsComputedPropertyName(container.Name()) {
		return IsImportAttributeKey(container.Name())
	}
	return false
}

// IsCommaOperator reports whether node is a BinaryExpression whose operator is
// the comma token — tsgo's collapsed form of ESLint's SequenceExpression.
func IsCommaOperator(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	return bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken
}

// IsLooseEqualityOperator reports whether kind is the == or != token, as
// opposed to their strict (===, !==) counterparts.
func IsLooseEqualityOperator(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsToken || kind == ast.KindExclamationEqualsToken
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
// override returns the final explicitly authored access for a name. When it
// turns the referenced name `off`, the identifier no longer resolves to a
// known global. Pass nil to skip that check (e.g. for callers whose upstream
// ESLint rule doesn't consult scope at all).
func IsGlobalParseIntCallee(callee *ast.Node, override func(string) GlobalAccess) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	if ast.IsIdentifier(callee) {
		if callee.AsIdentifier().Text != "parseInt" || IsShadowed(callee, "parseInt") {
			return false
		}
		return override == nil || override("parseInt") != GlobalAccessOff
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
	return override == nil || override("Number") != GlobalAccessOff
}

// globalObjectAliasNames are the well-known references to the global object
// through which a built-in constructor can also be reached (`globalThis.Foo`,
// `window.Foo`, …).
var globalObjectAliasNames = [...]string{"globalThis", "window", "self", "global"}

// IsBuiltinGlobalCallee reports whether callee references the built-in global
// named name — either directly (an unshadowed identifier) or through one of
// the well-known global-object aliases (`globalThis.RegExp`, `window.RegExp`,
// `self.RegExp`, `global.RegExp`). Outer parentheses and TS assertion
// wrappers are transparent on both the callee and the alias receiver.
//
// isDeclaredGlobal reports whether a name resolves to a known global (i.e.
// hasn't been turned `off` by a `/* global name: off */` comment or
// languageOptions.globals entry); pass ctx.Globals.Access(name).IsDeclared.
func IsBuiltinGlobalCallee(callee *ast.Node, name string, isDeclaredGlobal func(string) bool) bool {
	callee = SkipAssertionsAndParens(callee)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text != name || IsShadowed(callee, name) {
			return false
		}
		return isDeclaredGlobal(name)
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		if access.Name().AsIdentifier().Text != name {
			return false
		}
		return isGlobalObjectAlias(access.Expression, isDeclaredGlobal)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return false
		}
		value, ok := GetStaticExpressionValue(SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok || value != name {
			return false
		}
		return isGlobalObjectAlias(access.Expression, isDeclaredGlobal)
	}

	return false
}

func isGlobalObjectAlias(node *ast.Node, isDeclaredGlobal func(string) bool) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.AsIdentifier().Text
	for _, alias := range globalObjectAliasNames {
		if name == alias {
			return !IsShadowed(node, name) && isDeclaredGlobal(name)
		}
	}
	return false
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

// VisitDescendants walks node and everything beneath it, depth-first in source
// order. Returning false from visit leaves that node's own children unvisited
// and the walk carries on with its siblings, which is how a caller stops
// descending into a subtree it has already accounted for.
func VisitDescendants(node *ast.Node, visit func(*ast.Node) bool) {
	if node == nil || !visit(node) {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		VisitDescendants(child, visit)
		return false
	})
}

// IsJSDocSyntaxNode reports whether node is the root of syntax that tsgo
// synthesized from a JSDoc comment. Callers performing a depth-first walk can
// prune the whole subtree as soon as this returns true; ordinary source nodes
// do not require an ancestor walk.
func IsJSDocSyntaxNode(node *ast.Node) bool {
	return node != nil && (node.Flags&(ast.NodeFlagsJSDoc|ast.NodeFlagsReparsed) != 0 || ast.IsJSDocNode(node))
}

// JSDocTypeCastExpression returns the authored runtime expression inside
// a JavaScript JSDoc @type or @satisfies cast. ESTree exposes only that
// expression, while tsgo inserts AsExpression/SatisfiesExpression wrappers and
// a reparsed type. Walkers should visit the returned expression instead.
func JSDocTypeCastExpression(node *ast.Node) *ast.Node {
	if ast.IsJSDocTypeAssertion(node) {
		assertion := node.AsParenthesizedExpression().Expression
		if assertion == nil || assertion.Kind != ast.KindAsExpression {
			return nil
		}
		return assertion.AsAsExpression().Expression
	}
	if node == nil || !ast.IsInJSFile(node) {
		return nil
	}
	switch node.Kind {
	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		if IsJSDocSyntaxNode(expression.Type) {
			return expression.Expression
		}
	case ast.KindSatisfiesExpression:
		expression := node.AsSatisfiesExpression()
		if IsJSDocSyntaxNode(expression.Type) {
			return expression.Expression
		}
	}
	return nil
}

// IsJSDocTypeCastWrapper reports whether node is a wrapper that tsgo inserts
// around a JavaScript JSDoc @type or @satisfies cast.
func IsJSDocTypeCastWrapper(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsJSDocTypeAssertion(node) {
		return true
	}
	if JSDocTypeCastExpression(node) != nil {
		return true
	}
	return node.Kind == ast.KindAsExpression && node.Parent != nil && ast.IsJSDocTypeAssertion(node.Parent)
}

// ESTreeRuntimeExpression removes syntax that ESTree does not expose around a
// runtime expression: source parentheses and wrappers synthesized from JSDoc
// @type/@satisfies casts. Authored TypeScript assertions remain intact.
func ESTreeRuntimeExpression(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		if node.Kind != ast.KindAsExpression && node.Kind != ast.KindSatisfiesExpression {
			return node
		}
		expression := JSDocTypeCastExpression(node)
		if expression == nil {
			return node
		}
		node = expression
	}
	return nil
}

// ESTreeParent returns the first parent that ESTree exposes, skipping source
// parentheses and wrappers synthesized from JSDoc casts.
func ESTreeParent(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			parent = parent.Parent
		case ast.KindAsExpression, ast.KindSatisfiesExpression:
			if !IsJSDocTypeCastWrapper(parent) {
				return parent
			}
			parent = parent.Parent
		default:
			return parent
		}
	}
	return nil
}

// ESTreeMembers removes empty class elements, which tsgo exposes as
// SemicolonClassElement nodes but ESTree omits from ClassBody.body. The common
// no-semicolon path returns the AST-owned slice without allocating; filtering
// clones into a separate backing array so callers cannot rewrite that list.
func ESTreeMembers(members []*ast.Node) []*ast.Node {
	for index, member := range members {
		if member.Kind != ast.KindSemicolonClassElement {
			continue
		}
		filtered := make([]*ast.Node, 0, len(members)-1)
		filtered = append(filtered, members[:index]...)
		for _, remaining := range members[index+1:] {
			if remaining.Kind != ast.KindSemicolonClassElement {
				filtered = append(filtered, remaining)
			}
		}
		return filtered
	}
	return members
}

// ESTreeParameters returns only parameters authored in source. tsgo prepends
// a reparsed `this` parameter for JSDoc @this, but ESTree keeps the tag solely
// as a comment.
func ESTreeParameters(node *ast.Node) []*ast.Node {
	return slices.DeleteFunc(slices.Clone(node.Parameters()), IsJSDocSyntaxNode)
}

// ESTreeTypeParameters returns only type parameters authored in source. tsgo
// materializes JSDoc @template tags on function-like nodes, while ESTree keeps
// those tags solely as comments.
func ESTreeTypeParameters(node *ast.Node) []*ast.Node {
	return slices.DeleteFunc(slices.Clone(node.TypeParameters()), IsJSDocSyntaxNode)
}

// ESTreeType returns the source-authored type annotation, excluding a type
// that tsgo copied from JSDoc onto an otherwise ordinary declaration.
func ESTreeType(node *ast.Node) *ast.Node {
	typeNode := node.Type()
	if IsJSDocSyntaxNode(typeNode) {
		return nil
	}
	return typeNode
}

// ESTreeModifierFlags excludes modifiers synthesized from JSDoc tags such as
// @private and @override.
func ESTreeModifierFlags(node *ast.Node) ast.ModifierFlags {
	var flags ast.ModifierFlags
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if !IsJSDocSyntaxNode(modifier) {
				flags |= ast.ModifierToFlag(modifier.Kind)
			}
		}
	}
	return flags
}

// IsInJSDocSyntax reports whether node came from syntax parsed inside a JSDoc
// comment. TypeScript-Go deep-clones some of these nodes into the executable
// tree, preserving NodeFlagsJSDoc on the cloned subtree. Espree and
// typescript-eslint keep the same text as comments, so none of these names
// are references for scope-sensitive rules like no-undef and no-undefined.
func IsInJSDocSyntax(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Flags&ast.NodeFlagsJSDoc != 0 || ast.IsJSDocNode(current) {
			return true
		}
		if current.Kind == ast.KindSourceFile {
			return false
		}
	}
	return false
}

// IsImportTypeSyntax reports whether node belongs to the argument,
// attributes, or qualifier of an ImportType. Type arguments are siblings of
// those fields and remain references.
func IsImportTypeSyntax(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind != ast.KindImportType {
		current = current.Parent
	}
	if current.Parent == nil || current.Parent.Kind != ast.KindImportType {
		return false
	}
	importType := current.Parent.AsImportTypeNode()
	return importType != nil &&
		(importType.Argument == current || importType.Attributes == current || importType.Qualifier == current)
}
