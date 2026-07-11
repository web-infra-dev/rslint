package prefer_exponentiation_operator

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageUseExponentiation = "Use the '**' operator instead of 'Math.pow'."

var continuationChars = map[byte]bool{
	'(': true,
	'[': true,
	'/': true,
	'`': true,
}

func buildUseExponentiationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useExponentiation",
		Description: messageUseExponentiation,
	}
}

func staticExpressionString(node *ast.Node) (string, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return "", false
	}

	if value, ok := utils.GetStaticExpressionValue(node); ok {
		return value, true
	}

	switch node.Kind {
	case ast.KindTemplateExpression:
		tpl := node.AsTemplateExpression()
		var b strings.Builder
		if tpl.Head != nil {
			b.WriteString(tpl.Head.AsTemplateHead().Text)
		}
		if tpl.TemplateSpans != nil {
			for _, spanNode := range tpl.TemplateSpans.Nodes {
				span := spanNode.AsTemplateSpan()
				value, ok := staticExpressionString(span.Expression)
				if !ok {
					return "", false
				}
				b.WriteString(value)
				if span.Literal != nil {
					b.WriteString(span.Literal.TemplateLiteralLikeData().Text)
				}
			}
		}
		return b.String(), true
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindPlusToken {
			return "", false
		}
		left, ok := staticExpressionString(bin.Left)
		if !ok {
			return "", false
		}
		right, ok := staticExpressionString(bin.Right)
		if !ok {
			return "", false
		}
		return left + right, true
	}

	return "", false
}

func staticAccessName(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsAccessExpression(node) {
		return "", false
	}

	if name, ok := utils.AccessExpressionStaticName(node); ok {
		return name, true
	}

	// ESLint's static property matcher also folds simple string templates and
	// string concatenations, which rslint's shared literal-only helper does not.
	if node.Kind == ast.KindElementAccessExpression {
		return staticExpressionString(node.AsElementAccessExpression().ArgumentExpression)
	}

	return "", false
}

func staticAccessObject(node *ast.Node) *ast.Node {
	return utils.AccessExpressionObject(ast.SkipParentheses(node))
}

// globals is ctx.Globals; a config `/* global Math: off */` /
// `languageOptions.globals` entry un-declares the name, so it no longer
// resolves to a known global — mirrors ESLint's ReferenceTracker finding no
// global-scope variable to track references through.
func isGlobalNameReference(node *ast.Node, name string, globals map[string]bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil ||
		node.Kind != ast.KindIdentifier ||
		node.AsIdentifier().Text != name ||
		utils.IsShadowed(node, name) {
		return false
	}
	if declared, ok := globals[name]; ok && !declared {
		return false
	}
	return true
}

func isGlobalMathExpression(node *ast.Node, globals map[string]bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	if isGlobalNameReference(node, "Math", globals) {
		return true
	}

	name, ok := staticAccessName(node)
	if !ok || name != "Math" {
		return false
	}

	object := staticAccessObject(node)
	if object == nil {
		return false
	}
	return isGlobalNameReference(object, "globalThis", globals)
}

func isMathPowCall(node *ast.Node, globals map[string]bool) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}

	callee := utils.SkipAssertionsAndParens(call.Expression)
	name, ok := staticAccessName(callee)
	if !ok || name != "pow" {
		return false
	}

	object := staticAccessObject(callee)
	if object == nil {
		return false
	}
	return isGlobalMathExpression(object, globals)
}

func unparenthesized(node *ast.Node) *ast.Node {
	return ast.SkipParentheses(node)
}

func expressionText(sf *ast.SourceFile, node *ast.Node) string {
	return utils.TrimmedNodeText(sf, unparenthesized(node))
}

func expressionPrecedence(node *ast.Node) ast.OperatorPrecedence {
	return ast.GetExpressionPrecedence(unparenthesized(node))
}

func doesBaseNeedParens(base *ast.Node) bool {
	base = unparenthesized(base)
	if base == nil {
		return false
	}

	isUnaryBase := false
	if base.Kind == ast.KindPrefixUnaryExpression {
		switch base.AsPrefixUnaryExpression().Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			isUnaryBase = false
		default:
			isUnaryBase = true
		}
	}

	return expressionPrecedence(base) <= ast.OperatorPrecedenceExponentiation ||
		base.Kind == ast.KindAwaitExpression ||
		isUnaryBase
}

