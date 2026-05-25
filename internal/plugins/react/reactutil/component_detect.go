package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// IsInsideReactComponent reports whether `node` is lexically inside a
// React component, applying the SCOPE-BASED detection semantic that
// upstream's `componentUtil.getParentES6Component(...) ||
// componentUtil.getParentES5Component(...)` use directly (the pattern
// of `no-string-refs` and `no-access-state-in-setstate`).
//
// **NOT equivalent to `GetEnclosingReactComponent != nil`**: the latter
// mimics `Components.set`'s free AST ancestor walk that crosses any
// non-React class. This helper applies the stricter ES6-stops-at-first-
// class rule. Pick based on the upstream rule's pattern:
//
//   - Rule uses `Components.detect((context, components, utils) => ...)`
//     and calls `components.set(node, ...)` / `components.get(...)` →
//     use `GetEnclosingReactComponent`.
//
//   - Rule calls `componentUtil.getParentES6Component` /
//     `componentUtil.getParentES5Component` directly → use this helper
//     (or `GetParentReactComponentScopeBased` for the node).
//
// Pass empty strings for pragma/createClass to fall back to defaults.
func IsInsideReactComponent(node *ast.Node, pragma, createClass string) bool {
	return GetParentReactComponentScopeBased(node, pragma, createClass) != nil
}

// GetParentReactComponentScopeBased mirrors upstream's
// `componentUtil.getParentES6Component(context, node) ||
// componentUtil.getParentES5Component(context, node)` exactly — the
// helper used directly by `no-string-refs` and `no-access-state-in-setstate`.
//
// **NOT equivalent to `GetEnclosingReactComponent`**: that one mimics
// `Components.set`'s free AST ancestor walk; this one applies the
// stricter scope-based rules:
//
//   - **ES6 path**: finds the FIRST enclosing class (innermost). If it
//     extends `Component` / `PureComponent` (bare or pragma-qualified),
//     returns it; otherwise stops searching outer classes (mirrors
//     upstream's `while scope.type !== 'class'` loop).
//
//   - **ES5 path**: walks each enclosing FunctionLike scope. For each,
//     checks whether its parent.parent reaches a `createReactClass(...)`
//     argument ObjectLiteralExpression. This crosses non-React classes
//     freely — only function-like scopes are inspected.
//
// Empirically verified equivalent to ESLint output. Pass empty strings
// for pragma/createClass to fall back to defaults.
func GetParentReactComponentScopeBased(node *ast.Node, pragma, createClass string) *ast.Node {
	if node == nil {
		return nil
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}

	// ES6 path: find FIRST enclosing class. If React, return; else
	// remember that ES6 has decided "not a React class" and don't
	// search outer classes (matches upstream's `while scope.type !==
	// 'class'` loop that stops at the first class scope).
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Kind == ast.KindClassDeclaration || p.Kind == ast.KindClassExpression {
			if ExtendsReactComponent(p, pragma) {
				return p
			}
			// First class is not React → ES6 detection returns null.
			// Fall through to ES5 detection below.
			break
		}
	}

	// ES5 path: walk each enclosing FunctionLike. For each, check
	// whether its parent / parent.parent is a createReactClass(...)
	// arg ObjectLiteralExpression. Mirrors upstream's per-scope walk:
	//   `node = scope.block && scope.block.parent && scope.block.parent.parent`
	// where scope.block is the FunctionLike.
	for p := node.Parent; p != nil; p = p.Parent {
		if !ast.IsFunctionLike(p) {
			continue
		}
		// `key: function() {...}` — FE wrapped in PropertyAssignment;
		// its parent is the ObjectLiteralExpression.
		// `key() {...}` shorthand — MethodDeclaration / GetAccessor /
		// SetAccessor directly inside ObjectLiteralExpression.
		var objLit *ast.Node
		switch p.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			objLit = p.Parent
		default:
			propEntry := p.Parent
			if propEntry == nil || propEntry.Kind != ast.KindPropertyAssignment {
				continue
			}
			objLit = propEntry.Parent
		}
		if objLit == nil || objLit.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		// Unwrap parens and verify createReactClass call.
		arg := objLit
		for arg.Parent != nil && arg.Parent.Kind == ast.KindParenthesizedExpression {
			arg = arg.Parent
		}
		callExpr := arg.Parent
		if callExpr == nil || callExpr.Kind != ast.KindCallExpression {
			continue
		}
		call := callExpr.AsCallExpression()
		if !isObjectArgumentOf(call, arg) {
			continue
		}
		if IsCreateClassCall(call, pragma, createClass) {
			return objLit
		}
	}

	return nil
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
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			// Mirror upstream's `Components.set` behavior: it walks
			// `node.parent` looking for any node in the `_list` of
			// already-detected components. Non-React classes are NOT
			// in that list — Components.detect only registers classes
			// that extend `Component` / `PureComponent` (or pragma-
			// qualified) — so an inner non-React class does NOT block
			// the walk from reaching an outer React component or a
			// `createReactClass({...})` arg above.
			//
			// Concretely: a `this.setState({})` inside `class Helper {
			// foo() {...} }`, where Helper is itself nested inside
			// `class Outer extends React.Component { render() {...} }`
			// or inside `createReactClass({ method: function() {
			// class Helper {...} } })`, MUST attribute to the outer
			// detected component. Both upstream eslint-plugin-react
			// and rslint match here.
			if ExtendsReactComponent(p, pragma) {
				return p
			}
		case ast.KindObjectLiteralExpression:
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
				// Empirically verified against ESLint master:
				// `createReactClass({ key: this.setState({}) })` —
				// even a setState call at the TOP-LEVEL property
				// position (not inside any method/function) attributes
				// to the createReactClass arg via Components.set's
				// free parent walk and reports.
				return p
			}
		}
	}
	return nil
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
func GetEnclosingReactComponentOrStateless(node *ast.Node, pragma, createClass string, wrappers []ComponentWrapperEntry) *ast.Node {
	if comp := GetEnclosingReactComponent(node, pragma, createClass); comp != nil {
		return comp
	}
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) && IsStatelessReactComponentWithWrappers(p, pragma, nil, wrappers) {
			return p
		}
	}
	return nil
}

