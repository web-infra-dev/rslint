package prefer_math_min_max

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageID    = "prefer-math-min-max"
	suggestionID = "replace-with-math-min-max"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-min-max.js
var PreferMathMinMaxRule = rule.Rule{
	Name:   "unicorn/prefer-math-min-max",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				conditional := node.AsConditionalExpression()
				test := utils.ESTreeRuntimeExpression(conditional.Condition)
				if !ast.IsBinaryExpression(test) {
					return
				}
				comparison := test.AsBinaryExpression()
				method := mathMethod(ctx.SourceFile, comparison, conditional)
				if method == "" {
					return
				}

				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}
				left, right := comparison.Left, comparison.Right
				if skipOperand(&ctx, evaluator, left) || skipOperand(&ctx, evaluator, right) {
					return
				}

				span := node.Loc.WithPos(scanner.GetTokenPosOfNode(node, ctx.SourceFile, false))
				data := map[string]string{"method": method}
				ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
					Id:          messageID,
					Description: "Prefer `Math." + method + "()` to simplify ternary expressions.",
					Data:        data,
				}, func() []rule.RuleSuggestion {
					fixes := mathCallFixes(&ctx, evaluator, node, span, left, right, method)
					if len(fixes) == 0 {
						return nil
					}
					return []rule.RuleSuggestion{{
						Message: rule.RuleMessage{
							Id:          suggestionID,
							Description: "Replace with `Math." + method + "()`.",
							Data:        data,
						},
						FixesArr: fixes,
					}}
				})
			},
		}
	},
}

func mathMethod(file *ast.SourceFile, comparison *ast.BinaryExpression, conditional *ast.ConditionalExpression) string {
	var whenOrdered, whenReversed string
	switch comparison.OperatorToken.Kind {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken:
		whenOrdered, whenReversed = "min", "max"
	case ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		whenOrdered, whenReversed = "max", "min"
	default:
		return ""
	}

	// Upstream compares source spelling after unwrapping casts. Keep distinct
	// spellings such as obj.key and obj["key"] distinct.
	source := file.Text()
	left := scanner.GetTextOfNodeFromSourceText(source, unwrapOperand(comparison.Left), false)
	right := scanner.GetTextOfNodeFromSourceText(source, unwrapOperand(comparison.Right), false)
	consequent := scanner.GetTextOfNodeFromSourceText(source, unwrapOperand(conditional.WhenTrue), false)
	alternate := scanner.GetTextOfNodeFromSourceText(source, unwrapOperand(conditional.WhenFalse), false)
	if left == consequent && right == alternate {
		// When all four expressions match, upstream prefers min.
		if left == right {
			return "min"
		}
		return whenOrdered
	}
	if right == consequent && left == alternate {
		return whenReversed
	}
	return ""
}

func unwrapOperand(node *ast.Node) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKTypeAssertions|ast.OEKNonNullAssertions)
}
