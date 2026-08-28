package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// directlyReportableReceiverTypes is the set of ESTree node kinds that
// should be reported even when the receiver is "statically known not to be
// an array" (a string literal, a literal object, a function expression, …).
// The mismatch is visible at the call site, so reporting it is the right
// thing to do.
//
// This mirrors the upstream `shouldSkipKnownNonArrayReceiver` helper in
// eslint-plugin-unicorn. The kinds listed here correspond to ESTree's
// "directly visible" syntactic node categories: array / object / function
// literals and template literals. Critically, `ArrowFunctionExpression` and
// `ClassExpression` are NOT in the set — they fall through to
// `isKnownNonIndexedCollection`, which classifies them as non-array
// expressions and therefore skips them. That mirrors upstream's deliberate
// `nonArrayExpressionTypes` set in `is-array.js`.
func isDirectlyReportableReceiver(node *ast.Node) bool {
	if node == nil {
		return false
	}
	// ESTree / upstream transparently unwraps parens; tsgo keeps
	// ParenthesizedExpression nodes, so strip them before classifying.
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// isSourceOnlyUnknownShape reports whether node is a shape that upstream's
// `getTypeFromStaticValue` / `getTypeFromVariable` classify as "unknown"
// without a resolvable static value. For these shapes upstream reports the
// rule rather than skipping, even when the type checker would have a
// definitive answer (because a `TypeChecker` typically has global type
// information that over-classifies a source-only file's bare `undefined` /
// `NaN` / `Math.PI` / `Symbol()` as a known non-array).
//
// The Identifier branch checks `ctx.Refs`: a non-`const` binding or a
// `const IDENT = MEMBER` initializer recursively fall through to
// "unknown" exactly the way upstream's `getTypeFromVariable` does. The
// MemberExpression and unresolved CallExpression branches follow the
// upstream source-only rule directly.
func isSourceOnlyUnknownShape(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		return isSourceOnlyUnknownIdentifier(ctx, node)
	}
	if ast.IsAccessExpression(node) {
		return true
	}
	if ast.IsCallExpression(node) {
		return isSourceOnlyUnknownCall(ctx, node)
	}
	return false
}

// isSourceOnlyUnknownIdentifier mirrors upstream's `getTypeFromVariable`
// path: only `const IDENT = EXPR` resolves the binding to EXPR's type.
// Anything else (no binding, `let`, `var`, destructuring patterns,
// non-ident binding) stays "unknown" for source-only purposes.
func isSourceOnlyUnknownIdentifier(ctx rule.RuleContext, idNode *ast.Node) bool {
	if ctx.Refs == nil {
		return true
	}
	symbol := ctx.Refs.ResolveInFile(idNode)
	if symbol == nil {
		return true
	}
	if len(symbol.Declarations) != 1 {
		return true
	}
	declaration := symbol.Declarations[0]
	if declaration == nil {
		return true
	}
	// Type annotation, if present, would let the type checker resolve the
	// binding definitively; defer to the outer IsKnownNonIndexedCollection
	// path by reporting "not source-only unknown" here.
	if arrayBindingTypeAnnotation(declaration) != nil {
		return false
	}
	if !ast.IsVariableDeclaration(declaration) {
		return true
	}
	variable := declaration.AsVariableDeclaration()
	if variable == nil {
		return true
	}
	declarationList := declaration.Parent
	if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) {
		return true
	}
	if declarationList.Flags&ast.NodeFlagsConst == 0 {
		return true
	}
	if !ast.IsIdentifier(variable.Name()) {
		// Destructuring binding like `const {value = 1} = {};` — the
		// binding name is a pattern, not an Identifier. Upstream keeps this
		// "unknown".
		return true
	}
	if variable.Initializer == nil {
		return true
	}
	// `const value = EXPR` — recurse into EXPR to mirror upstream's
	// `getTypeFromVariable` recursive call to `getType(init, ...)`. If EXPR
	// is itself an Identifier or MemberExpression, upstream classifies it
	// as "unknown" and so do we.
	initializer := variable.Initializer
	if ast.IsIdentifier(initializer) || ast.IsAccessExpression(initializer) {
		return true
	}
	// For other initializer shapes (call expressions, literals, etc.),
	// fall through to the type checker / static evaluator via
	// IsKnownNonIndexedCollection, which can correctly handle
	// `const value = Number(1)`, `const value = [1, 2]`, etc.
	return false
}

