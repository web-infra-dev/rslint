// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package prefer_math_abs

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-abs.js
var PreferMathAbsRule = rule.Rule{
	Name:   "unicorn/prefer-math-abs",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		isBigInt := func(node *ast.Node) bool {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			return canBeBigInt(&ctx, evaluator, node)
		}
		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				if value := ternaryOperand(node.AsConditionalExpression()); value != nil && !isBigInt(value) {
					report(&ctx, node, value, nil, ast.KindUnknown)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind != ast.KindBarBarToken {
					return
				}
				positive, negative := rangeComparisons(binary)
				if positive.greater != nil && !isBigInt(positive.greater) && !isBigInt(positive.lesser) &&
					!isBigInt(negative.greater) && !isBigInt(negative.lesser) {
					report(&ctx, node, positive.greater, positive.lesser, positive.operator)
				}
			},
		}
	},
}

func orderingComparison(node *ast.Node) *ast.BinaryExpression {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsBinaryExpression(node) {
		return nil
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		return binary
	}
	return nil
}

func ternaryOperand(node *ast.ConditionalExpression) *ast.Node {
	binary := orderingComparison(node.Condition)
	if binary == nil {
		return nil
	}
	value, operator := binary.Left, binary.OperatorToken.Kind
	if !isNumber(binary.Right, "0") {
		if !isNumber(binary.Left, "0") {
			return nil
		}
		value, operator = binary.Right, unicornutil.FlipComparisonOperator(operator)
	}
	unchanged, negated := node.WhenTrue, node.WhenFalse
	if operator == ast.KindLessThanToken || operator == ast.KindLessThanEqualsToken {
		unchanged, negated = negated, unchanged
	}
	if isNegationOf(negated, value) && unicornutil.IsSameComparisonOperand(unchanged, value) {
		return value
	}
	return nil
}

// comparison expresses either operand order using > or >=.
type comparison struct {
	greater, lesser *ast.Node
	operator        ast.Kind
}

func greaterComparison(binary *ast.BinaryExpression) comparison {
	operator := binary.OperatorToken.Kind
	if operator == ast.KindGreaterThanToken || operator == ast.KindGreaterThanEqualsToken {
		return comparison{binary.Left, binary.Right, operator}
	}
	return comparison{binary.Right, binary.Left, unicornutil.FlipComparisonOperator(operator)}
}

func rangeComparisons(binary *ast.BinaryExpression) (positive, negative comparison) {
	left, right := orderingComparison(binary.Left), orderingComparison(binary.Right)
	if left == nil || right == nil {
		return comparison{}, comparison{}
	}
	first, second := greaterComparison(left), greaterComparison(right)
	if first.operator == second.operator {
		if isOutsideRange(first, second) {
			return first, second
		}
		if isOutsideRange(second, first) {
			return second, first
		}
	}
	return comparison{}, comparison{}
}

func isOutsideRange(positive, negative comparison) bool {
	return isNegationOf(negative.greater, positive.lesser) &&
		unicornutil.IsSameComparisonOperand(positive.greater, negative.lesser)
}

func isNegationOf(node, value *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if ast.IsPrefixUnaryExpression(node) {
		unary := node.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindMinusToken && unicornutil.IsSameComparisonOperand(unary.Operand, value)
	}
	if !ast.IsBinaryExpression(node) {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindMinusToken:
		return isNumber(binary.Left, "0") && unicornutil.IsSameComparisonOperand(binary.Right, value)
	case ast.KindAsteriskToken:
		return (isNegativeOne(binary.Left) && unicornutil.IsSameComparisonOperand(binary.Right, value)) ||
			(isNegativeOne(binary.Right) && unicornutil.IsSameComparisonOperand(binary.Left, value))
	}
	return false
}

func isNegativeOne(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsPrefixUnaryExpression(node) && node.AsPrefixUnaryExpression().Operator == ast.KindMinusToken &&
		isNumber(node.AsPrefixUnaryExpression().Operand, "1")
}

func isNumber(node *ast.Node, value string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	// tsgo normalizes numeric token text to the parsed JavaScript Number value.
	return node != nil && ast.IsNumericLiteral(node) && node.Text() == value
}

