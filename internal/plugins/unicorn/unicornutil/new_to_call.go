package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NewExpressionToCallFixes removes new while preserving comments and the
// parentheses required when a multiline new expression is returned, thrown, or
// yielded.
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
	if needsOperandParentheses(sourceFile, node, nodeRange.Pos(), expressionRange.Pos()) {
		if opening, closing, ok := operandParenthesesRanges(sourceFile, node.Parent); ok {
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

func needsOperandParentheses(sourceFile *ast.SourceFile, node *ast.Node, newPos int, expressionPos int) bool {
	if node.Parent == nil || node.Parent.Kind == ast.KindParenthesizedExpression {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement, ast.KindYieldExpression:
	default:
		return false
	}
	return !sameLine(sourceFile, newPos, expressionPos)
}

func operandParenthesesRanges(sourceFile *ast.SourceFile, statement *ast.Node) (int, int, bool) {
	if sourceFile == nil || statement == nil {
		return 0, 0, false
	}

	statementRange := utils.TrimNodeTextRange(sourceFile, statement)
	keywordLength := len("return")
	switch statement.Kind {
	case ast.KindThrowStatement:
		keywordLength = len("throw")
	case ast.KindYieldExpression:
		keywordLength = len("yield")
	}

	opening := statementRange.Pos() + keywordLength
	if statement.Kind == ast.KindYieldExpression {
		yield := statement.AsYieldExpression()
		if yield != nil && yield.AsteriskToken != nil {
			opening = yield.AsteriskToken.End()
		}
	}
	closing := statementRange.End()
	source := sourceFile.Text()
	for pos := closing - 1; pos >= statementRange.Pos(); pos-- {
		if isWhitespace(source[pos]) {
			continue
		}
		if statement.Kind != ast.KindYieldExpression && source[pos] == ';' {
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
		if char == '\n' || char == '\r' || char == '\u2028' || char == '\u2029' {
			return false
		}
	}
	return true
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}