// isSourceOnlyUnknownCall handles CallExpression receivers. Upstream
// resolves a number of common builtin call patterns to a known non-array
// value via its static evaluator; the rslint evaluator only handles a
// subset, so we replicate the upstream shape explicitly here. Calls that
// don't match any known pattern are treated as source-only unknown and
// reported — matching upstream's behavior for `Symbol()`, which returns a
// unique Symbol each call and therefore cannot be folded.
func isSourceOnlyUnknownCall(ctx rule.RuleContext, node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil {
		return true
	}
	if ast.IsIdentifier(callee) {
		return isSourceOnlyUnknownIdentifierCall(ctx, node, callee)
	}
	if ast.IsAccessExpression(callee) {
		return isSourceOnlyUnknownMemberCall(ctx, node, callee)
	}
	// Other callee shapes (IIFEs, optional call chains, …): upstream
	// leaves them unknown.
	return true
}

// isSourceOnlyUnknownIdentifierCall handles calls of the form
// `FUNC(args)` where `FUNC` is a bare Identifier. The rslint static
// evaluator does not fold these, but the shape is enough to decide:
// globals the user can't shadow (`parseInt`, `parseFloat`, `Object`, …)
// and the four coercion constructors return a known non-array value.
// `Symbol()` is excluded because every call returns a unique Symbol.
func isSourceOnlyUnknownIdentifierCall(ctx rule.RuleContext, call *ast.Node, callee *ast.Node) bool {
	// If the callee resolves to a local binding, the user's declaration
	// wins; fall through to IsKnownNonIndexedCollection.
	if ctx.Refs != nil && ctx.Refs.ResolveInFile(callee) != nil {
		return false
	}
	switch callee.AsIdentifier().Text {
	case "Number", "String", "Boolean", "BigInt",
		"parseInt", "parseFloat", "Object":
		return false
	}
	return true
}

// isSourceOnlyUnknownMemberCall handles calls of the form
// `RECEIVER.METHOD(args)`. `String.fromCharCode` is already handled by the
// static evaluator; the rest of the upstream-resolved set (Math.abs,
// Math.max, Array.isArray, …) is too large to enumerate, so we use a
// receiver-name heuristic. A receiver that is a known global with a
// method that returns a primitive resolves to a known non-array value
// and the rule legitimately skips.
func isSourceOnlyUnknownMemberCall(ctx rule.RuleContext, call *ast.Node, callee *ast.Node) bool {
	// First let the shared static evaluator take a swing: it knows how to
	// fold `String.fromCharCode`, `Array.of`, and a few string methods.
	// Anything it resolves falls through to IsKnownNonIndexedCollection.
	staticEvaluator := arrayReceiverStaticEvaluator(ctx)
	if _, known := staticEvaluator.EvalArrayValue(call); known {
		// Resolved to a known value: let the outer
		// IsKnownNonIndexedCollection path decide.
		return false
	}
	receiver := accessExpressionObject(callee)
	if receiver == nil {
		return true
	}
	if !ast.IsIdentifier(receiver) {
		// Nested member access like `Foo.Bar.baz()`; upstream can't
		// statically resolve these.
		return true
	}
	// If the receiver resolves to a local binding, the user's
	// declaration wins; fall through.
	if ctx.Refs != nil && ctx.Refs.ResolveInFile(receiver) != nil {
		return false
	}
	// Known global receivers that return non-array values for any
	// well-known method call. `Symbol` is excluded because Symbol values
	// cannot be folded.
	switch receiver.AsIdentifier().Text {
	case "Math", "Number", "String", "Boolean", "BigInt", "Array", "Object", "JSON":
		return false
	}
	return true
}

