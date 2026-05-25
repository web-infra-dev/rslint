package no_multi_comp

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options carries the parsed rule options. Mirrors upstream's schema:
//
//	[{
//	  type: 'object',
//	  properties: {
//	    ignoreStateless: { default: false, type: 'boolean' },
//	  },
//	  additionalProperties: false,
//	}]
type Options struct {
	IgnoreStateless bool
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["ignoreStateless"].(bool); ok {
			opts.IgnoreStateless = v
		}
	}
	return opts
}

// componentEntry pairs a detected component's reporting node with its sort
// key. `pos` is the trimmed position of the node (its source-file location
// excluding any leading trivia / modifiers); we sort by `pos` so the report
// order tracks the order in which upstream's listeners would have added the
// nodes to `components.list()`. The `kind` is cached for the
// `isIgnored`-equivalent classification done after sorting.
type componentEntry struct {
	node *ast.Node
	pos  int
}

// isAsyncGenerator reports whether `node` is a function expression /
// declaration / object-literal shorthand method that is BOTH `async` AND a
// generator (`function*` or `async *Foo() {}`). Mirrors upstream's
// `node.async && node.generator` gate inside the `Components.detect` FE / FD
// listeners — the only path that adds a node with `confidence 0` and excludes
// it from `components.list()`. Arrow functions cannot be generators, so this
// never fires on KindArrowFunction.
//
// MethodDeclaration is included because ESTree's
// `Property { method: true, value: FunctionExpression }` shape funnels object-
// literal shorthand methods through the same FE listener upstream — so a
// `{ async *Foo() {...} }` declaration is banned upstream as an async-generator
// FE. tsgo collapses this into a MethodDeclaration whose AsteriskToken / async
// modifier carry the same signal; we honor it here.
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

// outermostWrapperCall walks up through nested pragma-wrapper calls (paren /
// TS-wrapper transparent) and returns the outer-most CallExpression that
// matches `wrappers`. Mirrors upstream's `getPragmaComponentWrapper` which
// loops while `isPragmaComponentWrapper(currentNode.parent)` keeps yielding
// truthy and tracks the last truthy ancestor. Returns nil when the immediate
// effective parent is not a matching wrapper.
//
// Why upstream walks up: for `React.memo(React.forwardRef(arrow))`, the inner
// arrow's getStatelessComponent returns `pragmaComponentWrapper` — the
// OUTER-MOST wrapper, not the inner forwardRef call. The component is then
// registered against the memo call's range, which on multi-line source
// changes the report's line number compared to picking the inner wrapper.
func outermostWrapperCall(fn *ast.Node, pragma string, wrappers []reactutil.ComponentWrapperEntry, tc *checker.Checker) *ast.Node {
	cur := reactutil.SkipExpressionWrappersUp(fn)
	if cur == nil || cur.Kind != ast.KindCallExpression ||
		!reactutil.MatchesAnyComponentWrapperWithChecker(cur, fn, wrappers, pragma, tc) {
		return nil
	}
	for {
		// Try ascending one more level: the parent of the current
		// wrapper call (paren/TS-wrapper transparent) must itself be a
		// CallExpression matching the wrappers list, with `cur` as its
		// FIRST argument.
		next := reactutil.SkipExpressionWrappersUp(cur)
		if next == nil || next.Kind != ast.KindCallExpression {
			return cur
		}
		if !reactutil.MatchesAnyComponentWrapperWithChecker(next, cur, wrappers, pragma, tc) {
			return cur
		}
		cur = next
	}
}

