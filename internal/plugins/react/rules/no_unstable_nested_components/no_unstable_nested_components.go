package no_unstable_nested_components

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message text mirrors eslint-plugin-react v7.37.x byte-for-byte, including
// the typographic apostrophe (U+2019) in "subtree's" and the typographic
// double quotation marks (U+201C / U+201D) wrapping the parent name. ASCII
// straight quotes here would silently diverge from upstream and break
// `errors: [{message: ERROR_MESSAGE}]` assertions written against the
// upstream rule.
const (
	messageBase = "Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component"
	asPropsInfo = " If you want to allow component creation in props, set allowAsProps option to true."
	openQuote   = "“"
	closeQuote  = "”"
)

// options holds the parsed rule options. Mirrors upstream's schema:
//
//	[{
//	  type: 'object',
//	  properties: {
//	    customValidators: { type: 'array', items: { type: 'string' } },
//	    allowAsProps: { type: 'boolean' },
//	    propNamePattern: { type: 'string' },
//	  },
//	  additionalProperties: false,
//	}]
//
// Schema validation itself is a framework-layer concern — upstream relies
// on ESLint's central ajv pass; rslint's config loader is the equivalent
// home. The rule body simply reads the well-typed fields it needs and
// trusts the surrounding tooling to reject malformed configs upstream of
// this point. `customValidators` is declared in the schema but unused by
// both upstream and us, so we accept and discard it.
type options struct {
	allowAsProps    bool
	propNamePattern string
}

func parseOptions(raw any) options {
	opts := options{propNamePattern: "render*"}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowAsProps"].(bool); ok {
		opts.allowAsProps = v
	}
	if v, ok := optsMap["propNamePattern"].(string); ok && v != "" {
		opts.propNamePattern = v
	}
	return opts
}

// unwrap peels off ParenthesizedExpression and TS-only expression wrappers
// (`as`, `satisfies`, `<T>x` type assertions, `x!` non-null assertions) so
// that downstream checks never need to know about them. Mirrors what ESTree
// implicitly hands to ESLint after the parser strips parens.
func unwrap(node *ast.Node) *ast.Node {
	return reactutil.SkipExpressionWrappers(node)
}

// componentEnv bundles the per-rule-invocation context every helper needs.
// Threading a single struct keeps signatures readable and avoids drift when
// new fields (TypeChecker, settings derivatives) are added later.
type componentEnv struct {
	pragma      string
	createClass string
	wrappers    []reactutil.ComponentWrapperEntry
	tc          *checker.Checker
}

// isDetectedComponent is a thin env-aware adapter over
// reactutil.IsDetectedComponent — see that function's doc for the canonical
// component-classification semantics.
func isDetectedComponent(node *ast.Node, env componentEnv) bool {
	return reactutil.IsDetectedComponent(node, env.pragma, env.createClass, env.wrappers, env.tc)
}

// isInsideWrapperCall reports whether `node` is the FunctionLike argument of
// a wrapper call (memo / forwardRef / configured wrapper). When true, the
// outer CallExpression listener handles reporting at the wrapper's Pos —
// this listener should skip to avoid double-reporting at a different column.
func isInsideWrapperCall(node *ast.Node, env componentEnv) bool {
	parent := reactutil.SkipExpressionWrappersUp(node)
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	return reactutil.MatchesAnyComponentWrapperWithChecker(parent, node, env.wrappers, env.pragma, env.tc)
}

// isValueOfObjectProperty mirrors upstream's
// `node.parent.type === 'Property'` check. In tsgo, a function / arrow used
// as an object-literal value has parent `PropertyAssignment`; an
// object-literal shorthand method has parent `ObjectLiteralExpression`
// directly. Both correspond to upstream's "Property" filter. Wrappers
// (`()` / TS) between the FunctionLike and the PropertyAssignment are
// peeled via `SkipExpressionWrappersUp` so `{ Foo: ((arrow)) }` is
// recognized the same as `{ Foo: arrow }`.
func isValueOfObjectProperty(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Parent.Kind == ast.KindObjectLiteralExpression
	}
	parent := reactutil.SkipExpressionWrappersUp(node)
	return parent != nil && parent.Kind == ast.KindPropertyAssignment
}

