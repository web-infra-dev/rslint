package id_match

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed id_match.schema.json
var schemaJSON []byte

// constructorKeyword is the name ESLint's AST gives a class constructor.
const constructorKeyword = "constructor"

// IdMatchRule requires identifiers to match a specified regular expression.
// https://eslint.org/docs/latest/rules/id-match
var IdMatchRule = rule.Rule{
	Name:   "id-match",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		pattern, err := esregexp.Compile(opts.pattern, "u")
		if err != nil {
			// The configured pattern is not a regexp; nothing can be checked
			// against it.
			return rule.RuleListeners{}
		}

		r := &idMatch{ctx: ctx, opts: opts, pattern: pattern}
		return rule.RuleListeners{
			ast.KindIdentifier:        r.checkIdentifier,
			ast.KindPrivateIdentifier: r.checkPrivateIdentifier,
			ast.KindConstructor:       r.checkConstructor,
		}
	},
}

type idMatchOptions struct {
	pattern             string
	classFields         bool
	ignoreDestructuring bool
	onlyDeclarations    bool
	properties          bool
}

func parseOptions(options []any) idMatchOptions {
	opts := idMatchOptions{pattern: "^.+$"}
	if len(options) == 0 {
		return opts
	}
	if s, ok := options[0].(string); ok {
		opts.pattern = s
	}
	if len(options) < 2 {
		return opts
	}
	optsMap, _ := options[1].(map[string]any)
	if v, ok := optsMap["classFields"].(bool); ok {
		opts.classFields = v
	}
	if v, ok := optsMap["ignoreDestructuring"].(bool); ok {
		opts.ignoreDestructuring = v
	}
	if v, ok := optsMap["onlyDeclarations"].(bool); ok {
		opts.onlyDeclarations = v
	}
	if v, ok := optsMap["properties"].(bool); ok {
		opts.properties = v
	}
	return opts
}

type idMatch struct {
	ctx     rule.RuleContext
	opts    idMatchOptions
	pattern *esregexp.RegExp
}

func messageNotMatch(name, pattern string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notMatch",
		Description: fmt.Sprintf("Identifier '%s' does not match the pattern '%s'.", name, pattern),
	}
}

func messageNotMatchPrivate(name, pattern string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notMatchPrivate",
		Description: fmt.Sprintf("Identifier '#%s' does not match the pattern '%s'.", name, pattern),
	}
}

func (r *idMatch) isInvalid(name string) bool {
	return !r.pattern.TestOrTimeout(name)
}

// report emits the diagnostic unless the name belongs to something the author
// does not declare. Upstream decides that before dispatching on the parent
// shape; every branch it dispatches to does nothing but report, so asking here
// gives the same answer and only asks about names that are already invalid,
// which keeps the scope lookup off the path every identifier takes.
func (r *idMatch) report(node *ast.Node, name string) {
	if r.isReferenceToGlobalVariable(node, name) || r.isExternallyDeclaredType(node, name) {
		return
	}
	r.ctx.ReportNode(node, messageNotMatch(name, r.opts.pattern))
}

func (r *idMatch) reportIfInvalid(node *ast.Node, name string) {
	if r.isInvalid(name) {
		r.report(node, name)
	}
}

func (r *idMatch) checkPrivateIdentifier(node *ast.Node) {
	parent := node.Parent
	// An auto-accessor (`accessor #x = 1`) is not a class field; it is always
	// checked, as ESLint checks its `AccessorProperty`.
	isClassField := parent != nil && parent.Kind == ast.KindPropertyDeclaration &&
		!ast.HasAccessorModifier(parent) && !ast.HasAbstractModifier(parent)
	if isClassField && !r.opts.classFields {
		return
	}
	name := strings.TrimPrefix(node.Text(), "#")
	if r.isInvalid(name) {
		r.ctx.ReportNode(node, messageNotMatchPrivate(name, r.opts.pattern))
	}
}

