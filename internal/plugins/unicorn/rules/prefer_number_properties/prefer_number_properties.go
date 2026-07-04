package prefer_number_properties

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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

var safeGlobalProperties = map[string]bool{
	"parseInt":   true,
	"parseFloat": true,
	"NaN":        true,
	"Infinity":   true,
	"isNaN":      false,
	"isFinite":   false,
}

var globalObjectNames = map[string]struct{}{
	"globalThis": {},
	"global":     {},
	"window":     {},
	"self":       {},
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-number-properties.md
var PreferNumberPropertiesRule = rule.Rule{
	Name: "unicorn/prefer-number-properties",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		report := func(ref globalReference) {
			if !enabled(ref.name, opts) {
				return
			}
			if isLeftHandSide(ref.node) {
				return
			}

			msg := buildErrorMessage(ref.property, ref.description)
			fixes := buildFixes(ctx.SourceFile, ref)
			if safeGlobalProperties[ref.name] {
				ctx.ReportRangeWithFixes(utils.TrimNodeTextRange(ctx.SourceFile, ref.node), msg, fixes...)
				return
			}

			ctx.ReportRangeWithSuggestions(
				utils.TrimNodeTextRange(ctx.SourceFile, ref.node),
				msg,
				rule.RuleSuggestion{
					Message:  buildSuggestionMessage(ref.property, ref.description),
					FixesArr: fixes,
				},
			)
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.AsIdentifier().Text
				if !isTrackedGlobalName(name) || utils.IsNonReferenceIdentifier(node) || utils.IsShadowed(node, name) {
					return
				}
				report(referenceFromNode(node, name))
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if ref, ok := globalMemberReference(node); ok {
					report(ref)
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ref, ok := globalMemberReference(node); ok {
					report(ref)
				}
			},
		}
	},
}

func parseOptions(options any) preferNumberPropertiesOptions {
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
	_, ok := safeGlobalProperties[name]
	return ok
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

func globalMemberReference(node *ast.Node) (globalReference, bool) {
	propertyName, ok := utils.AccessExpressionStaticName(node)
	if !ok || !isTrackedGlobalName(propertyName) {
		return globalReference{}, false
	}

	object := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
	if object == nil || !ast.IsIdentifier(object) {
		return globalReference{}, false
	}
	objectName := object.AsIdentifier().Text
	if _, ok := globalObjectNames[objectName]; !ok || utils.IsShadowed(object, objectName) {
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

func buildFixes(sf *ast.SourceFile, ref globalReference) []rule.RuleFix {
	replacement := "Number." + ref.property
	textRange := utils.TrimNodeTextRange(sf, ref.node)

	if ast.IsIdentifier(ref.node) {
		if shorthand := ref.node.Parent; shorthand != nil && shorthand.Kind == ast.KindShorthandPropertyAssignment {
			if shorthand.AsShorthandPropertyAssignment().Name() == ref.node {
				replacement = ref.name + ": " + replacement
			}
		}
	}

	replacement = utils.SafeReplacementText(sf, ref.node, replacement)
	return []rule.RuleFix{rule.RuleFixReplaceRange(textRange, replacement)}
}
