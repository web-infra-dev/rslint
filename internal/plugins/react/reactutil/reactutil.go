package reactutil

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// DefaultReactPragma is the fallback object name for createElement calls
// when `settings.react.pragma` is not configured, matching eslint-plugin-react.
const DefaultReactPragma = "React"

// DefaultReactCreateClass is the fallback ES5 factory name when
// `settings.react.createClass` is not configured, matching
// eslint-plugin-react.
const DefaultReactCreateClass = "createReactClass"

// GetReactPragma reads `settings.react.pragma` from the config settings map.
// Returns DefaultReactPragma when the setting is absent, not a string, or empty.
func GetReactPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactPragma
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactPragma
	}
	pragma, ok := reactSettings["pragma"].(string)
	if !ok || pragma == "" {
		return DefaultReactPragma
	}
	return pragma
}

// DefaultReactFragment is the fallback fragment name for JSX shorthand
// fragment diagnostics when `settings.react.fragment` is not configured,
// matching eslint-plugin-react.
const DefaultReactFragment = "Fragment"

// GetReactFragmentPragma reads `settings.react.fragment` from the config
// settings map. Returns DefaultReactFragment when the setting is absent,
// not a string, or empty.
func GetReactFragmentPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactFragment
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactFragment
	}
	v, ok := reactSettings["fragment"].(string)
	if !ok || v == "" {
		return DefaultReactFragment
	}
	return v
}

// GetReactCreateClass reads `settings.react.createClass` from the config
// settings map. Returns DefaultReactCreateClass when the setting is absent,
// not a string, or empty.
func GetReactCreateClass(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactCreateClass
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactCreateClass
	}
	v, ok := reactSettings["createClass"].(string)
	if !ok || v == "" {
		return DefaultReactCreateClass
	}
	return v
}

