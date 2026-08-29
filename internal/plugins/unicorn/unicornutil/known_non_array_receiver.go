package unicornutil

import (
	"path/filepath"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	ext := ecmascript.StringToLowerCase(filepath.Ext(name))
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
	directNode := node
	if isSourceOnlyFile(ctx) {
		// tsgo materializes a JSDoc cast in JavaScript as an assertion wrapper;
		// ESTree leaves the visible expression as the receiver node.
		directNode = ast.SkipOuterExpressions(node, ast.OEKAll)
	}
	if isDirectlyReportableReceiver(directNode) {
		return false
	}
	if isSourceOnlyFile(ctx) {
		return classifySourceOnlyArrayReceiver(
			ctx, node, indexedCollectionTargets, keyedCollectionNames,
		) == arrayClassNonTarget
	}
	return IsKnownNonIndexedCollection(ctx, node)
}