func (r *idMatch) checkIdentifier(node *ast.Node) {
	parent := estreeParent(node)
	if parent == nil {
		return
	}
	name := node.Text()

	// JSX names are `JSXIdentifier` in ESLint's AST and never reach its
	// `Identifier` listener, so a tag or attribute name is never checked.
	if isJsxNamePosition(node) {
		return
	}
	// A name written inside a JSDoc comment stays a comment to ESLint; tsgo
	// clones it into the tree of a checked JavaScript file.
	if utils.IsInJSDocSyntax(node) {
		return
	}
	// `x as const` reads `const` as a type name, but it names no declaration.
	if parent.Kind == ast.KindTypeReference && ast.IsConstTypeReference(parent) {
		return
	}
	if parent.Kind == ast.KindMetaProperty || isImportAttributeKey(node) {
		return
	}

	if ast.IsAccessExpression(parent) {
		r.checkMemberAccess(node, name, parent)
		return
	}

	if roles, ok := describeRoles(node, parent); ok {
		for _, ro := range roles {
			// A non-computed key of an object literal has its own branch
			// upstream, gated only on `properties`.
			if ro.kind == roleProperty && !ro.grandIsPattern && ro.isKey && !ro.computed {
				if r.opts.properties && r.isInvalid(name) {
					r.report(node, name)
				}
				return
			}
		}
		for _, ro := range roles {
			if r.checkPropertyRole(node, name, ro) {
				r.report(node, name)
				return
			}
		}
		return
	}

	switch {
	case isImportBinding(parent):
		// Only the local imported identifier is checked; the name the module
		// exports it under is not the reader's to choose. Upstream compares the
		// two by name, so `import { a as a }` reports both halves.
		if local := parent.Name(); local != nil && local.Kind == ast.KindIdentifier &&
			local.Text() == name {
			r.reportIfInvalid(node, name)
		}
	case parent.Kind == ast.KindPropertyDeclaration && !ast.HasAccessorModifier(parent) &&
		!ast.HasAbstractModifier(parent):
		if r.opts.classFields && r.isInvalid(name) {
			r.report(node, name)
		}
	default:
		if r.shouldReport(node, parent, name) {
			r.report(node, name)
		}
	}
}

// checkMemberAccess mirrors upstream's `parent.type === "MemberExpression"`
// branch: the object of a member access is always checked, and its property is
// checked only when the access is the assigned-to side of an assignment.
func (r *idMatch) checkMemberAccess(node *ast.Node, name string, member *ast.Node) {
	if !r.opts.properties {
		return
	}
	if object := ast.SkipParentheses(member.Expression()); object.Kind == ast.KindIdentifier &&
		object.Text() == name {
		r.reportIfInvalid(node, name)
		return
	}
	// ESTree wraps the complete optional chain in ChainExpression. Its member
	// therefore is not the assignment target even when tsgo places the access
	// directly under the assignment after flattening that wrapper.
	if ast.IsOptionalChain(member) {
		return
	}

	effective := estreeParent(member)
	if effective == nil || effective.Kind != ast.KindBinaryExpression ||
		!ast.IsAssignmentOperator(effective.AsBinaryExpression().OperatorToken.Kind) {
		return
	}
	assignment := effective.AsBinaryExpression()

	if left := ast.SkipParentheses(assignment.Left); ast.IsAccessExpression(left) {
		if property, ok := memberPropertyName(left); ok && property == name {
			r.reportIfInvalid(node, name)
			return
		}
	}
	if !ast.IsAccessExpression(ast.SkipParentheses(assignment.Right)) {
		r.reportIfInvalid(node, name)
	}
}

// checkPropertyRole mirrors upstream's `parent.type === "Property" ||
// parent.type === "AssignmentPattern"` branch for one of the ESTree positions
// the identifier occupies, and reports whether that position produces a
// diagnostic.
func (r *idMatch) checkPropertyRole(node *ast.Node, name string, ro role) bool {
	if ro.kind == roleProperty && ro.grandIsPattern {
		if !r.opts.ignoreDestructuring && ro.shorthand && ro.valueIsAssignPattern &&
			r.isInvalid(name) {
			return true
		}
		keyEqualsValue := ro.keyName != "" && ro.keyName == ro.valueName
		if !keyEqualsValue && ro.isKey {
			// The renamed-from side of a destructuring pattern names a
			// property of the source object, not a new binding.
			return false
		}
		// A binding that merely repeats a property name is what
		// ignoreDestructuring lets through; a renamed one is not.
		if ro.valueName != "" && r.isInvalid(name) &&
			(!keyEqualsValue || !r.opts.ignoreDestructuring) {
			return true
		}
	}

	computed := ro.kind == roleProperty && ro.computed
	if (!r.opts.properties && !computed) ||
		(r.opts.ignoreDestructuring && isInsideObjectPattern(node)) {
		return false
	}
	if ro.isDefaultValue {
		// The right-hand side of a default is checked where it is declared.
		return false
	}
	// A `Property` / `AssignmentPattern` is neither a declaration nor a call,
	// so only `onlyDeclarations` can suppress the report here.
	return !r.opts.onlyDeclarations && r.isInvalid(name)
}