// reactVersionRe captures the leading major[.minor[.patch]] numeric triple of
// a semver-ish string. Prerelease / build metadata / range qualifiers are
// ignored — matching eslint-plugin-react's `semver.coerce`-like behavior for
// the simple comparisons this package performs.
var reactVersionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// ParseReactVersion returns the (major, minor, patch) triple of
// `settings.react.version`. When the setting is missing, not a string, empty,
// or not recognizable as a version, it defaults to (999, 999, 999) — matching
// eslint-plugin-react's `getReactVersionFromContext`, which treats an absent
// version as "latest".
func ParseReactVersion(settings map[string]interface{}) (int, int, int) {
	if settings == nil {
		return 999, 999, 999
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return 999, 999, 999
	}
	raw, _ := reactSettings["version"].(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 999, 999, 999
	}
	m := reactVersionRe.FindStringSubmatch(raw)
	if m == nil {
		return 999, 999, 999
	}
	toInt := func(s string) int {
		if s == "" {
			return 0
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return n
	}
	return toInt(m[1]), toInt(m[2]), toInt(m[3])
}

// ReactVersionLessThan reports whether `settings.react.version` is strictly
// less than the given major.minor.patch. See ParseReactVersion for the default
// when the setting is missing.
func ReactVersionLessThan(settings map[string]interface{}, major, minor, patch int) bool {
	a, b, c := ParseReactVersion(settings)
	if a != major {
		return a < major
	}
	if b != minor {
		return b < minor
	}
	return c < patch
}

// IsCreateClassCall reports whether the given CallExpression's callee is
// `<createClass>(...)` or `<pragma>.<createClass>(...)`. Parentheses are
// skipped on both the callee and the pragma identifier. Pass the empty string
// for pragma/createClass to fall back to `DefaultReactPragma` /
// `DefaultReactCreateClass`.
func IsCreateClassCall(call *ast.CallExpression, pragma, createClass string) bool {
	if call == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}
	callee := ast.SkipParentheses(call.Expression)
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

// ExtendsReactComponent reports whether `classNode` (a ClassDeclaration or
// ClassExpression) has an `extends` clause referencing `Component` or
// `PureComponent` — either as a bare identifier or qualified by the
// configured pragma (e.g. `React.Component`). Parentheses are skipped. Pass
// the empty string for pragma to default to `DefaultReactPragma`.
//
// NOTE: Matches the name regex used by eslint-plugin-react's
// `componentUtil.isES6Component` (`/^(Pure)?Component$/`). Aliased imports
// (e.g. `import { Component as C }`) are not resolved — same as the upstream
// rule.
func ExtendsReactComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	switch expr.Kind {
	case ast.KindIdentifier:
		return isComponentName(expr.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		return isComponentName(nameNode.AsIdentifier().Text)
	}
	return false
}

func isComponentName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

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

// IsInsideReactComponent reports whether `node` is lexically contained within
// a React component — either an ES5 component (object literal passed as an
// argument to `<createClass>(...)` / `<pragma>.<createClass>(...)`) or an
// ES6 component (ClassDeclaration / ClassExpression extending Component or
// PureComponent, optionally qualified by pragma).
//
// Pass the empty string for pragma/createClass to fall back to
// `DefaultReactPragma` / `DefaultReactCreateClass`.
//
// Mirrors eslint-plugin-react's componentUtil:
//
//   - ES6 path: only the nearest enclosing class decides component status
//     (matching `getParentES6Component`'s `while scope.type !== 'class'`).
//     A non-component class does not "pass through" to let an outer component
//     class match — this prevents false positives like a non-React inner
//     class nested inside a React class.
//
//   - ES5 path: `this` / `this.refs` must occur inside some function, whose
//     ObjectExpression parent is the argument to `<createClass>(...)`. We
//     approximate this by requiring that an enclosing function has been
//     passed on the walk up before an ObjectExpression is accepted — which
//     rules out pathological cases like `createReactClass({ x: this.refs.y })`
//     where `this` is not inside any function (ESLint's scope walk returns
//     null for that too).
func IsInsideReactComponent(node *ast.Node, pragma, createClass string) bool {
	return GetEnclosingReactComponent(node, pragma, createClass) != nil
}

// GetEnclosingReactComponent is IsInsideReactComponent's sibling that returns
// the component node itself (the ClassDeclaration / ClassExpression, or the
// ObjectLiteralExpression passed to createReactClass) rather than a bool.
// Returns nil when `node` is not inside a React component. See
// IsInsideReactComponent for the detection rules.
func GetEnclosingReactComponent(node *ast.Node, pragma, createClass string) *ast.Node {
	if node == nil {
		return nil
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}
	seenNearestClass := false
	seenEnclosingFunction := false
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) {
			seenEnclosingFunction = true
		}
		switch p.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if seenNearestClass {
				// The nearest class already decided ES6 classification;
				// outer classes do not get a second chance (matches ESLint's
				// scope-walk that stops at the first class scope).
				continue
			}
			seenNearestClass = true
			if ExtendsReactComponent(p, pragma) {
				return p
			}
		case ast.KindObjectLiteralExpression:
			if !seenEnclosingFunction {
				continue
			}
			// The ObjectLiteralExpression may be wrapped in one or more
			// ParenthesizedExpressions before reaching the CallExpression
			// (ESTree would flatten these; tsgo preserves them), e.g.
			// `createReactClass(({...}))`. Walk up through paren wrappers
			// to find the actual argument position.
			arg := p
			for arg.Parent != nil && arg.Parent.Kind == ast.KindParenthesizedExpression {
				arg = arg.Parent
			}
			parent := arg.Parent
			if parent == nil || parent.Kind != ast.KindCallExpression {
				continue
			}
			call := parent.AsCallExpression()
			if !isObjectArgumentOf(call, arg) {
				continue
			}
			if IsCreateClassCall(call, pragma, createClass) {
				return p
			}
		}
	}
	return nil
}

func isObjectArgumentOf(call *ast.CallExpression, obj *ast.Node) bool {
	if call.Arguments == nil {
		return false
	}
	for _, arg := range call.Arguments.Nodes {
		if arg == obj {
			return true
		}
	}
	return false
}

// GetEnclosingReactComponentOrStateless is GetEnclosingReactComponent extended
// with eslint-plugin-react's `getParentStatelessComponent` fallback: when no
// enclosing ES6 class / ES5 createReactClass component is found, the nearest
// FunctionLike ancestor that looks like a functional component (capital-cased
// name + returns JSX/null) is returned.
//
// Priority matches upstream's `getParentComponent`:
//
//	getParentES6Component || getParentES5Component || getParentStatelessComponent
//
// so when a mutation node is inside an inner function nested within an outer
// class component, the OUTER class component is returned (preventing the
// inner stateless candidate from masking the class boundary).
//
// Only a restricted subset of upstream's heuristics is implemented — the
// patterns covering production React code: named FunctionDeclaration,
// FunctionExpression / ArrowFunction assigned to a capital-cased
// VariableDeclarator, PropertyAssignment, or ExportAssignment (default export),
// plus function expression in a CallExpression (e.g. React.memo wrapper —
// approximate match). This is intentionally conservative: missed detection
// causes a rule miss, over-detection would cause false-positive reports in
// non-component functions.
func GetEnclosingReactComponentOrStateless(node *ast.Node, pragma, createClass string) *ast.Node {
	if comp := GetEnclosingReactComponent(node, pragma, createClass); comp != nil {
		return comp
	}
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) && IsStatelessReactComponent(p) {
			return p
		}
	}
	return nil
}

