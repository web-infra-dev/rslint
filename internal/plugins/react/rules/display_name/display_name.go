package display_name

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options carries the parsed rule options. Mirrors upstream's schema:
//
//	[{
//	  type: 'object',
//	  properties: {
//	    ignoreTranspilerName: { type: 'boolean' },
//	    checkContextObjects:  { type: 'boolean' },
//	  },
//	  additionalProperties: false,
//	}]
type Options struct {
	IgnoreTranspilerName bool
	CheckContextObjects  bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["ignoreTranspilerName"].(bool); ok {
			opts.IgnoreTranspilerName = v
		}
		if v, ok := optsMap["checkContextObjects"].(bool); ok {
			opts.CheckContextObjects = v
		}
	}
	return opts
}

// reactVersionAtLeast reports whether settings.react.version is >= the given
// major.minor.patch. Uses the same parsing as ReactVersionLessThan.
func reactVersionAtLeast(settings map[string]interface{}, major, minor, patch int) bool {
	return !reactutil.ReactVersionLessThan(settings, major, minor, patch)
}

// supportsNestedMemo reports whether settings.react.version is in the range
// `^0.14.10 || ^15.7.0 || >= 16.12.0`. Mirrors upstream's
// `testReactVersion(context, '^0.14.10 || ^15.7.0 || >= 16.12.0')` gate that
// suppresses the inner forwardRef-inside-memo report.
func supportsNestedMemo(settings map[string]interface{}) bool {
	major, minor, patch := reactutil.ParseReactVersion(settings)
	// `^0.14.10` — 0.14.x with patch ≥ 10.
	if major == 0 && minor == 14 && patch >= 10 {
		return true
	}
	// `^15.7.0` — 15.x with minor ≥ 7.
	if major == 15 && minor >= 7 {
		return true
	}
	// `>= 16.12.0`.
	return reactVersionAtLeast(settings, 16, 12, 0)
}

// detectedComponent pairs a registered component node with its sort key and
// the running `hasDisplayName` flag mirrored from upstream's
// `components.set(node, { hasDisplayName: ... })`. The sort key is the
// node's trimmed source position so we can produce reports in the same order
// as upstream's `components.list()` traversal.
type detectedComponent struct {
	node           *ast.Node
	pos            int
	hasDisplayName bool
}

// contextEntry mirrors upstream's `contextObjects.set(name, { node, hasDisplayName })`.
// The `name` (binding identifier of the context variable) is the lookup key
// used by the `MemberExpression` listener to flip `hasDisplayName`.
type contextEntry struct {
	node           *ast.Node
	hasDisplayName bool
}

// nodeWalker carries per-source-file state for the display-name analysis.
// A single walker fully owns the components / context maps and the helper
// closures that mutate them; the rule's `Run` constructs one per file.
type nodeWalker struct {
	ctx                 rule.RuleContext
	opts                Options
	pragma              string
	createClass         string
	wrappers            []reactutil.ComponentWrapperEntry
	tc                  *checker.Checker
	checkContextObjects bool
	nestedMemoSupported bool

	// Component registry. `byNode` is the source-of-truth for has-displayName
	// state; `order` preserves discovery order for deterministic reporting.
	byNode map[*ast.Node]*detectedComponent
	order  []*detectedComponent

	// nameToComponent indexes the registry by the binding identifier of the
	// owning declaration / property assignment (string name → component).
	// First-discovered wins; collides across scopes when the same name is
	// bound multiple times. Used as the fallback when the TypeChecker
	// can't resolve a reference precisely (or isn't available at all).
	nameToComponent map[string]*detectedComponent

	// Top-level binding map: identifier name → its initializer (the RHS of
	// `var X = ...`, `let X = ...`, `const X = ...`). Used as the fallback
	// for deep `Mixins.Greetings.Hello.displayName` resolution when the
	// TypeChecker isn't available; with TC, the resolver prefers
	// `reactutil.ResolveIdentifierInitializer` which works in any scope.
	topBindings map[string]*ast.Node

	// contextObjects mirrors upstream's `contextObjects` Map keyed by the
	// binding identifier of the createContext target.
	contextObjects map[string]*contextEntry
}

// addComponent registers `node` in the component registry, deduping on the
// node pointer. Returns the entry so callers can flip `hasDisplayName`
// inline. `bindingName` is optional — when non-empty it's also indexed in
// `nameToComponent` so MemberExpression listeners can find the component
// by the variable / class / function name that owns it.
func (w *nodeWalker) addComponent(node *ast.Node, bindingName string) *detectedComponent {
	if node == nil {
		return nil
	}
	if existing, ok := w.byNode[node]; ok {
		if bindingName != "" {
			if _, exists := w.nameToComponent[bindingName]; !exists {
				w.nameToComponent[bindingName] = existing
			}
		}
		return existing
	}
	trimmed := utils.TrimNodeTextRange(w.ctx.SourceFile, node)
	entry := &detectedComponent{node: node, pos: trimmed.Pos()}
	w.byNode[node] = entry
	w.order = append(w.order, entry)
	if bindingName != "" {
		if _, exists := w.nameToComponent[bindingName]; !exists {
			w.nameToComponent[bindingName] = entry
		}
	}
	return entry
}

// isDisplayNameKey reports whether `key` (a property/method/getter key node)
// is the literal identifier or string `displayName`. Mirrors upstream's
// `propsUtil.isDisplayNameDeclaration` — Identifier `displayName` and any
// string-typed Literal `'displayName'`. Computed keys whose expression is a
// string literal (`['displayName']`) collapse onto the same shape in tsgo
// via the ComputedPropertyName wrapper, so we peek through it.
func isDisplayNameKey(key *ast.Node) bool {
	if key == nil {
		return false
	}
	if key.Kind == ast.KindComputedPropertyName {
		key = key.AsComputedPropertyName().Expression
		if key == nil {
			return false
		}
	}
	switch key.Kind {
	case ast.KindIdentifier:
		return key.AsIdentifier().Text == "displayName"
	case ast.KindStringLiteral:
		return key.AsStringLiteral().Text == "displayName"
	case ast.KindNoSubstitutionTemplateLiteral:
		return key.AsNoSubstitutionTemplateLiteral().Text == "displayName"
	}
	return false
}

