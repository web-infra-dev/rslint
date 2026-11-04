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
				if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) || !withinEvaluationBudget(ctx, node) {
					return nil
				}
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}
				value, known := evaluator.EvalControlFlowValue(node)
				number, numeric := value.(jsnum.Number)
				// Object.is(value, 0) rejects negative zero as well as NaN.
				if !known || !numeric || number != 0 || math.Signbit(float64(number)) ||
					hasUnknownDiscardedOperand(node, evaluator) {
					return nil
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "no-constant-zero-expression/suggestion",
						Description: "Replace this expression with `0`.",
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(span, utils.SafeReplacementText(ctx.SourceFile, node, "0"))},
				}}
			})
		}}
	},
}

func isZero(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && node.Kind == ast.KindNumericLiteral && jsnum.FromString(node.Text()) == 0
}

// Count repeated constant expansion before entering the shared safety walks.
// Exhausting this per-expression budget only suppresses the optional suggestion.
func withinEvaluationBudget(ctx rule.RuleContext, node *ast.Node) bool {
	remaining := 4096
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return true
		}
		remaining--
		if remaining < 0 {
			return false
		}
		if ast.IsTypeNode(node) || ast.IsFunctionLike(node) {
			return true
		}
		if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
			if ctx.Refs == nil {
				return false
			}
			if symbol := ctx.Refs.ResolveInFile(node); symbol != nil && len(symbol.Declarations) == 1 {
				declaration := symbol.Declarations[0]
				if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && ast.IsVarConst(declaration.Parent) {
					return visit(declaration.AsVariableDeclaration().Initializer)
				}
			}
		}
		return !node.ForEachChild(func(child *ast.Node) bool { return !visit(child) })
	}
	return visit(node)
}

// A folded result does not prove that discarded operands complete normally.
// Keep the diagnostic but withhold a suggestion when their values are unknown.
func hasUnknownDiscardedOperand(node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) || ast.IsFunctionLike(node) {
		return false
	}
	var discarded *ast.Node
	if node.Kind == ast.KindVoidExpression {
		discarded = node.Expression()
	} else if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		discarded = node.AsBinaryExpression().Left
	}
	if discarded != nil {
		if _, known := evaluator.EvalValue(discarded); !known {
			return true
		}
	}
	return node.ForEachChild(func(child *ast.Node) bool { return hasUnknownDiscardedOperand(child, evaluator) })
}
