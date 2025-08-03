package consistent_type_assertions

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options represents the rule configuration
type Options struct {
	AssertionStyle              string `json:"assertionStyle"`
	ObjectLiteralTypeAssertions string `json:"objectLiteralTypeAssertions,omitempty"`
	ArrayLiteralTypeAssertions  string `json:"arrayLiteralTypeAssertions,omitempty"`
}

// Default options - when no options are provided, both styles are allowed
var defaultOptions = Options{
	AssertionStyle:              "", // Empty means both styles are allowed
	ObjectLiteralTypeAssertions: "allow",
	ArrayLiteralTypeAssertions:  "allow",
}

func buildAngleBracketMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "angle-bracket",
		Description: fmt.Sprintf("Use '<%s>' instead of 'as %s'.", cast, cast),
	}
}

func buildAsMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "as",
		Description: fmt.Sprintf("Use 'as %s' instead of '<%s>'.", cast, cast),
	}
}

func buildNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "never",
		Description: "Do not use any type assertions.",
	}
}

func buildUseAsAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "use-as-assertion",
		Description: "Use as assertion instead.",
	}
}

func buildUseAngleBracketAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "use-angle-bracket-assertion",
		Description: "Use angle bracket assertion instead.",
	}
}

func buildReplaceArrayTypeAssertionWithAnnotationMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "array-literal-assertion-suggestion",
		Description: fmt.Sprintf("Use const x: %s = [ ... ] instead.", cast),
	}
}

func buildReplaceArrayTypeAssertionWithSatisfiesMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "array-literal-assertion-suggestion",
		Description: fmt.Sprintf("Use const x = [ ... ] satisfies %s instead.", cast),
	}
}

func buildReplaceObjectTypeAssertionWithAnnotationMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object-literal-assertion-suggestion",
		Description: fmt.Sprintf("Use const x: %s = { ... } instead.", cast),
	}
}

func buildReplaceObjectTypeAssertionWithSatisfiesMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object-literal-assertion-suggestion",
		Description: fmt.Sprintf("Use const x = { ... } satisfies %s instead.", cast),
	}
}

func buildUnexpectedArrayTypeAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "array-literal-assertion",
		Description: "Always prefer const x: T[] = [ ... ].",
	}
}

func buildUnexpectedObjectTypeAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object-literal-assertion",
		Description: "Always prefer const x: T = { ... }.",
	}
}

func isConst(node *ast.Node) bool {
	if node == nil || !ast.IsTypeReferenceNode(node) {
		return false
	}

	typeRef := node.AsTypeReferenceNode()
	if typeRef == nil {
		return false
	}

	typeName := typeRef.TypeName
	if typeName == nil {
		return false
	}

	return ast.IsIdentifier(typeName) && typeName.Text() == "const"
}

func checkType(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindAnyKeyword, ast.KindUnknownKeyword:
		return false
	case ast.KindTypeReference:
		// For type references, check if it's `const`
		if isConst(node) {
			return false
		}
		// Also check for qualified names with dots (e.g., Foo.Bar)
		typeRef := node.AsTypeReferenceNode()
		if typeRef != nil && typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
			return true
		}
		return true
	default:
		return true
	}
}

func isAsParameter(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// Direct check for common parameter contexts
	switch parent.Kind {
	case ast.KindNewExpression, ast.KindCallExpression, ast.KindThrowStatement:
		return true
	case ast.KindJsxExpression:
		return true
	case ast.KindTemplateSpan:
		// Check if this is part of a tagged template expression
		if parent.Parent != nil && parent.Parent.Kind == ast.KindTemplateExpression {
			if parent.Parent.Parent != nil {
				return ast.IsTaggedTemplateExpression(parent.Parent.Parent)
			}
		}
		return true
	case ast.KindParameter:
		return true
	case ast.KindBinaryExpression:
		// Check if this is a default parameter initialization
		binExpr := parent.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
			if parent.Parent != nil && parent.Parent.Kind == ast.KindParameter {
				return true
			}
		}
	}

	// Special handling for optional chaining (print?.({ bar: 5 } as Foo))
	if parent.Kind == ast.KindParenthesizedExpression {
		grandParent := parent.Parent
		if grandParent != nil {
			switch grandParent.Kind {
			case ast.KindCallExpression:
				return true
			case ast.KindPropertyAccessExpression:
				// Check if this is part of an optional call chain
				if grandParent.Parent != nil && grandParent.Parent.Kind == ast.KindCallExpression {
					return true
				}
			}
		}
	}

	return false
}