// isModuleExportsLeft reports whether `expr` (the LHS of an
// AssignmentExpression) is `module.exports`. Mirrors upstream's
// `node.parent.parent.left.object.name !== 'module' || ... !== 'exports'`
// check inside `hasTranspilerName`. Optional-chain / paren wrappers don't
// occur in this position in real code; we still SkipParentheses for safety.
func isModuleExportsLeft(expr *ast.Node) bool {
	if expr == nil {
		return false
	}
	expr = ast.SkipParentheses(expr)
	if expr.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := expr.AsPropertyAccessExpression()
	obj := ast.SkipParentheses(pa.Expression)
	if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != "module" {
		return false
	}
	name := pa.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return name.AsIdentifier().Text == "exports"
}

// skipParenParentsUp walks up through ParenthesizedExpression wrappers
// starting from `node` and returns (resolvedNode, parentNode) — the
// last node still inside the paren chain and its non-paren parent.
// Mirrors upstream's implicit ESTree paren flattening for parent-walks.
// Used by the binding / hasTranspilerName helpers that look at
// `node.parent` after paren-flattening.
func skipParenParentsUp(node *ast.Node) (*ast.Node, *ast.Node) {
	cur := node
	for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
		cur = cur.Parent
	}
	return cur, cur.Parent
}

// bindingFromOwner extracts the binding-identifier name from a parent
// declaration / assignment, returning "" when `parent` is not a recognized
// owner shape OR `cur` is not in the owner's RHS / Initializer slot.
//
// Recognized shapes (matching the cases upstream's `hasTranspilerName`
// inspects via `node.parent.parent`):
//
//   - `var X = ...` / `const X = ...` / `let X = ...` (VariableDeclaration
//     whose Initializer is `cur`).
//   - `X = ...` (BinaryExpression `=` whose Right is `cur` and Left is a
//     paren-stripped Identifier).
//
// `module.exports = ...` LHS is intentionally NOT filtered here — callers
// that need the module.exports gate (`hasTranspilerNameForObject`) check it
// separately to keep this helper single-purpose.
func bindingFromOwner(parent, cur *ast.Node) string {
	if parent == nil {
		return ""
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		if parent.AsVariableDeclaration().Initializer != cur {
			return ""
		}
		nm := parent.Name()
		if nm != nil && nm.Kind == ast.KindIdentifier {
			return nm.AsIdentifier().Text
		}
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken || bin.Right != cur {
			return ""
		}
		left := ast.SkipParentheses(bin.Left)
		if left.Kind == ast.KindIdentifier {
			return left.AsIdentifier().Text
		}
	}
	return ""
}

// hasTranspilerNameForObject mirrors the ObjectExpression arm of upstream's
// `hasTranspilerName`. Returns true when the object literal sits as the
// argument of a createReactClass call whose enclosing context is a
// VariableDeclaration or non-module-exports AssignmentExpression — i.e.
// `var X = createReactClass({...})` / `X = createReactClass({...})`.
// Upstream walks `node.parent.parent` from the ObjectExpression — the parent
// is the createReactClass CallExpression, the parent.parent is the
// VariableDeclarator / AssignmentExpression. Parenthesized wrappers around
// either layer are transparent.
func hasTranspilerNameForObject(node *ast.Node) bool {
	cur, parent := skipParenParentsUp(node)
	// Walk up through the createReactClass CallExpression; upstream reads
	// `node.parent.parent` to find the VariableDeclarator / AssignmentExpression.
	if parent != nil && parent.Kind == ast.KindCallExpression {
		cur, parent = skipParenParentsUp(parent)
	}
	if parent == nil {
		return false
	}
	// `var X = createReactClass({...})` / `const X = ...` / `let X = ...`
	if parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == cur {
		return true
	}
	// `X = createReactClass({...})` — accepted unless LHS is `module.exports`.
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == cur {
			if !isModuleExportsLeft(bin.Left) {
				return true
			}
		}
	}
	return false
}

// hasTranspilerNameForFunctionLike mirrors the FunctionLike arms of
// upstream's `hasTranspilerName`:
//
//   - FunctionDeclaration / FunctionExpression with own `id.name` → true
//     (a generic capitalized id is enough; the rule checks for ANY name)
//   - FunctionExpression / ArrowFunction whose direct effective parent is
//     a VariableDeclarator (`var X = fn`), an object PropertyAssignment
//     (`{ X: fn }`), or an ObjectLiteralExpression shorthand method
//     (`{ X() { ... } }` — collapsed to MethodDeclaration in tsgo) — but
//     only when that owner is NOT itself the argument of an ES5
//     createReactClass call, since the createReactClass property names
//     don't contribute a transpiler-derivable name.
func (w *nodeWalker) hasTranspilerNameForFunctionLike(node *ast.Node) bool {
	if name := functionExpressionOwnName(node); name != "" {
		return true
	}
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	cur, parent := skipParenParentsUp(node)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// `var X = fn` — qualifies. Upstream also walks parent.parent (the
		// `VariableDeclaration`), but the createReactClass guard only fires
		// when the FunctionLike's grandparent is itself a `Property`
		// (object-literal value), not a VariableDeclaration.
		return parent.AsVariableDeclaration().Initializer == cur
	case ast.KindPropertyAssignment:
		// `{ X: fn }` — disqualifies if the owning ObjectLiteralExpression
		// is the argument of a createReactClass call (the property names
		// inside an ES5 component object don't transpile to component names).
		grand := parent.Parent
		if grand != nil && grand.Kind == ast.KindObjectLiteralExpression && reactutil.IsCreateReactClassObjectArg(grand, w.pragma, w.createClass) {
			return false
		}
		return true
	}
	// Object-literal shorthand method `{ X() { ... } }` collapses to a
	// MethodDeclaration whose parent IS the ObjectLiteralExpression — but
	// the AST kinds (FunctionExpression / ArrowFunction) we're checking
	// here exclude shorthand methods anyway. Shorthand methods are handled
	// directly by the MethodDeclaration arm of `collectComponents` /
	// detection logic.
	return false
}

// functionExpressionOwnName returns the FN's own id text (`function Foo()`
// → "Foo") or "" when the FN is anonymous / not a FunctionDeclaration or
// FunctionExpression. Used both for transpiler-name detection and binding
// indexing.
func functionExpressionOwnName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if name := fn.Name(); name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text != "" {
			return name.AsIdentifier().Text
		}
	case ast.KindFunctionExpression:
		fe := node.AsFunctionExpression()
		if name := fe.Name(); name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text != "" {
			return name.AsIdentifier().Text
		}
	}
	return ""
}

