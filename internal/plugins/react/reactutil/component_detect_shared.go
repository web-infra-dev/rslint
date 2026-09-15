package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// IsAsyncGeneratorFunction reports whether `node` is a function expression /
// declaration / object-literal shorthand method that is BOTH `async` AND a
// generator (`async function*` or `async *Foo() {}`).
//
// Mirrors upstream's `node.async && node.generator` gate inside the
// `Components.detect` FunctionExpression / FunctionDeclaration listeners —
// the only path that registers a node with `confidence 0`, which permanently
// bans it from `components.get()` and `components.list()`. Arrow functions
// cannot be generators (syntax), so this never fires on KindArrowFunction.
//
// MethodDeclaration is included because ESTree's
// `Property { method: true, value: FunctionExpression }` shape funnels
// object-literal shorthand methods through the same FunctionExpression
// listener upstream, so `{ async *Foo() {...} }` is banned there as an
// async-generator function expression. tsgo collapses that form into a
// MethodDeclaration whose AsteriskToken and `async` modifier carry the same
// signal.
func IsAsyncGeneratorFunction(node *ast.Node) bool {
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

// OutermostComponentWrapperCall walks up through nested pragma-wrapper calls
// (paren / TS-wrapper transparent) and returns the OUTER-MOST CallExpression
// that wraps `fn` and matches `wrappers`. Returns nil when the immediate
// effective parent is not a matching wrapper.
//
// Mirrors upstream's `getPragmaComponentWrapper`, which loops while
// `isPragmaComponentWrapper(currentNode.parent)` keeps yielding truthy and
// remembers the last truthy ancestor. Why the ascent matters: for
// `React.memo(React.forwardRef(arrow))`, the inner arrow's
// `getStatelessComponent` returns the OUTER-MOST wrapper (the memo call),
// not the inner forwardRef call — so the component is registered against the
// memo call's range, which on multi-line source changes the report's line
// number compared to picking the inner wrapper.
func OutermostComponentWrapperCall(fn *ast.Node, pragma string, wrappers []ComponentWrapperEntry, tc *checker.Checker) *ast.Node {
	return OutermostComponentWrapperCallFunc(fn, func(call, arg *ast.Node) bool {
		return MatchesAnyComponentWrapperWithChecker(call, arg, wrappers, pragma, tc)
	})
}

// OutermostComponentWrapperCallFunc is OutermostComponentWrapperCall with a
// caller-supplied wrapper matcher. Callers that memoize the match decision
// per CallExpression (rules that visit the same wrapper call many times) pass
// their cached predicate here instead of paying for a fresh
// MatchesAnyComponentWrapperWithChecker resolution on every ascent.
//
// `matches(call, arg)` must answer "is `call` a component-wrapper call whose
// first argument is `arg`".
func OutermostComponentWrapperCallFunc(fn *ast.Node, matches func(call, arg *ast.Node) bool) *ast.Node {
	cur := SkipExpressionWrappersUp(fn)
	if cur == nil || cur.Kind != ast.KindCallExpression || !matches(cur, fn) {
		return nil
	}
	for {
		// Try ascending one more level: the parent of the current wrapper
		// call (paren / TS-wrapper transparent) must itself be a
		// CallExpression matching the wrapper list, with `cur` as its FIRST
		// argument.
		next := SkipExpressionWrappersUp(cur)
		if next == nil || next.Kind != ast.KindCallExpression || !matches(next, cur) {
			return cur
		}
		cur = next
	}
}