// IsStatelessReactComponent reports whether `fn` (a FunctionLike) looks like a
// React functional component — i.e. it has a capital-cased associated name
// AND returns JSX or null on at least one path. Nested functions are not
// considered when searching for returns.
//
// See GetEnclosingReactComponentOrStateless for the priority rules.
func IsStatelessReactComponent(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
	default:
		return false
	}
	name := functionAssociatedName(fn)
	if name == "" || !isFirstLetterCapitalized(name) {
		return false
	}
	return functionReturnsJSXOrNull(fn)
}

// functionAssociatedName picks the name that identifies a function-like node
// for React component detection:
//
//   - the function's own Identifier (FunctionDeclaration / named
//     FunctionExpression)
//   - otherwise the Identifier of the enclosing VariableDeclaration /
//     PropertyAssignment
//
// Returns "" when no usable Identifier is found (e.g. computed keys,
// destructuring patterns, anonymous default exports).
func functionAssociatedName(fn *ast.Node) string {
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		name := fn.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	case ast.KindFunctionExpression:
		name := fn.Name()
		if name != nil && name.Kind == ast.KindIdentifier && isFirstLetterCapitalized(name.AsIdentifier().Text) {
			return name.AsIdentifier().Text
		}
	}
	parent := fn.Parent
	if parent == nil {
		return ""
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		binding := parent.AsVariableDeclaration().Name()
		if binding != nil && binding.Kind == ast.KindIdentifier {
			return binding.AsIdentifier().Text
		}
	case ast.KindPropertyAssignment:
		name := parent.AsPropertyAssignment().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == fn {
			left := ast.SkipParentheses(bin.Left)
			if left.Kind == ast.KindIdentifier {
				return left.AsIdentifier().Text
			}
		}
	}
	return ""
}

// functionReturnsJSXOrNull reports whether the function's body contains a
// `return <jsx/>` / `return null` at depth ≤ 1 (nested functions excluded),
// OR — for an arrow with expression body — whether that expression is JSX or
// `null`. ConditionalExpression is traversed so `return cond ? <jsx/> : null`
// qualifies.
func functionReturnsJSXOrNull(fn *ast.Node) bool {
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body = fn.AsArrowFunction().Body
		if body != nil && body.Kind != ast.KindBlock {
			return isJSXOrNullExpression(body)
		}
	}
	if body == nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			rs := n.AsReturnStatement()
			if rs.Expression != nil && isJSXOrNullExpression(rs.Expression) {
				found = true
				return true
			}
		case ast.KindFunctionExpression,
			ast.KindFunctionDeclaration,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return false
		}
		n.ForEachChild(visit)
		return found
	}
	visit(body)
	return found
}

// isJSXOrNullExpression reports whether `expr` is a JSX element, JSX fragment,
// or the `null` literal — walking through ParenthesizedExpression,
// ConditionalExpression branches, and comma-sequence (BinaryExpression with
// `,`) right-most operands so `(<div/>)`, `cond ? <div/> : null`, and
// `(side, <div/>)` all qualify (matches eslint-plugin-react's
// jsxUtil.isReturningJSXOrNull).
func isJSXOrNullExpression(expr *ast.Node) bool {
	expr = ast.SkipParentheses(expr)
	switch expr.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindNullKeyword:
		return true
	case ast.KindConditionalExpression:
		ce := expr.AsConditionalExpression()
		return isJSXOrNullExpression(ce.WhenTrue) || isJSXOrNullExpression(ce.WhenFalse)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken {
			return isJSXOrNullExpression(bin.Right)
		}
	}
	return false
}

func isFirstLetterCapitalized(s string) bool {
	return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z'
}

// IsCreateElementCall reports whether the callee is `<pragma>.createElement`.
// Pass an empty pragma to default to "React"; pass GetReactPragma(ctx.Settings)
// to honor the user's `settings.react.pragma` configuration.
//
// Parentheses are transparently skipped on both the callee itself and the
// pragma identifier (e.g. `(React).createElement` / `(React.createElement)()`),
// matching ESTree's flattened shape.
func IsCreateElementCall(callee *ast.Node, pragma string) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	callee = ast.SkipParentheses(callee)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "createElement" {
		return false
	}
	pragmaExpr := ast.SkipParentheses(prop.Expression)
	if pragmaExpr.Kind != ast.KindIdentifier || pragmaExpr.AsIdentifier().Text != pragma {
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
