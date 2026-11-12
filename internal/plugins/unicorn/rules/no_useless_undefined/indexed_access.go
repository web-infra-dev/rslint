package no_useless_undefined

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var useIndexedAccess = rule.RuleMessage{
	Id:          "no-useless-undefined/suggestion",
	Description: "Use the indexed access directly.",
}

func indexedAccessListener(ctx rule.RuleContext) func(*ast.Node) {
	var effects *utils.StaticStringEvaluator
	return func(node *ast.Node) {
		conditional := node.AsConditionalExpression()
		consequent := utils.ESTreeRuntimeExpression(conditional.WhenTrue)
		alternate := utils.ESTreeRuntimeExpression(conditional.WhenFalse)
		trueIsUndefined := utils.IsUndefinedIdentifier(consequent)
		if trueIsUndefined == utils.IsUndefinedIdentifier(alternate) {
			return
		}
		branch, undefined := consequent, alternate
		if trueIsUndefined {
			branch, undefined = alternate, consequent
		}
		access := unwrapTypeScript(branch)
		if access == nil || !ast.IsElementAccessExpression(access) {
			return
		}
		element := access.AsElementAccessExpression()
		test := unwrapTypeScript(conditional.Condition)
		matchesWhenTrue, matched := indexGuard(test, element)
		if !matched || matchesWhenTrue == trueIsUndefined ||
			unicornutil.ContainsOptionalChain(access) || unicornutil.ContainsOptionalChain(test) {
			return
		}
		if effects == nil {
			effects = utils.NewStaticStringEvaluatorWithoutScope()
		}
		if effects.HasSideEffect(element.Expression, true) || effects.HasSideEffect(element.ArgumentExpression, true) ||
			!unicornutil.IsGlobalReference(ctx, undefined) {
			return
		}
		ctx.ReportNodeWithDeferredSuggestions(undefined, uselessUndefined, func() []rule.RuleSuggestion {
			source := ctx.SourceFile
			span := utils.TrimNodeTextRange(source, node)
			if unicornutil.IsReferenceContext(node) || utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
				return nil
			}
			// Keep TypeScript assertions on the selected branch.
			replacement := utils.TrimmedNodeText(source, branch)
			if unicornutil.NeedsSemicolonBefore(source, node, replacement) {
				replacement = ";" + replacement
			}
			fixes := []rule.RuleFix{rule.RuleFixReplaceRange(span, replacement)}
			fixes = append(fixes, unicornutil.SpaceAroundKeywordFixes(source, node)...)
			return []rule.RuleSuggestion{{Message: useIndexedAccess, FixesArr: fixes}}
		})
	}
}

// indexGuard reports which outcome of the test passes a recognized bound.
func indexGuard(test *ast.Node, access *ast.ElementAccessExpression) (whenTrue, matched bool) {
	if test == nil || !ast.IsBinaryExpression(test) {
		return false, false
	}
	comparison := test.AsBinaryExpression()
	operator := comparison.OperatorToken.Kind
	switch operator {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
	default:
		return false, false
	}
	index, offset := unwrapTypeScript(access.ArgumentExpression), float64(0)
	if subtraction := subtractionExpression(index); subtraction != nil {
		if value, ok := safeInteger(subtraction.Right); ok && value > 0 {
			index, offset = subtraction.Left, value
		}
	}

	bound := comparison.Right
	if !unicornutil.IsSameComparisonOperand(comparison.Left, index) {
		if !unicornutil.IsSameComparisonOperand(comparison.Right, index) {
			return false, false
		}
		bound = comparison.Left
		operator = unicornutil.FlipComparisonOperator(operator)
	}

	// An index of i - offset is nonnegative at i >= offset (or i > offset - 1).
	if value, ok := safeInteger(bound); ok {
		switch operator {
		case ast.KindGreaterThanEqualsToken, ast.KindLessThanToken:
			return operator == ast.KindGreaterThanEqualsToken, value == offset
		case ast.KindGreaterThanToken, ast.KindLessThanEqualsToken:
			return operator == ast.KindGreaterThanToken, value == offset-1
		}
	}
	if offset != 0 {
		return false, false
	}
	// An index without an offset is below the upper bound at i < length (or i <= length - 1).
	switch operator {
	case ast.KindLessThanToken, ast.KindGreaterThanEqualsToken:
		return operator == ast.KindLessThanToken, isLengthOf(bound, access.Expression)
	case ast.KindLessThanEqualsToken, ast.KindGreaterThanToken:
		if subtraction := subtractionExpression(bound); subtraction != nil {
			if value, ok := safeInteger(subtraction.Right); ok && value == 1 {
				return operator == ast.KindLessThanEqualsToken, isLengthOf(subtraction.Left, access.Expression)
			}
		}
	}
	return false, false
}

func isLengthOf(node, object *ast.Node) bool {
	node = unwrapTypeScript(node)
	return node != nil && ast.IsPropertyAccessExpression(node) && ast.IsIdentifier(node.Name()) && node.Name().Text() == "length" &&
		unicornutil.IsSameComparisonOperand(node.Expression(), object)
}

func subtractionExpression(node *ast.Node) *ast.BinaryExpression {
	node = unwrapTypeScript(node)
	if node != nil && ast.IsBinaryExpression(node) && node.AsBinaryExpression().OperatorToken.Kind == ast.KindMinusToken {
		return node.AsBinaryExpression()
	}
	return nil
}

func safeInteger(node *ast.Node) (float64, bool) {
	node = unwrapTypeScript(node)
	if node != nil && ast.IsPrefixUnaryExpression(node) {
		prefix := node.AsPrefixUnaryExpression()
		switch prefix.Operator {
		case ast.KindPlusToken:
			return safeInteger(prefix.Operand)
		case ast.KindMinusToken:
			value, ok := safeInteger(prefix.Operand)
			return -value, ok
		}
	}
	return unicornutil.SafeIntegerLiteral(node)
}

func unwrapTypeScript(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
}