func getTypeAnnotationText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil || ctx.SourceFile == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if !textRange.IsValid() {
		return ""
	}
	text := ctx.SourceFile.Text()
	if text == "" || textRange.Pos() < 0 || textRange.End() > len(text) || textRange.Pos() >= textRange.End() {
		return ""
	}
	return text[textRange.Pos():textRange.End()]
}

func getExpressionText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil || ctx.SourceFile == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if !textRange.IsValid() {
		return ""
	}
	text := ctx.SourceFile.Text()
	if text == "" || textRange.Pos() < 0 || textRange.End() > len(text) || textRange.Pos() >= textRange.End() {
		return ""
	}
	return text[textRange.Pos():textRange.End()]
}

func getSuggestions(ctx rule.RuleContext, node *ast.Node, isAsExpression bool, annotationMessageId, satisfiesMessageId string) []rule.RuleSuggestion {
	var suggestions []rule.RuleSuggestion
	if node == nil {
		return suggestions
	}

	var typeAnnotation *ast.Node
	var expression *ast.Node

	if isAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr == nil {
			return suggestions
		}
		typeAnnotation = asExpr.Type
		expression = asExpr.Expression
	} else {
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return suggestions
		}
		typeAnnotation = typeAssertion.Type
		expression = typeAssertion.Expression
	}

	if typeAnnotation == nil || expression == nil {
		return suggestions
	}

	cast := getTypeAnnotationText(ctx, typeAnnotation)

	// Check if this is a variable declarator that can have type annotation
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindVariableDeclaration {
		varDecl := parent.AsVariableDeclaration()
		if varDecl != nil && varDecl.Type == nil && varDecl.Name() != nil {
			// Add annotation suggestion
			annotationMsg := rule.RuleMessage{
				Id:          annotationMessageId,
				Description: fmt.Sprintf("Use const x: %s = ... instead.", cast),
			}
			if annotationMessageId == "replaceObjectTypeAssertionWithAnnotation" {
				annotationMsg = buildReplaceObjectTypeAssertionWithAnnotationMessage(cast)
			} else if annotationMessageId == "replaceArrayTypeAssertionWithAnnotation" {
				annotationMsg = buildReplaceArrayTypeAssertionWithAnnotationMessage(cast)
			}

			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: annotationMsg,
				FixesArr: []rule.RuleFix{
					rule.RuleFixInsertAfter(varDecl.Name(), fmt.Sprintf(": %s", cast)),
					rule.RuleFixReplace(ctx.SourceFile, node, getExpressionText(ctx, expression)),
				},
			})
		}
	}

	// Always add satisfies suggestion
	satisfiesMsg := rule.RuleMessage{
		Id:          satisfiesMessageId,
		Description: fmt.Sprintf("Use ... satisfies %s instead.", cast),
	}
	if satisfiesMessageId == "replaceObjectTypeAssertionWithSatisfies" {
		satisfiesMsg = buildReplaceObjectTypeAssertionWithSatisfiesMessage(cast)
	} else if satisfiesMessageId == "replaceArrayTypeAssertionWithSatisfies" {
		satisfiesMsg = buildReplaceArrayTypeAssertionWithSatisfiesMessage(cast)
	}

	suggestions = append(suggestions, rule.RuleSuggestion{
		Message: satisfiesMsg,
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, node, getExpressionText(ctx, expression)),
			rule.RuleFixInsertAfter(node, fmt.Sprintf(" satisfies %s", cast)),
		},
	})

	return suggestions
}