// GetParentReactComponentScopeBasedOrStateless mirrors upstream's
// `utils.getParentComponent(node)` =
// `getParentES6Component || getParentES5Component || getParentStatelessComponent`.
//
// **NOT equivalent to `GetEnclosingReactComponentOrStateless`**: that one
// uses `Components.set`'s free AST ancestor walk. This helper applies
// the stricter scope-based ES6+ES5 detection (see
// `GetParentReactComponentScopeBased`) and falls back to stateless
// component detection. Use this for rules that call
// `utils.getParentComponent(node)` directly inside a listener and gate
// their report on the result being non-null — e.g.
// `no-direct-mutation-state`'s `shouldIgnoreComponent(component)`
// check, which bails when the result is undefined.
//
// Pass empty strings for pragma/createClass to fall back to defaults.
func GetParentReactComponentScopeBasedOrStateless(node *ast.Node, pragma, createClass string, wrappers []ComponentWrapperEntry) *ast.Node {
	if comp := GetParentReactComponentScopeBased(node, pragma, createClass); comp != nil {
		return comp
	}
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) && IsStatelessReactComponentWithWrappers(p, pragma, nil, wrappers) {
			return p
		}
	}
	return nil
}

// GetParentStatelessComponent mirrors eslint-plugin-react's
// `utils.getParentStatelessComponent(node)`: walk up enclosing FunctionLike
// scopes from `node` and return the first one classified as a stateless
// functional component. A non-component function does NOT stop the walk —
// the next outer FunctionLike still gets a chance, matching upstream's
// `scope.upper` traversal.
//
// ES6 class scopes / class field initializers / module scope are
// non-FunctionLike nodes that are simply skipped during the walk.
//
// Pass empty `pragma` to default to `DefaultReactPragma`. `wrappers` should
// be the resolved `settings.componentWrapperFunctions` list (plus the
// built-in `memo` / `forwardRef` defaults) so user-configured HOCs are
// recognized as component-wrapping calls.
func GetParentStatelessComponent(node *ast.Node, pragma string, wrappers []ComponentWrapperEntry) *ast.Node {
	for p := node.Parent; p != nil; p = p.Parent {
		if !ast.IsFunctionLike(p) {
			continue
		}
		if IsStatelessReactComponentWithWrappers(p, pragma, nil, wrappers) {
			return p
		}
	}
	return nil
}

