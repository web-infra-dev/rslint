package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NewExpressionToCallFixes removes new while preserving comments and the
// parentheses required when a multiline new expression is returned or thrown.
// suffix is inserted after the unparenthesized callee before the expression is
// called, for example ".from".
func NewExpressionToCallFixes(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	nodeRange core.TextRange,
	newExpression *ast.NewExpression,
	callee *ast.Node,
	suffix string,
) []rule.RuleFix {
	if sourceFile == nil || node == nil || newExpression == nil || callee == nil {
		return nil
	}

	expressionRange := utils.TrimNodeTextRange(sourceFile, newExpression.Expression)
	calleeRange := utils.TrimNodeTextRange(sourceFile, callee)
	if nodeRange.Pos() >= expressionRange.Pos() || calleeRange.End() > expressionRange.End() {
		return nil
	}

	source := sourceFile.Text()
	removeEnd := nodeRange.Pos() + len("new")
	for removeEnd < expressionRange.Pos() && isWhitespace(source[removeEnd]) {
		removeEnd++
	}

	insertAfterExpression := ""
	if newExpression.Arguments == nil {
		insertAfterExpression = "()"
	}
	if calleeRange.End() == expressionRange.End() {
		suffix += insertAfterExpression
		insertAfterExpression = ""
	}

	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(nodeRange.Pos(), removeEnd)),
	}
	if needsReturnOrThrowParentheses(sourceFile, node, nodeRange.Pos(), expressionRange.Pos()) {
		if opening, closing, ok := returnOrThrowParenthesesRanges(sourceFile, node.Parent); ok {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(opening, opening), " ("))
			if closing == expressionRange.End() {
				if calleeRange.End() == expressionRange.End() {
					suffix += ")"
				} else {
					insertAfterExpression += ")"
				}
			} else {
				fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(closing, closing), ")"))
			}
		}
	}

	if suffix != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(calleeRange.End(), calleeRange.End()), suffix))
	}
	if insertAfterExpression != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(expressionRange.End(), expressionRange.End()), insertAfterExpression))
	}
	return fixes
}

func needsReturnOrThrowParentheses(sourceFile *ast.SourceFile, node *ast.Node, newPos int, expressionPos int) bool {
	if node.Parent == nil || node.Parent.Kind == ast.KindParenthesizedExpression {
		return false
	}
	if node.Parent.Kind != ast.KindReturnStatement && node.Parent.Kind != ast.KindThrowStatement {
		return false
	}
	return !sameLine(sourceFile, newPos, expressionPos)
}

func returnOrThrowParenthesesRanges(sourceFile *ast.SourceFile, statement *ast.Node) (int, int, bool) {
	if sourceFile == nil || statement == nil {
		return 0, 0, false
	}

	statementRange := utils.TrimNodeTextRange(sourceFile, statement)
	keywordLength := len("return")
	if statement.Kind == ast.KindThrowStatement {
		keywordLength = len("throw")
	}

	opening := statementRange.Pos() + keywordLength
	closing := statementRange.End()
	source := sourceFile.Text()
	for pos := closing - 1; pos >= statementRange.Pos(); pos-- {
		if isWhitespace(source[pos]) {
			continue
		}
		if source[pos] == ';' {
			closing = pos
		}
		break
	}
	return opening, closing, true
}

func sameLine(sourceFile *ast.SourceFile, left, right int) bool {
	if left > right {
		left, right = right, left
	}
	for _, char := range sourceFile.Text()[left:right] {
		if char == '\n' || char == '\r' {
			return false
		}
	}
	return true
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}
