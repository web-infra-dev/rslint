package prefer_stateless_function

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options carries the user-supplied options for this rule.
type Options struct {
	IgnorePureComponents bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["ignorePureComponents"].(bool); ok {
			opts.IgnorePureComponents = v
		}
	}
	return opts
}

// componentFlags accumulates the per-component disqualifiers checked by
// upstream's `prefer-stateless-function`. A component is reported only when
// every flag remains false (and the basic ES5/ES6 component shape is met).
type componentFlags struct {
	hasOtherProperty     bool
	useThis              bool
	useRef               bool
	invalidReturn        bool
	hasChildContextTypes bool
	useDecorators        bool
	hasSCU               bool
}

// classKeywordStart returns the start position of the `class` keyword for a
// ClassDeclaration / ClassExpression, skipping any leading `export` /
// `export default` modifiers tsgo inlines into the node's modifier list.
//
// ESTree wraps `export class …` in `ExportNamedDeclaration` /
// `ExportDefaultDeclaration` and exposes the inner `ClassDeclaration`
// starting at the `class` keyword. tsgo flattens this; we recover the same
// position to match upstream's report range exactly. Decorators / TS
// `abstract` / `declare` modifiers are PART of the class's range upstream
// (they are kept on TSESTree's ClassDeclaration), so we stop trimming the
// moment a non-export/default modifier appears.
func classKeywordStart(text string, node *ast.Node) int {
	mods := node.Modifiers()
	pos := node.Pos()
	if mods != nil {
		for _, mod := range mods.Nodes {
			switch mod.Kind {
			case ast.KindExportKeyword, ast.KindDefaultKeyword:
				pos = mod.End()
			default:
				return scanner.SkipTrivia(text, mod.Pos())
			}
		}
	}
	return scanner.SkipTrivia(text, pos)
}

// hasDecorators reports whether `node` carries one or more `@decorator`
// modifiers. tsgo inlines decorators into the modifier list (ESTree exposes
// them via `node.decorators`).
func hasDecorators(node *ast.Node) bool {
	mods := node.Modifiers()
	if mods == nil {
		return false
	}
	for _, mod := range mods.Nodes {
		if mod.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}

// componentBindingName returns the identifier text the component is bound
// under — class's own name, parent VariableDeclaration's binding for
// anonymous ClassExpression, or surrounding `var X = createReactClass({...})`
// binding for ES5. Returns "" when no static name is available. Used as the
// fallback identity when no TypeChecker is present (see externalCCT).
func componentBindingName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			if binding := parent.AsVariableDeclaration().Name(); binding != nil && binding.Kind == ast.KindIdentifier {
				return binding.AsIdentifier().Text
			}
		}
	case ast.KindObjectLiteralExpression:
		cur := node
		for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
			cur = cur.Parent
		}
		if cur.Parent == nil {
			return ""
		}
		switch cur.Parent.Kind {
		case ast.KindCallExpression, ast.KindNewExpression:
		default:
			return ""
		}
		vd := cur.Parent.Parent
		if vd == nil || vd.Kind != ast.KindVariableDeclaration {
			return ""
		}
		if binding := vd.AsVariableDeclaration().Name(); binding != nil && binding.Kind == ast.KindIdentifier {
			return binding.AsIdentifier().Text
		}
	}
	return ""
}

// memberName returns the static text of a class-member / object-property key
// for the allow-list comparisons. Mirrors upstream's `astUtils.getPropertyName`:
//
//   - Identifier `foo`            → "foo"
//   - StringLiteral `"foo"`       → "foo"
//   - NumericLiteral `0`          → "0" (decimal-normalised)
//   - Template “ `foo` “        → "foo"
//   - PrivateIdentifier `#foo`    → "foo" (strip `#`, matching ESTree's
//     PrivateIdentifier.name semantics that upstream's `getPropertyName`
//     reads via `nameNode.name`)
//   - ComputedPropertyName `[…]`  → static text of the inner expression when
//     it resolves to a literal-like value, "" otherwise
//
// Dynamic computed keys (`[expr]`) and shapes without a static name fall
// through to "" — the allow-list comparisons then all evaluate to false and
// the property is treated as "other", matching upstream where
// `getPropertyName` returns `undefined`.
func memberName(member *ast.Node) string {
	key := member.Name()
	if key == nil {
		return ""
	}
	if key.Kind == ast.KindPrivateIdentifier {
		return strings.TrimPrefix(key.AsPrivateIdentifier().Text, "#")
	}
	if name, ok := utils.GetStaticPropertyName(key); ok {
		return name
	}
	return ""
}

