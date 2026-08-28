package unicornutil

import (
	"path/filepath"
	"strings"

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

// isSourceOnlyFile reports whether the file is a non-TypeScript source
// (`.mjs`, `.js`, `.cjs`, …) where the type checker's global type
// information would over-classify a bare identifier / member / call
// receiver that the upstream parser would leave "unknown". When the file
// is TypeScript, the type annotation / type-service path stays in charge.
func isSourceOnlyFile(ctx rule.RuleContext) bool {
	if ctx.SourceFile == nil {
		return true
	}
	name := ctx.SourceFile.FileName()
	if name == "" {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	return ext != ".ts" && ext != ".tsx" && ext != ".mts" && ext != ".cts"
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
//
// The rslint tsgo type checker has global type information for built-in
// identifiers (e.g. `undefined` → Undefined, `Symbol` → Symbol,
// `Math.random()` → number) that the upstream parser does NOT have. For a
// source-only `.mjs` input, the type checker would over-classify these as
// known non-arrays and silence reports that upstream emits. The helper
// therefore takes a source-only path that mirrors upstream's
// `getTypeFromStaticValue` returning "unknown" for those shapes, and only
// falls back to `IsKnownNonIndexedCollection` (and therefore the type
// checker) for `.ts` inputs where type annotations make the classification
// defensible.
func ShouldSkipKnownNonArrayReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if isDirectlyReportableReceiver(node) {
		return false
	}

	// Coercion constructors: the only CallExpression callees upstream's
	// `getStaticValueForControlFlow` reliably folds to a known non-array
	// value (a number, string, boolean, or bigint). `Symbol()` is
	// deliberately excluded because every call returns a unique Symbol
	// and cannot be folded.
	if isNonArrayCoercionConstructorCall(ctx, node) {
		return true
	}

	// For CallExpression, the shared static evaluator folds a handful
	// of cases like `String.fromCharCode(65)` and `Array.of(...)`. If
	// it resolves to a non-array value, trust that result.
	if isSourceOnlyFile(ctx) && ast.IsCallExpression(node) {
		staticEvaluator := arrayReceiverStaticEvaluator(ctx)
		if _, known := staticEvaluator.EvalArrayValue(node); known {
			return true
		}
	}

	// Source-only path: trust upstream's "unknown" classification for
	// identifier / member / call receivers. We deliberately skip the
	// shared static evaluator for those shapes because it folds
	// identifiers like `undefined` to a known non-array value that the
	// upstream parser does NOT fold.
	if isSourceOnlyFile(ctx) {
		if isShapeUpstreamUnknown(node) {
			return false
		}
	}

	return IsKnownNonIndexedCollection(ctx, node)
}

// isShapeUpstreamUnknown reports whether node is one of the shapes
// upstream's parser leaves "unknown" for source-only files. The shared
// rslint static evaluator would also fold these (e.g. `undefined` →
// Undefined), but the upstream parser does not, so the rule fires.
func isShapeUpstreamUnknown(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	// Strip type annotations / assertions / non-null so a TS-only
	// expression like `(x as number)` is treated by its underlying shape.
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		return true
	}
	if ast.IsAccessExpression(node) {
		return true
	}
	if ast.IsCallExpression(node) {
		return true
	}
	return false
}

// isNonArrayCoercionConstructorCall reports whether node is a bare call
// to one of the four coercion constructors whose result is statically
// known to be a non-array primitive: `Number(...)`, `String(...)`,
// `Boolean(...)`, `BigInt(...)`. Upstream's
// `getStaticValueForControlFlow` folds these to a primitive value; the
// rslint static evaluator does not, so we mirror the upstream shape here.
//
// `Symbol()` is excluded: every call returns a unique Symbol and the
// static evaluator cannot fold it. `parseInt` / `parseFloat` / `Object()`
// are also excluded for the same reason — they aren't on the upstream
// `staticGlobalProperties` allowlist, and the rslint type checker would
// over-classify their return type for `.mjs` inputs.
func isNonArrayCoercionConstructorCall(ctx rule.RuleContext, node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, ast.OEKParentheses|ast.OEKAssertions)
	if !ast.IsIdentifier(callee) {
		return false
	}
	// If the callee resolves to a local binding, the user's declaration
	// wins; fall through to IsKnownNonIndexedCollection.
	if ctx.Refs != nil && ctx.Refs.ResolveInFile(callee) != nil {
		return false
	}
	switch callee.AsIdentifier().Text {
	case "Number", "String", "Boolean", "BigInt":
		return true
	}
	return false
}
