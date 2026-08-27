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
// literals and template literals.
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
		ast.KindArrowFunction,
		ast.KindClassExpression,
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
	return IsKnownNonIndexedCollection(ctx, node)
}