func (r *idMatch) shouldReport(node *ast.Node, effectiveParent *ast.Node, name string) bool {
	switch classifyParent(node, effectiveParent) {
	case parentCall, parentNew:
		return false
	case parentFunctionDeclaration, parentVariableDeclarator:
	default:
		if r.opts.onlyDeclarations {
			return false
		}
	}
	return r.isInvalid(name)
}

// isReferenceToGlobalVariable reports whether node reads a global that the
// linted file does not declare. Those names are outside the author's control,
// so upstream leaves them alone.
func (r *idMatch) isReferenceToGlobalVariable(node *ast.Node, name string) bool {
	// A name this file declares is the author's, however global the spelling.
	return r.ctx.Globals.Access(name).IsDeclared() && !isNonReferenceIdentifier(node) &&
		!r.isDeclaredInFile(node, name)
}

// isDeclaredInFile reports whether the linted file itself declares or imports
// the name node reads. Asking the reference index costs a scope-chain lookup,
// where scanning for a shadowing declaration costs a walk of the whole file —
// and this question is asked of every unmatched occurrence of a global name.
func (r *idMatch) isDeclaredInFile(node *ast.Node, name string) bool {
	if r.ctx.Refs == nil {
		return utils.IsShadowed(node, name)
	}
	return r.ctx.Refs.ResolveInFile(node) != nil
}

// isNonReferenceIdentifier reports whether an identifier names something rather
// than reading it. Import-type qualifiers name exports of another module, not
// globals in this file. It also extends utils.IsNonReferenceIdentifier with the
// two places tsgo writes one identifier where ESLint's AST writes two over the
// same range: a shorthand property, whose key half is no reference, and an
// un-aliased export specifier, whose exported half is no reference. Upstream
// dedupes such a pair by range, so the half that is not a reference is the one
// that decides.
func isNonReferenceIdentifier(node *ast.Node) bool {
	if utils.IsImportTypeSyntax(node) {
		return true
	}
	if parent := node.Parent; parent != nil && parent.Name() == node {
		switch parent.Kind {
		case ast.KindShorthandPropertyAssignment:
			return true
		case ast.KindExportSpecifier:
			if parent.AsExportSpecifier().PropertyName == nil {
				return true
			}
		}
	}
	return utils.IsNonReferenceIdentifier(node)
}

// isExternallyDeclaredType reports whether a name used in a type position is
// declared somewhere other than the linted file — `Record`, `Omit` and the
// rest of the TypeScript standard library, an ambient declaration, a global
// another file contributes.
//
// NOTE: Unlike ESLint, whose scope model has no types in it at all, rslint has
// to answer this to stay usable on TypeScript: without it every `Record<K, V>`
// in the project is reported against a pattern its author never chose. A name
// the file itself declares or imports still counts as the author's.
func (r *idMatch) isExternallyDeclaredType(node *ast.Node, name string) bool {
	if r.ctx.SourceFile == nil || !ast.IsPartOfTypeNode(node) || !isTypeGlobalReference(node) {
		return false
	}
	// The standard library's own names are answered without asking for a
	// symbol, because a lib declaration reaches Resolve only through its
	// TypeChecker fallback: asking there alone reports `Record` in a file no
	// tsconfig owns and stays quiet in one a tsconfig does. This list is the
	// type-capable global scope typescript-eslint seeds, which is where its
	// own scope model finds these names.
	if rule.IsDefaultTypeScriptTypeGlobal(name) && !r.isDeclaredInFile(node, name) {
		return true
	}
	symbol := r.ctx.Refs.Resolve(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		// A name nothing declares is still the author's to fix.
		return false
	}
	return !utils.IsSymbolDeclaredInFile(symbol, r.ctx.SourceFile)
}

// isTypeGlobalReference reports whether node is a reference that can resolve
// through TypeScript's global type scope. Declaration and property names such
// as tuple labels are authored names, while an import-type qualifier names an
// export of the imported module; neither inherits an exemption merely because
// it is spelled `Record`, `Array`, or another standard-library type name.
func isTypeGlobalReference(node *ast.Node) bool {
	return !isNonReferenceIdentifier(node)
}

