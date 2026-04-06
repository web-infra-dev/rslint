package no_unsafe_negation

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unsafe-negation
var NoUnsafeNegationRule = rule.Rule{
	Name: "no-unsafe-negation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}

				op := bin.OperatorToken.Kind

				// Check if operator is relational (always checked)
				isRelational := op == ast.KindInKeyword || op == ast.KindInstanceOfKeyword

				// Check if operator is ordering (only checked with option)
				isOrdering := op == ast.KindLessThanToken || op == ast.KindGreaterThanToken ||
					op == ast.KindLessThanEqualsToken || op == ast.KindGreaterThanEqualsToken

				if !isRelational && (!opts.enforceForOrderingRelations || !isOrdering) {
					return
				}

				left := bin.Left
				if left == nil || left.Kind != ast.KindPrefixUnaryExpression {
					return
				}

				prefix := left.AsPrefixUnaryExpression()
				if prefix == nil || prefix.Operator != ast.KindExclamationToken {
					return
				}

				// If left is directly a PrefixUnaryExpression (not wrapped in
				// ParenthesizedExpression), the negation is not parenthesized.
				// A parenthesized `(!a)` would make left.Kind == KindParenthesizedExpression,
				// so we would not reach this point.

				operatorText := scanner.TokenToString(op)

				// Get the range of the `!` token to find where it ends
				negTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, left.Pos())
				negTokenEnd := negTokenRange.End()

				// Suggestion 1: Negate the whole expression. !a in b → !(a in b)
				afterNegText := ctx.SourceFile.Text()[negTokenEnd:node.End()]
				suggestion1 := rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          "suggestNegatedExpression",
						Description: fmt.Sprintf("Negate '%s' expression instead of its left operand. This changes the current behavior.", operatorText),
					},
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplaceRange(
							core.NewTextRange(negTokenEnd, node.End()),
							"("+afterNegText+")",
						),
					},
				}

				// Suggestion 2: Parenthesize the negation. !a in b → (!a) in b
				leftText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, left, false)
				suggestion2 := rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          "suggestParenthesisedNegation",
						Description: "Wrap negation in '()' to make the intention explicit. This preserves the current behavior.",
					},
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, left, "("+leftText+")"),
					},
				}

				ctx.ReportNodeWithSuggestions(left, rule.RuleMessage{
					Id:          "unexpected",
					Description: fmt.Sprintf("Unexpected negating the left operand of '%s' operator.", operatorText),
				}, suggestion1, suggestion2)
			},
		}
	},
}

type noUnsafeNegationOptions struct {
	enforceForOrderingRelations bool
}

func parseOptions(opts any) noUnsafeNegationOptions {
	result := noUnsafeNegationOptions{
		enforceForOrderingRelations: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if enforce, ok := optsMap["enforceForOrderingRelations"].(bool); ok {
			result.enforceForOrderingRelations = enforce
		}
	}

	return result
}