// hasTranspilerNameForClass mirrors upstream's classNamed arm — true iff the
// class has a non-empty `id.name`. Both ClassDeclaration and ClassExpression
// expose `Name()` that returns the binding identifier.
func hasTranspilerNameForClass(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		name := node.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text != ""
	}
	return false
}

// outermostWrapperCall walks up through nested pragma-wrapper calls and
// returns the outer-most CallExpression that matches `wrappers`. Mirrors
// upstream's `getPragmaComponentWrapper` loop. Used for the
// `Components.detect` redirect from inner FunctionLike to its outer wrapper
// call, so that the report node's range tracks upstream's
// `getStatelessComponent` output.
func (w *nodeWalker) outermostWrapperCall(fn *ast.Node) *ast.Node {
	cur := reactutil.SkipExpressionWrappersUp(fn)
	if cur == nil || cur.Kind != ast.KindCallExpression ||
		!reactutil.MatchesAnyComponentWrapperWithChecker(cur, fn, w.wrappers, w.pragma, w.tc) {
		return nil
	}
	for {
		next := reactutil.SkipExpressionWrappersUp(cur)
		if next == nil || next.Kind != ast.KindCallExpression {
			return cur
		}
		if !reactutil.MatchesAnyComponentWrapperWithChecker(next, cur, w.wrappers, w.pragma, w.tc) {
			return cur
		}
		cur = next
	}
}

// bindingNameForFunctionLike returns the binding-identifier name of the
// declaration that owns this FunctionLike, or "" when there isn't one.
// Covers: own FN id (`function Foo()`), VariableDeclaration owner
// (`var Foo = fn`), Identifier-LHS assignment (`Foo = fn`).
func bindingNameForFunctionLike(node *ast.Node) string {
	if name := functionExpressionOwnName(node); name != "" {
		return name
	}
	cur, parent := skipParenParentsUp(node)
	return bindingFromOwner(parent, cur)
}

// bindingNameForClass returns the binding-identifier name of a class
// declaration / expression, or "" when none. ClassExpression in
// `var Foo = class {}` doesn't have its own id (anonymous) — but its
// VariableDeclaration parent does, so we walk up. Class*Declaration*
// always carries its own id.
func bindingNameForClass(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text != "" {
		return name.AsIdentifier().Text
	}
	cur, parent := skipParenParentsUp(node)
	// Class binding only flows from a VariableDeclaration. Identifier-LHS
	// assignment (`Foo = class {}`) is upstream-rejected because
	// `namedClass` requires `node.id.name`.
	if parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == cur {
		nm := parent.Name()
		if nm != nil && nm.Kind == ast.KindIdentifier {
			return nm.AsIdentifier().Text
		}
	}
	return ""
}

// bindingNameForObjectExpression returns the binding identifier of the
// containing VariableDeclaration / Identifier-LHS assignment, or "" when
// the object literal isn't directly bound. Mirrors the shape needed by
// `Foo.displayName = ...` resolution for ES5 createReactClass components.
func bindingNameForObjectExpression(node *ast.Node) string {
	if node == nil {
		return ""
	}
	cur, parent := skipParenParentsUp(node)
	// For ES5 components, the object is the argument of createReactClass —
	// walk up through the wrapping CallExpression to find the variable name.
	if parent != nil && parent.Kind == ast.KindCallExpression {
		cur, parent = skipParenParentsUp(parent)
	}
	return bindingFromOwner(parent, cur)
}

// bindingNameForCallExpression returns the binding identifier when the
// pragma-wrapper call is directly bound (e.g. `const Foo = React.memo(...)`),
// else "". Used so `Foo.displayName = 'Foo'` can find the wrapper-call
// component. Skips TS expression wrappers on the way up so
// `Comp = React.forwardRef(...) as SomeComponent` still resolves to `Comp`.
func bindingNameForCallExpression(node *ast.Node) string {
	if node == nil {
		return ""
	}
	cur := node
	for cur.Parent != nil {
		switch cur.Parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			cur = cur.Parent
			continue
		}
		break
	}
	return bindingFromOwner(cur.Parent, cur)
}

// hasOwnDisplayNameProperty scans an ObjectLiteralExpression for an own
// property whose key is `displayName`. Mirrors upstream's
//
//	node.properties.forEach((property) => {
//	  if (!property.key || !propsUtil.isDisplayNameDeclaration(property.key)) return;
//	  markDisplayNameAsDeclared(node);
//	});
//
// Both PropertyAssignment and shorthand-method (MethodDeclaration) /
// accessor (Get/SetAccessor) keys are recognized.
func hasOwnDisplayNameProperty(obj *ast.Node) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		key := prop.Name()
		if isDisplayNameKey(key) {
			return true
		}
	}
	return false
}

// findClassMemberDisplayName scans a class body for any member whose key is
// `displayName`. Mirrors the combined upstream `ClassProperty,
// PropertyDefinition` and `MethodDefinition` listeners — both fire on
// class-body members and call `markDisplayNameAsDeclared(node)`. The owning
// class node is returned implicitly: callers iterate class declarations and
// flip the displayName flag when this returns true.
func findClassMemberDisplayName(class *ast.Node) bool {
	if class == nil {
		return false
	}
	members := class.Members()
	if members == nil {
		return false
	}
	for _, m := range members {
		switch m.Kind {
		case ast.KindPropertyDeclaration,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			if isDisplayNameKey(m.Name()) {
				return true
			}
		}
	}
	return false
}

// resolveCreateContextCall reports whether `expr` (the RHS of a
// VariableDeclaration initializer or AssignmentExpression) is a
// `createContext(...)` or `<X>.createContext(...)` call. Mirrors upstream's
// `isCreateContext` predicate, peeling parens / TS-wrappers off both the
// call and its callee.
func resolveCreateContextCall(expr *ast.Node) bool {
	if expr == nil {
		return false
	}
	expr = reactutil.SkipExpressionWrappers(expr)
	if expr.Kind != ast.KindCallExpression {
		return false
	}
	callee := reactutil.SkipExpressionWrappers(expr.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == "createContext"
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "createContext"
	}
	return false
}

