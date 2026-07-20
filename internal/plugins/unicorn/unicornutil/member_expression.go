package unicornutil

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
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
		ast.KindCallExpression,
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
	case ast.KindNewExpression:
		return node.AsNewExpression().Arguments == nil
	case ast.KindNumericLiteral:
		return sourceFile == nil ||
			decimalIntegerPattern.MatchString(utils.TrimmedNodeText(sourceFile, node))
	default:
		return true
	}
}
