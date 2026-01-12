package no_explicit_any

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoExplicitAnyOptions struct {
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

func parseOptions(options any) NoExplicitAnyOptions {
	opts := NoExplicitAnyOptions{}
	if options == nil {
		return opts
	}
	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				if v, ok := m["fixToUnknown"].(bool); ok {
					opts.FixToUnknown = v
				}
				if v, ok := m["ignoreRestArgs"].(bool); ok {
					opts.IgnoreRestArgs = v
				}
			}
		}
		return opts
	}
	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		if v, ok := m["fixToUnknown"].(bool); ok {
			opts.FixToUnknown = v
		}
		if v, ok := m["ignoreRestArgs"].(bool); ok {
			opts.IgnoreRestArgs = v
		}
	}
	return opts
}

func isAnyInRestParameter(node *ast.Node) bool {
	// Check if the any keyword is inside a rest parameter with array type
	// We need to check if the any is part of an array type in a rest parameter
	// Valid patterns to ignore: ...args: any[], ...args: readonly any[], ...args: Array<any>, ...args: ReadonlyArray<any>

	// First check if we're inside an ArrayType
	inArrayType := false
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Kind == ast.KindArrayType {
			inArrayType = true
			break
		}
		if p.Kind == ast.KindTypeReference {
			typeRef := p.AsTypeReference()
			if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
				identifier := typeRef.TypeName.AsIdentifier()
				if identifier != nil && (identifier.Text == "Array" || identifier.Text == "ReadonlyArray") {
					inArrayType = true
					break
				}
			}
		}
	}

	if !inArrayType {
		return false
	}

	// Then check if we're in a rest parameter
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Kind == ast.KindParameter {
			param := p.AsParameterDeclaration()
			return param.DotDotDotToken != nil
		}
	}
	return false
}

func isWithinKeyofAny(node *ast.Node) bool {
	if node.Parent == nil || node.Parent.Kind != ast.KindTypeOperator {
		return false
	}
	typeOp := node.Parent.AsTypeOperatorNode()
	return typeOp != nil && typeOp.Operator == ast.KindKeyOfKeyword
}

var NoExplicitAnyRule = rule.CreateRule(rule.Rule{
	Name: "no-explicit-any",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindAnyKeyword: func(node *ast.Node) {
				if opts.IgnoreRestArgs && isAnyInRestParameter(node) {
					return
				}
				if isWithinKeyofAny(node) {
					if opts.FixToUnknown {
						ctx.ReportNodeWithFixes(node, buildUnexpectedAnyMessage(), rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey"))
					} else {
						ctx.ReportNodeWithSuggestions(node, buildUnexpectedAnyMessage(), rule.RuleSuggestion{
							Message:  buildSuggestPropertyKeyMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey")},
						})
					}
					return
				}

				if opts.FixToUnknown {
					ctx.ReportNodeWithFixes(node, buildUnexpectedAnyMessage(), rule.RuleFixReplace(ctx.SourceFile, node, "unknown"))
				} else {
					ctx.ReportNodeWithSuggestions(node, buildUnexpectedAnyMessage(),
						rule.RuleSuggestion{
							Message:  buildSuggestUnknownMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "unknown")},
						},
						rule.RuleSuggestion{
							Message:  buildSuggestNeverMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "never")},
						},
					)
				}
			},
		}
	},
})