// recordTopBinding remembers that `name` is bound to `init` at the
// source-file top level. Subsequent calls with the same name are ignored
// (first-binding-wins) so a re-assignment doesn't shadow the original
// initializer for `Foo.Bar.displayName` resolution.
func (w *nodeWalker) recordTopBinding(name string, init *ast.Node) {
	if name == "" || init == nil {
		return
	}
	if _, exists := w.topBindings[name]; !exists {
		w.topBindings[name] = init
	}
}

// resolveAndMarkComponentRef tries to resolve `obj` (the receiver of a
// `.displayName` property access) to a registered component via the
// TypeChecker, and marks the component's `hasDisplayName` when found.
// Returns true on a successful match. Always returns false when the
// TypeChecker is nil — callers must fall back to the syntactic indexes
// (`nameToComponent` / `topBindings`) in that case, which is the design
// for non-type-aware rule runs.
//
// Resolution rules — checked in order against `obj`'s value declaration:
//
//  1. The declaration node itself is a registered component (covers
//     `function Foo() {}` and `class Foo {}` shapes).
//  2. The declaration is a VariableDeclaration whose initializer (after
//     `SkipExpressionWrappers`) is a registered component (covers
//     `const Foo = arrow / FE / class-expression / wrapper-call`).
//  3. ES5 indirection: the initializer is a `createReactClass({...})`
//     CallExpression and the inner ObjectLiteralExpression argument is
//     the registered component.
//
// All TC accesses are guarded — `w.tc == nil` short-circuits to false.
func (w *nodeWalker) resolveAndMarkComponentRef(obj *ast.Node) bool {
	if w.tc == nil || obj == nil || obj.Kind != ast.KindIdentifier {
		return false
	}
	symbol := w.tc.GetSymbolAtLocation(obj)
	if symbol == nil {
		return false
	}
	decl := symbol.ValueDeclaration
	if decl == nil {
		return false
	}
	// Rule 1: declaration is the registered component.
	if entry, ok := w.byNode[decl]; ok {
		entry.hasDisplayName = true
		return true
	}
	// Rule 2 & 3: VariableDeclaration → init.
	if decl.Kind != ast.KindVariableDeclaration {
		return false
	}
	init := decl.AsVariableDeclaration().Initializer
	if init == nil {
		return false
	}
	stripped := reactutil.SkipExpressionWrappers(init)
	if entry, ok := w.byNode[stripped]; ok {
		entry.hasDisplayName = true
		return true
	}
	// ES5 indirection: `var X = createReactClass({...})`. The init is the
	// call; the registered component is the inner ObjectLiteralExpression.
	if stripped.Kind == ast.KindCallExpression {
		call := stripped.AsCallExpression()
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			arg := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
			if entry, ok := w.byNode[arg]; ok {
				entry.hasDisplayName = true
				return true
			}
		}
	}
	return false
}

// resolveDeepRelatedComponent mirrors a pragmatic subset of upstream's
// `getRelatedComponent`. Given a MemberExpression like
// `Mixins.Greetings.Hello.displayName`, it walks `Mixins`'s top-level
// initializer through ObjectLiteralExpression properties (`Greetings`,
// `Hello`) and returns the leaf node — typically a FunctionExpression or
// ArrowFunction — so the caller can flip its component's `hasDisplayName`.
//
// The shallow case (`Foo.displayName = ...` where `Foo` is a directly
// declared component) is handled by `nameToComponent` ahead of this; this
// helper only kicks in for paths of length ≥ 2.
func (w *nodeWalker) resolveDeepRelatedComponent(member *ast.Node) *ast.Node {
	if member == nil || member.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	// Walk down to the root identifier, collecting property names along
	// the way (excluding the trailing `displayName`, which the caller
	// already verified). Append-then-reverse keeps the slice ops O(N) —
	// prepending in the loop would re-allocate every iteration.
	var path []string
	cur := ast.SkipParentheses(member.AsPropertyAccessExpression().Expression)
	for cur.Kind == ast.KindPropertyAccessExpression {
		pa := cur.AsPropertyAccessExpression()
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil
		}
		path = append(path, name.AsIdentifier().Text)
		cur = ast.SkipParentheses(pa.Expression)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	if cur.Kind != ast.KindIdentifier {
		return nil
	}
	if len(path) == 0 {
		return nil
	}
	// TC-aware resolution (preferred): `ResolveIdentifierInitializer`
	// uses `tc.GetSymbolAtLocation` when `tc` is non-nil, finding the
	// initializer of the resolved binding regardless of scope. Falls
	// back to a syntactic local-block / SourceFile scan when `tc` is
	// nil — `reactutil` already encodes the defensive pattern, so we
	// just call through. When even that fails, drop to `topBindings`.
	init := reactutil.ResolveIdentifierInitializer(cur, w.tc)
	if init == nil {
		root := cur.AsIdentifier().Text
		if v, ok := w.topBindings[root]; ok {
			init = v
		}
	}
	if init == nil {
		return nil
	}
	node := init
	for _, key := range path {
		node = reactutil.SkipExpressionWrappers(node)
		if node.Kind != ast.KindObjectLiteralExpression {
			return nil
		}
		var found *ast.Node
		for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
			pn := prop.Name()
			if pn == nil {
				continue
			}
			if pn.Kind == ast.KindIdentifier && pn.AsIdentifier().Text == key {
				found = prop
				break
			}
			if pn.Kind == ast.KindStringLiteral && pn.AsStringLiteral().Text == key {
				found = prop
				break
			}
		}
		if found == nil {
			return nil
		}
		switch found.Kind {
		case ast.KindPropertyAssignment:
			node = found.AsPropertyAssignment().Initializer
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			node = found
		default:
			return nil
		}
	}
	return reactutil.SkipExpressionWrappers(node)
}