// objectPropertyKey returns the Identifier-like key of the
// PropertyAssignment / shorthand-method that owns `node`, or nil when the
// owner shape doesn't match. The returned node is always an Identifier when
// non-nil; computed keys / numeric / string-literal keys return nil since
// upstream's render-prop matching only fires on Identifier keys. Wrappers
// (`()` / TS) between the FunctionLike and the owning PropertyAssignment
// are walked through.
func objectPropertyKey(node *ast.Node) *ast.Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if node.Parent.Kind != ast.KindObjectLiteralExpression {
			return nil
		}
		name := node.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name
		}
		return nil
	}
	parent := reactutil.SkipExpressionWrappersUp(node)
	if parent == nil || parent.Kind != ast.KindPropertyAssignment {
		return nil
	}
	name := parent.AsPropertyAssignment().Name()
	if name != nil && name.Kind == ast.KindIdentifier {
		return name
	}
	return nil
}

// isComponentInsideCreateElementProp mirrors upstream's
// `isComponentInsideCreateElementsProp`: the node (a detected component)
// sits inside the second argument (the props object) of a
// `<pragma>.createElement(...)` call. The closest enclosing
// ObjectLiteralExpression must be the call's second argument — deeper
// nesting (object inside an object inside the props object) does NOT match,
// matching upstream's strict-equality check.
func isComponentInsideCreateElementProp(node *ast.Node, env componentEnv) bool {
	if !isDetectedComponent(node, env) {
		return false
	}
	// Start the ancestor walks from the wrapper-skipped parent so paren /
	// TS wrappers between the FunctionLike and its semantic ESTree
	// ancestor don't introduce slack. Matches upstream's flattened tree.
	start := reactutil.SkipExpressionWrappersUp(node)
	if start == nil {
		return false
	}
	objectExpr := ast.FindAncestor(start, func(n *ast.Node) bool {
		return n.Kind == ast.KindObjectLiteralExpression
	})
	if objectExpr == nil {
		return false
	}
	createEl := ast.FindAncestor(start, func(n *ast.Node) bool {
		return n.Kind == ast.KindCallExpression &&
			reactutil.IsCreateElementCallWithChecker(n.AsCallExpression().Expression, env.pragma, env.tc)
	})
	if createEl == nil {
		return false
	}
	call := createEl.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return false
	}
	return unwrap(call.Arguments.Nodes[1]) == objectExpr
}

// isComponentInProp mirrors upstream's `isComponentInProp`: reports whether
// the node is used as the value of an object-literal property (including
// props on a JSX element or the props object of `React.createElement`) AND
// strictly returns JSX. Matches upstream's `utils.isReturningJSX` which
// passes `ignoreNull=true` — functions that only return `null` do NOT
// qualify here (Components.js isReturningJSX wrapper).
func isComponentInProp(node *ast.Node, env componentEnv) bool {
	if isValueOfObjectProperty(node) {
		return reactutil.FunctionReturnsJSXWithChecker(node, env.pragma, env.tc)
	}
	// Walk up to a JsxAttribute whose value is a JsxExpression — the tsgo
	// equivalent of ESTree's "JSXAttribute with JSXExpressionContainer
	// value". The first such ancestor decides; objects nested deeper still
	// count because upstream's `getClosestMatchingParent` walks past
	// non-matching ancestors.
	if hasAncestorJsxAttributeWithExpression(node) {
		return reactutil.FunctionReturnsJSXWithChecker(node, env.pragma, env.tc)
	}
	return isComponentInsideCreateElementProp(node, env)
}