func doesExponentNeedParens(exponent *ast.Node) bool {
	exponent = unparenthesized(exponent)
	if exponent == nil {
		return false
	}
	return expressionPrecedence(exponent) < ast.OperatorPrecedenceExponentiation
}

func isParenthesized(node *ast.Node) bool {
	return node != nil && node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression
}

func isBinaryExponentRightChild(parent *ast.Node, node *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	return bin.OperatorToken != nil &&
		bin.OperatorToken.Kind == ast.KindAsteriskAsteriskToken &&
		bin.Right == node
}

func isArgumentOfCallOrNew(parent *ast.Node, node *ast.Node) bool {
	if parent == nil {
		return false
	}
	if ast.IsCallExpression(parent) {
		args := parent.AsCallExpression().Arguments
		if args == nil {
			return false
		}
		for _, arg := range args.Nodes {
			if arg == node {
				return true
			}
		}
		return false
	}
	if parent.Kind == ast.KindNewExpression {
		args := parent.AsNewExpression().Arguments
		if args == nil {
			return false
		}
		for _, arg := range args.Nodes {
			if arg == node {
				return true
			}
		}
	}
	return false
}

func isTypeWrapperParent(parent *ast.Node, node *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindAsExpression:
		return parent.AsAsExpression().Expression == node
	case ast.KindSatisfiesExpression:
		return parent.AsSatisfiesExpression().Expression == node
	case ast.KindTypeAssertionExpression:
		return parent.AsTypeAssertion().Expression == node
	}
	return false
}

func doesExponentiationExpressionNeedParens(node *ast.Node) bool {
	if isParenthesized(node) {
		return false
	}

	parent := node.Parent
	if parent != nil && (ast.IsExpressionWithTypeArgumentsInClassExtendsClause(parent) || isTypeWrapperParent(parent, node)) {
		return true
	}
	if parent == nil || !ast.IsExpression(parent) {
		return false
	}

	parentPrecedence := ast.GetExpressionPrecedence(parent)
	needsParens := parentPrecedence == ast.OperatorPrecedenceInvalid ||
		parentPrecedence >= ast.OperatorPrecedenceExponentiation
	if !needsParens {
		return false
	}

	return !isBinaryExponentRightChild(parent, node) &&
		!isArgumentOfCallOrNew(parent, node) &&
		!ast.IsArgumentExpressionOfElementAccess(node) &&
		!ast.IsArrayLiteralExpression(parent)
}

func parenthesizeIfShould(text string, shouldParenthesize bool) string {
	if shouldParenthesize {
		return "(" + text + ")"
	}
	return text
}

func firstTokenKind(sf *ast.SourceFile, node *ast.Node) ast.Kind {
	node = unparenthesized(node)
	if node == nil {
		return ast.KindUnknown
	}
	return scanner.ScanTokenAtPosition(sf, utils.TrimNodeTextRange(sf, node).Pos())
}

func isBaseStartThatNeedsWholeParens(sf *ast.SourceFile, base *ast.Node) bool {
	switch firstTokenKind(sf, base) {
	case ast.KindOpenBraceToken, ast.KindFunctionKeyword, ast.KindClassKeyword:
		return true
	default:
		return false
	}
}

func trimTokenText(text string) string {
	return strings.TrimFunc(text, unicode.IsSpace)
}

func canTokenTextsBeAdjacent(left string, right string) bool {
	left = trimTokenText(left)
	right = trimTokenText(right)
	if left == "" || right == "" {
		return true
	}

	leftRune, _ := utf8.DecodeLastRuneInString(left)
	rightRune, _ := utf8.DecodeRuneInString(right)
	if scanner.IsIdentifierPart(leftRune) && scanner.IsIdentifierPart(rightRune) {
		return false
	}
	if (leftRune == '+' && rightRune == '+') || (leftRune == '-' && rightRune == '-') {
		return false
	}
	if leftRune == '/' && (rightRune == '/' || rightRune == '*') {
		return false
	}
	return true
}