func canBeBigInt(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	if isBigIntExpression(ctx, evaluator, node) {
		return true
	}
	// A let/var initializer or parameter default can supply a BigInt even
	// when the expression cannot be statically evaluated at the use site.
	symbol := ctx.Refs.ResolveInFile(utils.SkipAssertionsAndParens(node))
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if (ast.IsVariableDeclaration(declaration) || ast.IsParameterDeclaration(declaration)) &&
				ast.IsIdentifier(declaration.Name()) &&
				(mentionsBigInt(declaration.Type()) || isBigIntExpression(ctx, evaluator, declaration.Initializer())) {
				return true
			}
		}
	}
	return false
}

func isBigIntExpression(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	if node == nil {
		return false
	}
	const outer = ast.OEKParentheses | ast.OEKNonNullAssertions
	assertion := ast.SkipOuterExpressions(node, outer)
	for ast.IsPrefixUnaryExpression(assertion) && assertion.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		assertion = ast.SkipOuterExpressions(assertion.AsPrefixUnaryExpression().Operand, outer)
	}
	if (ast.IsAssertionExpression(assertion) || ast.IsSatisfiesExpression(assertion)) && mentionsBigInt(assertion.Type()) {
		return true
	}
	return unicornutil.IsBigInt(ctx, evaluator, node)
}

// An authored union or intersection can mention bigint even when the checker
// simplifies the whole type to any. Math.abs can still throw in that case.
func mentionsBigInt(annotation *ast.Node) bool {
	if annotation == nil {
		return false
	}
	annotation = ast.SkipTypeParentheses(annotation)
	switch annotation.Kind {
	case ast.KindBigIntKeyword:
		return true
	case ast.KindLiteralType:
		return ast.IsBigIntLiteral(annotation.AsLiteralTypeNode().Literal)
	case ast.KindUnionType, ast.KindIntersectionType:
		return annotation.ForEachChild(mentionsBigInt)
	}
	return false
}

func report(ctx *rule.RuleContext, node, value, threshold *ast.Node, operator ast.Kind) {
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
		Id:          "prefer-math-abs",
		Description: "Prefer `Math.abs()` to simplify this expression.",
	}, func() []rule.RuleSuggestion {
		if !isSimpleReference(value) || (threshold != nil && !isSimpleReference(threshold)) {
			return nil
		}
		// Keep the diagnostic when the replacement would use a different Math.
		if !ctx.Globals.Access("Math").IsDeclared() ||
			!ctx.Refs.IsGlobalNameReference(node, "Math", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
			return nil
		}

		source := ctx.SourceFile.Text()
		valueSpan := utils.TrimNodeTextRange(ctx.SourceFile, value)
		var thresholdSpan core.TextRange
		if threshold != nil {
			thresholdSpan = utils.TrimNodeTextRange(ctx.SourceFile, threshold)
		}
		for _, comment := range utils.CommentsInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
			if !comment.ContainedBy(valueSpan) && !comment.ContainedBy(thresholdSpan) {
				return nil
			}
		}

		valueText := source[valueSpan.Pos():valueSpan.End()]
		var replacement string
		if threshold == nil {
			replacement = "Math.abs(" + valueText + ")"
		} else {
			text := source[thresholdSpan.Pos():thresholdSpan.End()]
			if ast.IsAsExpression(threshold) || ast.IsSatisfiesExpression(threshold) {
				text = "(" + text + ")"
			}
			replacement = "Math.abs(" + valueText + ") " + scanner.TokenToString(operator) + " " + text
		}
		// Math.abs coerces the operand and normalizes signed zero, so offer
		// this rewrite as a suggestion.
		fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
		return []rule.RuleSuggestion{{
			Message: rule.RuleMessage{
				Id:          "prefer-math-abs/suggestion",
				Description: "Replace with `Math.abs()`.",
			},
			FixesArr: append(fixes, rule.RuleFixReplaceRange(span, replacement)),
		}}
	})
}

func isSimpleReference(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	return node != nil && (ast.IsIdentifier(node) || node.Kind == ast.KindThisKeyword || utils.IsESTreeLiteralKind(node.Kind))
}
