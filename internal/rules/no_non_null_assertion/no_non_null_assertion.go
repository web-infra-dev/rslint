package no_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
)

var NoNonNullAssertionRule = rule.Rule{
	Name: "no-non-null-assertion",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				nonNullExpr := node.AsNonNullExpression()
				
				// Build suggestions based on the parent context
				suggestions := buildSuggestions(ctx, node, nonNullExpr)
				
				ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
					Id:          "noNonNull",
					Description: "Forbidden non-null assertion.",
				}, suggestions...)
			},
		}
	},
}

func buildSuggestions(ctx rule.RuleContext, node *ast.Node, nonNullExpr *ast.NonNullExpression) []rule.RuleSuggestion {
	var suggestions []rule.RuleSuggestion
	
	parent := node.Parent
	if parent == nil {
		return suggestions
	}
	
	// Only provide suggestions if this non-null assertion is immediately followed by a chaining operation
	// This means the non-null assertion should be the direct expression of the parent access/call
	shouldProvideSuggestions := false
	
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		propAccess := parent.AsPropertyAccessExpression()
		shouldProvideSuggestions = propAccess.Expression == node
		
	case ast.KindElementAccessExpression:
		elemAccess := parent.AsElementAccessExpression()
		shouldProvideSuggestions = elemAccess.Expression == node
		
	case ast.KindCallExpression:
		callExpr := parent.AsCallExpression()
		shouldProvideSuggestions = callExpr.Expression == node
	}
	
	if !shouldProvideSuggestions {
		return suggestions
	}
	
	// Calculate the position of the '!' to remove it
	// The '!' is at the end of the non-null expression
	nonNullEnd := node.End()
	exclamationStart := nonNullEnd - 1
	
	// Helper function to create suggestion for removing '!'
	removeExclamation := func() rule.RuleFix {
		exclamationRange := core.NewTextRange(exclamationStart, nonNullEnd)
		return rule.RuleFixRemoveRange(exclamationRange)
	}
	
	// Helper function to create suggestion for replacing '!' with '?.'
	// For property access (x!.y), we replace '!' with '?' to get x?.y
	// For element access (x![y]), we replace '!' with '?.' to get x?.[y]  
	// For call expressions (x!()), we replace '!' with '?.' to get x?.()
	replaceWithOptional := func() rule.RuleFix {
		exclamationRange := core.NewTextRange(exclamationStart, nonNullEnd)
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			// x!.y -> x?.y (replace ! with ? since . is already there)
			return rule.RuleFixReplaceRange(exclamationRange, "?")
		default:
			// x![y] -> x?.[y] or x!() -> x?.() (replace ! with ?.)
			return rule.RuleFixReplaceRange(exclamationRange, "?.")
		}
	}
	
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// x!.y or x!?.y
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess.QuestionDotToken == nil {
			// x!.y -> x?.y (replace ! with ?.)
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{
					replaceWithOptional(),
				},
			})
		} else {
			// x!?.y -> x?.y (just remove !)
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{removeExclamation()},
			})
		}
		
	case ast.KindElementAccessExpression:
		// x![y] or x!?.[y]
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess.QuestionDotToken == nil {
			// x![y] -> x?.[y]
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{replaceWithOptional()},
			})
		} else {
			// x!?.[y] -> x?.[y] (just remove !)
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{removeExclamation()},
			})
		}
		
	case ast.KindCallExpression:
		// x!() or x!?.()
		callExpr := parent.AsCallExpression()
		if callExpr.QuestionDotToken == nil {
			// x!() -> x?.()
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{replaceWithOptional()},
			})
		} else {
			// x!?.() -> x?.() (just remove !)
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          "suggestOptionalChain",
					Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
				},
				FixesArr: []rule.RuleFix{removeExclamation()},
			})
		}
	}
	
	return suggestions
}