// IsStatelessReactComponent reports whether `fn` (a FunctionLike) looks like a
// React functional component. Mirrors eslint-plugin-react's
// `getStatelessComponent` decision tree:
//
//   - FunctionDeclaration — component iff returns JSX/null AND either:
//     (a) its own Identifier is capitalized, OR
//     (b) it is anonymous AND carries the `export default` modifier (ESLint's
//     `!node.id || capitalized(node.id.name)` condition).
//
//   - FunctionExpression / ArrowFunction — component iff returns JSX/null AND
//     either wrapped in a pragma component call OR in an "allowed position"
//     AND the position-specific capitalization check passes:
//
//   - Wrapped in `<pragma>.memo(...)` / `<pragma>.forwardRef(...)` / bare
//     `memo(...)` / bare `forwardRef(...)` — always a component.
//
//   - Allowed positions (VariableDeclarator, AssignmentExpression,
//     PropertyAssignment, ReturnStatement, ExportAssignment, outer
//     ArrowFunction body) gate everything else. A bare IIFE or any other
//     CallExpression argument position is NOT allowed, matching upstream's
//     `isInAllowedPositionForComponent` default-false branch.
//
//   - Within an allowed position, specific capitalization rules apply per
//     upstream: VariableDeclarator/PropertyAssignment use the binding name;
//     `Id = fn` assignments use the LHS Identifier; MemberExpression LHS
//     uses the rightmost property name (with `module.exports = ...` as a
//     special blanket-true case); a named FunctionExpression defers to its
//     own Identifier.
//
// Pass the empty string for `pragma` to default to `DefaultReactPragma`.
//
// This wrapper preserves the historical "no checker" call shape used by
// every other React rule. Pass a non-nil checker via
// `IsStatelessReactComponentWithChecker` to enable Identifier-through-scope
// resolution inside the JSX-return checks (relevant for any input where
// the function returns a name bound elsewhere — `return view` ↔
// `let view = <div/>` etc).
func IsStatelessReactComponent(fn *ast.Node, pragma string) bool {
	return isStatelessReactComponentCore(fn, pragma, nil, nil)
}

// IsStatelessReactComponentWithChecker mirrors IsStatelessReactComponent and
// additionally threads `tc` into every isReturningJSX / isReturningJSXOrNull
// gate inside the decision tree. When `tc` is nil, all behavior matches
// `IsStatelessReactComponent` exactly (local-block initializer scan only).
//
// The pragma-component-wrapper branch (Branch 11) uses the hardcoded
// default wrappers (`memo` / `forwardRef`, pragma-qualified or bare). To
// honor `settings.componentWrapperFunctions` here, use
// `IsStatelessReactComponentWithWrappers` instead.
func IsStatelessReactComponentWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return isStatelessReactComponentCore(fn, pragma, tc, nil)
}

// IsStatelessReactComponentWithWrappers is the variant that consults a
// user-provided wrapper list when classifying the inner function of
// pragma-component-wrapper calls (Branch 11 of the decision tree).
//
// Why this matters: `myMemo(() => null)` with
// `settings.componentWrapperFunctions: ['myMemo']` should classify the
// inner arrow as a stateless component (via the wrapper-arm of upstream's
// `getStatelessComponent`), so that the null-only return correctly
// triggers `isStatelessComponentReturningNull` and the rule SKIPs. With
// the hardcoded variant above, `myMemo` isn't recognized → the arrow
// isn't classified → the null-only skip never fires → the rule reports
// where upstream would not.
//
// Pass `wrappers = nil` for hardcoded defaults; pass the configured
// `GetComponentWrapperFunctions(...)` list to honor user settings.
func IsStatelessReactComponentWithWrappers(fn *ast.Node, pragma string, tc *checker.Checker, wrappers []ComponentWrapperEntry) bool {
	return isStatelessReactComponentCore(fn, pragma, tc, wrappers)
}

