package operator_assignment

import (
	_ "embed"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed operator_assignment.schema.json
var schemaJSON []byte

// compoundAssignmentPlainOperator maps a compound-assignment operator token
// (e.g. `*=`) to its plain binary-operator token (`*`). Only the 12 operators
// with a shorthand form are present; logical assignment (`&&=`, `||=`, `??=`)
// and plain `=` are intentionally absent.
var compoundAssignmentPlainOperator = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:                              ast.KindPlusToken,
	ast.KindMinusEqualsToken:                             ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:                          ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:                             ast.KindSlashToken,
	ast.KindPercentEqualsToken:                           ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken:                  ast.KindAsteriskAsteriskToken,
	ast.KindLessThanLessThanEqualsToken:                  ast.KindLessThanLessThanToken,
	ast.KindGreaterThanGreaterThanEqualsToken:            ast.KindGreaterThanGreaterThanToken,
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: ast.KindGreaterThanGreaterThanGreaterThanToken,
	ast.KindAmpersandEqualsToken:                         ast.KindAmpersandToken,
	ast.KindCaretEqualsToken:                             ast.KindCaretToken,
	ast.KindBarEqualsToken:                               ast.KindBarToken,
}

func isCommutativeShorthandOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindAsteriskToken, ast.KindAmpersandToken, ast.KindCaretToken, ast.KindBarToken:
		return true
	}
	return false
}

func isNonCommutativeShorthandOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindSlashToken, ast.KindPercentToken,
		ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken, ast.KindAsteriskAsteriskToken:
		return true
	}
	return false
}

// isLiteralPropertyKey reports whether node is a literal usable as a computed
// property key without side effects (matches ESTree's unified "Literal" type:
// string / numeric / bigint / regex literals, plus the `true` / `false` /
// `null` keyword literals).
func isLiteralPropertyKey(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

// skipReceiverWrappers unwraps parentheses and TS-only wrappers (`as`,
// `satisfies`, the non-null `!` operator) that have no runtime effect of
// their own, so a receiver like `(x)`, `x!`, or `(x as T)` is treated the
// same as a bare `x`.
func skipReceiverWrappers(node *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
}

// canBeFixed reports whether node can be safely rewritten between `x = x op y`
// and `x op= y` without changing how many times a getter/setter or a computed
// key's `toString()` runs. Parentheses and TS-only wrappers (`as`,
// `satisfies`, `!`) are transparent — they have no runtime effect, so the
// fixer's raw source-text slicing preserves them verbatim regardless of this
// check's outcome. Any node that is part of an optional chain is rejected
// outright: ESLint's own canBeFixed checks the raw ESTree node type, and an
// optional chain is always wrapped in a ChainExpression there, which matches
// neither of its two accepted shapes.
func canBeFixed(node *ast.Node) bool {
	node = skipReceiverWrappers(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		object := skipReceiverWrappers(node.AsPropertyAccessExpression().Expression)
		return object != nil && (object.Kind == ast.KindIdentifier || object.Kind == ast.KindThisKeyword)
	case ast.KindElementAccessExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		elementAccess := node.AsElementAccessExpression()
		object := skipReceiverWrappers(elementAccess.Expression)
		if object == nil || (object.Kind != ast.KindIdentifier && object.Kind != ast.KindThisKeyword) {
			return false
		}
		argument := skipReceiverWrappers(elementAccess.ArgumentExpression)
		return argument != nil && isLiteralPropertyKey(argument)
	}
	return false
}

func replacedMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaced",
		Description: "Assignment (=) can be replaced with operator assignment (" + operator + ").",
		Data:        map[string]string{"operator": operator},
	}
}

func unexpectedMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected operator assignment (" + operator + ") shorthand.",
		Data:        map[string]string{"operator": operator},
	}
}

// checkAlways implements the default "always" mode: `x = x op y` should be
// written as `x op= y` where possible.
func checkAlways(ctx rule.RuleContext, node *ast.Node) {
	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	rhs := ast.SkipParentheses(binExpr.Right)
	if rhs == nil || rhs.Kind != ast.KindBinaryExpression {
		return
	}
	rhsExpr := rhs.AsBinaryExpression()
	operatorKind := rhsExpr.OperatorToken.Kind
	commutative := isCommutativeShorthandOperator(operatorKind)
	if !commutative && !isNonCommutativeShorthandOperator(operatorKind) {
		return
	}

	left := binExpr.Left
	replacementOperator := scanner.TokenToString(operatorKind) + "="

	if utils.IsSameReference(left, rhsExpr.Left, true) {
		reportReplaced(ctx, node, binExpr, left, rhs, rhsExpr, replacementOperator)
		return
	}
	if commutative && utils.IsSameReference(left, rhsExpr.Right, true) {
		// This case can't be fixed safely: if `a` and `b` both have custom
		// valueOf() behavior, fixing `a = b * a` to `a *= b` would change the
		// order the valueOf() functions run in.
		ctx.ReportNode(node, replacedMessage(replacementOperator))
	}
}