// isAsyncGenerator reports whether `node` is a function expression /
// declaration / shorthand method that is BOTH `async` AND a generator.
// Mirrors upstream's `Components.detect` listener gate
// `node.async && node.generator → components.add(node, 0)` — confidence 0
// nodes are PERMANENTLY banned from `components.list()`. Arrow functions
// cannot be generators (syntax) so this never fires on KindArrowFunction.
//
// In tsgo, the syntactic generator marker is `AsteriskToken` on the
// FunctionDeclaration / FunctionExpression / MethodDeclaration; the
// `async` modifier is in the modifiers list.
func isAsyncGenerator(node *ast.Node) bool {
	if node == nil {
		return false
	}
	mods := node.Modifiers()
	hasAsync := false
	if mods != nil {
		for _, m := range mods.Nodes {
			if m.Kind == ast.KindAsyncKeyword {
				hasAsync = true
				break
			}
		}
	}
	if !hasAsync {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// classifyAndRegisterFunctionLike runs the upstream FunctionLike component
// classification (Branch 11 of `getStatelessComponent` plus the pragma
// wrapper redirect) and registers the node. `componentNode` is the node
// upstream `components.list()` would surface — either the FunctionLike
// itself or the outer-most pragma-wrapper call when the FunctionLike sits
// inside one.
func (w *nodeWalker) classifyAndRegisterFunctionLike(n *ast.Node) {
	// Async-generator ban — upstream's `Components.detect` adds these at
	// confidence 0, which is permanently excluded from `components.list()`.
	// Without this gate, `async function* Foo() { return <div/>; }` would
	// silently classify here as a component and report. Arrow functions
	// can't be generators by syntax, so the gate is naturally a no-op for
	// KindArrowFunction.
	if isAsyncGenerator(n) {
		return
	}
	directParent := reactutil.SkipExpressionWrappersUp(n)
	directInWrapper := directParent != nil && directParent.Kind == ast.KindCallExpression &&
		reactutil.MatchesAnyComponentWrapperWithChecker(directParent, n, w.wrappers, w.pragma, w.tc)
	// Wrap-known-sibling gate: when the outer wrapper call returns JSX
	// whose root tag names a sibling/outer detected component, upstream's
	// `isPragmaComponentWrapper` short-circuits to false. The inner
	// FunctionLike is then NOT a component.
	if directInWrapper && reactutil.WrapperWrapsKnownSiblingComponent(directParent, n) {
		return
	}
	classifies := reactutil.IsStatelessReactComponentWithWrappers(n, w.pragma, w.tc, w.wrappers)
	if !classifies {
		// User-configured wrapper fallback — mirrors
		// `IsDetectedComponent` FunctionLike arm. Some wrappers don't
		// imply Branch 11 but still register the inner function as a
		// component.
		if directInWrapper && reactutil.FunctionReturnsJSXOrNullWithChecker(n, w.pragma, w.tc) {
			outer := w.outermostWrapperCall(n)
			if outer != nil {
				w.addComponent(outer, bindingNameForCallExpression(outer))
			}
		}
		return
	}
	// Pragma-wrapper redirect — mirrors upstream's
	// `getStatelessComponent → getPragmaComponentWrapper` ascent.
	if outer := w.outermostWrapperCall(n); outer != nil {
		w.addComponent(outer, bindingNameForCallExpression(outer))
		return
	}
	w.addComponent(n, bindingNameForFunctionLike(n))
}

// classifyAndRegisterCallExpression mirrors the CallExpression arm of
// upstream's `Components.detect`. Registers a pragma-wrapper call as a
// component when its first argument is a FunctionLike that classifies (or
// the outer wrapper itself classifies) and the wrapper isn't wrapping a
// known sibling component.
func (w *nodeWalker) classifyAndRegisterCallExpression(n *ast.Node) {
	call := n.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}
	inner := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
	if inner == nil || !reactutil.IsFunctionLikeForComponent(inner) {
		return
	}
	if !reactutil.MatchesAnyComponentWrapperWithChecker(n, inner, w.wrappers, w.pragma, w.tc) {
		return
	}
	if reactutil.WrapperWrapsKnownSiblingComponent(n, inner) {
		return
	}
	// When the call is itself nested inside another wrapper, redirect to
	// the outer-most wrapper for component identity (matches the
	// FunctionLike redirect).
	if outer := w.outermostWrapperCall(inner); outer != nil && outer != n {
		w.addComponent(outer, bindingNameForCallExpression(outer))
		return
	}
	w.addComponent(n, bindingNameForCallExpression(n))
}

// collect walks the source file once and registers every component node /
// context object, plus indexes top-level bindings for later
// MemberExpression resolution. Mirrors the per-listener semantics of
// upstream's `Components.detect` plus `display-name`'s context-tracking
// listeners.
func (w *nodeWalker) collect() {
	sf := w.ctx.SourceFile

	// First pass — top-level binding index. Used to resolve
	// `Mixins.Greetings.Hello.displayName = ...`-style deep references.
	for _, stmt := range sf.Statements.Nodes {
		switch stmt.Kind {
		case ast.KindVariableStatement:
			vs := stmt.AsVariableStatement()
			if vs.DeclarationList == nil {
				continue
			}
			for _, decl := range vs.DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				vd := decl.AsVariableDeclaration()
				name := vd.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					continue
				}
				if vd.Initializer != nil {
					w.recordTopBinding(name.AsIdentifier().Text, vd.Initializer)
				}
			}
		case ast.KindFunctionDeclaration:
			fd := stmt.AsFunctionDeclaration()
			name := fd.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				w.recordTopBinding(name.AsIdentifier().Text, stmt)
			}
		case ast.KindClassDeclaration:
			cd := stmt.AsClassDeclaration()
			name := cd.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				w.recordTopBinding(name.AsIdentifier().Text, stmt)
			}
		}
	}

	// Second pass — visit every node, register components and contexts,
	// and apply the various display-name-marker listeners.
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if reactutil.ExtendsReactComponent(n, w.pragma) {
				entry := w.addComponent(n, bindingNameForClass(n))
				if entry != nil {
					if !w.opts.IgnoreTranspilerName && hasTranspilerNameForClass(n) {
						entry.hasDisplayName = true
					}
					if findClassMemberDisplayName(n) {
						entry.hasDisplayName = true
					}
				}
			}

		case ast.KindObjectLiteralExpression:
			if reactutil.IsCreateReactClassObjectArg(n, w.pragma, w.createClass) {
				entry := w.addComponent(n, bindingNameForObjectExpression(n))
				if entry != nil {
					if !w.opts.IgnoreTranspilerName && hasTranspilerNameForObject(n) {
						entry.hasDisplayName = true
					}
					if hasOwnDisplayNameProperty(n) {
						entry.hasDisplayName = true
					}
				}
			}

		case ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			// Object-literal shorthand-method classification — when the
			// shorthand method itself classifies as a stateless React
			// component (capitalized owner key, returns JSX, not in a
			// disqualifying position). Excludes class-body members
			// (handled by the ClassDeclaration / ClassExpression arm).
			if n.Parent != nil && n.Parent.Kind != ast.KindObjectLiteralExpression {
				break
			}
			if reactutil.IsStatelessReactComponentWithWrappers(n, w.pragma, w.tc, w.wrappers) {
				w.classifyAndRegisterFunctionLike(n)
				// Shorthand methods carry a transpiler-derivable name —
				// upstream's `hasTranspilerName` recognizes
				// `node.parent.method === true` (ESTree's shape for
				// `{ Foo() { ... } }`). tsgo collapses that into a
				// MethodDeclaration directly under the
				// ObjectLiteralExpression; `parent.method` doesn't exist as
				// a separate signal but the kind itself is the equivalent.
				// Don't apply the `outermostWrapperCall == nil` gate here —
				// shorthand methods can't be inside a pragma-wrapper call
				// argument anyway (the wrapper would receive the surrounding
				// ObjectLiteralExpression, not the bare method).
				if !w.opts.IgnoreTranspilerName {
					if entry, ok := w.byNode[n]; ok {
						entry.hasDisplayName = true
					}
				}
			}

		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction:
			w.classifyAndRegisterFunctionLike(n)
			// Mark the FunctionLike's component when its transpiler-derivable
			// name is in scope. Upstream's listener does
			// `if (components.get(node)) markDisplayNameAsDeclared(node)` —
			// `components.get(fn)` only resolves when `fn` itself was the
			// registered component. When `getStatelessComponent(fn)`
			// redirected to an outer pragma-wrapper, `components.get(fn)`
			// returns undefined and no mark happens. Mirror that: only mark
			// when the FunctionLike was registered without a wrapper redirect
			// (i.e. `outermostWrapperCall` returned nil for this FN).
			if !w.opts.IgnoreTranspilerName && w.hasTranspilerNameForFunctionLike(n) {
				if w.outermostWrapperCall(n) == nil {
					if entry, ok := w.byNode[n]; ok {
						entry.hasDisplayName = true
					}
				}
			}

		case ast.KindCallExpression:
			w.classifyAndRegisterCallExpression(n)
			// Upstream CallExpression listener: when the wrapper's first
			// argument is a FunctionLike, register the wrapper call as a
			// component and flip displayName when:
			//   - the wrapper is itself nested inside another wrapper, OR
			//   - the inner FunctionLike has a transpiler name.
			//
			// The first arm fires ONLY when no outer wrapper triggers; if
			// there IS an outer wrapper, upstream's listener returns early
			// without flipping (so the outer wrapper's component, if any,
			// is the one that needs the transpiler-name check).
			call := n.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				break
			}
			inner := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
			if inner == nil {
				break
			}
			if inner.Kind != ast.KindFunctionExpression && inner.Kind != ast.KindArrowFunction {
				break
			}
			if !reactutil.MatchesAnyComponentWrapperWithChecker(n, inner, w.wrappers, w.pragma, w.tc) {
				break
			}
			isWrappedInAnother := false
			if outer := reactutil.SkipExpressionWrappersUp(n); outer != nil && outer.Kind == ast.KindCallExpression {
				if reactutil.MatchesAnyComponentWrapperWithChecker(outer, n, w.wrappers, w.pragma, w.tc) {
					isWrappedInAnother = true
				}
			}
			if isWrappedInAnother || (!w.opts.IgnoreTranspilerName && w.hasTranspilerNameForFunctionLike(inner)) {
				if entry, ok := w.byNode[n]; ok {
					entry.hasDisplayName = true
				}
			}

		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			// `foo.displayName` and `foo["displayName"]` references — mark
			// the related component. Both AST kinds carry an Identifier
			// (or string-literal) "name"; upstream's listener fires on
			// ESTree's unified `MemberExpression` for both.
			var memberKey *ast.Node
			var memberObject *ast.Node
			if n.Kind == ast.KindPropertyAccessExpression {
				pa := n.AsPropertyAccessExpression()
				memberKey = pa.Name()
				memberObject = ast.SkipParentheses(pa.Expression)
			} else {
				ea := n.AsElementAccessExpression()
				memberKey = ea.ArgumentExpression
				memberObject = ast.SkipParentheses(ea.Expression)
			}
			if !isDisplayNameKey(memberKey) {
				break
			}
			obj := memberObject
			// Context-object case: `Hello.displayName = "..."` where
			// `Hello = createContext(...)`.
			if w.checkContextObjects && obj.Kind == ast.KindIdentifier {
				if ctxEntry, ok := w.contextObjects[obj.AsIdentifier().Text]; ok {
					ctxEntry.hasDisplayName = true
				}
			}
			// Component case: `Foo.displayName = ...`. Try TC-aware
			// resolution first (handles cross-scope bindings precisely);
			// fall back to the syntactic indexes when TC is unavailable
			// or doesn't resolve. The TC path mirrors upstream's
			// `getRelatedComponent` ↔ ESLint scope manager interaction.
			if w.resolveAndMarkComponentRef(obj) {
				break
			}
			// Fallback 1 (syntactic): `nameToComponent` first-discovered
			// match. Coarse but works without a TypeChecker.
			if obj.Kind == ast.KindIdentifier {
				if entry, ok := w.nameToComponent[obj.AsIdentifier().Text]; ok {
					entry.hasDisplayName = true
					break
				}
			}
			// Fallback 2: deep path case (`Mixins.Greetings.Hello.displayName = ...`).
			// Only the dot-access form is supported (mirrors upstream's
			// `getRelatedComponent` MemberExpression walk through `.object`
			// chains; bracket-access shapes don't appear in real-world
			// React APIs and upstream doesn't probe them).
			if n.Kind == ast.KindPropertyAccessExpression {
				if leaf := w.resolveDeepRelatedComponent(n); leaf != nil {
					if entry, ok := w.byNode[leaf]; ok {
						entry.hasDisplayName = true
					}
				}
			}

		case ast.KindVariableDeclaration:
			if !w.checkContextObjects {
				break
			}
			vd := n.AsVariableDeclaration()
			if vd.Initializer == nil {
				break
			}
			name := vd.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				break
			}
			if resolveCreateContextCall(vd.Initializer) {
				w.contextObjects[name.AsIdentifier().Text] = &contextEntry{node: n}
			}

		case ast.KindBinaryExpression:
			if !w.checkContextObjects {
				break
			}
			bin := n.AsBinaryExpression()
			if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
				break
			}
			left := ast.SkipParentheses(bin.Left)
			if left.Kind != ast.KindIdentifier {
				break
			}
			if !resolveCreateContextCall(bin.Right) {
				break
			}
			// Record only when the assignment is a top-level expression
			// statement, mirroring upstream's `ExpressionStatement`
			// listener entry point.
			parent := n.Parent
			if parent == nil || parent.Kind != ast.KindExpressionStatement {
				break
			}
			w.contextObjects[left.AsIdentifier().Text] = &contextEntry{node: parent}
		}
		n.ForEachChild(visit)
		return false
	}
	sf.Node.ForEachChild(visit)
}

