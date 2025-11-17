package no_non_null_asserted_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoNonNullOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNullOptionalChain",
		Description: "Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong.",
	}
}

func buildSuggestRemovingNonNullMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemovingNonNull",
		Description: "Remove the non-null assertion operator.",
	}
}

// Helper function to check if a node contains an optional chain
func containsOptionalChain(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if the current node is an optional chain
	if ast.IsPropertyAccessExpression(node) && node.AsPropertyAccessExpression().QuestionDotToken != nil {
		return true
	}
	if ast.IsElementAccessExpression(node) && node.AsElementAccessExpression().QuestionDotToken != nil {
		return true
	}
	if ast.IsCallExpression(node) && node.AsCallExpression().QuestionDotToken != nil {
		return true
	}

	// Check child nodes recursively
	if ast.IsPropertyAccessExpression(node) {
		return containsOptionalChain(node.AsPropertyAccessExpression().Expression)
	}
	if ast.IsElementAccessExpression(node) {
		return containsOptionalChain(node.AsElementAccessExpression().Expression)
	}
	if ast.IsCallExpression(node) {
		return containsOptionalChain(node.AsCallExpression().Expression)
	}
	if ast.IsParenthesizedExpression(node) {
		return containsOptionalChain(node.AsParenthesizedExpression().Expression)
	}

	return false
}

var NoNonNullAssertedOptionalChainRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-asserted-optional-chain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				if node.Kind != ast.KindNonNullExpression {
					return
				}

				nonNullExpr := node.AsNonNullExpression()
				expression := nonNullExpr.Expression

				// Check if the expression or any of its ancestors contains an optional chain
				if !containsOptionalChain(expression) {
					return
				}

				// Get the position of the ! token (the exclamation mark itself)
				// The ! token is at the end of the NonNullExpression node
				exprEnd := utils.TrimNodeTextRange(ctx.SourceFile, expression).End()
				nonNullEnd := utils.TrimNodeTextRange(ctx.SourceFile, node).End()

				// Create a fix that removes the ! token
				fix := rule.RuleFixRemoveRange(core.NewTextRange(exprEnd, nonNullEnd))

				ctx.ReportNodeWithSuggestions(
					node,
					buildNoNonNullOptionalChainMessage(),
					rule.RuleSuggestion{
						Message:  buildSuggestRemovingNonNullMessage(),
						FixesArr: []rule.RuleFix{fix},
					},
				)
			},
		}
	},
})