// isUselessConstructor mirrors upstream's `isRedundantSuperCall`. A
// constructor counts as useless iff:
//
//   - body has exactly one statement that is `super(...)`
//   - every constructor parameter is an Identifier or RestElement (no
//     destructuring / defaults — those have side effects)
//   - the super-call args are either `...arguments` or a 1:1 pass-through of
//     the constructor's parameters (Identifier→Identifier or
//     RestElement→SpreadElement, same name)
func isUselessConstructor(member *ast.Node) bool {
	if member.Kind != ast.KindConstructor {
		return false
	}
	cd := member.AsConstructorDeclaration()
	if cd.Body == nil {
		return false
	}
	body := cd.Body.AsBlock()
	if body == nil || body.Statements == nil || len(body.Statements.Nodes) != 1 {
		return false
	}
	stmt := body.Statements.Nodes[0]
	if stmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := ast.SkipParentheses(stmt.AsExpressionStatement().Expression)
	if expr.Kind != ast.KindCallExpression {
		return false
	}
	call := expr.AsCallExpression()
	if ast.SkipParentheses(call.Expression).Kind != ast.KindSuperKeyword {
		return false
	}

	params := cd.Parameters
	var paramNodes []*ast.Node
	if params != nil {
		paramNodes = params.Nodes
	}
	for _, p := range paramNodes {
		pd := p.AsParameterDeclaration()
		if pd == nil {
			return false
		}
		if pd.Initializer != nil {
			return false
		}
		nameNode := pd.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
	}

	var argNodes []*ast.Node
	if call.Arguments != nil {
		argNodes = call.Arguments.Nodes
	}

	if len(argNodes) == 1 {
		a := ast.SkipParentheses(argNodes[0])
		if a.Kind == ast.KindSpreadElement {
			inner := ast.SkipParentheses(a.AsSpreadElement().Expression)
			if inner.Kind == ast.KindIdentifier && inner.AsIdentifier().Text == "arguments" {
				return true
			}
		}
	}

	if len(paramNodes) != len(argNodes) {
		return false
	}
	for i, p := range paramNodes {
		pd := p.AsParameterDeclaration()
		paramName := pd.Name().AsIdentifier().Text
		paramIsRest := pd.DotDotDotToken != nil
		arg := ast.SkipParentheses(argNodes[i])
		if paramIsRest {
			if arg.Kind != ast.KindSpreadElement {
				return false
			}
			inner := ast.SkipParentheses(arg.AsSpreadElement().Expression)
			if inner.Kind != ast.KindIdentifier || inner.AsIdentifier().Text != paramName {
				return false
			}
			continue
		}
		if arg.Kind != ast.KindIdentifier || arg.AsIdentifier().Text != paramName {
			return false
		}
	}
	return true
}

// isAllowedClassMember reports whether `member` of an ES6 class is part of
// upstream's allow-list (displayName / propTypes / contextTypes / defaultProps
// / render / `props` with type annotation / useless constructor). Constructors
// outside that "useless" niche are NOT allowed — they're counted as "other".
func isAllowedClassMember(member *ast.Node) bool {
	if member.Kind == ast.KindConstructor {
		return isUselessConstructor(member)
	}
	name := memberName(member)
	switch name {
	case "displayName", "propTypes", "contextTypes", "defaultProps", "render":
		return true
	case "props":
		if member.Kind == ast.KindPropertyDeclaration {
			pd := member.AsPropertyDeclaration()
			if pd.Type != nil {
				return true
			}
		}
	}
	return false
}

// isAllowedES5Property mirrors `isAllowedClassMember` for an ES5
// createReactClass object literal: the same allow-list, minus useless
// constructor (no constructor concept) and minus the `props` typeAnnotation
// branch (no type annotations on object-literal properties in ESLint's model).
func isAllowedES5Property(prop *ast.Node) bool {
	name := memberName(prop)
	switch name {
	case "displayName", "propTypes", "contextTypes", "defaultProps", "render":
		return true
	}
	return false
}

// isThisExpression reports whether `node` (after stripping parens) is `this`.
func isThisExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.SkipParentheses(node).Kind == ast.KindThisKeyword
}

