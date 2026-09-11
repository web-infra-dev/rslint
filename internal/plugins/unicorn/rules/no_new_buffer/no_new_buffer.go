package no_new_buffer

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDError        = "error"
	messageIDErrorUnknown = "error-unknown"
	messageIDSuggestion   = "suggestion"
)

var NoNewBufferRule = rule.Rule{
	Name:   "unicorn/no-new-buffer",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		staticEvaluator := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()
				if newExpression == nil {
					return
				}

				callee := skipParentheses(newExpression.Expression)
				if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Buffer" {
					return
				}

				method := inferMethod(newExpression.Arguments, staticEvaluator, ctx)
				if method == "" {
					ctx.ReportNodeWithDeferredSuggestions(node, messageUnknown(), func() []rule.RuleSuggestion {
						return suggestions(ctx, node, newExpression, callee)
					})
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, messageForMethod(method), func() []rule.RuleFix {
					return fixes(ctx, node, newExpression, callee, method)
				})
			},
		}
	},
}

func inferMethod(arguments *ast.NodeList, staticEvaluator *utils.StaticStringEvaluator, ctx rule.RuleContext) string {
	if arguments == nil || len(arguments.Nodes) != 1 {
		return "from"
	}

	argument := arguments.Nodes[0]
	if argument == nil || argument.Kind == ast.KindSpreadElement {
		return ""
	}
	argument = skipParentheses(argument)
	if argument == nil {
		return ""
	}

	switch argument.Kind {
	case ast.KindArrayLiteralExpression, ast.KindTemplateExpression, ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return "from"
	}

	if isNumber(argument, staticEvaluator, ctx) {
		return "alloc"
	}

	if value, ok := staticEvaluator.EvalControlFlowValue(argument); ok {
		if _, isNumber := value.(interface{ IsNaN() bool }); isNumber {
			return "alloc"
		}
		if _, isString := value.(string); isString {
			return "from"
		}
	}
	if isArray, known := staticEvaluator.EvalControlFlowArrayValue(argument); known && isArray {
		return "from"
	}
	return ""
}

// isNumber ports the portion of unicorn's rules/utils/is-number.js that
// determines Buffer's safe numeric constructor arguments. Keep this separate
// from static value classification: a known non-number must remain a
// suggestion, not silently become Buffer.from().
func isNumber(node *ast.Node, staticEvaluator *utils.StaticStringEvaluator, ctx rule.RuleContext) bool {
	node = skipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNumericLiteral:
		return true
	case ast.KindPropertyAccessExpression:
		if isNumberProperty(node) {
			return true
		}
	case ast.KindCallExpression:
		if isNumberCall(node, staticEvaluator, ctx) {
			return true
		}
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindPlusToken ||
			((prefix.Operator == ast.KindMinusToken || prefix.Operator == ast.KindTildeToken) && isNumber(prefix.Operand, staticEvaluator, ctx)) ||
			((prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken) && isNumber(prefix.Operand, staticEvaluator, ctx)) {
			return true
		}
	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		if (postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken) && isNumber(postfix.Operand, staticEvaluator, ctx) {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		leftNumber := isNumber(binary.Left, staticEvaluator, ctx)
		rightNumber := isNumber(binary.Right, staticEvaluator, ctx)
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			return rightNumber
		case ast.KindPlusToken:
			return leftNumber && rightNumber
		case ast.KindGreaterThanGreaterThanGreaterThanToken:
			return true
		case ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken,
			ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
			ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
			ast.KindBarToken, ast.KindCaretToken, ast.KindAmpersandToken:
			return leftNumber || rightNumber
		case ast.KindCommaToken:
			return rightNumber
		case ast.KindPlusEqualsToken:
			return leftNumber && rightNumber
		case ast.KindMinusEqualsToken, ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken, ast.KindPercentEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken, ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken, ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken, ast.KindAmpersandEqualsToken:
			return leftNumber || rightNumber
		case ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
			return true
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		trueIsNumber := isNumber(conditional.WhenTrue, staticEvaluator, ctx)
		falseIsNumber := isNumber(conditional.WhenFalse, staticEvaluator, ctx)
		if trueIsNumber && falseIsNumber {
			return true
		}
		if truthy, known := staticEvaluator.EvalControlFlowTruthiness(conditional.Condition); known {
			return (truthy && trueIsNumber) || (!truthy && falseIsNumber)
		}
	case ast.KindAsExpression:
		if isNumberType(node.AsAsExpression().Type) {
			return true
		}
	case ast.KindSatisfiesExpression:
		if isNumberType(node.AsSatisfiesExpression().Type) {
			return true
		}
	case ast.KindTypeAssertionExpression:
		if isNumberType(node.AsTypeAssertion().Type) {
			return true
		}
	case ast.KindNonNullExpression:
		if isNumber(node.AsNonNullExpression().Expression, staticEvaluator, ctx) {
			return true
		}
	case ast.KindIdentifier:
		if hasNumberTypeAnnotation(ctx, node) {
			return true
		}
	}

	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false
	}
	_, isNumber := value.(interface{ IsNaN() bool })
	return isNumber
}