// hasAncestorJsxAttributeWithExpression walks up from the node's semantic
// parent (skipping `()` / TS wrappers) and returns true once it sees a
// JsxExpression whose parent is a JsxAttribute. Starting the walk from the
// wrapper-skipped parent matches upstream's ESTree-flattened ancestor
// chain.
func hasAncestorJsxAttributeWithExpression(node *ast.Node) bool {
	start := reactutil.SkipExpressionWrappersUp(node)
	if start == nil {
		return false
	}
	return ast.FindAncestor(start, func(n *ast.Node) bool {
		return n.Kind == ast.KindJsxExpression &&
			n.Parent != nil && n.Parent.Kind == ast.KindJsxAttribute
	}) != nil
}

// isComponentInRenderProp mirrors upstream's `isComponentInRenderProp`:
// returns true when the node is hosted by a render-prop-shaped position —
// either a property whose key matches the configured glob, a JSX child
// expression, or a JSX attribute whose name matches the glob / equals
// `children`. Wrapper-transparent at every parent-walk site so paren / TS
// wrappers (which ESTree flattens but tsgo preserves) don't break alignment.
func isComponentInRenderProp(node *ast.Node, propNamePattern string) bool {
	// Direct value of a PropertyAssignment / shorthand method whose key
	// matches the pattern.
	if key := objectPropertyKey(node); key != nil {
		if reactutil.MatchGlob(key.AsIdentifier().Text, propNamePattern) {
			return true
		}
	}
	// Direct child of a JsxExpression whose parent is a JsxElement /
	// JsxFragment — i.e. a render-prop passed as JSX children
	// (`<C>{() => <div/>}</C>` / `<>{() => <div/>}</>` / paren-wrapped
	// `<C>{((() => <div/>))}</C>`). Walk up through wrappers so the
	// JsxExpression-as-direct-parent shape matches upstream regardless of
	// intermediate `()` / TS-wrapper nodes.
	if jsxParent := reactutil.SkipExpressionWrappersUp(node); jsxParent != nil &&
		jsxParent.Kind == ast.KindJsxExpression && jsxParent.Parent != nil {
		switch jsxParent.Parent.Kind {
		case ast.KindJsxElement, ast.KindJsxFragment:
			return true
		}
	}
	// Closest JsxExpression ancestor whose parent is a JsxAttribute.
	container := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return n.Kind == ast.KindJsxExpression
	})
	if container == nil || container.Parent == nil || container.Parent.Kind != ast.KindJsxAttribute {
		return false
	}
	nameNode := container.Parent.AsJsxAttribute().Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return false
	}
	propName := nameNode.AsIdentifier().Text
	if reactutil.MatchGlob(propName, propNamePattern) {
		return true
	}
	return propName == "children"
}