// staticAccessName returns the name expressed by a member access on `this`
// when that name resolves to a static identifier or string literal — matching
// upstream's `(node.property.name || node.property.value)` pattern.
//
//   - `this.foo`        → ("foo", true)        — Identifier key
//   - `this['foo']`     → ("foo", true)        — string-literal key
//   - `this[bar]`       → ("bar", true)        — Identifier key (variable)
//   - `this[0]`         → (any, false)         — numeric / non-name key
//   - `this[expr]`      → ("", false)          — dynamic key
//
// The rationale: upstream's rule treats Identifier-keyed `this[bar]` AS-IF
// the static name were `bar`, which is the same as `this.bar`. That's why
// `this[bar]` qualifies as "useThis" (matches upstream test).
func staticAccessName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		pa := node.AsPropertyAccessExpression()
		nm := pa.Name()
		if nm == nil || nm.Kind != ast.KindIdentifier {
			return "", false
		}
		return nm.AsIdentifier().Text, true
	case ast.KindElementAccessExpression:
		ea := node.AsElementAccessExpression()
		arg := ast.SkipParentheses(ea.ArgumentExpression)
		switch arg.Kind {
		case ast.KindStringLiteral:
			return arg.AsStringLiteral().Text, true
		case ast.KindNoSubstitutionTemplateLiteral:
			return arg.AsNoSubstitutionTemplateLiteral().Text, true
		case ast.KindIdentifier:
			return arg.AsIdentifier().Text, true
		}
	}
	return "", false
}

// componentBoundary reports whether `node` is the start of a separate React
// component context — its body must be analyzed independently and must not
// be counted as part of an outer component's body. ClassDeclaration /
// ClassExpression and createReactClass(...) ObjectLiteralExpressions both
// qualify. The body walk uses this to stop at nested boundaries.
func componentBoundary(node *ast.Node, pragma, createClass string) bool {
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return true
	case ast.KindObjectLiteralExpression:
		return isES5Component(node, pragma, createClass)
	}
	return false
}

// isES5Component reports whether `obj` (an ObjectLiteralExpression) is an
// argument of a createReactClass / `<pragma>.createClass` call.
//
// Mirrors upstream's `componentUtil.isES5Component` exactly: the check is
// `node.parent && node.parent.callee` followed by a callee identity match.
// Two notable consequences we preserve for parity:
//
//   - BOTH `CallExpression` and `NewExpression` parents qualify, because
//     `.callee` exists on both ESTree kinds. `new createReactClass({...})`
//     is non-idiomatic but syntactically valid and upstream registers it.
//   - The check does NOT verify which argument position `obj` occupies —
//     any ObjectLiteralExpression whose direct parent is a createClass call
//     qualifies, even non-first arguments. This matches upstream's
//     permissive behavior; in practice non-first-arg uses are vanishingly
//     rare and would still be filtered out by the rule's other gates
//     (hasOtherProperty etc.).
//
// Pass empty strings for pragma / createClass to default to
// `DefaultReactPragma` / `DefaultReactCreateClass`.
func isES5Component(obj *ast.Node, pragma, createClass string) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	cur := obj
	for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
		cur = cur.Parent
	}
	parent := cur.Parent
	if parent == nil {
		return false
	}
	var callee *ast.Node
	switch parent.Kind {
	case ast.KindCallExpression:
		callee = parent.AsCallExpression().Expression
	case ast.KindNewExpression:
		callee = parent.AsNewExpression().Expression
	default:
		return false
	}
	return isCreateClassCallee(callee, pragma, createClass)
}

// isCreateClassCallee mirrors `reactutil.IsCreateClassCall` but operates on a
// callee node directly so the same logic is shared between Call- and
// New-expression code paths. Parentheses are skipped on both the callee
// itself and the pragma identifier.
func isCreateClassCallee(callee *ast.Node, pragma, createClass string) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = reactutil.DefaultReactPragma
	}
	if createClass == "" {
		createClass = reactutil.DefaultReactCreateClass
	}
	callee = ast.SkipParentheses(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == createClass
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == createClass
	}
	return false
}

// flagsAnalyzer walks a component body and accumulates the disqualifier flags.
// The walk stops at nested component boundaries (ClassDeclaration /
// ClassExpression / createReactClass object arg) so a nested component's
// `this.X`, JSX refs, or invalid render returns don't pollute the outer one.
type flagsAnalyzer struct {
	flags       *componentFlags
	pragma      string
	createClass string
	allowNull   bool
}

