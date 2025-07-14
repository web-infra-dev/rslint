package no_explicit_any

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// RuleOptions represents the configuration options for the no-explicit-any rule
type RuleOptions struct {
	FixToUnknown   bool `json:"fixToUnknown"`
	IgnoreRestArgs bool `json:"ignoreRestArgs"`
}

func buildUnexpectedAnyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAny",
		Description: "Unexpected any. Specify a different type.",
	}
}

func buildSuggestUnknownMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestUnknown",
		Description: "Use `unknown` instead, this will force you to explicitly, and safely assert the type is correct.",
	}
}

func buildSuggestNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestNever",
		Description: "Use `never` instead, this is useful when instantiating generic type parameters that you don't need to know the type of.",
	}
}

func buildSuggestPropertyKeyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestPropertyKey",
		Description: "Use `PropertyKey` instead, this is more explicit than `keyof any`.",
	}
}

var NoExplicitAnyRule = rule.Rule{
	Name: "no-explicit-any",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default options
		ruleOptions := RuleOptions{
			FixToUnknown:   false,
			IgnoreRestArgs: false,
		}

		// TODO: Parse options if provided
		// if options != nil {
		//     // Parse options from any to RuleOptions
		// }

		// Helper function to check if a node is within a keyof any expression
		isNodeWithinKeyofAny := func(node *ast.Node) bool {
			if node.Parent == nil {
				return false
			}
			// Check if parent is a type operator with keyof operator
			return ast.IsTypeOperatorNode(node.Parent) && 
				node.Parent.Kind == ast.KindTypeOperator
		}

		// Helper function to check if node is a rest element in a function
		isNodeValidFunction := func(node *ast.Node) bool {
			functionKinds := []ast.Kind{
				ast.KindArrowFunction,
				ast.KindFunctionDeclaration,
				ast.KindFunctionExpression,
				ast.KindMethodDeclaration,
				ast.KindConstructor,
				ast.KindCallSignature,
				ast.KindConstructSignature,
			}
			for _, kind := range functionKinds {
				if node.Kind == kind {
					return true
				}
			}
			return false
		}

		isNodeRestElementInFunction := func(node *ast.Node) bool {
			if node.Kind != ast.KindParameter {
				return false
			}
			param := node.AsParameterDeclaration()
			return param.DotDotDotToken != nil &&
				node.Parent != nil &&
				isNodeValidFunction(node.Parent)
		}

		// Helper function to check if node is descendant of rest element in function
		isNodeDescendantOfRestElementInFunction := func(node *ast.Node) bool {
			// Check ancestors up to 4 levels deep
			current := node
			for i := 0; i < 4 && current != nil; i++ {
				if isNodeRestElementInFunction(current) {
					return true
				}
				current = current.Parent
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindAnyKeyword: func(node *ast.Node) {
				// Check if we should ignore this any due to ignoreRestArgs option
				if ruleOptions.IgnoreRestArgs && isNodeDescendantOfRestElementInFunction(node) {
					return
				}

				isKeyofAny := isNodeWithinKeyofAny(node)

				var fixes []rule.RuleFix
				var suggestions []rule.RuleSuggestion

				if ruleOptions.FixToUnknown {
					// Provide auto-fix
					if isKeyofAny {
						fixes = []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey"),
						}
					} else {
						fixes = []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, "unknown"),
						}
					}
				} else {
					// Provide suggestions
					if isKeyofAny {
						suggestions = []rule.RuleSuggestion{
							{
								Message: buildSuggestPropertyKeyMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey"),
								},
							},
						}
					} else {
						suggestions = []rule.RuleSuggestion{
							{
								Message: buildSuggestUnknownMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixReplace(ctx.SourceFile, node, "unknown"),
								},
							},
							{
								Message: buildSuggestNeverMessage(),
								FixesArr: []rule.RuleFix{
									rule.RuleFixReplace(ctx.SourceFile, node, "never"),
								},
							},
						}
					}
				}

				if len(fixes) > 0 {
					ctx.ReportNodeWithFixes(node, buildUnexpectedAnyMessage(), fixes...)
				} else {
					ctx.ReportNodeWithSuggestions(node, buildUnexpectedAnyMessage(), suggestions...)
				}
			},
		}
	},
}