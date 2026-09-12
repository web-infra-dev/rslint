package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	removeEnd = ecmascript.SkipLeadingWhitespace(source, removeEnd, expressionRange.Pos())

	insertAfterExpression := ""
	if newExpression.Arguments == nil {
		insertAfterExpression = "()"
	}
	callEnd := expressionRange.End()
	if typeArguments := newExpression.TypeArguments; typeArguments != nil && len(typeArguments.Nodes) > 0 {
		closeAngle := scanner.GetRangeOfTokenAtPosition(sourceFile, typeArguments.End())
		if closeAngle.Pos() < len(source) && source[closeAngle.Pos()] == '>' {
			callEnd = closeAngle.End()
		} else {
			return nil
		}
	}
	if calleeRange.End() == expressionRange.End() && callEnd == expressionRange.End() {
		suffix += insertAfterExpression
		insertAfterExpression = ""
	}

	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(nodeRange.Pos(), removeEnd)),
	}
	if opening, closing, ok := operandParenthesesRanges(sourceFile, node, nodeRange.Pos(), expressionRange.Pos()); ok {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(opening, opening), " ("))
		if closing == callEnd {
			if calleeRange.End() == expressionRange.End() && callEnd == expressionRange.End() {
				suffix += ")"
			} else {
				insertAfterExpression += ")"
			}
		} else {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(closing, closing), ")"))
		}
	}

	if suffix != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(calleeRange.End(), calleeRange.End()), suffix))
	}
	if insertAfterExpression != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(callEnd, callEnd), insertAfterExpression))
	}
	return fixes
}

func operandParenthesesRanges(sourceFile *ast.SourceFile, node *ast.Node, newPos, expressionPos int) (int, int, bool) {
	if !ecmascript.ContainsLineTerminator(sourceFile.Text(), newPos, expressionPos) {
		return 0, 0, false
	}

	// A constructor can lead a larger operand, such as `return new Buffer().length`.
	// Follow only ancestors with the same first token: parentheses, a unary
	// operator, or another preceding token already protect against ASI.
	for node.Parent != nil && utils.TrimNodeTextRange(sourceFile, node.Parent).Pos() == newPos {
		node = node.Parent
	}
	statement := node.Parent
	if statement == nil {
		return 0, 0, false
	}
	switch statement.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement, ast.KindYieldExpression:
	default:
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
	closing = ecmascript.SkipTrailingWhitespace(source, statementRange.Pos(), closing)
	if statement.Kind != ast.KindYieldExpression && closing > statementRange.Pos() && source[closing-1] == ';' {
		closing--
	}
	return opening, closing, true
}
