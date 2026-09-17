package unicornutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func startsWithSemicolonHazard(text string) bool {
	return text != "" && strings.ContainsRune("[(/`+-*,.<", rune(text[0]))
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
	if sourceFile == nil || node == nil || !startsWithSemicolonHazard(replacement) {
		return false
	}

	// A replaced receiver or left operand can lead a larger expression. Keep
	// the source start fixed so parentheses and embedded expressions stop us.
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	for node.Parent != nil && !ast.IsExpressionStatement(node.Parent) &&
		utils.TrimNodeTextRange(sourceFile, node.Parent).Pos() == nodeRange.Pos() {
		node = node.Parent
	}
	if node.Parent == nil || !ast.IsExpressionStatement(node.Parent) || isEmbeddedStatement(node.Parent) {
		return false
	}
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
	case ast.KindCloseBraceToken, ast.KindGreaterThanToken, ast.KindExclamationToken:
		// These tokens can end a declaration or a runtime value, including
		// TypeScript instantiations and non-null assertions. Only values can
		// absorb a following expression across the statement boundary.
		for previousNode := ast.GetNodeAtPosition(sourceFile, previous.Start, false); previousNode != nil && previousNode.End() == previous.End; previousNode = previousNode.Parent {
			switch previousNode.Kind {
			case ast.KindObjectLiteralExpression, ast.KindFunctionExpression,
				ast.KindArrowFunction, ast.KindClassExpression,
				ast.KindExpressionWithTypeArguments, ast.KindNonNullExpression:
				return true
			}
		}
		return false
	default:
		return false
	}
}