// analyzeBody walks the body of a class/object member.
func (a *flagsAnalyzer) analyzeBody(body *ast.Node) {
	if body == nil {
		return
	}
	a.walk(body)
}

// walk visits every descendant of `node` until it hits a nested component
// boundary. Per-kind handlers route to the right flag; ReturnStatement is
// handled inline because its render-scope membership is decided per-statement
// via `returnIsInsideRender`.
func (a *flagsAnalyzer) walk(node *ast.Node) {
	if node == nil {
		return
	}
	if componentBoundary(node, a.pragma, a.createClass) {
		return
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		a.handleMemberAccess(node)
	case ast.KindVariableDeclaration:
		a.handleVariableDeclaration(node)
	case ast.KindJsxAttribute:
		a.handleJsxAttribute(node)
	case ast.KindReturnStatement:
		if returnIsInsideRender(node) {
			a.handleReturnStatement(node)
		}
	}
	// Recurse — a method's body may contain nested arrow functions whose JSX
	// or `this` accesses still belong to the enclosing component.
	node.ForEachChild(func(child *ast.Node) bool {
		a.walk(child)
		return false
	})
}

// returnIsInsideRender mirrors upstream's scope-walk that finds the nearest
// enclosing `MethodDefinition` or `Property` whose key.name is "render".
//
// Upstream walks ESLint scope chain; we approximate by walking AST parents
// and treating each FunctionLike encountered the same way. The intent (per
// upstream) is to handle BOTH ES6 `render() {…}` and ES5
// `render: function(){}` / shorthand methods. A subtle consequence — also
// present in upstream — is that ANY object literal whose property is named
// `render` qualifies, even when the object is unrelated to React (e.g. a
// nested config object inside a non-render method). We mirror this exactly:
// it is upstream's intentional `Property` arm that produces this side
// effect.
//
// The walk stops at the first FunctionLike whose parent is a recognized
// method-style container (the equivalent of ESLint's `scope.block.parent`
// match). FunctionExpression / ArrowFunction / FunctionDeclaration whose
// parent is NOT a property-assignment container do NOT terminate the walk —
// upstream's `scope.upper` continues past them, so we do too. This covers
// inline `<div onClick={function(){return X;}}/>` patterns where the
// enclosing render container is several scopes up.
func returnIsInsideRender(node *ast.Node) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			// Direct class-body or object-shorthand method. tsgo represents
			// these as the FunctionLike node itself (no FunctionExpression
			// wrapper), so this is the analog of upstream's
			// `scope.block.parent === MethodDefinition` (class body) or
			// `... === Property && method:true` (object shorthand) match.
			return memberName(p) == "render"
		case ast.KindConstructor:
			// Constructor has no `key.name === 'render'` — upstream's
			// `blockNode.key.name === 'render'` returns false.
			return false
		case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindFunctionDeclaration:
			// Only counts as a method-style container when the FunctionLike
			// is the value of an object-literal PropertyAssignment
			// (ESTree's `Property` with non-method `value: FunctionExpression`).
			// Other parents (PropertyDeclaration class-field, ReturnStatement
			// for a returned closure, JsxExpression for inline event
			// handlers, …) keep the scope walk going outward — match
			// upstream's `scope.upper` continuation.
			if p.Parent != nil && p.Parent.Kind == ast.KindPropertyAssignment {
				key := p.Parent.AsPropertyAssignment().Name()
				if key != nil && key.Kind == ast.KindIdentifier {
					return key.AsIdentifier().Text == "render"
				}
				return false
			}
		}
	}
	return false
}

func (a *flagsAnalyzer) handleMemberAccess(node *ast.Node) {
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
	}
	if !isThisExpression(receiver) {
		return
	}
	name, ok := staticAccessName(node)
	if !ok {
		// Dynamic computed key (e.g. `this[expr]`): upstream's
		// `(node.property.name || node.property.value)` falls back to
		// undefined, which is neither "props" nor "context", so it
		// flips markThisAsUsed.
		a.flags.useThis = true
		return
	}
	if name == "props" || name == "context" {
		// markPropsOrContextAsUsed — does NOT disqualify.
		return
	}
	a.flags.useThis = true
}