func reportReplaced(
	ctx rule.RuleContext,
	node *ast.Node,
	binExpr *ast.BinaryExpression,
	left *ast.Node,
	rhs *ast.Node,
	rhsExpr *ast.BinaryExpression,
	replacementOperator string,
) {
	msg := replacedMessage(replacementOperator)
	if !canBeFixed(left) || !canBeFixed(rhsExpr.Left) {
		ctx.ReportNode(node, msg)
		return
	}

	ctx.ReportNodeWithDeferredFixes(node, msg, func() []rule.RuleFix {
		sourceFile := ctx.SourceFile
		eqRange := utils.TrimNodeTextRange(sourceFile, binExpr.OperatorToken)
		opRange := utils.TrimNodeTextRange(sourceFile, rhsExpr.OperatorToken)

		if utils.HasCommentInSpan(ctx.Comments.All(), eqRange.End(), opRange.Pos()) {
			return nil
		}

		text := sourceFile.Text()
		nodeRange := utils.TrimNodeTextRange(sourceFile, node)
		leftText := text[nodeRange.Pos():eqRange.Pos()]
		rightText := text[opRange.End():rhs.End()]
		replacement := leftText + replacementOperator + rightText

		return []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, replacement)}
	})
}

// checkNever implements the "never" mode: any of the 12 shorthand compound
// assignment operators should be written as `x = x op y` instead.
func checkNever(ctx rule.RuleContext, node *ast.Node) {
	binExpr := node.AsBinaryExpression()
	plainOperatorKind, ok := compoundAssignmentPlainOperator[binExpr.OperatorToken.Kind]
	if !ok {
		return
	}

	operatorText := scanner.TokenToString(binExpr.OperatorToken.Kind)
	msg := unexpectedMessage(operatorText)

	left := binExpr.Left
	if !canBeFixed(left) {
		ctx.ReportNode(node, msg)
		return
	}

	ctx.ReportNodeWithDeferredFixes(node, msg, func() []rule.RuleFix {
		sourceFile := ctx.SourceFile
		nodeRange := utils.TrimNodeTextRange(sourceFile, node)
		opRange := utils.TrimNodeTextRange(sourceFile, binExpr.OperatorToken)

		if utils.HasCommentInSpan(ctx.Comments.All(), nodeRange.Pos(), opRange.Pos()) {
			return nil
		}

		text := sourceFile.Text()
		leftText := text[nodeRange.Pos():opRange.Pos()]
		plainOperatorText := scanner.TokenToString(plainOperatorKind)

		rightPrecedence := ast.GetExpressionPrecedence(binExpr.Right)
		newOperatorPrecedence := ast.GetBinaryOperatorPrecedence(plainOperatorKind)

		var rightText string
		if rightPrecedence <= newOperatorPrecedence {
			// A lower- (or equal-) precedence right side needs parentheses to
			// preserve grouping (e.g. `foo *= bar + 1` -> `foo * (bar + 1)`).
			// An already-parenthesized right side reports the maximum
			// precedence here, so this branch is never hit for it — the
			// existing parentheses are instead preserved verbatim by the
			// plain-slice branch below.
			rightRange := utils.TrimNodeTextRange(sourceFile, binExpr.Right)
			between := text[opRange.End():rightRange.Pos()]
			rightText = between + "(" + text[rightRange.Pos():rightRange.End()] + ")"
		} else {
			rest := text[opRange.End():node.End()]
			prefix := ""
			if firstRune, size := utf8.DecodeRuneInString(rest); firstRune != utf8.RuneError && size > 0 {
				if !isJSWhitespace(firstRune) && !utils.CanTokenTextsBeAdjacent(plainOperatorText, rest[:size]) {
					prefix = " "
				}
			}
			rightText = prefix + rest
		}

		replacement := leftText + "= " + leftText + plainOperatorText + rightText
		return []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, replacement)}
	})
}

func isJSWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// https://eslint.org/docs/latest/rules/operator-assignment
var OperatorAssignmentRule = rule.Rule{
	Name:   "operator-assignment",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		never := false
		if len(options) > 0 {
			if s, ok := options[0].(string); ok && s == "never" {
				never = true
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if never {
					checkNever(ctx, node)
				} else {
					checkAlways(ctx, node)
				}
			},
		}
	},
}
