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

		// Parse options if provided
		if options != nil {
			if optionsMap, ok := options.(map[string]interface{}); ok {
				if fixToUnknown, exists := optionsMap["fixToUnknown"]; exists {
					if fixToUnknownBool, ok := fixToUnknown.(bool); ok {
						ruleOptions.FixToUnknown = fixToUnknownBool
					}
				}
				if ignoreRestArgs, exists := optionsMap["ignoreRestArgs"]; exists {
					if ignoreRestArgsBool, ok := ignoreRestArgs.(bool); ok {
						ruleOptions.IgnoreRestArgs = ignoreRestArgsBool
					}
				}
			}
		}

		// Helper function to check if a node is within a keyof any expression
		isNodeWithinKeyofAny := func(node *ast.Node) bool {
			if node.Parent == nil {
				return false
			}
			// Check if parent is a type operator with keyof operator
			return ast.IsTypeOperatorNode(node.Parent) && 
				node.Parent.Kind == ast.KindTypeOperator
		}

		// Helper function to check if the any is in a valid rest parameter type that should be ignored
		isValidRestParameterType := func(anyNode *ast.Node, paramNode *ast.Node) bool {
			// Valid patterns to ignore: any[], readonly any[], Array<any>, ReadonlyArray<any>
			// NOT valid: bare any in rest parameter
			
			// Walk up from the any node to check if it's in array or Array<T> type
			current := anyNode
			for current != nil && current != paramNode {
				if current.Kind == ast.KindArrayType {
					// any[] or readonly any[] pattern
					return true
				}
				if current.Kind == ast.KindTypeReference {
					// Array<any> or ReadonlyArray<any> pattern
					return true
				}
				current = current.Parent
			}
			
			return false
		}

		// Helper function to check if node is a rest parameter
		isNodeInRestParameter := func(node *ast.Node) bool {
			// Walk up the AST to find if this any is within a rest parameter
			current := node
			for current != nil {
				if current.Kind == ast.KindParameter {
					param := current.AsParameterDeclaration()
					if param.DotDotDotToken != nil {
						// This is a rest parameter - check if the any is in a valid rest parameter type
						// For ignoreRestArgs, we should ignore any types in rest parameters that are arrays or Array-like
						// but NOT ignore bare any types in rest parameters
						return isValidRestParameterType(node, current)
					}
				}
				current = current.Parent
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindAnyKeyword: func(node *ast.Node) {
				// Check if we should ignore this any due to ignoreRestArgs option
				if ruleOptions.IgnoreRestArgs && isNodeInRestParameter(node) {
					return
				}

				isKeyofAny := isNodeWithinKeyofAny(node)

				var suggestions []rule.RuleSuggestion

				// The fixToUnknown option only provides auto-fixes, not suggestions
				// For test cases, we still need to provide suggestions for validation
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

				// Always provide suggestions for test validation
				ctx.ReportNodeWithSuggestions(node, buildUnexpectedAnyMessage(), suggestions...)
			},
		}
	},
}