func (a *flagsAnalyzer) handleVariableDeclaration(node *ast.Node) {
	vd := node.AsVariableDeclaration()
	if vd.Initializer == nil || !isThisExpression(vd.Initializer) {
		return
	}
	nameNode := vd.Name()
	if nameNode == nil || nameNode.Kind != ast.KindObjectBindingPattern {
		return
	}
	bp := nameNode.AsBindingPattern()
	if bp.Elements == nil {
		return
	}
	useThis := false
	for _, elem := range bp.Elements.Nodes {
		if elem.Kind != ast.KindBindingElement {
			continue
		}
		be := elem.AsBindingElement()
		// Rest binding (`...rest = this`): upstream's
		// `astUtil.getPropertyName(RestElement)` returns null/undefined; the
		// `name !== 'props' && name !== 'context'` predicate evaluates to
		// `true`, so a rest binding flips useThis the same way any non-
		// allowed key would. (We previously mistakenly skipped these — the
		// effect was that `let { ...rest } = this` failed to mark useThis
		// and the component slipped through as a "should be pure"
		// candidate.)
		if be.DotDotDotToken != nil {
			useThis = true
			break
		}
		var keyName string
		if be.PropertyName != nil {
			keyName = staticBindingKeyName(be.PropertyName)
		} else {
			local := be.Name()
			if local != nil && local.Kind == ast.KindIdentifier {
				keyName = local.AsIdentifier().Text
			}
		}
		if keyName != "props" && keyName != "context" {
			useThis = true
			break
		}
	}
	if useThis {
		a.flags.useThis = true
	}
}

// staticBindingKeyName returns the static text of an ObjectBindingPattern
// property name (`{ foo: localBinding }` → "foo"). Returns "" for dynamic
// computed keys, mirroring upstream's `getPropertyName` undefined fallthrough.
func staticBindingKeyName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindComputedPropertyName:
		// `{ [expr]: x }` — only matches when expr resolves statically to a
		// recognised name; otherwise leave empty so the property is treated
		// as "useThis"-disqualifying.
		inner := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		return staticBindingKeyName(inner)
	}
	return ""
}

func (a *flagsAnalyzer) handleJsxAttribute(node *ast.Node) {
	attr := node.AsJsxAttribute()
	if attr == nil {
		return
	}
	nameNode := attr.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	if nameNode.AsIdentifier().Text == "ref" {
		a.flags.useRef = true
	}
}

func (a *flagsAnalyzer) handleReturnStatement(node *ast.Node) {
	rs := node.AsReturnStatement()
	// strict mirrors upstream's `!allowNull` argument to `isReturningJSX`:
	// when null is NOT allowed (React < 15), conditional / logical branches
	// must ALL be JSX; when null IS allowed, ANY branch being JSX qualifies
	// the return. This is critical — `return cond ? <jsx/> : null` flips
	// classification across the React 15 boundary, gating multiple upstream
	// tests.
	strict := !a.allowNull
	isReturningJSX := isJSXLike(rs.Expression, strict)
	isReturningNull := rs.Expression != nil && isReturnNullOrFalse(rs.Expression)
	if a.allowNull {
		if isReturningJSX || isReturningNull {
			return
		}
		a.flags.invalidReturn = true
		return
	}
	if isReturningJSX {
		return
	}
	a.flags.invalidReturn = true
}

// isJSXLike reports whether `expr` is JSX or contains JSX along all/any
// branches of a ternary or short-circuiting logical expression. Mirrors
// upstream's `jsxUtil.isJSX` + `isReturning` recursion: in strict mode,
// EVERY branch must be JSX (AND); in non-strict mode, ANY branch suffices
// (OR). The Comma operator always uses the right-hand side (sequence value).
func isJSXLike(expr *ast.Node, strict bool) bool {
	if expr == nil {
		return false
	}
	expr = ast.SkipParentheses(expr)
	switch expr.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindConditionalExpression:
		ce := expr.AsConditionalExpression()
		l := isJSXLike(ce.WhenTrue, strict)
		r := isJSXLike(ce.WhenFalse, strict)
		if strict {
			return l && r
		}
		return l || r
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindCommaToken:
			return isJSXLike(bin.Right, strict)
		case ast.KindAmpersandAmpersandToken,
			ast.KindBarBarToken,
			ast.KindQuestionQuestionToken:
			l := isJSXLike(bin.Left, strict)
			r := isJSXLike(bin.Right, strict)
			if strict {
				return l && r
			}
			return l || r
		}
	}
	return false
}