func isNumberProperty(node *ast.Node) bool {
	access := node.AsPropertyAccessExpression()
	if access == nil || access.QuestionDotToken != nil || ast.IsOptionalChainRoot(node) {
		return false
	}
	if access.Name().Text() == "length" {
		return true
	}
	object := skipParentheses(access.Expression)
	if object == nil || !ast.IsIdentifier(object) {
		return false
	}
	property := access.Name().Text()
	switch object.Text() {
	case "Math":
		switch property {
		case "E", "LN2", "LN10", "LOG2E", "LOG10E", "PI", "SQRT1_2", "SQRT2":
			return true
		}
	case "Number":
		switch property {
		case "EPSILON", "MAX_SAFE_INTEGER", "MAX_VALUE", "MIN_SAFE_INTEGER", "MIN_VALUE", "NaN", "NEGATIVE_INFINITY", "POSITIVE_INFINITY":
			return true
		}
	}
	return false
}

func isNumberCall(node *ast.Node, staticEvaluator *utils.StaticStringEvaluator, ctx rule.RuleContext) bool {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil || ast.IsOptionalChainRoot(node) {
		return false
	}
	callee := skipParentheses(call.Expression)
	if ast.IsIdentifier(callee) {
		switch callee.Text() {
		case "Number", "parseInt", "parseFloat":
			return true
		}
		return false
	}
	if !ast.IsPropertyAccessExpression(callee) || ast.IsOptionalChainRoot(callee) {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil {
		return false
	}
	object := skipParentheses(access.Expression)
	if object != nil && ast.IsIdentifier(object) {
		switch object.Text() {
		case "Math":
			switch access.Name().Text() {
			case "abs", "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh", "cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor", "fround", "hypot", "imul", "log", "log1p", "log10", "log2", "max", "min", "pow", "random", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc":
				return true
			}
		case "Number":
			return access.Name().Text() == "parseInt" || access.Name().Text() == "parseFloat"
		}
	}
	if !isString(object, staticEvaluator, ctx) {
		return false
	}
	switch access.Name().Text() {
	case "charCodeAt", "codePointAt", "indexOf", "lastIndexOf", "localeCompare", "search":
		return true
	}
	return false
}

func isNumberType(typeNode *ast.TypeNode) bool {
	if typeNode == nil {
		return false
	}
	return isNumberTypeNode(typeNode.AsNode())
}

func isNumberTypeNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return false
		}
	}
	if node.Kind == ast.KindNumberKeyword {
		return true
	}
	return node.Kind == ast.KindLiteralType && node.AsLiteralTypeNode().Literal.Kind == ast.KindNumericLiteral
}

func hasNumberTypeAnnotation(ctx rule.RuleContext, node *ast.Node) bool {
	return hasTypeAnnotation(ctx, node, isNumberTypeNode)
}

func hasStringTypeAnnotation(ctx rule.RuleContext, node *ast.Node) bool {
	return hasTypeAnnotation(ctx, node, isStringTypeNode)
}

