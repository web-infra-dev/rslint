package no_extra_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoExtraNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noExtraNonNullAssertion",
		Description: "Forbidden extra non-null assertion.",
	}
}

var NoExtraNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-extra-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.Expression()

				// Check for double non-null assertion: foo!!
				if ast.IsNonNullExpression(expression) {
					// Report the outer non-null assertion
					// Get trimmed ranges for consistent replacement
					outerRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
					innerRange := utils.TrimNodeTextRange(ctx.SourceFile, expression)
					expressionText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

					// Create error range that includes the full expression including both ! characters
					errorRange := outerRange.WithEnd(node.End())

					ctx.ReportRangeWithFixes(
						errorRange,
						buildNoExtraNonNullAssertionMessage(),
						// Fix: replace the outer expression with the inner expression
						rule.RuleFixReplaceRange(outerRange, expressionText),
					)
					return
				}

				// Check for non-null assertion before optional chaining: foo!?.bar or foo!?.()
				// The parent of the NonNullExpression should be a PropertyAccessExpression, CallExpression,
				// or ElementAccessExpression with a QuestionDotToken, and the NonNullExpression must be
				// the Expression (left side) of that parent, not an argument
				if expression != nil {
					parent := node.Parent
					if parent != nil {
						// Check if parent has QuestionDotToken AND this node is the Expression (not ArgumentExpression)
						hasQuestionDot := false

						// For property access: obj!?.prop
						if ast.IsPropertyAccessExpression(parent) {
							propAccess := parent.AsPropertyAccessExpression()
							if propAccess != nil && propAccess.QuestionDotToken != nil && propAccess.Expression == node {
								hasQuestionDot = true
							}
						}

						// For call expression: obj!?.()
						if ast.IsCallExpression(parent) {
							callExpr := parent.AsCallExpression()
							if callExpr != nil && callExpr.QuestionDotToken != nil && callExpr.Expression == node {
								hasQuestionDot = true
							}
						}

						// For element access: obj!?.[prop]
						if ast.IsElementAccessExpression(parent) {
							elemAccess := parent.AsElementAccessExpression()
							if elemAccess != nil && elemAccess.QuestionDotToken != nil && elemAccess.Expression == node {
								hasQuestionDot = true
							}
						}

						if hasQuestionDot {
							// Report the non-null assertion as unnecessary
							// Get trimmed ranges for consistent replacement
							outerRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
							innerRange := utils.TrimNodeTextRange(ctx.SourceFile, expression)
							expressionText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

							// Create error range that includes the full expression including the ! character
							errorRange := outerRange.WithEnd(node.End())

							ctx.ReportRangeWithFixes(
								errorRange,
								buildNoExtraNonNullAssertionMessage(),
								// Fix: remove the non-null assertion, keeping just the expression
								rule.RuleFixReplaceRange(outerRange, expressionText),
							)
						}
					}
				}
			},
		}
	},
})