func reportIncorrectAssertionType(ctx rule.RuleContext, node *ast.Node, options Options, isAsExpression bool) {
	if node == nil {
		return
	}

	var typeAnnotation *ast.Node
	if isAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr == nil {
			return
		}
		typeAnnotation = asExpr.Type
	} else {
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return
		}
		typeAnnotation = typeAssertion.Type
	}

	// If this node is `as const`, then don't report an error when style is 'never'
	if isConst(typeAnnotation) && options.AssertionStyle == "never" {
		return
	}

	switch options.AssertionStyle {
	case "angle-bracket":
		cast := getTypeAnnotationText(ctx, typeAnnotation)
		ctx.ReportNode(node, buildAngleBracketMessage(cast))
	case "as":
		cast := getTypeAnnotationText(ctx, typeAnnotation)
		// For angle-bracket to as conversion, we'd need complex fix logic
		// For now, just report without fix
		ctx.ReportNode(node, buildAsMessage(cast))
	case "never":
		ctx.ReportNode(node, buildNeverMessage())
	}
}

func checkExpressionForObjectAssertion(ctx rule.RuleContext, node *ast.Node, options Options, isAsExpression bool) {
	if node == nil || options.AssertionStyle == "never" ||
		options.ObjectLiteralTypeAssertions == "allow" {
		return
	}

	var expression *ast.Node
	var typeAnnotation *ast.Node

	if isAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr == nil {
			return
		}
		expression = asExpr.Expression
		typeAnnotation = asExpr.Type
	} else {
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return
		}
		expression = typeAssertion.Expression
		typeAnnotation = typeAssertion.Type
	}

	if expression == nil || expression.Kind != ast.KindObjectLiteralExpression {
		return
	}

	if options.ObjectLiteralTypeAssertions == "allow-as-parameter" && isAsParameter(node) {
		return
	}

	if checkType(typeAnnotation) {
		suggestions := getSuggestions(ctx, node, isAsExpression,
			"replaceObjectTypeAssertionWithAnnotation",
			"replaceObjectTypeAssertionWithSatisfies")

		ctx.ReportNodeWithSuggestions(node, buildUnexpectedObjectTypeAssertionMessage(), suggestions...)
	}
}

func checkExpressionForArrayAssertion(ctx rule.RuleContext, node *ast.Node, options Options, isAsExpression bool) {
	if node == nil || options.AssertionStyle == "never" ||
		options.ArrayLiteralTypeAssertions == "allow" {
		return
	}

	var expression *ast.Node
	var typeAnnotation *ast.Node

	if isAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr == nil {
			return
		}
		expression = asExpr.Expression
		typeAnnotation = asExpr.Type
	} else {
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return
		}
		expression = typeAssertion.Expression
		typeAnnotation = typeAssertion.Type
	}

	if expression == nil || expression.Kind != ast.KindArrayLiteralExpression {
		return
	}

	if options.ArrayLiteralTypeAssertions == "allow-as-parameter" && isAsParameter(node) {
		return
	}

	if checkType(typeAnnotation) {
		suggestions := getSuggestions(ctx, node, isAsExpression,
			"replaceArrayTypeAssertionWithAnnotation",
			"replaceArrayTypeAssertionWithSatisfies")

		ctx.ReportNodeWithSuggestions(node, buildUnexpectedArrayTypeAssertionMessage(), suggestions...)
	}
}

var ConsistentTypeAssertionsRule = rule.Rule{
	Name: "consistent-type-assertions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := defaultOptions

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if val, ok := optsMap["assertionStyle"].(string); ok {
					opts.AssertionStyle = val
				}
				if val, ok := optsMap["objectLiteralTypeAssertions"].(string); ok {
					opts.ObjectLiteralTypeAssertions = val
				}
				if val, ok := optsMap["arrayLiteralTypeAssertions"].(string); ok {
					opts.ArrayLiteralTypeAssertions = val
				}
			}
		}

		return rule.RuleListeners{
			ast.KindAsExpression: func(node *ast.Node) {
				// Only report style violation if a specific style is configured
				if opts.AssertionStyle == "angle-bracket" || opts.AssertionStyle == "never" {
					reportIncorrectAssertionType(ctx, node, opts, true)
					return
				}

				checkExpressionForObjectAssertion(ctx, node, opts, true)
				checkExpressionForArrayAssertion(ctx, node, opts, true)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				// Only report style violation if a specific style is configured
				if opts.AssertionStyle == "as" || opts.AssertionStyle == "never" {
					reportIncorrectAssertionType(ctx, node, opts, false)
					return
				}

				checkExpressionForObjectAssertion(ctx, node, opts, false)
				checkExpressionForArrayAssertion(ctx, node, opts, false)
			},
		}
	},
}