// isShadowedComponent reports whether the wrapper identifier (`React`,
// `memo`, or `forwardRef`) used by `node` (a CallExpression) is shadowed in
// `node`'s lexical scope. Mirrors upstream's `isShadowedComponent` —
// preventing false positives where users intentionally shadow the React
// wrapper APIs in nested scopes (the canonical examples appear at the top
// of upstream's `valid` block).
func (w *nodeWalker) isShadowedComponent(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		obj := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "React" {
			return isIdentifierShadowed(node, "React")
		}
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		text := callee.AsIdentifier().Text
		if text == "memo" || text == "forwardRef" {
			return isIdentifierShadowed(node, text)
		}
	}
	return false
}

// isIdentifierShadowed reports whether `name` is bound in any enclosing
// function / block / parameter scope between `node` and the source-file
// root. Deliberately NOT a thin wrapper over `utils.IsShadowed`:
//
//   - upstream's `isIdentifierShadowed` only looks for INNER bindings that
//     would shadow the wrapper identifier in a nested scope (e.g. `const
//     memo = …` inside a function body). It does not consider the
//     SourceFile-level binding (the very import that BRINGS `memo` in)
//     "shadowing" — which would invert the shadow-check's purpose and
//     suppress every memo / forwardRef report against any file that
//     actually imports React.
//   - upstream uses a SHALLOW pattern scan (one level of object / array
//     destructuring) — `{ a: { React } }` does NOT shadow `React`.
//     `utils.IsShadowed` / `HasShadowingDeclaration` use
//     `CollectBindingNames`, which recurses into nested patterns and
//     would over-suppress.
//
// Keep the bespoke walk to preserve byte-for-byte alignment with upstream.
func isIdentifierShadowed(node *ast.Node, name string) bool {
	cur := node
	for cur.Parent != nil {
		cur = cur.Parent
		if ast.IsFunctionLike(cur) {
			if body := cur.Body(); body != nil && bodyHasVariableDeclaration(body, name) {
				return true
			}
			if functionParamsBindName(cur, name) {
				return true
			}
			continue
		}
		if cur.Kind == ast.KindBlock {
			if bodyHasVariableDeclaration(cur, name) {
				return true
			}
		}
	}
	return false
}