// isMapCall mirrors upstream's `isMapCall`: `<anything>.map(...)`. Only the
// property name is inspected; the receiver is ignored.
//
// The `node` parameter is unwrapped first so callers passing
// `node.Parent` get a hit even when the parent is a ParenthesizedExpression
// / TS-wrapper around the actual map call — e.g.
// `<C foo={(items.map((arrow)))}/>`, where the arrow's direct parent is the
// inner ParenExpr around the arrow, whose own parent is the ParenExpr
// wrapping the call result. ESTree flattens these wrappers; tsgo preserves
// them, so we unwrap here to keep the skip-on-map behavior aligned with
// upstream byte-for-byte. Same applies to the callee.
func isMapCall(node *ast.Node) bool {
	node = unwrap(node)
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := unwrap(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "map"
}

// isReturnStatementOfHook mirrors upstream's `isReturnStatementOfHook`: the
// node is the direct value of a `return` that lives inside a hook call
// (Identifier callee whose name matches the hook regex). The closest
// enclosing CallExpression decides — hook detection does NOT recurse past
// the first call.
//
// tsgo preserves `()` / TS-wrapper nodes between the FunctionLike and its
// owning ReturnStatement that ESTree flattens — e.g.
// `useEffect(() => { return (() => null); })` has the inner arrow's
// parent = ParenExpr in tsgo but the inner arrow's parent = ReturnStatement
// in ESTree. Walk up through wrappers so the skip behavior aligns with
// upstream byte-for-byte.
func isReturnStatementOfHook(node *ast.Node) bool {
	parent := reactutil.SkipExpressionWrappersUp(node)
	if parent == nil || parent.Kind != ast.KindReturnStatement {
		return false
	}
	enclosing := ast.FindAncestor(parent, func(n *ast.Node) bool {
		return n.Kind == ast.KindCallExpression
	})
	if enclosing == nil {
		return false
	}
	callee := unwrap(enclosing.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	return reactutil.IsHookName(callee.AsIdentifier().Text)
}

// isDirectValueOfRenderProperty mirrors upstream's
// `isDirectValueOfRenderProperty`: the node's owning property has an
// Identifier key matching the render-prop pattern.
//
// tsgo preserves `()` / TS-wrapper nodes between the FunctionLike and its
// owning PropertyAssignment that ESTree flattens — e.g.
// `{ render: ((props) => <Row/>) }` has arrow.Parent = ParenExpr in tsgo
// but arrow.parent = PropertyAssignment in ESTree. Walk up through
// wrappers so the skip behavior aligns with upstream byte-for-byte.
func isDirectValueOfRenderProperty(node *ast.Node, propNamePattern string) bool {
	parent := reactutil.SkipExpressionWrappersUp(node)
	if parent == nil || parent.Kind != ast.KindPropertyAssignment {
		return false
	}
	name := parent.AsPropertyAssignment().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return reactutil.MatchGlob(name.AsIdentifier().Text, propNamePattern)
}

// isStatelessComponentReturningNull mirrors upstream's sibling check: the
// node is classified as a stateless component but strictly does NOT return
// JSX (e.g. `() => undefined` or `() => null` under ignoreNull=true). These
// are excluded because they cannot render as React components. Upstream's
// `utils.isReturningJSX` negated — ignoreNull=true — is our
// `FunctionReturnsJSX`.
func isStatelessComponentReturningNull(node *ast.Node, env componentEnv) bool {
	// Use the wrappers-aware classification so user wrappers (myMemo /
	// MyLib.observer / ...) correctly participate in stateless detection.
	// Without this, `myMemo(() => null)` would not classify as a
	// stateless component → this skip wouldn't fire → we'd over-report
	// where upstream skips.
	if !reactutil.IsStatelessReactComponentWithWrappers(node, env.pragma, env.tc, env.wrappers) {
		return false
	}
	return !reactutil.FunctionReturnsJSXWithChecker(node, env.pragma, env.tc)
}

// isFunctionComponentInsideClassComponent mirrors upstream's safety net of
// the same name — cases where Components.detect classifies a parent as an
// ES6 class component but the inner FunctionLike returning JSX wasn't
// picked up via the stateless path on its own. Returns true when:
//
//  1. the nearest enclosing component is a class (ES6 path),
//  2. there's a FunctionLike ancestor between the class and `node`
//     (the "parentStatelessComponent" upstream walks for),
//  3. that FunctionLike ancestor itself classifies as a stateless
//     component (matching upstream's `getStatelessComponent(parentStatelessComponent)`
//     gate), and
//  4. `node` strictly returns JSX.
//
// All four conditions matter — without (3) we'd misclassify any
// lowercase-named helper inside a class render method's nested function
// as a "function component inside class component" purely on the basis
// of returning JSX.
func isFunctionComponentInsideClassComponent(node *ast.Node, env componentEnv) bool {
	if !reactutil.IsFunctionLikeForComponent(node) {
		return false
	}
	parent := getClosestParentComponent(node, env)
	if parent == nil {
		return false
	}
	if parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindClassExpression {
		return false
	}
	enclosingFn := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return reactutil.IsFunctionLikeForComponent(n)
	})
	if enclosingFn == nil {
		return false
	}
	// Upstream's gate (3): the enclosing FunctionLike must itself
	// classify as a stateless component for this safety net to fire.
	// Use the wrappers-aware variant for user-configured wrappers
	// participating in stateless detection.
	if !reactutil.IsStatelessReactComponentWithWrappers(enclosingFn, env.pragma, env.tc, env.wrappers) {
		return false
	}
	return reactutil.FunctionReturnsJSXWithChecker(node, env.pragma, env.tc)
}

