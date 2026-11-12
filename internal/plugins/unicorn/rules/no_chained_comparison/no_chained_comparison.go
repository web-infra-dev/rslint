// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_chained_comparison

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoChainedComparisonRule = rule.Rule{
	Name:   "unicorn/no-chained-comparison",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		booleans := unicornutil.NewBooleanAnalyzer(&ctx)
		return rule.RuleListeners{ast.KindBinaryExpression: func(node *ast.Node) {
			if !isComparison(node) {
				return
			}
			binary := node.AsBinaryExpression()
			left, right := utils.ESTreeRuntimeExpression(binary.Left), utils.ESTreeRuntimeExpression(binary.Right)
			nested, sibling := left, binary.Right
			if !isComparison(nested) {
				nested, sibling = right, binary.Left
			}
			if !isComparison(nested) {
				return
			}
			operator := binary.OperatorToken.Kind
			if !isOrdering(operator) && booleans.IsExpression(sibling) {
				return
			}
			ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
				Id:          "no-chained-comparison/error",
				Description: "Comparison operators cannot be chained. The inner comparison evaluates to a boolean, which is then compared instead of the operands.",
			}, func() []rule.RuleSuggestion {
				inner := nested.AsBinaryExpression()
				// The shared side-effect check also rejects tag calls, which
				// Unicorn v79's eslint-utils check overlooks (see the rule docs).
				if nested != left || isOrdering(operator) != isOrdering(inner.OperatorToken.Kind) ||
					right.Kind == ast.KindTrueKeyword || right.Kind == ast.KindFalseKeyword ||
					isComparison(right) || isComparison(utils.ESTreeRuntimeExpression(inner.Left)) ||
					!safeLogicalReplacement(node) || booleans.StaticEvaluator().HasSideEffect(inner.Right, false) {
					return nil
				}
				// Only the first and last operand are preserved once. Comments on
				// the middle operand would be duplicated; other comments may be lost.
				firstRange := utils.TrimNodeTextRange(ctx.SourceFile, inner.Left)
				lastRange := utils.TrimNodeTextRange(ctx.SourceFile, binary.Right)
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				comments := ctx.Comments.All()
				if utils.HasCommentInSpan(comments, nodeRange.Pos(), firstRange.Pos()) ||
					utils.HasCommentInSpan(comments, firstRange.End(), lastRange.Pos()) ||
					utils.HasCommentInSpan(comments, lastRange.End(), nodeRange.End()) {
					return nil
				}
				replacement := utils.TrimmedNodeText(ctx.SourceFile, left) + " && " +
					utils.TrimmedNodeText(ctx.SourceFile, inner.Right) + " " +
					utils.TrimmedNodeText(ctx.SourceFile, binary.OperatorToken) + " " +
					utils.TrimmedNodeText(ctx.SourceFile, binary.Right)
				if needsStartParentheses(ctx, node, left) {
					replacement = "(" + replacement + ")"
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "no-chained-comparison/suggestion",
						Description: "Replace with `" + replacement + "`.",
						Data:        map[string]string{"replacement": replacement},
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
				}}
			})
		}}
	},
}

func isOrdering(operator ast.Kind) bool {
	return ast.IsRelationalOperator(operator) && operator != ast.KindInKeyword && operator != ast.KindInstanceOfKeyword
}

func isComparison(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken.Kind
	return isOrdering(operator) || ast.IsEqualityOperator(operator)
}

func safeLogicalReplacement(node *ast.Node) bool {
	if ast.IsParenthesizedExpression(node.Parent) {
		return true
	}
	parent := utils.ESTreeParent(node)
	switch parent.Kind {
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		return false
	case ast.KindBinaryExpression:
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		// tsgo also represents logical, assignment and sequence expressions as
		// binary nodes. Only &&, ||, assignments and sequences are safe here.
		return operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken ||
			operator == ast.KindCommaToken || ast.IsAssignmentOperator(operator)
	}
	return true
}

// Removing the inner comparison's parentheses can expose an object, function
// or class at the start of a statement, default export or arrow body.
func needsStartParentheses(ctx rule.RuleContext, node, left *ast.Node) bool {
	start := utils.TrimNodeTextRange(ctx.SourceFile, left).Pos()
	first, _ := utils.TokenAtOrAfter(ctx.SourceFile, start)
	next, _ := utils.TokenAtOrAfter(ctx.SourceFile, first.End)
	if first.Text != "{" && first.Text != "class" && first.Text != "function" &&
		(first.Text != "async" || next.Text != "function") && (first.Text != "let" || next.Text != "[") {
		return false
	}
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			return false
		}
		if parent.Kind == ast.KindArrowFunction && parent.Body() == current {
			return first.Text == "{"
		}
		if parent.Kind == ast.KindExportAssignment {
			return first.Text != "{" && first.Text != "let"
		}
		if parent.Kind == ast.KindForStatement && parent.AsForStatement().Initializer == current {
			return first.Text == "let"
		}
		if current.Kind == ast.KindExpressionStatement {
			return true
		}
		if utils.TrimNodeTextRange(ctx.SourceFile, current).Pos() != utils.TrimNodeTextRange(ctx.SourceFile, parent).Pos() {
			return false
		}
	}
	return false
}
