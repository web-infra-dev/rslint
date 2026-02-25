package no_non_null_asserted_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// unwrapParens returns the inner expression after stripping parentheses.
func unwrapParens(node *ast.Node) *ast.Node {
	for node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// isPartOfOptionalChain checks if a node (after unwrapping parens) is part of an optional chain.
func isPartOfOptionalChain(node *ast.Node) bool {
	unwrapped := unwrapParens(node)
	return ast.IsOptionalChain(unwrapped)
}

var NoNonNullAssertedOptionalChainRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-asserted-optional-chain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "noNonNullOptionalChain",
			Description: "Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong.",
		}
		suggestMsg := rule.RuleMessage{
			Id:          "suggestRemovingNonNull",
			Description: "You should remove the non-null assertion.",
		}

		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.AsNonNullExpression().Expression

				// Case 1: NonNullExpression is part of an optional chain and is outermost
				// e.g., foo?.bar! or (foo?.bar!)
				if ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node) {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					fix := rule.RuleFixRemoveRange(s.TokenRange())
					ctx.ReportNodeWithSuggestions(node, msg,
						rule.RuleSuggestion{
							Message:  suggestMsg,
							FixesArr: []rule.RuleFix{fix},
						},
					)
					return
				}

				// Case 2: NonNullExpression wraps an optional chain (possibly through parens)
				// e.g., (foo?.bar)! or (foo?.bar)!.baz
				if !ast.IsOptionalChain(node) && isPartOfOptionalChain(expression) {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					fix := rule.RuleFixRemoveRange(s.TokenRange())
					ctx.ReportNodeWithSuggestions(node, msg,
						rule.RuleSuggestion{
							Message:  suggestMsg,
							FixesArr: []rule.RuleFix{fix},
						},
					)
					return
				}
			},
		}
	},
})