// collectComponents walks the source file once and returns every node the
// upstream `Components.detect` pipeline would register at confidence ≥ 2,
// deduplicated by node pointer and sorted by source position. Mirrors
// upstream's per-listener add semantics:
//
//   - ClassDeclaration / ClassExpression / ObjectLiteralExpression / a
//     pragma-wrapper CallExpression: register the node itself.
//   - FunctionDeclaration / FunctionExpression / ArrowFunction inside a
//     pragma-wrapper call: upstream's `getStatelessComponent` redirects to
//     the outer wrapper call, so we register the wrapper instead of the
//     inner function. This avoids double-counting when both the
//     FunctionLike listener and the CallExpression listener would
//     otherwise add (and identify) the same outer node — matching
//     upstream's `getId(node)` deduping inside `components.add`.
//   - Bare FunctionLike that classifies as a stateless component on its
//     own (capitalized name, returns JSX, in an allowed position): register
//     the FunctionLike itself.
func collectComponents(sf *ast.SourceFile, pragma, createClass string, wrappers []reactutil.ComponentWrapperEntry, tc *checker.Checker) []componentEntry {
	seen := map[*ast.Node]bool{}
	var entries []componentEntry

	add := func(node *ast.Node) {
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		trimmed := utils.TrimNodeTextRange(sf, node)
		entries = append(entries, componentEntry{node: node, pos: trimmed.Pos()})
	}

	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if reactutil.ExtendsReactComponent(n, pragma) {
				add(n)
			}
		case ast.KindObjectLiteralExpression:
			if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
				add(n)
			}
		case ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			// Object-literal shorthand methods / accessors. Upstream's
			// FunctionExpression listener fires on these via ESTree's
			// `Property { method: true, value: FunctionExpression }`
			// shape. tsgo collapses both forms into a MethodDeclaration
			// (or GetAccessor / SetAccessor) child of an
			// ObjectLiteralExpression — which is what
			// `IsStatelessReactComponentWithWrappers`'s first switch arm
			// already classifies. Class-body occurrences are excluded
			// inside that helper because their parent is a ClassLike.
			//
			// Banned-confidence gate: upstream's FE listener treats an
			// `async *Foo() {...}` shorthand method as an async
			// generator (`node.async && node.generator`) and registers
			// it with confidence 0 — never surfacing in
			// `components.list()`. We mirror that here so the
			// MethodDeclaration form does not slip past
			// IsStatelessReactComponentWithWrappers (which lacks this
			// gate). Getters / setters cannot be generators by syntax,
			// so `isAsyncGenerator` always returns false on them.
			if isAsyncGenerator(n) {
				break
			}
			if reactutil.IsStatelessReactComponentWithWrappers(n, pragma, tc, wrappers) {
				add(n)
			}
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction:
			// Banned-confidence gate (upstream `Components.detect` FE/FD
			// listeners): `node.async && node.generator` immediately
			// adds the node with confidence 0, which never surfaces in
			// `components.list()`. Arrow functions can't be generators
			// (syntax) so this gate only fires on FE/FD.
			if n.Kind != ast.KindArrowFunction && isAsyncGenerator(n) {
				break
			}
			directParent := reactutil.SkipExpressionWrappersUp(n)
			directInWrapper := directParent != nil && directParent.Kind == ast.KindCallExpression &&
				reactutil.MatchesAnyComponentWrapperWithChecker(directParent, n, wrappers, pragma, tc)
			// Wrap-known-sibling gate: when the FunctionLike's enclosing
			// pragma-wrapper call's body returns JSX whose root tag names
			// a sibling/outer detected component, upstream's
			// `isPragmaComponentWrapper` short-circuits to false. The
			// inner FunctionLike then fails Branch 12's allowed-position
			// gate (its parent is the CallExpression) and also returns
			// undefined from `getStatelessComponent`. Net effect: the
			// inner FunctionLike is NOT classified as a component on its
			// own, and the outer call is NOT classified as a wrapper.
			//
			// rslint's `IsStatelessReactComponentWithWrappers` does not
			// thread the nodeWrapsComponent gate into Branch 11, so we
			// guard here before consulting it. (The CallExpression arm
			// below applies the same gate via `IsDetectedComponent`.)
			if directInWrapper && reactutil.WrapperWrapsKnownSiblingComponent(directParent, n) {
				break
			}
			if reactutil.IsStatelessReactComponentWithWrappers(n, pragma, tc, wrappers) {
				// Pragma-wrapper redirect: upstream's
				// `getStatelessComponent` returns
				// `getPragmaComponentWrapper(node)` — the OUTER-MOST
				// wrapper in a chain like `memo(forwardRef(arrow))`. We
				// must mirror that ascent to keep the report node's
				// line number aligned with upstream when wrappers span
				// multiple lines.
				if outer := outermostWrapperCall(n, pragma, wrappers, tc); outer != nil {
					add(outer)
				} else {
					add(n)
				}
			} else if directInWrapper && reactutil.FunctionReturnsJSXOrNullWithChecker(n, pragma, tc) {
				// User-configured wrapper fallback — same shape as
				// reactutil.IsDetectedComponent's FunctionLike arm. The
				// `IsStatelessReactComponentWithWrappers` decision tree
				// only honors wrappers in its Branch 11; user wrappers
				// applied to a FunctionLike that doesn't satisfy a
				// branch's structural gates need this explicit check.
				if outer := outermostWrapperCall(n, pragma, wrappers, tc); outer != nil {
					add(outer)
				}
			}
		case ast.KindCallExpression:
			call := n.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				break
			}
			inner := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
			if inner == nil || !reactutil.IsFunctionLikeForComponent(inner) {
				break
			}
			if !reactutil.MatchesAnyComponentWrapperWithChecker(n, inner, wrappers, pragma, tc) {
				break
			}
			if reactutil.WrapperWrapsKnownSiblingComponent(n, inner) {
				break
			}
			add(n)
		}
		n.ForEachChild(visit)
		return false
	}
	sf.Node.ForEachChild(visit)

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].pos < entries[j].pos
	})
	return entries
}

