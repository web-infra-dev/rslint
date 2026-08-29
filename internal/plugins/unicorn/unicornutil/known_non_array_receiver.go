package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isDirectlyReportableReceiver mirrors the ESTree node kinds Unicorn reports
// even when the expression is visibly not an array. Arrow and class
// expressions are deliberately absent: upstream classifies those as known
// non-arrays and skips them.
func isDirectlyReportableReceiver(node *ast.Node) bool {
	if node = ast.SkipParentheses(node); node == nil {
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

// isSourceOnlyFile reports whether Unicorn receives the file without parser
// type services. The source-only classifier must not use tsgo's lib.d.ts
// knowledge, which would turn upstream-unknown JavaScript expressions into
// known non-arrays.
func isSourceOnlyFile(ctx rule.RuleContext) bool {
	if ctx.SourceFile == nil {
		return true
	}
	return ctx.SourceFile.ScriptKind == core.ScriptKindJS || ctx.SourceFile.ScriptKind == core.ScriptKindJSX
}

// ShouldSkipKnownNonArrayReceiver mirrors Unicorn's helper of the same name.
// JavaScript uses the source-only recursive/static classifier; TypeScript keeps
// the existing annotation and type-information path.
func ShouldSkipKnownNonArrayReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if isDirectlyReportableReceiver(node) {
		return false
	}
	if isSourceOnlyFile(ctx) {
		return classifySourceOnlyIndexedCollectionReceiver(ctx, node) == arrayClassNonTarget
	}
	return IsKnownNonIndexedCollection(ctx, node)
}
