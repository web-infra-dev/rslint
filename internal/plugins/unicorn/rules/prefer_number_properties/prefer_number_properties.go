package prefer_number_properties

import (
	_ "embed"
	"fmt"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer_number_properties.schema.json
var schemaJSON []byte

const (
	messageIDError      = "error"
	messageIDSuggestion = "suggestion"
)

type preferNumberPropertiesOptions struct {
	checkInfinity bool
	checkNaN      bool
}

type globalReference struct {
	node        *ast.Node
	name        string
	description string
	property    string
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-number-properties.md
var PreferNumberPropertiesRule = rule.Rule{
	Name:   "unicorn/prefer-number-properties",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		report := func(ref globalReference) {
			if isLeftHandSide(ref.node) {
				return
			}

			textRange := utils.TrimNodeTextRange(ctx.SourceFile, ref.node)
			msg := buildErrorMessage(ref.property, ref.description)
			if isSafeGlobalProperty(ref.name) {
				ctx.ReportRangeWithDeferredFixes(textRange, msg, func() []rule.RuleFix {
					return buildFixes(ctx.SourceFile, ref, textRange)
				})
				return
			}

			ctx.ReportRangeWithDeferredSuggestions(textRange, msg, func() []rule.RuleSuggestion {
				return []rule.RuleSuggestion{{
					Message:  buildSuggestionMessage(ref.property, ref.description),
					FixesArr: buildFixes(ctx.SourceFile, ref, textRange),
				}}
			})
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.AsIdentifier().Text
				if !isTrackedGlobalName(name) || !enabled(name, opts) ||
					utils.IsNonReferenceIdentifier(node) || isShadowed(&ctx, node, name) {
					return
				}
				report(referenceFromNode(node, name))
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if ref, ok := globalMemberReference(&ctx, node, opts); ok {
					report(ref)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ref, ok := globalMemberReference(&ctx, node, opts); ok {
					report(ref)
				}
			},
		}
	},
}

func parseOptions(options []any) preferNumberPropertiesOptions {
	opts := preferNumberPropertiesOptions{
		checkInfinity: false,
		checkNaN:      true,
	}
	optionsMap := utils.GetOptionsMap(options)
	if optionsMap == nil {
		return opts
	}
	if value, ok := optionsMap["checkInfinity"].(bool); ok {
		opts.checkInfinity = value
	}
	if value, ok := optionsMap["checkNaN"].(bool); ok {
		opts.checkNaN = value
	}
	return opts
}

func enabled(name string, opts preferNumberPropertiesOptions) bool {
	if name == "Infinity" && !opts.checkInfinity {
		return false
	}
	if name == "NaN" && !opts.checkNaN {
		return false
	}
	return true
}

func isTrackedGlobalName(name string) bool {
	switch name {
	case "parseInt", "parseFloat", "NaN", "Infinity", "isNaN", "isFinite":
		return true
	default:
		return false
	}
}

func isSafeGlobalProperty(name string) bool {
	switch name {
	case "parseInt", "parseFloat", "NaN", "Infinity":
		return true
	default:
		return false
	}
}

func isGlobalObjectName(name string) bool {
	switch name {
	case "globalThis", "global", "window", "self":
		return true
	default:
		return false
	}
}

func isShadowed(ctx *rule.RuleContext, node *ast.Node, name string) bool {
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			return utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return utils.IsShadowed(node, name)
}

func referenceFromNode(node *ast.Node, name string) globalReference {
	property := name
	description := name
	reportNode := node
	if name == "Infinity" && isNegative(node) {
		property = "NEGATIVE_INFINITY"
		description = "-Infinity"
		reportNode = node.Parent
	} else if name == "Infinity" {
		property = "POSITIVE_INFINITY"
	}
	return globalReference{
		node:        reportNode,
		name:        name,
		description: description,
		property:    property,
	}
}

func globalMemberReference(ctx *rule.RuleContext, node *ast.Node, opts preferNumberPropertiesOptions) (globalReference, bool) {
	propertyName, ok := utils.AccessExpressionStaticName(node)
	if !ok || !isTrackedGlobalName(propertyName) || !enabled(propertyName, opts) {
		return globalReference{}, false
	}

	object := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
	if object == nil || !ast.IsIdentifier(object) {
		return globalReference{}, false
	}
	objectName := object.AsIdentifier().Text
	if !isGlobalObjectName(objectName) || isShadowed(ctx, object, objectName) {
		return globalReference{}, false
	}

	return referenceFromNode(node, propertyName), true
}

func isNegative(node *ast.Node) bool {
	// ESTree wraps optional chains in ChainExpression, so upstream treats
	// `-globalThis?.Infinity` as replacing the positive member only.
	if ast.IsOptionalChain(node) {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	prefix := parent.AsPrefixUnaryExpression()
	return prefix != nil && prefix.Operator == ast.KindMinusToken && prefix.Operand == node
}

func isLeftHandSide(node *ast.Node) bool {
	if utils.IsWriteReference(node) {
		return true
	}
	parent := node.Parent
	// `delete globalThis.NaN` is filtered as a non-read. Optional chains are
	// not: upstream sees `delete globalThis?.NaN` through ChainExpression.
	return parent != nil &&
		parent.Kind == ast.KindDeleteExpression &&
		parent.Expression() == node &&
		node.Kind != ast.KindPrefixUnaryExpression &&
		!ast.IsOptionalChain(node)
}

func buildErrorMessage(property, description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDError,
		Description: fmt.Sprintf("Prefer `Number.%s` over `%s`.", property, description),
		Data: map[string]string{
			"property":    property,
			"description": description,
		},
	}
}

func buildSuggestionMessage(property, description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDSuggestion,
		Description: fmt.Sprintf("Replace `%s` with `Number.%s`.", description, property),
		Data: map[string]string{
			"property":    property,
			"description": description,
		},
	}
}

func buildFixes(sf *ast.SourceFile, ref globalReference, textRange core.TextRange) []rule.RuleFix {
	replacement := "Number." + ref.property

	if ast.IsIdentifier(ref.node) {
		if shorthand := ref.node.Parent; shorthand != nil && shorthand.Kind == ast.KindShorthandPropertyAssignment {
			if shorthand.AsShorthandPropertyAssignment().Name() == ref.node {
				replacement = ref.name + ": " + replacement
			}
		}
	}

	replacement = safeNumberReplacementText(sf, textRange, replacement)
	return []rule.RuleFix{rule.RuleFixReplaceRange(textRange, replacement)}
}

// safeNumberReplacementText is the rule-specific fast path for
// utils.SafeReplacementText. Every replacement starts and ends with an
// identifier character, so only an immediately adjacent identifier character
// can merge with it. Looking at the two boundary runes avoids scanning all
// preceding tokens again for every reported reference.
func safeNumberReplacementText(sf *ast.SourceFile, textRange core.TextRange, replacement string) string {
	text := sf.Text()
	if textRange.Pos() > 0 {
		before, _ := utf8.DecodeLastRuneInString(text[:textRange.Pos()])
		if scanner.IsIdentifierPart(before) {
			replacement = " " + replacement
		}
	}
	if textRange.End() < len(text) {
		after, _ := utf8.DecodeRuneInString(text[textRange.End():])
		if scanner.IsIdentifierPart(after) {
			replacement += " "
		}
	}
	return replacement
}
