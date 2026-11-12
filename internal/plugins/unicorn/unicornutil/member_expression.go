package unicornutil

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var decimalIntegerPattern = regexp.MustCompile(`^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`)

// IsReferenceTarget reports whether an expression is written, updated, or
// deleted through any enclosing parentheses and TypeScript assertions.
func IsReferenceTarget(node *ast.Node) bool {
	for parent := node.Parent; parent != nil &&
		ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments); parent = node.Parent {
		node = parent
	}
	return node.Parent != nil && (ast.IsAssignmentTarget(node) || node.Parent.Kind == ast.KindDeleteExpression)
}

// IsReferenceContext reports whether replacing a value with a member access
// could change its call receiver or expose a reference to an enclosing operator.
func IsReferenceContext(node *ast.Node) bool {
	const wrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments
	for node.Parent != nil && ast.IsOuterExpression(node.Parent, wrappers) && node.Parent.Expression() == node {
		node = node.Parent
	}
	if utils.IsWriteReference(node) {
		return true
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		return ast.IsCallExpressionTarget(node, false, false)
	case ast.KindDeleteExpression:
		return parent.Expression() == node
	case ast.KindTaggedTemplateExpression:
		return ast.IsTaggedTemplateTag(node, false, false)
	}
	return false
}

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
	prefix := declarationLikePrefix(source, node)
	if prefix != ast.KindUnknown {
		for current := replacement; current.Parent != nil; current = current.Parent {
			parent := current.Parent
			if ast.IsParenthesizedExpression(parent) {
				break
			}
			if parent.Kind == ast.KindExpressionStatement ||
				(parent.Kind == ast.KindArrowFunction && parent.Body() == current && prefix == ast.KindOpenBraceToken) ||
				(parent.Kind == ast.KindForStatement && parent.Initializer() == current && prefix == ast.KindLetKeyword) ||
				(parent.Kind == ast.KindExportAssignment && !parent.AsExportAssignment().IsExportEquals && parent.Expression() == current &&
					(prefix == ast.KindFunctionKeyword || prefix == ast.KindClassKeyword)) {
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
	case ast.KindTaggedTemplateExpression, ast.KindNonNullExpression, ast.KindMetaProperty, ast.KindSuperKeyword:
		return false
	case ast.KindCallExpression:
		return false
	default:
		return ShouldAddParenthesesToMemberExpressionObject(source, node)
	}
}

// StartsWithDeclarationLikeToken reports whether the raw expression text
// needs parentheses to start a standalone expression statement.
func StartsWithDeclarationLikeToken(source *ast.SourceFile, node *ast.Node) bool {
	return declarationLikePrefix(source, node) != ast.KindUnknown
}

func declarationLikePrefix(source *ast.SourceFile, node *ast.Node) ast.Kind {
	first, _ := utils.TokenAtOrAfter(source, utils.TrimNodeTextRange(source, node).Pos())
	switch first.Kind {
	case ast.KindOpenBraceToken, ast.KindClassKeyword, ast.KindFunctionKeyword:
		return first.Kind
	}
	switch first.Text {
	case "async":
		if next, _ := utils.TokenAtOrAfter(source, first.End); next.Kind == ast.KindFunctionKeyword {
			return ast.KindFunctionKeyword
		}
	case "let":
		if next, _ := utils.TokenAtOrAfter(source, first.End); next.Kind == ast.KindOpenBracketToken {
			return ast.KindLetKeyword
		}
	}
	return ast.KindUnknown
}