// isStatelessReactComponentCore is the shared decision tree. `wrappers`
// nil means "use hardcoded memo/forwardRef defaults" (matching the legacy
// public API); non-nil means "consult this list in Branch 11 instead".
func isStatelessReactComponentCore(fn *ast.Node, pragma string, tc *checker.Checker, wrappers []ComponentWrapperEntry) bool {
	if fn == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}

	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// Object-literal shorthand method / accessor. Upstream's Property
		// branch (method && !computed) | (!id && !computed) classifies the
		// inner FE as a component when the property key is a capitalized
		// Identifier AND the function returns strict JSX (isReturningJSX).
		// Setters naturally fail functionReturnsJSX (no return value).
		// Class-body occurrences have a ClassLike parent — NOT
		// ObjectLiteralExpression — and are excluded so they continue to go
		// through the ES6-class path.
		parent := fn.Parent
		if parent == nil || parent.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		name := fn.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return isFirstLetterCapitalized(name.AsIdentifier().Text) && functionReturnsJSXInternal(fn, false, pragma, tc)
	case ast.KindFunctionDeclaration:
		// Branch: FunctionDeclaration requires isReturningJSXOrNull AND
		// (no id || capitalized). Anonymous FD is only legal as
		// `export default function() {...}`.
		if !functionReturnsJSXInternal(fn, true, pragma, tc) {
			return false
		}
		name := fn.Name()
		if name == nil {
			return ast.GetCombinedModifierFlags(fn)&ast.ModifierFlagsDefault != 0
		}
		return name.Kind == ast.KindIdentifier && isFirstLetterCapitalized(name.AsIdentifier().Text)
	case ast.KindFunctionExpression, ast.KindArrowFunction:
	default:
		return false
	}

	parent := fn.Parent
	if parent == nil {
		return false
	}

	// Derived flags mirroring upstream's local `isPropertyAssignment` /
	// `isModuleExportsAssignment`.
	isMEAssign := false
	isModuleExportsAssign := false
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == fn {
			left := ast.SkipParentheses(bin.Left)
			if left.Kind == ast.KindPropertyAccessExpression {
				isMEAssign = true
				pa := left.AsPropertyAccessExpression()
				obj := ast.SkipParentheses(pa.Expression)
				name := pa.Name()
				if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "module" &&
					name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "exports" {
					isModuleExportsAssign = true
				}
			}
		}
	}

	// Branch 1 — ExportDefault (strict isReturningJSX).
	if parent.Kind == ast.KindExportAssignment {
		return functionReturnsJSXInternal(fn, false, pragma, tc)
	}

	// Branch 2 — VariableDeclarator.
	if parent.Kind == ast.KindVariableDeclaration {
		if !functionReturnsJSXInternal(fn, true, pragma, tc) {
			return false
		}
		binding := parent.AsVariableDeclaration().Name()
		if binding != nil && binding.Kind == ast.KindIdentifier {
			return isFirstLetterCapitalized(binding.AsIdentifier().Text)
		}
		return false
	}

	// Branch 3 — early-reject in ReturnStatement / arrow-expression-body
	// when not strictly returning JSX.
	if parent.Kind == ast.KindReturnStatement ||
		(parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body == fn) {
		if !functionReturnsJSXInternal(fn, false, pragma, tc) {
			return false
		}
	}

	// Branch 4 — AssignmentExpression with non-MemberExpression LHS
	// (handled; Identifier LHS path).
	if parent.Kind == ast.KindBinaryExpression && !isMEAssign {
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == fn {
			if !functionReturnsJSXInternal(fn, true, pragma, tc) {
				return false
			}
			// Named FE defers to its own id (matches upstream's final
			// `if (node.id)` check, which runs before the lowercase-LHS
			// reject in the property-assignment tail).
			if fn.Kind == ast.KindFunctionExpression {
				name := fn.Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
			}
			left := ast.SkipParentheses(bin.Left)
			if left.Kind == ast.KindIdentifier {
				return isFirstLetterCapitalized(left.AsIdentifier().Text)
			}
			return false
		}
	}

	// Branches 5 & 6 — nested Arrow whose outer Arrow is itself in an
	// AssignmentExpression / PropertyAssignment position.
	if parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body == fn {
		grand := parent.Parent
		if grand != nil && !isMEAssign && functionReturnsJSXInternal(fn, true, pragma, tc) {
			switch grand.Kind {
			case ast.KindBinaryExpression:
				bin := grand.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == parent {
					left := ast.SkipParentheses(bin.Left)
					if left.Kind == ast.KindIdentifier {
						return isFirstLetterCapitalized(left.AsIdentifier().Text)
					}
					return false
				}
			case ast.KindPropertyAssignment:
				name := grand.AsPropertyAssignment().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
				return false
			}
		}
	}

	// Branches 7 & 8 — inner function in a ReturnStatement whose enclosing
	// function itself sits in an AssignmentExpression / PropertyAssignment
	// position. Upstream first checks the inner FE's own id (if capitalized
	// return it), then walks functionExpr = parent.parent.parent.
	if parent.Kind == ast.KindReturnStatement {
		if fn.Kind == ast.KindFunctionExpression {
			name := fn.Name()
			if name != nil && name.Kind == ast.KindIdentifier && isFirstLetterCapitalized(name.AsIdentifier().Text) {
				return true
			}
		}
		// functionExpr = ReturnStatement.parent (Block) . parent (functionExpr)
		funcExpr := parent.Parent
		if funcExpr != nil {
			funcExpr = funcExpr.Parent
		}
		if funcExpr != nil && funcExpr.Parent != nil && !isMEAssign && functionReturnsJSXInternal(fn, true, pragma, tc) {
			gp := funcExpr.Parent
			switch gp.Kind {
			case ast.KindBinaryExpression:
				bin := gp.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == funcExpr {
					left := ast.SkipParentheses(bin.Left)
					if left.Kind == ast.KindIdentifier {
						return isFirstLetterCapitalized(left.AsIdentifier().Text)
					}
					return false
				}
			case ast.KindPropertyAssignment:
				name := gp.AsPropertyAssignment().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
				return false
			}
		}
	}

	// Branch 9 — parent has a MemberExpression-style key
	// (e.g. `{ [obj.prop]: fn }` computed key resolving to a member access).
	if parent.Kind == ast.KindPropertyAssignment {
		nameNode := parent.AsPropertyAssignment().Name()
		if nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName {
			keyExpr := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
			if keyExpr.Kind == ast.KindPropertyAccessExpression || keyExpr.Kind == ast.KindElementAccessExpression {
				if !functionReturnsJSXInternal(fn, false, pragma, tc) && !functionReturnsOnlyNull(fn) {
					return false
				}
			}
		}
	}

	// Branch 10 — Property method/no-id + !computed form.
	// In tsgo, the `method: true` arm is handled via the MethodDeclaration
	// path above. Here we handle the `!id && !computed` arm — an anonymous
	// FE/Arrow assigned as a PropertyAssignment initializer with Identifier
	// key. Strict isReturningJSX applies.
	if parent.Kind == ast.KindPropertyAssignment {
		pa := parent.AsPropertyAssignment()
		name := pa.Name()
		isComputed := name != nil && name.Kind == ast.KindComputedPropertyName
		hasId := fn.Kind == ast.KindFunctionExpression && fn.Name() != nil
		if !hasId && !isComputed {
			if name != nil && name.Kind == ast.KindIdentifier {
				if !isFirstLetterCapitalized(name.AsIdentifier().Text) {
					return false
				}
				return functionReturnsJSXInternal(fn, false, pragma, tc)
			}
			return false
		}
	}

	// Branch 11 — pragma component wrapper. tsgo preserves `()`, `as`,
	// `satisfies`, `<T>x`, and `x!` wrappers around the arrow argument
	// (ESTree flattens parens and has no equivalent for the TS-only
	// forms), so we walk up through every such wrapper before looking for
	// the enclosing CallExpression.
	effectiveParent := SkipExpressionWrappersUp(fn)
	if effectiveParent != nil && effectiveParent.Kind == ast.KindCallExpression {
		// When the caller threaded `wrappers`, consult the configured
		// list so user `settings.componentWrapperFunctions` entries
		// (`myMemo`, `MyLib.observer`, etc.) participate in stateless-
		// component classification — which in turn makes
		// `isStatelessComponentReturningNull` fire correctly on
		// null-only inner functions of user-configured wrappers. With
		// `wrappers == nil` we fall back to the hardcoded default
		// (memo / forwardRef, pragma-qualified or bare), preserving
		// every legacy caller's behavior.
		matched := false
		if wrappers != nil {
			matched = MatchesAnyComponentWrapperWithChecker(effectiveParent, fn, wrappers, pragma, tc)
		} else {
			matched = isPragmaComponentWrapperCall(effectiveParent, fn, pragma)
		}
		if matched && functionReturnsJSXInternal(fn, true, pragma, tc) {
			return true
		}
	}

	// Branch 12 — require allowed position AND isReturningJSXOrNull.
	if !isInAllowedPositionForComponent(fn) || !functionReturnsJSXInternal(fn, true, pragma, tc) {
		return false
	}

	// Branch 13 — isParentComponentNotStatelessComponent carve-out.
	if parent.Kind == ast.KindPropertyAssignment {
		name := parent.AsPropertyAssignment().Name()
		if name != nil && name.Kind == ast.KindIdentifier &&
			!isFirstLetterCapitalized(name.AsIdentifier().Text) &&
			len(fn.Parameters()) > 0 {
			return false
		}
	}

	// Branch 14 — `if (node.id) return capitalized(node.id.name)`.
	if fn.Kind == ast.KindFunctionExpression {
		name := fn.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return isFirstLetterCapitalized(name.AsIdentifier().Text)
		}
	}

	// Branch 15 — isPropertyAssignment (MemberExpression LHS) but not
	// module.exports: reject when rightmost property name is lowercase.
	if isMEAssign && !isModuleExportsAssign {
		bin := parent.AsBinaryExpression()
		left := ast.SkipParentheses(bin.Left)
		if left.Kind == ast.KindPropertyAccessExpression {
			pa := left.AsPropertyAccessExpression()
			name := pa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && !isFirstLetterCapitalized(name.AsIdentifier().Text) {
				return false
			}
		}
	}

	// Branch 16 — Property parent + returns only null ⇒ undefined.
	// Upstream's tail check:
	//
	//   if (parent.type === 'Property' && utils.isReturningOnlyNull(node)) {
	//     return undefined;
	//   }
	//
	// This is reachable for shapes Branch 10 doesn't filter — anonymous
	// arrow with a COMPUTED key (`{ [k]: () => null }`) and named FE
	// values (`{ Foo: function Bar() { return null; } }` once Branch 14's
	// id-capitalization check has passed). Both cases must fall through
	// to here and get rejected when the body returns only `null`.
	//
	// Use SkipExpressionWrappersUp to make the check paren / TS-wrapper
	// transparent, mirroring ESTree's flattened parent (where
	// `{ [k]: (() => null) }` resolves the arrow's parent directly to
	// the Property node).
	if effective := SkipExpressionWrappersUp(fn); effective != nil &&
		effective.Kind == ast.KindPropertyAssignment && functionReturnsOnlyNull(fn) {
		return false
	}
	return true
}