// isReturnNullOrFalse mirrors upstream's
// `node.argument.value === null || node.argument.value === false` check on
// the return argument. The argument may be parenthesized; in tsgo `null` is
// `KindNullKeyword`, `false` is `KindFalseKeyword`.
func isReturnNullOrFalse(expr *ast.Node) bool {
	expr = ast.SkipParentheses(expr)
	switch expr.Kind {
	case ast.KindNullKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}

// externalCCT records who has had `.childContextTypes` declared on them via
// an external assignment / access. We index by Symbol when a TypeChecker is
// available (precise — handles `const C = Foo; C.childContextTypes = {...}`
// alias patterns the way upstream's `getRelatedComponent` does via scope
// resolution), and fall back to base-identifier name when the checker is nil
// (gap files in the program; same degradation pattern other rules use).
type externalCCT struct {
	bySymbol map[*ast.Symbol]bool
	byName   map[string]bool
}

func (e *externalCCT) hasSymbol(sym *ast.Symbol) bool {
	return sym != nil && e.bySymbol[sym]
}

func (e *externalCCT) hasName(name string) bool {
	return name != "" && e.byName[name]
}

// collectExternalChildContextTypes scans the source file for
// `*.childContextTypes` member access (anywhere — not gated on assignment)
// and records each receiver's base symbol (precise) and base name (fallback).
//
// Mirrors upstream's `MemberExpression` listener + `getRelatedComponent`:
// upstream uses ESLint's scope manager to map a member-expression receiver
// back to its variable binding, then to the binding's RHS. tsgo's checker
// gives us the same answer — and crucially handles cross-binding aliases
// like `const C = Foo` that the previous AST-only `leftmostIdentifierName`
// could not follow.
func collectExternalChildContextTypes(sf *ast.SourceFile, tc *checker.Checker) externalCCT {
	out := externalCCT{
		bySymbol: map[*ast.Symbol]bool{},
		byName:   map[string]bool{},
	}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		if n.Kind == ast.KindPropertyAccessExpression {
			pa := n.AsPropertyAccessExpression()
			receiver := ast.SkipParentheses(pa.Expression)
			if receiver.Kind != ast.KindThisKeyword {
				name := pa.Name()
				if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "childContextTypes" {
					base := leftmostIdentifier(receiver)
					if base != nil {
						out.byName[base.AsIdentifier().Text] = true
						if tc != nil {
							if sym := resolveBindingSymbol(tc, base); sym != nil {
								out.bySymbol[sym] = true
							}
						}
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	sf.Node.ForEachChild(visit)
	return out
}

// leftmostIdentifier walks a receiver expression to its leftmost Identifier
// node, peeling parentheses and PropertyAccessExpression wrappers. Returns
// nil for receivers that don't terminate in a bare Identifier (`this.X`,
// `getFoo().X`, `Foo['x'].X`, …) — same conservative skip upstream's
// `getRelatedComponent` does for those shapes.
func leftmostIdentifier(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindIdentifier:
			return node
		case ast.KindPropertyAccessExpression:
			node = node.AsPropertyAccessExpression().Expression
		default:
			return nil
		}
	}
	return nil
}

// resolveBindingSymbol returns the symbol an Identifier ultimately binds to,
// transparently following local `const X = Y` aliases and import-alias
// indirections. This mirrors upstream's reliance on
// `variableUtil.getVariableFromContext` + walking
// `variableInScope.references` — both of those resolve through aliasing.
//
// Returns nil when no TypeChecker is available (gap files in the program
// have a nil checker — see linter scheduling). We bound the alias chain
// length to avoid pathological / cyclic inputs.
func resolveBindingSymbol(tc *checker.Checker, ident *ast.Node) *ast.Symbol {
	if tc == nil || ident == nil {
		return nil
	}
	sym := tc.GetSymbolAtLocation(ident)
	for i := 0; i < 16 && sym != nil; i++ {
		// Resolve import / export aliases first.
		if sym.Flags&ast.SymbolFlagsAlias != 0 {
			aliased := tc.GetAliasedSymbol(sym)
			if aliased == nil || aliased == sym {
				return sym
			}
			sym = aliased
			continue
		}
		// Follow `const X = Y` (Identifier-only) chains.
		if len(sym.Declarations) == 0 {
			return sym
		}
		decl := sym.Declarations[0]
		if decl.Kind != ast.KindVariableDeclaration {
			return sym
		}
		init := decl.AsVariableDeclaration().Initializer
		if init == nil {
			return sym
		}
		init = ast.SkipParentheses(init)
		if init.Kind != ast.KindIdentifier {
			return sym
		}
		next := tc.GetSymbolAtLocation(init)
		if next == nil || next == sym {
			return sym
		}
		sym = next
	}
	return sym
}

// componentBindingSymbol returns the symbol identifying a component's binding:
//
//   - Named ClassDeclaration / ClassExpression — the class's own Identifier
//     symbol.
//   - Anonymous ClassExpression assigned via `var Foo = class …` — the
//     parent VariableDeclaration's binding Identifier symbol.
//   - createReactClass(...) ObjectLiteralExpression in `var Foo = createReactClass({...})`
//     — the surrounding VariableDeclaration's binding Identifier symbol.
//
// Returns nil when the component has no binding name available, or when no
// TypeChecker is present.
func componentBindingSymbol(tc *checker.Checker, node *ast.Node) *ast.Symbol {
	if tc == nil || node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
			if sym := tc.GetSymbolAtLocation(name); sym != nil {
				return sym
			}
		}
		// Anonymous ClassExpression — look at the enclosing VariableDeclaration.
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			if name := parent.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
				return tc.GetSymbolAtLocation(name)
			}
		}
	case ast.KindObjectLiteralExpression:
		// ES5 createReactClass — the object literal's symbol-bearing handle is
		// the surrounding `var Foo = ...` binding. The createClass call may
		// be either `createReactClass({...})` (CallExpression) or the rare
		// `new createReactClass({...})` (NewExpression); both parent kinds
		// are accepted to mirror upstream `componentUtil.isES5Component`.
		cur := node
		for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
			cur = cur.Parent
		}
		if cur.Parent == nil {
			return nil
		}
		switch cur.Parent.Kind {
		case ast.KindCallExpression, ast.KindNewExpression:
		default:
			return nil
		}
		vd := cur.Parent.Parent
		if vd == nil || vd.Kind != ast.KindVariableDeclaration {
			return nil
		}
		if name := vd.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
			return tc.GetSymbolAtLocation(name)
		}
	}
	return nil
}