// getClosestParentComponent walks up from `node.Parent` and returns the
// first ancestor that would be classified as a React component. Mirrors
// upstream's `getClosestMatchingParent(components.get)`.
//
// Starts the walk from the wrapper-skipped parent so paren / TS wrappers
// between the node and its semantic ESTree ancestor don't change the
// closest-component result. Matches upstream's flattened tree.
func getClosestParentComponent(node *ast.Node, env componentEnv) *ast.Node {
	start := reactutil.SkipExpressionWrappersUp(node)
	if start == nil {
		return nil
	}
	return ast.FindAncestor(start, func(n *ast.Node) bool {
		return isDetectedComponent(n, env)
	})
}

// resolveComponentName returns the display name of a detected component
// node, or "" when anonymous. Mirrors upstream's `resolveComponentName`
// byte-for-byte:
//
//   - Class / Function declarations and named FunctionExpressions: the
//     declaration's own Identifier (`node.id.name` upstream).
//   - Anonymous ArrowFunctionExpression: the binding name of its
//     VariableDeclarator parent (`node.parent.id.name` upstream).
//
// Every other parent component shape (ObjectLiteralExpression returned by
// `<createClass>(...)`, wrapper CallExpression `React.memo(arrow)`, etc.)
// returns "" — upstream never walks beyond the immediate `.id` lookup, and
// matching that quirk keeps the diagnostic message text byte-equal in
// corner cases like ES5 components and optional-chain wrappers.
func resolveComponentName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression:
		if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	case ast.KindArrowFunction:
		// Upstream's ArrowFunctionExpression branch: parent must be a
		// VariableDeclarator with an Identifier id. tsgo's
		// VariableDeclaration is the equivalent of ESTree's
		// VariableDeclarator (the per-declaration node, not the
		// const/let/var statement).
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			binding := node.Parent.AsVariableDeclaration().Name()
			if binding != nil && binding.Kind == ast.KindIdentifier {
				return binding.AsIdentifier().Text
			}
		}
	}
	return ""
}

// isOptionalPragmaWrapperCall reports whether `node` is a CallExpression
// whose optional-chain shape triggers upstream's empty-parent-name quirk:
//
//   - Member-level optional (`React?.memo(arrow)`) — Babel wraps in
//     ChainExpression sharing `range[0]` with the inner CE; upstream's
//     `components.add` storage is keyed by `range[0]` so
//     `components.get(ChainExpression)` returns the wrapper entry, the
//     parent walk stops there, `resolveComponentName(ChainExpression)`
//     returns falsy → empty parent name.
//
//   - Call-level optional (`myMemo?.(arrow)` with user wrapper) — same
//     ChainExpression wrapping shape; same `range[0]` collision; same
//     empty-parent-name observation.
//
// Either form requires the parentName-blanking quirk to keep diagnostics
// byte-aligned with upstream.
func isOptionalPragmaWrapperCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	c := node.AsCallExpression()
	if c.QuestionDotToken != nil {
		return true
	}
	callee := reactutil.SkipExpressionWrappers(c.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	return ast.IsOptionalChain(callee)
}

// generateErrorMessageWithParentName mirrors upstream's helper of the same
// name — the parent name, when known, is wrapped in typographic double
// quotes (U+201C / U+201D) and appears immediately after "parent component".
func generateErrorMessageWithParentName(parentName string) string {
	if parentName != "" {
		return messageBase + " " + openQuote + parentName + closeQuote + " and pass data as props."
	}
	return messageBase + " and pass data as props."
}

