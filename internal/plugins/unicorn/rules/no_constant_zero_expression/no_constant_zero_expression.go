// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_constant_zero_expression

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoConstantZeroExpressionRule = rule.Rule{
	Name:   "unicorn/no-constant-zero-expression",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindAsteriskToken, ast.KindAmpersandToken:
				if !isZero(binary.Left) && !isZero(binary.Right) {
					return
				}
			case ast.KindSlashToken:
				if !isZero(binary.Left) {
					return
				}
			default:
				return
			}

			span := utils.TrimNodeTextRange(ctx.SourceFile, node)
			ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
				Id:          "no-constant-zero-expression/error",
				Description: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.",
			}, func() []rule.RuleSuggestion {
				if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
					return nil
				}
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}
				value, known := evaluator.EvalControlFlowValue(node)
				number, numeric := value.(jsnum.Number)
				// Object.is(value, 0) rejects negative zero as well as NaN.
				if !known || !numeric || number != 0 || math.Signbit(float64(number)) {
					return nil
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "no-constant-zero-expression/suggestion",
						Description: "Replace this expression with `0`.",
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(span, "0")},
				}}
			})
		}}
	},
}

func isZero(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && node.Kind == ast.KindNumericLiteral && jsnum.FromString(node.Text()) == 0
}
