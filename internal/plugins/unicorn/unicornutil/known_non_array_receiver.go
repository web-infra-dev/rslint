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
// `getTypeFromStaticValue` classifies as "unknown" without a resolvable
// static value. For these shapes upstream reports the rule rather than
// skipping, even when the type checker would have a definitive answer
// (because a `TypeChecker` typically has global type information that
// over-classifies a source-only file's bare `undefined` / `NaN` /
// `Math.PI` / `Symbol()` as a known non-array).
//
// The Identifier branch checks `ctx.Refs`: a local binding resolves to a
// `const` array declaration and the outer `IsKnownNonIndexedCollection` will
// correctly return false. A global identifier that fails to resolve gets
// the upstream-style "unknown" treatment here so the rule fires. The
// MemberExpression and unresolved CallExpression branches follow the
// upstream source-only rule directly, since the static evaluator rarely
// resolves a member chain or constructor-style call to a known array
// shape and the type checker would over-classify built-in members like
// `Math.PI` or `Symbol()`.
func isSourceOnlyUnknownShape(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		if ctx.Refs == nil {
			return true
		}
		if ctx.Refs.ResolveInFile(node) == nil {
			return true
		}
		return false
	}
	if ast.IsAccessExpression(node) {
		return true
	}
	if ast.IsCallExpression(node) {
		// `Number(1)` / `String("x")` / `Boolean(0)` / `BigInt(1)` /
		// `Symbol()` — these constructor calls are folded by upstream's
		// static evaluator into a known non-array value. The rslint static
		// evaluator only handles method calls (e.g. `String.fromCharCode`,
		// `Array.of`), not bare constructor calls, so replicate the upstream
		// shape here: a bare Identifier callee that is a non-array
		// constructor resolves to a known non-array value, and the rule
		// legitimately skips. `Symbol()` returns a unique Symbol each call
		// and therefore does NOT resolve — keep reporting it.
		if isNonArrayConstructorCall(ctx, node) {
			return false
		}
		// For unknown call results like `Symbol()`, upstream reports; match
		// that by treating the unresolvable case as source-only unknown.
		return true
	}
	return false
}

// isNonArrayConstructorCall reports whether node is a bare constructor call
// (`Number(1)`, `String("x")`, `Boolean(0)`, `BigInt(1)`) that upstream's
// static evaluator folds to a known non-array value. `Symbol()` is excluded
// because every call returns a unique Symbol and the static evaluator cannot
// fold it.
func isNonArrayConstructorCall(ctx rule.RuleContext, node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if callee == nil || !ast.IsIdentifier(callee) {
		return false
	}
	// If the identifier shadows a local binding, defer to the
	// type-checker / reference resolver path so the user's declaration wins.
	if ctx.Refs != nil && ctx.Refs.ResolveInFile(callee) != nil {
		return false
	}
	switch callee.AsIdentifier().Text {
	case "Number", "String", "Boolean", "BigInt":
		return true
	}
	return false
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
	// Source-only JS: identifiers without a local binding (e.g. `undefined`,
	// `NaN`) and bare member expressions are "unknown" in upstream's
	// classification. Skipping the type checker for these avoids
	// over-classifying them via global type info that the user's source
	// doesn't actually carry.
	if isSourceOnlyUnknownShape(ctx, node) {
		return false
	}
	return IsKnownNonIndexedCollection(ctx, node)
}