// functionReturnsOnlyNull mirrors jsxUtil.isReturningOnlyNull: every
// return statement (at depth ≤ 1) returns the `null` literal, and at
// least one return exists. Arrow expression bodies count as a single
// return. Functions without any returns don't qualify.
func functionReturnsOnlyNull(fn *ast.Node) bool {
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		body = af.Body
		if body != nil && body.Kind != ast.KindBlock {
			return ast.SkipParentheses(body).Kind == ast.KindNullKeyword
		}
	case ast.KindMethodDeclaration:
		body = fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		body = fn.AsGetAccessorDeclaration().Body
	}
	if body == nil {
		return false
	}
	sawReturn := false
	allNull := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			sawReturn = true
			rs := n.AsReturnStatement()
			if rs.Expression == nil || ast.SkipParentheses(rs.Expression).Kind != ast.KindNullKeyword {
				allNull = false
			}
			return false
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
		return false
	}
	visit(body)
	return sawReturn && allNull
}

// isInAllowedPositionForComponent mirrors eslint-plugin-react's
// `utils.isInAllowedPositionForComponent`: only parent node kinds in the
// allow-list may host a stateless functional component. Sequence expressions
// (`a, b`) pass through when `fn` is the last operand. ParenthesizedExpression
// wrappers (which ESTree flattens but tsgo preserves) are transparent so
// `const Hello = (init(), arrow)` — whose comma Sequence sits inside parens —
// still reaches the VariableDeclaration ancestor.
func isInAllowedPositionForComponent(fn *ast.Node) bool {
	parent := skipParenParents(fn)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindPropertyAssignment,
		ast.KindReturnStatement,
		ast.KindExportAssignment,
		ast.KindArrowFunction:
		return true
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsToken:
			// AssignmentExpression — always allowed when `fn` is the RHS.
			return bin.Right == fn
		case ast.KindCommaToken:
			// SequenceExpression — only the last operand inherits its parent's
			// allowed-ness.
			if bin.Right == fn {
				return isInAllowedPositionForComponent(parent)
			}
		}
	}
	return false
}

