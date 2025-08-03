package no_non_null_asserted_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoNonNullAssertedOptionalChainRule = rule.Rule{
	Name: "no-non-null-asserted-optional-chain",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				nonNullExpr := node.AsNonNullExpression()
				expression := nonNullExpr.Expression

				// Check if we're applying non-null assertion directly to an optional chain result
				if isDirectOptionalChainAssertion(expression, node) {
					reportError(ctx, node, node)
				}
			},
		}
	},
}

// isDirectOptionalChainAssertion checks if we're applying non-null assertion directly to an optional chain result
func isDirectOptionalChainAssertion(expression *ast.Node, nonNullNode *ast.Node) bool {
	if expression == nil {
		return false
	}

	// Handle the two main patterns from TypeScript-ESLint:

	// Pattern 1: NonNullExpression > ChainExpression (parenthesized optional chain)
	// Examples: (foo?.bar)!, (foo?.())!
	if expression.Kind == ast.KindParenthesizedExpression {
		parenExpr := expression.AsParenthesizedExpression()
		return hasOptionalChaining(parenExpr.Expression)
	}

	// Pattern 2: ChainExpression > NonNullExpression (direct optional chain)
	// Examples: foo?.bar!, foo?.()!
	// But NOT: foo?.bar!.baz (this is valid in TypeScript 3.9+)

	if hasOptionalChaining(expression) {
		// Check if this non-null assertion is "terminal" (not continued)
		return isTerminalAssertion(nonNullNode)
	}

	return false
}

// hasOptionalChaining checks if an expression has optional chaining at any level
func hasOptionalChaining(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Use the built-in IsOptionalChain check which should be more reliable
	return ast.IsOptionalChain(node)
}

// isTerminalAssertion checks if the non-null assertion is at the end of the expression chain
// For TypeScript 3.9+ compatibility: foo?.bar!.baz is valid (continued), but foo?.bar! is invalid (terminal)
func isTerminalAssertion(nonNullNode *ast.Node) bool {
	if nonNullNode.Parent == nil {
		return true // No parent means it's terminal
	}

	parent := nonNullNode.Parent

	// Check if the non-null assertion is being continued
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		propAccess := parent.AsPropertyAccessExpression()
		// If this non-null is the left side of a property access, it's continued (valid)
		return propAccess.Expression != nonNullNode

	case ast.KindElementAccessExpression:
		elemAccess := parent.AsElementAccessExpression()
		// If this non-null is the left side of an element access, it's continued (valid)
		return elemAccess.Expression != nonNullNode

	case ast.KindCallExpression:
		callExpr := parent.AsCallExpression()
		// If this non-null is the expression being called, it's continued (valid)
		return callExpr.Expression != nonNullNode

	default:
		// For other parents, it's terminal (invalid)
		return true
	}
}

func reportError(ctx rule.RuleContext, reportNode *ast.Node, nonNullNode *ast.Node) {
	message := rule.RuleMessage{
		Id:          "noNonNullOptionalChain",
		Description: "Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong.",
	}

	// Calculate the position of the '!' to remove it
	// The '!' is at the end of the non-null expression
	nonNullEnd := nonNullNode.End()
	exclamationStart := nonNullEnd - 1
	exclamationRange := core.NewTextRange(exclamationStart, nonNullEnd)

	suggestion := rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemovingNonNull",
			Description: "You should remove the non-null assertion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixRemoveRange(exclamationRange),
		},
	}

	ctx.ReportNodeWithSuggestions(reportNode, message, suggestion)
}
