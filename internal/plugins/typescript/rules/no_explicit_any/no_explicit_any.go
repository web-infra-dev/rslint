package no_explicit_any

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_explicit_any.schema.json
var schemaJSON []byte

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

func parseOptions(options []any) NoExplicitAnyOptions {
	opts := NoExplicitAnyOptions{}
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]interface{})
	if v, ok := m["fixToUnknown"].(bool); ok {
		opts.FixToUnknown = v
	}
	if v, ok := m["ignoreRestArgs"].(bool); ok {
		opts.IgnoreRestArgs = v
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
			typeRef := p.AsTypeReferenceNode()
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

func anyKeywordRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	const keyword = "any"
	end := node.End()
	start := end - len(keyword)
	text := sourceFile.Text()
	// Parsed keyword nodes end at the token boundary, so the validated suffix
	// avoids rescanning leading trivia. Reparsed or synthetic nodes keep the
	// scanner-backed behavior.
	if start >= 0 && end <= len(text) && text[start:end] == keyword {
		return core.NewTextRange(start, end)
	}
	return utils.TrimNodeTextRange(sourceFile, node)
}

var NoExplicitAnyRule = rule.CreateRule(rule.Rule{
	Name:   "no-explicit-any",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindAnyKeyword: func(node *ast.Node) {
				if opts.IgnoreRestArgs && isAnyInRestParameter(node) {
					return
				}

				reportRange := anyKeywordRange(ctx.SourceFile, node)
				isKeyofAny := isWithinKeyofAny(node)
				if opts.FixToUnknown {
					if isKeyofAny {
						ctx.ReportRangeWithDeferredFixes(reportRange, buildUnexpectedAnyMessage(), func() []rule.RuleFix {
							return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey")}
						})
						return
					}
					ctx.ReportRangeWithDeferredFixes(reportRange, buildUnexpectedAnyMessage(), func() []rule.RuleFix {
						return []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "unknown")}
					})
					return
				}

				if isKeyofAny {
					ctx.ReportRangeWithDeferredSuggestions(reportRange, buildUnexpectedAnyMessage(), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message:  buildSuggestPropertyKeyMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node.Parent, "PropertyKey")},
						}}
					})
					return
				}

				ctx.ReportRangeWithDeferredSuggestions(reportRange, buildUnexpectedAnyMessage(), func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{
						{
							Message:  buildSuggestUnknownMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "unknown")},
						},
						{
							Message:  buildSuggestNeverMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "never")},
						},
					}
				})
			},
		}
	},
})