// skipParenParents walks up through ParenthesizedExpression wrappers and
// returns the first non-paren ancestor of `node`, or nil.
func skipParenParents(node *ast.Node) *ast.Node {
	p := node.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		p = p.Parent
	}
	return p
}

// isPragmaComponentWrapperCall reports whether `call` is a React
// component-wrapping call — `<pragma>.memo(fn)` / `<pragma>.forwardRef(fn)` /
// bare `memo(fn)` / bare `forwardRef(fn)` — with `fn` as the first argument.
// Pragma defaults to `DefaultReactPragma` when empty. Mirrors upstream's
// default `wrapperFunctions` entries (`{property: 'memo', object: pragma}`,
// `{property: 'forwardRef', object: pragma}`); the user-configurable
// `settings.componentWrapperFunctions` is NOT honored.
//
// Call-level optional chains (`memo?.(fn)`) are rejected for the same
// reason as `MatchesAnyComponentWrapper` — upstream's
// `isPragmaComponentWrapper` reads `callee.name` on a plain Identifier
// callee, which fails on the OptionalCallExpression / ChainExpression
// shape Babel emits.
func isPragmaComponentWrapperCall(call, fn *ast.Node, pragma string) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	if c.QuestionDotToken != nil {
		return false
	}
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	// Paren- and TS-wrapper-transparent argument match: tsgo preserves
	// `()` / `as` / `satisfies` / `<T>x` / `x!` wrappers that ESTree
	// either flattens or doesn't have at all, so `React.memo((fn))` /
	// `React.memo(fn as Foo)` / `React.memo(fn!)` all surface the wrapper
	// as the first argument rather than `fn` itself.
	if SkipExpressionWrappers(c.Arguments.Nodes[0]) != fn {
		return false
	}
	callee := SkipExpressionWrappers(c.Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		text := callee.AsIdentifier().Text
		return text == "memo" || text == "forwardRef"
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
		text := name.AsIdentifier().Text
		return text == "memo" || text == "forwardRef"
	}
	return false
}