// analyzeES6Class fills `flags` for an ES6 class component. Each member's
// allow-list status is checked, and each method/property body is walked for
// `this.X` / VariableDeclaration / JSXAttribute / ReturnStatement features.
//
// Static members are still walked: upstream's getComponentProperties returns
// every body member regardless of `static`. A `static childContextTypes = {}`
// is therefore counted as an "other property", not via the external
// hasChildContextTypes path.
func analyzeES6Class(classNode *ast.Node, flags *componentFlags, pragma, createClass string, allowNull bool) {
	if hasDecorators(classNode) {
		flags.useDecorators = true
	}
	a := &flagsAnalyzer{flags: flags, pragma: pragma, createClass: createClass, allowNull: allowNull}
	members := classNode.Members()
	for _, member := range members {
		if !isAllowedClassMember(member) {
			flags.hasOtherProperty = true
		}
		switch member.Kind {
		case ast.KindMethodDeclaration:
			a.analyzeBody(member.AsMethodDeclaration().Body)
		case ast.KindGetAccessor:
			a.analyzeBody(member.AsGetAccessorDeclaration().Body)
		case ast.KindSetAccessor:
			a.analyzeBody(member.AsSetAccessorDeclaration().Body)
		case ast.KindConstructor:
			a.analyzeBody(member.AsConstructorDeclaration().Body)
		case ast.KindPropertyDeclaration:
			pd := member.AsPropertyDeclaration()
			if pd.Initializer != nil {
				a.walk(pd.Initializer)
			}
		}
	}
}

// analyzeES5Object fills `flags` for an ES5 createReactClass object literal.
func analyzeES5Object(obj *ast.Node, flags *componentFlags, pragma, createClass string, allowNull bool) {
	a := &flagsAnalyzer{flags: flags, pragma: pragma, createClass: createClass, allowNull: allowNull}
	ole := obj.AsObjectLiteralExpression()
	if ole.Properties == nil {
		return
	}
	for _, prop := range ole.Properties.Nodes {
		if !isAllowedES5Property(prop) {
			flags.hasOtherProperty = true
		}
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			pa := prop.AsPropertyAssignment()
			if pa.Initializer != nil {
				// Walk RHS so nested arrow / function bodies still get JSX
				// `ref` attribute and `this.X` accesses recorded.
				a.walk(pa.Initializer)
			}
		case ast.KindShorthandPropertyAssignment:
			// Nothing to walk — shorthand initializer is just an identifier
			// reference.
		case ast.KindMethodDeclaration:
			a.analyzeBody(prop.AsMethodDeclaration().Body)
		case ast.KindGetAccessor:
			a.analyzeBody(prop.AsGetAccessorDeclaration().Body)
		case ast.KindSetAccessor:
			a.analyzeBody(prop.AsSetAccessorDeclaration().Body)
		}
	}
}