// checkConstructor reports a class constructor. tsgo spells its name with a
// keyword and gives it no name node, where ESLint's AST holds an `Identifier`
// named `constructor` that its `Identifier` listener visits like any other.
func (r *idMatch) checkConstructor(node *ast.Node) {
	if r.opts.onlyDeclarations || !r.isInvalid(constructorKeyword) {
		return
	}
	start := node.Pos()
	if modifiers := node.ModifierNodes(); len(modifiers) > 0 {
		start = modifiers[len(modifiers)-1].End()
	}
	text := r.ctx.SourceFile.Text()
	start = scanner.SkipTrivia(text, start)
	end := start + len(constructorKeyword)
	// `class A { 'constructor'() {} }` is a constructor too, but its name is a
	// string literal, which ESLint's AST holds as a `Literal` and never visits.
	if end > len(text) || text[start:end] != constructorKeyword {
		return
	}
	r.ctx.ReportRange(
		core.NewTextRange(start, end),
		messageNotMatch(constructorKeyword, r.opts.pattern),
	)
}

// ---------------------------------------------------------------------------
// ESTree shape adapters
// ---------------------------------------------------------------------------

// estreeParent returns the node that would be the identifier's parent in
// ESLint's AST. Parentheses are nodes of their own in tsgo but not in ESTree,
// and a computed key's `ComputedPropertyName` wrapper is folded into the
// member it names, which ESTree marks with `computed: true` instead.
func estreeParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindComputedPropertyName {
		parent = parent.Parent
	}
	return parent
}

// isComputedKey reports whether node is the whole key expression of a computed
// member name (`{ [node]: … }`), as opposed to a part of one (`{ [node + 1]: … }`).
func isComputedKey(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindComputedPropertyName
}

// memberPropertyName returns the value ESTree exposes as
// `MemberExpression.property.name`, which exists only for an identifier-shaped
// property.
func memberPropertyName(member *ast.Node) (string, bool) {
	var property *ast.Node
	if member.Kind == ast.KindPropertyAccessExpression {
		property = member.AsPropertyAccessExpression().Name()
	} else {
		property = ast.SkipParentheses(member.AsElementAccessExpression().ArgumentExpression)
	}
	switch property.Kind {
	case ast.KindIdentifier:
		return property.Text(), true
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(property.Text(), "#"), true
	}
	return "", false
}

func isImportBinding(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
		return true
	}
	return false
}

type parentKind uint8

const (
	parentOther parentKind = iota
	parentCall
	parentNew
	parentFunctionDeclaration
	parentVariableDeclarator
)

func classifyParent(node *ast.Node, parent *ast.Node) parentKind {
	switch parent.Kind {
	case ast.KindCallExpression:
		// `import(…)` is an `ImportExpression` in ESTree, not a call.
		if parent.AsCallExpression().Expression.Kind == ast.KindImportKeyword {
			return parentOther
		}
		return parentCall
	case ast.KindNewExpression:
		return parentNew
	case ast.KindVariableDeclaration:
		// tsgo writes a catch parameter as a variable declaration; ESTree
		// hangs it off the catch clause, which is no declaration at all.
		if parent.Parent != nil && parent.Parent.Kind == ast.KindCatchClause {
			return parentOther
		}
		return parentVariableDeclarator
	case ast.KindFunctionDeclaration:
		// A bodyless TypeScript signature is a TSDeclareFunction in ESTree,
		// not one of the declarations onlyDeclarations admits.
		if parent.AsFunctionDeclaration().Body != nil {
			return parentFunctionDeclaration
		}
	case ast.KindParameter:
		// A plain parameter has no node of its own in ESTree; its name sits
		// directly among the function's parameters.
		if isPlainParameter(parent) && parent.Name() == node && parent.Parent != nil {
			return classifyParent(node, parent.Parent)
		}
	}
	return parentOther
}

func isPlainParameter(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	return declaration.DotDotDotToken == nil && declaration.Initializer == nil &&
		parameter.ModifierFlags()&ast.ModifierFlagsParameterPropertyModifier == 0
}

// isJsxNamePosition reports whether node names a JSX element or attribute.
func isJsxNamePosition(node *ast.Node) bool {
	for current := node; current.Parent != nil; current = current.Parent {
		if ast.IsJsxTagName(current) {
			return true
		}
		parent := current.Parent
		if parent.Kind == ast.KindJsxAttribute {
			return parent.Name() == current
		}
		// A member or namespaced tag name is spelled with ordinary nodes, so
		// keep climbing while the identifier can still be part of one.
		if parent.Kind != ast.KindPropertyAccessExpression &&
			parent.Kind != ast.KindJsxNamespacedName {
			return false
		}
	}
	return false
}