// SourceHasComponentNamedBefore scans `root`'s subtree for a sibling/outer
// component declaration whose name equals `name` and whose start position
// precedes `before`. Mirrors upstream's `getDetectedComponents` filter —
// only `class` declarations and arrow-assigned-to-VariableDeclarator
// declarations qualify; function declarations do NOT (upstream's filter
// in `Components.js getDetectedComponents` only retains those two kinds).
// The position gate replicates upstream's order-dependence: a sibling
// declared AFTER the use site has not yet been added to the components
// list when `isPragmaComponentWrapper` runs, so it must not match here
// either.
func SourceHasComponentNamedBefore(root *ast.Node, name string, before int) bool {
	if root == nil || name == "" {
		return false
	}
	var found bool
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if found || n == nil {
			return
		}
		if n.Pos() >= before {
			return
		}
		switch n.Kind {
		case ast.KindClassDeclaration:
			id := n.Name()
			if id != nil && id.Kind == ast.KindIdentifier && id.AsIdentifier().Text == name {
				found = true
				return
			}
		case ast.KindVariableDeclaration:
			vd := n.AsVariableDeclaration()
			if vd.Initializer == nil {
				break
			}
			init := SkipExpressionWrappers(vd.Initializer)
			if init == nil || init.Kind != ast.KindArrowFunction {
				break
			}
			id := vd.Name()
			if id != nil && id.Kind == ast.KindIdentifier && id.AsIdentifier().Text == name {
				found = true
				return
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(root)
	return found
}

// WrapperWrapsKnownSiblingComponent reports whether `call` is a
// MemberExpression-callee wrapper (e.g. `React.memo(arrow)`) whose
// FunctionLike argument's body returns JSX whose root tag-name matches a
// sibling/outer arrow-assigned-to-VariableDeclarator or ClassDeclaration in
// the same source file declared before `call`. Mirrors upstream's
// `nodeWrapsComponent` gate inside `isPragmaComponentWrapper`, which is
// intentionally name-based (not symbol-based) and only applied to the
// MemberExpression form of the wrapper. The bare-callee form
// (`memo(...)` after `import { memo } from 'react'`) is NOT gated this way
// upstream and must NOT be gated here either.
func WrapperWrapsKnownSiblingComponent(call *ast.Node, fn *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	// Paren / TS-wrapper transparent callee: `(R.memo)(arrow)` /
	// `(R.memo as any)(arrow)` should still trip the gate because
	// upstream's ESTree-flattened `node.callee.type === 'MemberExpression'`
	// check sees the inner MemberExpression directly. tsgo preserves the
	// wrapper, so we strip it before kind-checking.
	expr := SkipExpressionWrappers(call.AsCallExpression().Expression)
	if expr == nil || expr.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	tag := ReturnedJSXRootTagName(fn)
	if tag == "" {
		return false
	}
	src := ast.GetSourceFileOfNode(call)
	if src == nil {
		return false
	}
	return SourceHasComponentNamedBefore(src.AsNode(), tag, call.Pos())
}

// IsDetectedComponent reports whether `node` looks like a React component the
// upstream `Components.detect` pipeline would classify with confidence ≥ 2 —
// i.e. would surface in `components.list()`. Mirrors `components.get(node)`
// for the four AST kinds upstream's detection visits:
//
//   - FunctionDeclaration / FunctionExpression / ArrowFunction (and the
//     object-shorthand Method / Get / Set forms): defers to
//     IsStatelessReactComponentWithWrappers, with a fallback for
//     user-configured wrappers that the hardcoded memo/forwardRef branch
//     wouldn't catch on its own.
//   - ClassDeclaration / ClassExpression: an extends clause that resolves to
//     `<pragma>.Component` / `Component`.
//   - ObjectLiteralExpression: the argument shape of `<createClass>(...)`
//     (ES5 component).
//   - CallExpression: matches a configured wrapper, has a FunctionLike first
//     argument, and is not a MemberExpression wrapper around a body whose
//     root JSX tag refers to a sibling/outer detected component
//     (`nodeWrapsComponent` gate — see WrapperWrapsKnownSiblingComponent).
//
// Note that this function returns true for the inner FunctionLike of a
// pragma-wrapper call AND for the wrapper CallExpression itself — the same
// dual classification upstream produces (the inner arrow's
// `getStatelessComponent` redirects to the outer call, while the outer
// CallExpression listener also runs). Callers that need single-component
// identity must dedupe by node pointer or by remapping inner FunctionLike
// to its enclosing wrapper call (see no-multi-comp's collection pass for
// the canonical pattern).
func IsDetectedComponent(node *ast.Node, pragma, createClass string, wrappers []ComponentWrapperEntry, tc *checker.Checker) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		if IsStatelessReactComponentWithWrappers(node, pragma, tc, wrappers) {
			return true
		}
		parent := SkipExpressionWrappersUp(node)
		if parent != nil && parent.Kind == ast.KindCallExpression &&
			MatchesAnyComponentWrapperWithChecker(parent, node, wrappers, pragma, tc) &&
			FunctionReturnsJSXOrNullWithChecker(node, pragma, tc) {
			return true
		}
		return false
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return ExtendsReactComponent(node, pragma)
	case ast.KindObjectLiteralExpression:
		return IsCreateReactClassObjectArg(node, pragma, createClass)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return false
		}
		inner := SkipExpressionWrappers(call.Arguments.Nodes[0])
		if inner == nil || !IsFunctionLikeForComponent(inner) {
			return false
		}
		if !MatchesAnyComponentWrapperWithChecker(node, inner, wrappers, pragma, tc) {
			return false
		}
		if WrapperWrapsKnownSiblingComponent(node, inner) {
			return false
		}
		return true
	}
	return false
}