var PreferStatelessFunctionRule = rule.Rule{
	Name: "react/prefer-stateless-function",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		// React >= 15 lets stateless components return null. Default version
		// is "latest" (see reactutil.ParseReactVersion), so allowNull defaults
		// to true. Older settings (< 15.0.0) flip it.
		allowNull := !reactutil.ReactVersionLessThan(ctx.Settings, 15, 0, 0)

		// Use TypeChecker when present for symbol-precise resolution of
		// `<binding>.childContextTypes` references — required to honor
		// upstream's `getRelatedComponent` semantics for cross-binding
		// aliases (e.g. `const C = Foo; C.childContextTypes = {...}`).
		// On gap files (no TypeChecker), fall back to the leftmost
		// identifier name — covers the common unaliased pattern.
		externalCCT := collectExternalChildContextTypes(ctx.SourceFile, ctx.TypeChecker)

		// hasExternalCCT bridges the two indices: it returns true when an
		// external `.childContextTypes` declaration was observed for the
		// given component, preferring symbol-precise matching when
		// TypeChecker is available and falling back to name match otherwise.
		hasExternalCCT := func(componentNode *ast.Node) bool {
			if ctx.TypeChecker != nil {
				if sym := componentBindingSymbol(ctx.TypeChecker, componentNode); sym != nil {
					return externalCCT.hasSymbol(sym)
				}
				// TypeChecker available but binding has no symbol — fall
				// through to name-based check (e.g. anonymous class
				// expression with no var binding).
			}
			return externalCCT.hasName(componentBindingName(componentNode))
		}

		runOnClass := func(node *ast.Node) {
			if !reactutil.ExtendsReactComponent(node, pragma) {
				return
			}
			flags := &componentFlags{}
			if opts.IgnorePureComponents && reactutil.ExtendsReactPureComponent(node, pragma) {
				flags.hasSCU = true
			}
			analyzeES6Class(node, flags, pragma, createClass, allowNull)
			if hasExternalCCT(node) {
				flags.hasChildContextTypes = true
			}
			if shouldReport(flags) {
				// Report at the `class` keyword, not at any preceding
				// `export` / `export default` modifier. ESTree separates
				// `ExportNamedDeclaration` / `ExportDefaultDeclaration` from
				// the inner `ClassDeclaration` node, so upstream's
				// `report({ node: component.node })` lands on `class …`.
				// tsgo inlines `export` into the ClassDeclaration's modifier
				// list, so we trim it back out to match upstream's report
				// position.
				start := classKeywordStart(ctx.SourceFile.Text(), node)
				ctx.ReportRange(core.NewTextRange(start, node.End()), rule.RuleMessage{
					Id:          "componentShouldBePure",
					Description: "Component should be written as a pure function",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: runOnClass,
			ast.KindClassExpression:  runOnClass,
			// Listen on the ObjectLiteralExpression directly — same as
			// upstream's `ObjectExpression(node)` listener that calls
			// `componentUtil.isES5Component(node)`. This naturally covers
			// both `createReactClass({...})` (CallExpression parent) and
			// the rare `new createReactClass({...})` (NewExpression parent),
			// because `isES5Component` checks both shapes.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if !isES5Component(node, pragma, createClass) {
					return
				}
				flags := &componentFlags{}
				analyzeES5Object(node, flags, pragma, createClass, allowNull)
				if hasExternalCCT(node) {
					flags.hasChildContextTypes = true
				}
				if shouldReport(flags) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "componentShouldBePure",
						Description: "Component should be written as a pure function",
					})
				}
			},
		}
	},
}

func shouldReport(f *componentFlags) bool {
	return !f.hasOtherProperty &&
		!f.useThis &&
		!f.useRef &&
		!f.invalidReturn &&
		!f.hasChildContextTypes &&
		!f.useDecorators &&
		!f.hasSCU
}