// isImportAttributeKey mirrors ESLint's `astUtils.isImportAttributeKey`: the
// keys of `with { type: "json" }`, in both the static and the dynamic form,
// name an import attribute rather than anything the author declares.
func isImportAttributeKey(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindImportAttribute && parent.Name() == node {
		return true
	}

	for key := node; ; {
		parent := key.Parent
		if parent == nil {
			return false
		}
		// Upstream gates only the shorthand-value alternative on `!method`, so
		// a method or accessor key counts like any other.
		switch parent.Kind {
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		default:
			return false
		}
		// A computed key is wrapped in a `ComputedPropertyName`, so this also
		// rejects the `computed: true` form upstream rejects — on the outer key
		// the walk steps up to as well as on the one it started from.
		if parent.Name() != key || key.Kind == ast.KindComputedPropertyName {
			return false
		}
		object := parent.Parent
		if object == nil || object.Kind != ast.KindObjectLiteralExpression {
			return false
		}

		outer := ast.WalkUpParenthesizedExpressions(object.Parent)
		if outer == nil {
			return false
		}
		if outer.Kind == ast.KindCallExpression {
			call := outer.AsCallExpression()
			if call.Expression.Kind != ast.KindImportKeyword || call.Arguments == nil ||
				len(call.Arguments.Nodes) < 2 {
				return false
			}
			return ast.SkipParentheses(call.Arguments.Nodes[1]) == object
		}
		if outer.Kind == ast.KindPropertyAssignment &&
			ast.SkipParentheses(outer.AsPropertyAssignment().Initializer) == object {
			key = outer.Name()
			if key == nil {
				return false
			}
			continue
		}
		return false
	}
}

// isInsideObjectPattern reports whether any ancestor of node is a destructuring
// object pattern.
func isInsideObjectPattern(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindObjectBindingPattern:
			return true
		case ast.KindObjectLiteralExpression:
			if isPatternTarget(parent) {
				return true
			}
		}
	}
	return false
}

// isPatternTarget reports whether an object or array literal stands on the
// assigned-to side of an assignment, where ESTree parses it as a pattern
// instead of a literal.
func isPatternTarget(literal *ast.Node) bool {
	return literal.Parent != nil && ast.IsAssignmentTarget(literal)
}

// isAssignmentPattern reports whether an `x = y` expression is ESTree's
// `AssignmentPattern`, i.e. a default inside a destructuring target rather
// than an assignment of its own.
func isAssignmentPattern(node *ast.Node) bool {
	return node.Kind == ast.KindBinaryExpression &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken &&
		node.Parent != nil && ast.IsAssignmentTarget(node)
}

// ---------------------------------------------------------------------------
// ESTree Property / AssignmentPattern positions
// ---------------------------------------------------------------------------

type roleKind uint8

const (
	roleProperty roleKind = iota
	roleAssignPattern
)

// role describes one ESTree `Property` or `AssignmentPattern` position an
// identifier occupies. tsgo writes `{ a }` as a single identifier where ESTree
// writes two nodes over the same range — a property key and its value — so one
// identifier can stand in more than one position, and the rule reports when
// any of them would.
type role struct {
	kind roleKind

	// Property positions.
	grandIsPattern       bool
	computed             bool
	shorthand            bool
	keyName              string
	valueName            string
	valueIsAssignPattern bool
	isKey                bool

	// AssignmentPattern positions.
	isDefaultValue bool
}

func describeRoles(node *ast.Node, parent *ast.Node) ([]role, bool) {
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Parent == nil || parent.Parent.Kind != ast.KindObjectLiteralExpression {
			return nil, false
		}
		return objectLiteralRoles(node, parent)
	case ast.KindBindingElement:
		return bindingElementRoles(node, parent)
	case ast.KindParameter:
		return parameterRoles(node, parent)
	case ast.KindBinaryExpression:
		return assignmentPatternRoles(node, parent)
	}
	return nil, false
}

func objectLiteralRoles(node *ast.Node, member *ast.Node) ([]role, bool) {
	name := member.Name()
	property := role{
		kind:           roleProperty,
		grandIsPattern: isPatternTarget(member.Parent),
		computed:       name != nil && name.Kind == ast.KindComputedPropertyName,
		shorthand:      member.Kind == ast.KindShorthandPropertyAssignment,
		keyName:        propertyKeyName(name),
	}
	property.valueName, property.valueIsAssignPattern = objectLiteralValue(member)

	if isComputedKey(node) || name == node {
		property.isKey = true
		return []role{property}, true
	}
	if member.Kind == ast.KindPropertyAssignment &&
		ast.SkipParentheses(member.AsPropertyAssignment().Initializer) == node {
		return []role{property}, true
	}
	// `({ a = b } = o)` — `b` is the right-hand side of ESTree's
	// AssignmentPattern, which the rule checks where it is declared instead.
	if member.Kind == ast.KindShorthandPropertyAssignment {
		if def := member.AsShorthandPropertyAssignment().ObjectAssignmentInitializer; def != nil &&
			ast.SkipParentheses(def) == node {
			return []role{{kind: roleAssignPattern, isDefaultValue: true}}, true
		}
	}
	return nil, false
}

