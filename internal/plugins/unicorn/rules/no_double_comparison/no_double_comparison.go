// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_double_comparison

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoDoubleComparisonRule = rule.Rule{
	Name:   "unicorn/no-double-comparison",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		const wrappers = ast.OEKParentheses | ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			logical := node.AsBinaryExpression()
			join := logical.OperatorToken.Kind
			if !ast.IsLogicalBinaryOperator(join) {
				return
			}
			left, right := ast.SkipOuterExpressions(logical.Left, wrappers), ast.SkipOuterExpressions(logical.Right, wrappers)
			if !isComparison(left) || !isComparison(right) {
				return
			}
			first, second := left.AsBinaryExpression(), right.AsBinaryExpression()
			rightOperator := second.OperatorToken.Kind
			if unicornutil.IsSameComparisonOperand(first.Left, second.Left) && unicornutil.IsSameComparisonOperand(first.Right, second.Right) {
				// Keep the first comparison's operand order.
			} else if unicornutil.IsSameComparisonOperand(first.Left, second.Right) && unicornutil.IsSameComparisonOperand(first.Right, second.Left) {
				rightOperator = flipOperator(rightOperator)
			} else {
				return
			}
			operator := combinedOperator(join, first.OperatorToken.Kind, rightOperator)
			if operator == "" {
				return
			}
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			leftRegExp, _ := evaluator.EvalControlFlowRegExpValue(first.Left)
			rightRegExp, _ := evaluator.EvalControlFlowRegExpValue(first.Right)
			if leftRegExp || rightRegExp {
				return
			}

			leftRange := utils.TrimNodeTextRange(ctx.SourceFile, first.Left)
			rightRange := utils.TrimNodeTextRange(ctx.SourceFile, first.Right)
			leftText := ctx.SourceFile.Text()[leftRange.Pos():leftRange.End()]
			rightText := ctx.SourceFile.Text()[rightRange.Pos():rightRange.End()]
			// A relational operator otherwise captures an unparenthesized assertion.
			if operator != "===" && operator != "!==" &&
				(first.Right.Kind == ast.KindAsExpression || first.Right.Kind == ast.KindSatisfiesExpression) {
				rightText = "(" + rightText + ")"
			}
			replacement := leftText + " " + operator + " " + rightText
			ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
				Id: "no-double-comparison", Description: "These comparisons can be simplified to `" + replacement + "`.",
				Data: map[string]string{"replacement": replacement},
			}, func() []rule.RuleSuggestion {
				span := utils.TrimNodeTextRange(ctx.SourceFile, node)
				// Only the two first-comparison operands survive the replacement.
				comments := ctx.Comments.All()
				if utils.HasCommentInSpan(comments, span.Pos(), leftRange.Pos()) ||
					utils.HasCommentInSpan(comments, leftRange.End(), rightRange.Pos()) ||
					utils.HasCommentInSpan(comments, rightRange.End(), span.End()) {
					return nil
				}
				text := replacement
				if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, text) {
					text = ";" + text
				}
				return []rule.RuleSuggestion{{
					Message:  rule.RuleMessage{Id: "no-double-comparison/suggestion", Description: "Replace with `" + replacement + "`.", Data: map[string]string{"replacement": replacement}},
					FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, text)},
				}}
			})
		}}
	},
}

func isComparison(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	switch node.AsBinaryExpression().OperatorToken.Kind {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return !containsOptionalChain(node)
	}
	return false
}

func containsOptionalChain(node *ast.Node) bool {
	return ast.IsOptionalChain(node) || node.ForEachChild(containsOptionalChain)
}

func flipOperator(operator ast.Kind) ast.Kind {
	switch operator {
	case ast.KindLessThanToken:
		return ast.KindGreaterThanToken
	case ast.KindLessThanEqualsToken:
		return ast.KindGreaterThanEqualsToken
	case ast.KindGreaterThanToken:
		return ast.KindLessThanToken
	case ast.KindGreaterThanEqualsToken:
		return ast.KindLessThanEqualsToken
	}
	return operator
}

func combinedOperator(join, first, second ast.Kind) string {
	// Normalize equality first; the remaining relational pairs are symmetric.
	if second == ast.KindEqualsEqualsEqualsToken || second == ast.KindExclamationEqualsEqualsToken {
		first, second = second, first
	}
	if join == ast.KindBarBarToken {
		switch {
		case first == ast.KindEqualsEqualsEqualsToken && second == ast.KindLessThanToken:
			return "<="
		case first == ast.KindEqualsEqualsEqualsToken && second == ast.KindGreaterThanToken:
			return ">="
		case first == ast.KindLessThanToken && second == ast.KindGreaterThanToken,
			first == ast.KindGreaterThanToken && second == ast.KindLessThanToken:
			return "!=="
		}
	} else {
		switch {
		case first == ast.KindLessThanEqualsToken && second == ast.KindGreaterThanEqualsToken,
			first == ast.KindGreaterThanEqualsToken && second == ast.KindLessThanEqualsToken:
			return "==="
		case first == ast.KindExclamationEqualsEqualsToken && second == ast.KindLessThanEqualsToken:
			return "<"
		case first == ast.KindExclamationEqualsEqualsToken && second == ast.KindGreaterThanEqualsToken:
			return ">"
		}
	}
	return ""
}