// bodyHasVariableDeclaration mirrors upstream's `hasVariableDeclaration`
// recursive scan: a `var` / `let` / `const` declaration whose declarator id
// (Identifier, ArrayPattern element, ObjectPattern key) names `name`,
// recursively into BlockStatement children. tsgo wraps `let`/`const` in a
// `VariableStatement` containing a `VariableDeclarationList`, while a
// single arrow body might be an Expression — both shapes are covered.
func bodyHasVariableDeclaration(body *ast.Node, name string) bool {
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return false
	}
	for _, stmt := range body.AsBlock().Statements.Nodes {
		if statementHasVariableDeclaration(stmt, name) {
			return true
		}
	}
	return false
}

// statementHasVariableDeclaration scans a single statement for an own
// VariableDeclaration whose binding includes `name`. Block statements
// recurse to mirror upstream's `node.body.some(...)`.
func statementHasVariableDeclaration(stmt *ast.Node, name string) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		vs := stmt.AsVariableStatement()
		if vs.DeclarationList == nil {
			return false
		}
		for _, decl := range vs.DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			vd := decl.AsVariableDeclaration()
			if bindingIncludesName(vd.Name(), name) {
				return true
			}
		}
	case ast.KindBlock:
		for _, child := range stmt.AsBlock().Statements.Nodes {
			if statementHasVariableDeclaration(child, name) {
				return true
			}
		}
	}
	return false
}

// bindingIncludesName mirrors upstream's check across Identifier /
// ArrayPattern / ObjectPattern bindings: returns true when `name` appears
// at the top level of the binding (no recursion into nested patterns —
// matches upstream's one-level-deep traversal).
func bindingIncludesName(binding *ast.Node, name string) bool {
	if binding == nil {
		return false
	}
	switch binding.Kind {
	case ast.KindIdentifier:
		return binding.AsIdentifier().Text == name
	case ast.KindArrayBindingPattern:
		for _, el := range binding.AsBindingPattern().Elements.Nodes {
			if el.Kind != ast.KindBindingElement {
				continue
			}
			be := el.AsBindingElement()
			if be.Name() != nil && be.Name().Kind == ast.KindIdentifier && be.Name().AsIdentifier().Text == name {
				return true
			}
		}
	case ast.KindObjectBindingPattern:
		for _, el := range binding.AsBindingPattern().Elements.Nodes {
			if el.Kind != ast.KindBindingElement {
				continue
			}
			be := el.AsBindingElement()
			// `{ name }` shorthand — the name node itself is the binding.
			if be.Name() != nil && be.Name().Kind == ast.KindIdentifier && be.Name().AsIdentifier().Text == name {
				return true
			}
			// `{ key: name }` — `PropertyName` is the source key, `Name()`
			// is the local binding. Upstream matches on `prop.key.name`,
			// which for the shorthand-form maps to the binding itself
			// (already covered above). For renamed-form (`{ key: name }`),
			// upstream's `prop.key.name` matches the SOURCE key, not the
			// local — so a binding `{ React: r }` does NOT shadow `React`
			// in upstream. Mirror that by only matching when there is no
			// PropertyName (shorthand) OR when the PropertyName matches
			// `name` (renamed case where the SOURCE key is `name`).
			if be.PropertyName != nil {
				if be.PropertyName.Kind == ast.KindIdentifier && be.PropertyName.AsIdentifier().Text == name {
					return true
				}
			}
		}
	}
	return false
}