// objectLiteralValue returns what ESTree exposes as `Property.value.name`, and
// whether that value is an `AssignmentPattern`.
func objectLiteralValue(member *ast.Node) (string, bool) {
	switch member.Kind {
	case ast.KindPropertyAssignment:
		value := ast.SkipParentheses(member.AsPropertyAssignment().Initializer)
		if isAssignmentPattern(value) {
			return "", true
		}
		if value.Kind == ast.KindIdentifier {
			return value.Text(), false
		}
	case ast.KindShorthandPropertyAssignment:
		if member.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
			return "", true
		}
		if name := member.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), false
		}
	}
	// A method or accessor holds a function expression, which has no `name`.
	return "", false
}

func bindingElementRoles(node *ast.Node, element *ast.Node) ([]role, bool) {
	data := element.AsBindingElement()
	if data.DotDotDotToken != nil {
		// ESTree writes a rest element, not a property.
		return nil, false
	}
	name := element.Name()
	hasDefault := data.Initializer != nil

	if element.Parent == nil || element.Parent.Kind != ast.KindObjectBindingPattern {
		// An array pattern holds the binding directly, so only a default
		// creates an ESTree node around it.
		if !hasDefault {
			return nil, false
		}
		if name == node {
			return []role{{kind: roleAssignPattern}}, true
		}
		if ast.SkipParentheses(data.Initializer) == node {
			return []role{{kind: roleAssignPattern, isDefaultValue: true}}, true
		}
		return nil, false
	}

	property := role{
		kind:                 roleProperty,
		grandIsPattern:       true,
		computed:             data.PropertyName != nil && data.PropertyName.Kind == ast.KindComputedPropertyName,
		shorthand:            data.PropertyName == nil,
		valueIsAssignPattern: hasDefault,
	}
	if data.PropertyName != nil {
		property.keyName = propertyKeyName(data.PropertyName)
	} else {
		property.keyName = propertyKeyName(name)
	}
	if !hasDefault && name != nil && name.Kind == ast.KindIdentifier {
		property.valueName = name.Text()
	}

	if isComputedKey(node) || data.PropertyName == node {
		property.isKey = true
		return []role{property}, true
	}
	if name == node {
		roles := make([]role, 0, 2)
		if data.PropertyName == nil {
			key := property
			key.isKey = true
			roles = append(roles, key)
		}
		if hasDefault {
			roles = append(roles, role{kind: roleAssignPattern})
		} else if data.PropertyName != nil {
			roles = append(roles, property)
		}
		return roles, len(roles) > 0
	}
	if hasDefault && ast.SkipParentheses(data.Initializer) == node {
		return []role{{kind: roleAssignPattern, isDefaultValue: true}}, true
	}
	return nil, false
}

func parameterRoles(node *ast.Node, parameter *ast.Node) ([]role, bool) {
	data := parameter.AsParameterDeclaration()
	if data.Initializer == nil {
		return nil, false
	}
	if parameter.Name() == node {
		return []role{{kind: roleAssignPattern}}, true
	}
	if ast.SkipParentheses(data.Initializer) == node {
		return []role{{kind: roleAssignPattern, isDefaultValue: true}}, true
	}
	return nil, false
}

func assignmentPatternRoles(node *ast.Node, expression *ast.Node) ([]role, bool) {
	if !isAssignmentPattern(expression) {
		return nil, false
	}
	data := expression.AsBinaryExpression()
	if ast.SkipParentheses(data.Left) == node {
		return []role{{kind: roleAssignPattern}}, true
	}
	if ast.SkipParentheses(data.Right) == node {
		return []role{{kind: roleAssignPattern, isDefaultValue: true}}, true
	}
	return nil, false
}

// propertyKeyName returns what ESTree exposes as `Property.key.name`, which a
// string or numeric key does not have.
func propertyKeyName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if name.Kind == ast.KindIdentifier {
		return name.Text()
	}
	return ""
}
