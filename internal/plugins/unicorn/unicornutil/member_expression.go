package unicornutil

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var decimalIntegerPattern = regexp.MustCompile(`^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`)

// ShouldAddParenthesesToMemberExpressionObject mirrors Unicorn's conservative
// helper for turning node into the object of a member expression. This is not
// a general precedence check: decimal integers need parentheses for lexical
// validity (`1.flat()`), and uncommon expression kinds default to parentheses.
func ShouldAddParenthesesToMemberExpressionObject(
	sourceFile *ast.SourceFile,
	node *ast.Node,
) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindThisKeyword,
		ast.KindArrayLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindStringLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return false
	case ast.KindCallExpression:
		// ESTree represents dynamic import as ImportExpression, which falls
		// through to the conservative parenthesized form upstream. tsgo
		// represents it as a CallExpression whose callee is `import`.
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		return callee != nil && callee.Kind == ast.KindImportKeyword
	case ast.KindNewExpression:
		return node.AsNewExpression().Arguments == nil
	case ast.KindNumericLiteral:
		return sourceFile == nil ||
			decimalIntegerPattern.MatchString(utils.TrimmedNodeText(sourceFile, node))
	default:
		return true
	}
}

// ShouldAddParenthesesToMemberExpressionObjectAt also accounts for the source
// position an expression will occupy, including statement and arrow-body starts.
func ShouldAddParenthesesToMemberExpressionObjectAt(source *ast.SourceFile, node, replacement *ast.Node) bool {
	first, _ := utils.TokenAtOrAfter(source, utils.TrimNodeTextRange(source, node).Pos())
	next, _ := utils.TokenAtOrAfter(source, first.End)
	if first.Kind == ast.KindOpenBraceToken || first.Kind == ast.KindFunctionKeyword || first.Kind == ast.KindClassKeyword ||
		(first.Text == "async" && next.Kind == ast.KindFunctionKeyword) ||
		(first.Text == "let" && next.Kind == ast.KindOpenBracketToken) {
		for current := replacement; current.Parent != nil; current = current.Parent {
			parent := current.Parent
			if ast.IsParenthesizedExpression(parent) {
				break
			}
			if parent.Kind == ast.KindExpressionStatement || (parent.Kind == ast.KindArrowFunction && parent.Body() == current && first.Kind == ast.KindOpenBraceToken) {
				return true
			}
			if utils.TrimNodeTextRange(source, parent).Pos() != utils.TrimNodeTextRange(source, current).Pos() {
				break
			}
		}
	}
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindObjectLiteralExpression:
		return true
	case ast.KindTaggedTemplateExpression, ast.KindNonNullExpression, ast.KindMetaProperty:
		return false
	case ast.KindCallExpression:
		return false
	default:
		return ShouldAddParenthesesToMemberExpressionObject(source, node)
	}
}
