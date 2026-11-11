// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_misrefactored_assignment

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoMisrefactoredAssignmentRule = rule.Rule{
	Name:   "unicorn/no-misrefactored-assignment",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var keys unicornutil.ComputedKeyEvaluator
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			assignment := node.AsBinaryExpression()
			operator := assignment.OperatorToken.Kind
			if !ast.IsCompoundAssignment(operator) || ast.IsLogicalOrCoalescingAssignmentOperator(operator) {
				return
			}
			right := utils.ESTreeRuntimeExpression(assignment.Right)
			if right == nil || !ast.IsBinaryExpression(right) {
				return
			}
			binary := right.AsBinaryExpression()
			if scanner.TokenToString(binary.OperatorToken.Kind) != strings.TrimSuffix(scanner.TokenToString(operator), "=") {
				return
			}

			left := utils.ESTreeRuntimeExpression(assignment.Left)
			var kept *ast.Node
			if unicornutil.IsSameComparisonOperandWithKeyEvaluator(left, binary.Left, keys.Eval) {
				kept = binary.Right
			} else if matchesEitherOperand(operator) && unicornutil.IsSameComparisonOperandWithKeyEvaluator(left, binary.Right, keys.Eval) {
				kept = binary.Left
			} else {
				return
			}

			target := utils.TrimmedNodeText(ctx.SourceFile, left)
			data := map[string]string{"target": target}
			ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
				Id:          "no-misrefactored-assignment/error",
				Description: "`" + target + "` is duplicated in the right-hand side, which is likely a refactoring mistake.",
				Data:        data,
			}, func() []rule.RuleSuggestion {
				rightRange := utils.TrimNodeTextRange(ctx.SourceFile, right)
				keptRange := utils.TrimNodeTextRange(ctx.SourceFile, operandToCopy(kept))
				comments := ctx.Comments.All()
				if utils.HasCommentInSpan(comments, rightRange.Pos(), keptRange.Pos()) ||
					utils.HasCommentInSpan(comments, keptRange.End(), rightRange.End()) {
					return nil
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "no-misrefactored-assignment/suggestion",
						Description: "Remove the duplicated `" + target + "`.",
						Data:        data,
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(rightRange, ctx.SourceFile.Text()[keptRange.Pos():keptRange.End()])},
				}}
			})
		}}
	},
}

func matchesEitherOperand(operator ast.Kind) bool {
	switch operator {
	case ast.KindPlusEqualsToken, ast.KindAsteriskEqualsToken, ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken:
		return true
	}
	return false
}

// Unicorn retains the authored parentheses of operands that bind less tightly
// than addition.
func operandToCopy(operand *ast.Node) *ast.Node {
	expression := utils.ESTreeRuntimeExpression(operand)
	precedence := ast.GetExpressionPrecedence(expression)
	if precedence != ast.OperatorPrecedenceInvalid && precedence < ast.OperatorPrecedenceAdditive {
		return operand
	}
	return expression
}