var NoUnstableNestedComponentsRule = rule.Rule{
	Name: "react/no-unstable-nested-components",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// componentEnv bundles every per-invocation derivative the helpers
		// below need. Building it once at rule entry avoids both repeated
		// `ctx.Settings` lookups and the risk of helpers drifting on
		// pragma / createClass / wrapper / TypeChecker reads.
		env := componentEnv{
			pragma:      reactutil.GetReactPragma(ctx.Settings),
			createClass: reactutil.GetReactCreateClass(ctx.Settings),
			tc:          ctx.TypeChecker,
		}
		// `settings.componentWrapperFunctions` decides which CallExpression
		// wrappers count as "creating a component" — defaults to memo +
		// forwardRef (pragma-qualified and bare), users may add more.
		env.wrappers = reactutil.GetComponentWrapperFunctions(ctx.Settings, env.pragma)

		// reported tracks node positions we've already reported on, so
		// the FunctionLike + CallExpression listeners can both fire for a
		// React.memo-wrapped arrow without double-counting.
		reported := map[int]struct{}{}

		validate := func(node *ast.Node) {
			if node == nil || node.Parent == nil {
				return
			}
			// Class-body methods / accessors are NOT separate component
			// candidates — the enclosing class already decides component
			// status (and is reported on its own). Without this guard, a
			// nested-class-in-class case would emit one report for the
			// inner ClassDeclaration AND a second for its `render` method.
			// Mirrors upstream's `isInsideRenderMethod` skip plus the
			// fact that upstream does not listen to MethodDefinition.
			switch node.Kind {
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				if node.Parent.Kind != ast.KindObjectLiteralExpression {
					return
				}
			}
			// Wrapper-call arguments are reported by the CallExpression
			// listener at the wrapper's Pos (matching upstream's
			// `components.add(call, 2)` registration). Suppress the inner
			// FunctionLike report so the diagnostic position aligns with
			// upstream byte-for-byte.
			if reactutil.IsFunctionLikeForComponent(node) && isInsideWrapperCall(node, env) {
				return
			}

			isDeclaredInsideProps := isComponentInProp(node, env)
			isComponent := isDetectedComponent(node, env)
			isFnInClass := !isComponent && !isDeclaredInsideProps && isFunctionComponentInsideClassComponent(node, env)

			if !isComponent && !isDeclaredInsideProps && !isFnInClass {
				return
			}

			// Allow components in render-prop positions (and, when
			// configured, everywhere a prop hosts one).
			if isDeclaredInsideProps && (opts.allowAsProps || isComponentInRenderProp(node, opts.propNamePattern)) {
				return
			}
			// Skip nodes produced by Array#map callbacks. tsgo preserves
			// `()` / TS-wrapper nodes that ESTree flattens, so an arrow
			// inside `items.map((arrow))` has parent = ParenExpr, not the
			// CallExpression directly. Walking up through wrappers
			// recovers the ESTree-equivalent parent and matches upstream's
			// `isMapCall(node.parent)` byte-for-byte.
			if isMapCall(node) || isMapCall(reactutil.SkipExpressionWrappersUp(node)) {
				return
			}
			if isReturnStatementOfHook(node) {
				return
			}
			if isDirectValueOfRenderProperty(node, opts.propNamePattern) {
				return
			}
			if isStatelessComponentReturningNull(node, env) {
				return
			}

			parentComponent := getClosestParentComponent(node, env)
			if parentComponent == nil {
				return
			}

			parentName := resolveComponentName(parentComponent)
			// Lowercase factory/helper names are not React components at
			// runtime — React-DOM only treats Capitalized references as
			// components. Mirror upstream's EXACT check from
			// `lib/rules/no-unstable-nested-components.js`:
			//
			//     if (parentName && parentName[0] === parentName[0].toLowerCase()) return;
			//
			// This is **NOT** the same as `isFirstLetterCapitalized`:
			//   - upstream's component-detection helper strips leading `_`
			//     (so `_Foo` is treated as the component "Foo");
			//   - upstream's lowercase-skip check above does NOT strip — it
			//     just checks `firstLetter === firstLetter.toLowerCase()`,
			//     which is true for `_foo` AND `_Foo` (`_` is non-cased) AND
			//     CJK / digit prefixes.
			//
			// Both conditions co-exist: `_Foo` qualifies as a parent
			// component (detection passes) but the lowercase-skip ALSO
			// fires (so the inner nested component is silently allowed).
			// We replicate this with `isLowercaseFirstLetter` below.
			if parentName != "" && reactutil.IsLowercaseFirstLetter(parentName) {
				return
			}
			// Optional-chain pragma wrapper quirk: in Babel's ESTree, the
			// inner CallExpression of `<pragma>?.<wrapper>(arrow)` is
			// wrapped in a ChainExpression that shares the same `range[0]`
			// as the inner CallExpression. Upstream's `components.add` uses
			// `range[0]` as the storage key, so `components.get(ChainExpression)`
			// returns the wrapper entry — which makes
			// `getClosestMatchingParent` stop at the ChainExpression instead
			// of walking up to the enclosing component. ChainExpression has
			// no `.id`, so `resolveComponentName` returns empty.
			//
			// rslint's tsgo AST has no ChainExpression wrapper (optional is
			// flag-based), so our parent walk would naturally find the
			// outer FunctionDeclaration and produce a non-empty parentName.
			// To stay byte-aligned with upstream's observable output, blank
			// the parentName whenever the validate target is a wrapper call
			// whose callee carries an optional-chain mark.
			if isOptionalPragmaWrapperCall(node) {
				parentName = ""
			}

			pos := node.Pos()
			if _, seen := reported[pos]; seen {
				return
			}
			reported[pos] = struct{}{}

			msg := generateErrorMessageWithParentName(parentName)
			if isDeclaredInsideProps && !opts.allowAsProps {
				msg += asPropsInfo
			}
			// MessageId is intentionally empty: upstream's
			// eslint-plugin-react reports this rule's diagnostic with
			// `messageId: null` — it passes the raw message string to
			// `report(context, message, null, …)`. We mirror that with an
			// empty Id; tests matching by `message` text work in both
			// runners.
			ruleMsg := rule.RuleMessage{Id: "", Description: msg}
			// Object-literal shorthand methods (`{ Foo() {} }`,
			// `{ get Foo() {} }`, `{ async Foo() {} }`) are one
			// `MethodDeclaration` / `GetAccessor` / `SetAccessor` node in
			// tsgo, but ESTree wraps them as `Property { kind, value:
			// FunctionExpression }` and reports on the inner FE — whose
			// `loc.start` lands at the `(` of the parameter list, after
			// the `get`/`set`/`async` modifier and the property name.
			// Mirror that by narrowing the report range to start at the
			// parameter list's `(` for these node kinds when they live
			// inside an ObjectLiteralExpression.
			if reactutil.IsObjectLiteralShorthandFunction(node) {
				if start := reactutil.ParamListOpenParenPos(ctx.SourceFile, node); start >= 0 {
					ctx.ReportRange(core.NewTextRange(start, node.End()), ruleMsg)
					return
				}
			}
			ctx.ReportNode(node, ruleMsg)
		}

		// Listener set mirrors upstream's exact list:
		// FunctionDeclaration / ArrowFunctionExpression / FunctionExpression
		// / ClassDeclaration / CallExpression. ClassExpression is
		// intentionally NOT listened to — upstream's rule body declares
		// only those five `return { ... }` keys; `const X = class extends
		// React.Component {}` therefore does NOT report.
		//
		// MethodDeclaration / GetAccessor / SetAccessor are also listened
		// to here so we cover object-literal shorthand methods (`{ Foo() {
		// return <div/>; } }`); class-body methods are filtered out by the
		// `parent.Kind != ObjectLiteralExpression` guard inside `validate`.
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: validate,
			ast.KindFunctionExpression:  validate,
			ast.KindArrowFunction:       validate,
			ast.KindMethodDeclaration:   validate,
			ast.KindGetAccessor:         validate,
			ast.KindSetAccessor:         validate,
			ast.KindClassDeclaration:    validate,
			ast.KindCallExpression:      validate,
		}
	},
}