func hasTypeAnnotation(ctx rule.RuleContext, node *ast.Node, predicate func(*ast.Node) bool) bool {
	if node == nil || !ast.IsIdentifier(node) {
		return false
	}
	symbol := node.Symbol()
	if ctx.Refs != nil {
		symbol = ctx.Refs.ResolveInFile(node)
	}
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && predicate(declaration.Type()) {
			return true
		}
	}
	return false
}

func isString(node *ast.Node, staticEvaluator *utils.StaticStringEvaluator, ctx rule.RuleContext) bool {
	node = skipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	case ast.KindTypeOfExpression:
		return true
	case ast.KindAsExpression:
		if isStringType(node.AsAsExpression().Type) {
			return true
		}
	case ast.KindSatisfiesExpression:
		if isStringType(node.AsSatisfiesExpression().Type) {
			return true
		}
	case ast.KindTypeAssertionExpression:
		if isStringType(node.AsTypeAssertion().Type) {
			return true
		}
	case ast.KindIdentifier:
		if hasStringTypeAnnotation(ctx, node) {
			return true
		}
	case ast.KindCallExpression:
		if isStringConstructorCall(node) {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken != nil {
			switch binary.OperatorToken.Kind {
			case ast.KindPlusToken, ast.KindPlusEqualsToken:
				if isString(binary.Left, staticEvaluator, ctx) || isString(binary.Right, staticEvaluator, ctx) {
					return true
				}
			case ast.KindEqualsToken:
				if isString(binary.Right, staticEvaluator, ctx) {
					return true
				}
			}
		}
	}
	value, ok := staticEvaluator.EvalControlFlowValue(node)
	if !ok {
		return false
	}
	_, isString := value.(string)
	return isString
}

func isStringConstructorCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil || ast.IsOptionalChainRoot(node) {
		return false
	}
	callee := skipParentheses(call.Expression)
	if ast.IsIdentifier(callee) && callee.Text() == "String" {
		return true
	}
	if !ast.IsPropertyAccessExpression(callee) || ast.IsOptionalChainRoot(callee) {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || !ast.IsIdentifier(access.Expression) || access.Expression.Text() != "String" {
		return false
	}
	return access.Name().Text() == "fromCharCode" || access.Name().Text() == "fromCodePoint"
}

func isStringTypeNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return false
		}
	}
	if node.Kind == ast.KindStringKeyword {
		return true
	}
	return node.Kind == ast.KindLiteralType && node.AsLiteralTypeNode().Literal.Kind == ast.KindStringLiteral
}

func isStringType(typeNode *ast.TypeNode) bool {
	return typeNode != nil && isStringTypeNode(typeNode.AsNode())
}

func skipParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		parenthesized := node.AsParenthesizedExpression()
		if parenthesized == nil {
			return nil
		}
		node = parenthesized.Expression
	}
	return node
}

func fixes(ctx rule.RuleContext, node *ast.Node, newExpression *ast.NewExpression, callee *ast.Node, method string) []rule.RuleFix {
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return unicornutil.NewExpressionToCallFixes(ctx.SourceFile, node, nodeRange, newExpression, callee, "."+method)
}

func suggestions(ctx rule.RuleContext, node *ast.Node, newExpression *ast.NewExpression, callee *ast.Node) []rule.RuleSuggestion {
	return []rule.RuleSuggestion{
		{
			Message:  suggestionMessage("from"),
			FixesArr: fixes(ctx, node, newExpression, callee, "from"),
		},
		{
			Message:  suggestionMessage("alloc"),
			FixesArr: fixes(ctx, node, newExpression, callee, "alloc"),
		},
	}
}

func messageForMethod(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDError,
		Description: fmt.Sprintf("`new Buffer()` is deprecated, use `Buffer.%s()` instead.", method),
		Data:        map[string]string{"method": method},
	}
}

func messageUnknown() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDErrorUnknown,
		Description: "`new Buffer()` is deprecated, use `Buffer.alloc()` or `Buffer.from()` instead.",
	}
}

func suggestionMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDSuggestion,
		Description: fmt.Sprintf("Switch to `Buffer.%s()`.", replacement),
		Data:        map[string]string{"replacement": replacement},
	}
}