func previousAdjacentTokenText(sf *ast.SourceFile, pos int) (string, bool) {
	text := sf.Text()
	if pos <= 0 || pos > len(text) {
		return "", false
	}
	if text[pos-1] == '/' && pos >= 2 && text[pos-2] == '*' {
		return "", false
	}
	if unicode.IsSpace(rune(text[pos-1])) {
		return "", false
	}

	start := pos - 1
	for start > 0 && !unicode.IsSpace(rune(text[start-1])) {
		ch := text[start-1]
		if strings.ContainsRune("()[]{};,.?:", rune(ch)) {
			break
		}
		start--
	}
	return text[start:pos], true
}

func nextAdjacentTokenText(sf *ast.SourceFile, pos int) (string, bool) {
	text := sf.Text()
	if pos < 0 || pos >= len(text) {
		return "", false
	}
	if unicode.IsSpace(rune(text[pos])) {
		return "", false
	}
	if text[pos] == '/' && pos+1 < len(text) && (text[pos+1] == '*' || text[pos+1] == '/') {
		return "", false
	}

	end := pos + 1
	for end < len(text) && !unicode.IsSpace(rune(text[end])) {
		ch := text[end]
		if strings.ContainsRune("()[]{};,.?:", rune(ch)) {
			break
		}
		end++
	}
	return text[pos:end], true
}

func buildFix(sf *ast.SourceFile, node *ast.Node) *rule.RuleFix {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil {
		return nil
	}
	args := call.Arguments.Nodes
	if len(args) != 2 {
		return nil
	}
	for _, arg := range args {
		if arg.Kind == ast.KindSpreadElement {
			return nil
		}
	}
	if utils.HasCommentInsideNode(sf, node) {
		return nil
	}

	base := args[0]
	exponent := args[1]
	baseText := expressionText(sf, base)
	exponentText := expressionText(sf, exponent)
	shouldParenthesizeBase := doesBaseNeedParens(base)
	shouldParenthesizeExponent := doesExponentNeedParens(exponent)
	isStart := utils.IsStartOfExpressionStatement(sf, node)
	shouldParenthesizeAll := doesExponentiationExpressionNeedParens(node)

	if !shouldParenthesizeAll && !shouldParenthesizeBase && isStart && isBaseStartThatNeedsWholeParens(sf, base) {
		shouldParenthesizeAll = true
	}

	prefix := ""
	suffix := ""
	nodeRange := utils.TrimNodeTextRange(sf, node)

	if !shouldParenthesizeAll {
		if !shouldParenthesizeBase {
			if before, ok := previousAdjacentTokenText(sf, nodeRange.Pos()); ok {
				if !canTokenTextsBeAdjacent(before, baseText) {
					prefix = " "
				}
			}
		}
		if !shouldParenthesizeExponent {
			if after, ok := nextAdjacentTokenText(sf, nodeRange.End()); ok {
				if !canTokenTextsBeAdjacent(exponentText, after) {
					suffix = " "
				}
			}
		}
	}

	baseReplacement := parenthesizeIfShould(baseText, shouldParenthesizeBase)
	exponentReplacement := parenthesizeIfShould(exponentText, shouldParenthesizeExponent)
	replacement := parenthesizeIfShould(baseReplacement+"**"+exponentReplacement, shouldParenthesizeAll)

	if prefix == "" && isStart && len(replacement) > 0 && continuationChars[replacement[0]] && utils.NeedsPrecedingSemicolon(sf, node) {
		prefix = ";"
	}

	fix := rule.RuleFixReplaceRange(core.NewTextRange(nodeRange.Pos(), nodeRange.End()), prefix+replacement+suffix)
	return &fix
}

// https://eslint.org/docs/latest/rules/prefer-exponentiation-operator
var PreferExponentiationOperatorRule = rule.Rule{
	Name: "prefer-exponentiation-operator",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !isMathPowCall(node, ctx.Globals) {
					return
				}

				if fix := buildFix(ctx.SourceFile, node); fix != nil {
					ctx.ReportNodeWithFixes(node, buildUseExponentiationMessage(), *fix)
					return
				}

				ctx.ReportNode(node, buildUseExponentiationMessage())
			},
		}
	},
}