// isStatelessKindForIgnore returns true for the upstream `isIgnored` filter:
// a component is ignored under `ignoreStateless: true` when its registered
// node is a `Function*`-typed AST node OR a pragma-wrapper CallExpression.
//
// Upstream check (`Components.js` `no-multi-comp.isIgnored`):
//
//	ignoreStateless && (
//	  /Function/.test(component.node.type)
//	  || utils.isPragmaComponentWrapper(component.node)
//	)
//
// `/Function/.test` matches `FunctionDeclaration`, `FunctionExpression`,
// and `ArrowFunctionExpression` (which all contain "Function" in their
// ESTree type name). It does NOT match `MethodDefinition` or class shapes.
//
// AST-shape note: tsgo collapses ESTree's
// `Property { method: true, value: FunctionExpression }` and
// `Property { value: ArrowFunctionExpression }` shapes into a single
// `MethodDeclaration` (or GetAccessor / SetAccessor) child of an
// ObjectLiteralExpression. Upstream registers `Property.value` (a
// FunctionExpression) as the component node, so `/Function/.test`
// matches there. To preserve the SAME observable `ignoreStateless`
// behavior in rslint, the object-literal method/accessor kinds we
// register in `collectComponents` count as stateless here.
//
// For CallExpression nodes we trust the caller: every CallExpression we
// added via `collectComponents` already passed the
// `MatchesAnyComponentWrapperWithChecker` gate (and the
// `WrapperWrapsKnownSiblingComponent` carve-out), so re-classifying it
// here as a wrapper would be redundant.
func isStatelessKindForIgnore(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindCallExpression:
		return true
	}
	return false
}

var NoMultiCompRule = rule.Rule{
	Name: "react/no-multi-comp",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

		entries := collectComponents(ctx.SourceFile, pragma, createClass, wrappers, ctx.TypeChecker)

		kept := entries
		if opts.IgnoreStateless {
			filtered := kept[:0]
			for _, e := range entries {
				if isStatelessKindForIgnore(e.node) {
					continue
				}
				filtered = append(filtered, e)
			}
			kept = filtered
		}

		if len(kept) <= 1 {
			return rule.RuleListeners{}
		}

		for _, e := range kept[1:] {
			ctx.ReportNode(e.node, rule.RuleMessage{
				Id:          "onlyOneComponent",
				Description: "Declare only one React component per file",
			})
		}
		return rule.RuleListeners{}
	},
}