// functionParamsBindName mirrors upstream's `currentNode.params.some(...)`
// scan. Looks at the FunctionLike's parameters and returns true when any
// parameter (or shallow destructured parameter property) binds `name`.
// Uses tsgo's standard `node.Parameters()` accessor which covers every
// FunctionLike kind uniformly.
func functionParamsBindName(fn *ast.Node, name string) bool {
	for _, p := range fn.Parameters() {
		if p == nil || p.Kind != ast.KindParameter {
			continue
		}
		if bindingIncludesName(p.AsParameterDeclaration().Name(), name) {
			return true
		}
	}
	return false
}

// isNestedMemo reports whether `node` is a `<wrapper>(<wrapper>(...))`
// shape — specifically a CallExpression whose first argument is itself a
// pragma-component-wrapper CallExpression. Mirrors upstream's `isNestedMemo`,
// which suppresses the inner forwardRef-inside-memo report when the React
// version supports nested memo (`^0.14.10 || ^15.7.0 || >= 16.12.0`).
func (w *nodeWalker) isNestedMemo(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	first := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
	if first == nil || first.Kind != ast.KindCallExpression {
		return false
	}
	// Both the outer and inner calls must be pragma-component-wrappers.
	innerInnerArg := first.AsCallExpression()
	if innerInnerArg.Arguments == nil || len(innerInnerArg.Arguments.Nodes) == 0 {
		return false
	}
	innerFn := reactutil.SkipExpressionWrappers(innerInnerArg.Arguments.Nodes[0])
	if !reactutil.MatchesAnyComponentWrapperWithChecker(first, innerFn, w.wrappers, w.pragma, w.tc) {
		return false
	}
	return reactutil.MatchesAnyComponentWrapperWithChecker(node, first, w.wrappers, w.pragma, w.tc)
}

// DisplayNameRule is the registered rule. Use the `react/` prefix in
// registration.
var DisplayNameRule = rule.Rule{
	Name: "react/display-name",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

		// upstream: `(config.checkContextObjects || false) &&
		// testReactVersion(context, '>= 16.3.0')` — the React-version gate
		// is independent of the option and silently disables it on older
		// configured versions.
		checkContextObjects := opts.CheckContextObjects && reactVersionAtLeast(ctx.Settings, 16, 3, 0)
		nestedMemoSupported := supportsNestedMemo(ctx.Settings)

		w := &nodeWalker{
			ctx:                 ctx,
			opts:                opts,
			pragma:              pragma,
			createClass:         createClass,
			wrappers:            wrappers,
			tc:                  ctx.TypeChecker,
			checkContextObjects: checkContextObjects,
			nestedMemoSupported: nestedMemoSupported,
			byNode:              map[*ast.Node]*detectedComponent{},
			nameToComponent:     map[string]*detectedComponent{},
			topBindings:         map[string]*ast.Node{},
			contextObjects:      map[string]*contextEntry{},
		}
		w.collect()

		// Report missing displayNames in source order — mirrors upstream's
		// `values(list).filter(...).forEach(reportMissingDisplayName)`.
		for _, comp := range w.order {
			if comp.hasDisplayName {
				continue
			}
			if w.nestedMemoSupported && w.isNestedMemo(comp.node) {
				continue
			}
			if w.isShadowedComponent(comp.node) {
				continue
			}
			ctx.ReportRange(reportRangeFor(ctx, comp.node), rule.RuleMessage{
				Id:          "noDisplayName",
				Description: "Component definition is missing display name",
			})
		}
		// Report missing displayNames for context objects — preserve
		// declaration order via a stable iteration (Go's `range` over map
		// is intentionally randomized; we recover order by replaying the
		// source-file walk).
		if checkContextObjects {
			reportContextObjects(ctx, w)
		}

		return rule.RuleListeners{}
	},
}

// reportContextObjects emits the `noContextDisplayName` diagnostic for every
// tracked context object whose `hasDisplayName` is false, in source-file
// declaration order. Iterating Go's map directly would randomize the
// reports; we replay the AST walk so the report order matches upstream's
// `forEach(filter(contextObjects.values(), ...))`.
func reportContextObjects(ctx rule.RuleContext, w *nodeWalker) {
	type pendingContext struct {
		node *ast.Node
		pos  int
	}
	var pending []pendingContext
	for _, e := range w.contextObjects {
		if e.hasDisplayName {
			continue
		}
		trimmed := utils.TrimNodeTextRange(ctx.SourceFile, e.node)
		pending = append(pending, pendingContext{node: e.node, pos: trimmed.Pos()})
	}
	sort.SliceStable(pending, func(i, j int) bool {
		return pending[i].pos < pending[j].pos
	})
	for _, p := range pending {
		ctx.ReportRange(reportRangeFor(ctx, p.node), rule.RuleMessage{
			Id:          "noContextDisplayName",
			Description: "Context definition is missing display name",
		})
	}
}

// reportRangeFor returns the diagnostic range for `node`, aligned with
// upstream's ESTree-style coordinates.
//
// For ClassDeclaration / ClassExpression: ESTree splits `export default
// class …` into `ExportDefaultDeclaration > ClassDeclaration`, so the class
// node's range starts past the `export default` flow modifiers but keeps
// `abstract` / `declare` / decorators on the class itself. tsgo flattens
// all modifiers onto the ClassDeclaration; we recover ESTree's range by
// skipping `export` / `default` keywords only and stopping at the first
// other modifier (or at the `class` keyword when there is none).
//
// For every other node, the standard trimmed-trivia range matches upstream.
func reportRangeFor(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	defaultRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if node == nil {
		return defaultRange
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		text := ctx.SourceFile.Text()
		mods := node.Modifiers()
		pos := node.Pos()
		if mods != nil {
			for _, mod := range mods.Nodes {
				switch mod.Kind {
				case ast.KindExportKeyword, ast.KindDefaultKeyword:
					pos = mod.End()
				default:
					// Decorators / abstract / declare — stop skipping but
					// keep the modifier in the range. Anchor `pos` at the
					// modifier's start so SkipTrivia lands on the modifier
					// (matching ESTree, which keeps `abstract` / decorators
					// on the ClassDeclaration node itself).
					pos = mod.Pos()
					start := scanner.SkipTrivia(text, pos)
					return core.NewTextRange(start, node.End())
				}
			}
		}
		start := scanner.SkipTrivia(text, pos)
		return core.NewTextRange(start, node.End())
	}
	return defaultRange
}