// accessExpressionObject returns the object of a PropertyAccessExpression
// or ElementAccessExpression, with parentheses stripped. Returns nil for
// other access shapes.
func accessExpressionObject(callee *ast.Node) *ast.Node {
	if callee == nil {
		return nil
	}
	if ast.IsPropertyAccessExpression(callee) {
		return ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	}
	if ast.IsElementAccessExpression(callee) {
		return ast.SkipParentheses(callee.AsElementAccessExpression().Expression)
	}
	return nil
}

// ShouldSkipKnownNonArrayReceiver mirrors the upstream helper of the same
// name. A receiver that is statically known to be a non-array (a typed Set,
// a custom class with a same-named method, etc.) should be skipped UNLESS
// the receiver is one of the directly-reportable kinds above, in which case
// the mismatch between the receiver and `Array#flat()` is visible at the
// call site and worth reporting.
//
// A typed-array receiver is still reported, since it shares most of
// `Array`'s method surface and `flat()` is not on its prototype at all.
// `require-array-sort-compare` (which needs typed arrays to be skipped)
// should call `IsKnownNonArray` directly instead.
func ShouldSkipKnownNonArrayReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if isDirectlyReportableReceiver(node) {
		return false
	}
	// Calls that the source-only path recognizes as folding to a known
	// non-array value skip the rule, mirroring upstream's static
	// evaluator. The type checker would over-classify or under-classify
	// these in source-only files, so we short-circuit here.
	if ast.IsCallExpression(node) {
		if isNonArrayResolvingCall(ctx, node) {
			return true
		}
	}
	// Source-only JS: identifiers without a local binding, identifiers
	// bound via `let`/`var`/destructuring, bare member expressions, and
	// call expressions the static evaluator cannot fold are "unknown" in
	// upstream's classification. Skipping the type checker for these
	// avoids over-classifying them via global type info that the user's
	// source doesn't actually carry.
	if isSourceOnlyUnknownShape(ctx, node) {
		return false
	}
	return IsKnownNonIndexedCollection(ctx, node)
}

// isNonArrayResolvingCall reports whether node is a CallExpression that
// upstream's static evaluator folds to a known non-array value, including
// the four coercion constructors (`Number`, `String`, `Boolean`, `BigInt`),
// the `parseInt` / `parseFloat` / `Object` global functions, and the
// well-known global method receivers (`Math.*`, `Array.isArray`, etc.).
// `Symbol()` is excluded because every call returns a unique Symbol and
// the static evaluator cannot fold it.
func isNonArrayResolvingCall(ctx rule.RuleContext, node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil {
		return false
	}
	if ast.IsIdentifier(callee) {
		// If the callee resolves to a local binding, the user's
		// declaration wins; fall through to IsKnownNonIndexedCollection.
		if ctx.Refs != nil && ctx.Refs.ResolveInFile(callee) != nil {
			return false
		}
		switch callee.AsIdentifier().Text {
		case "Number", "String", "Boolean", "BigInt",
			"parseInt", "parseFloat", "Object":
			return true
		}
		return false
	}
	if ast.IsAccessExpression(callee) {
		// First let the shared static evaluator take a swing: it knows
		// how to fold `String.fromCharCode`, `Array.of`, and a few
		// string methods.
		staticEvaluator := arrayReceiverStaticEvaluator(ctx)
		if _, known := staticEvaluator.EvalArrayValue(node); known {
			// Resolved: let the outer IsKnownNonIndexedCollection path
			// decide.
			return false
		}
		receiver := accessExpressionObject(callee)
		if receiver == nil || !ast.IsIdentifier(receiver) {
			return false
		}
		// If the receiver resolves to a local binding, the user's
		// declaration wins; fall through.
		if ctx.Refs != nil && ctx.Refs.ResolveInFile(receiver) != nil {
			return false
		}
		// Known global receivers that return non-array values for any
		// well-known method call. `Symbol` is excluded because Symbol
		// values cannot be folded.
		switch receiver.AsIdentifier().Text {
		case "Math", "Number", "String", "Boolean", "BigInt", "Array", "Object", "JSON":
			return true
		}
	}
	return false
}
