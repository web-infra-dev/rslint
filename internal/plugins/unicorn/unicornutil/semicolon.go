package unicornutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func startsWithSemicolonHazard(text string) bool {
	return text != "" && strings.ContainsRune("[(/`+-*,.", rune(text[0]))
}

func isEmbeddedStatement(statement *ast.Node) bool {
	if statement == nil || statement.Parent == nil {
		return false
	}
	parent := statement.Parent
	switch parent.Kind {
	case ast.KindIfStatement:
		ifStatement := parent.AsIfStatement()
		return ifStatement.ThenStatement == statement ||
			ifStatement.ElseStatement == statement
	case ast.KindForStatement:
		return parent.AsForStatement().Statement == statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		return parent.AsForInOrOfStatement().Statement == statement
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Statement == statement
	case ast.KindDoStatement:
		return parent.AsDoStatement().Statement == statement
	case ast.KindWithStatement:
		return parent.AsWithStatement().Statement == statement
	default:
		return false
	}
}

// NeedsSemicolonBefore reports whether replacing node with replacement could
// continue the preceding statement through automatic semicolon insertion.
// Calls inside parentheses and unbraced control-flow bodies are excluded to
// match Unicorn's token- and enclosing-node-aware needsSemicolon helper.
func NeedsSemicolonBefore(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	replacement string,
) bool {
	if sourceFile == nil || node == nil ||
		!startsWithSemicolonHazard(replacement) ||
		utils.OutermostParenthesizedExpression(node) != node ||
		node.Parent == nil || !ast.IsExpressionStatement(node.Parent) ||
		isEmbeddedStatement(node.Parent) {
		return false
	}

	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	previous, ok := utils.TokenBeforePosition(sourceFile, nodeRange.Pos())
	if !ok {
		return false
	}

	switch previous.Kind {
	case ast.KindCloseBracketToken,
		ast.KindCloseParenToken,
		ast.KindIdentifier,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateTail,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	case ast.KindCloseBraceToken:
		// `value = {}` can be followed by an identifier-starting statement,
		// but a replacement beginning with `[` would instead index the object.
		// Blocks, classes, and function bodies ending in `}` do not need this.
		return ast.IsObjectLiteralExpression(
			ast.GetNodeAtPosition(sourceFile, previous.Start, false),
		)
	default:
		return false
	}
}
