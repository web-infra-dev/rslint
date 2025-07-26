package consistent_type_assertions

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

// Options represents the rule configuration
type Options struct {
	AssertionStyle               string `json:"assertionStyle"`
	ObjectLiteralTypeAssertions  string `json:"objectLiteralTypeAssertions,omitempty"`
	ArrayLiteralTypeAssertions   string `json:"arrayLiteralTypeAssertions,omitempty"`
}

// Default options
var defaultOptions = Options{
	AssertionStyle:              "as",
	ObjectLiteralTypeAssertions: "allow",
	ArrayLiteralTypeAssertions:  "allow",
}

func buildAngleBracketMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "angleBracket",
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

func buildReplaceArrayTypeAssertionWithAnnotationMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceArrayTypeAssertionWithAnnotation",
		Description: fmt.Sprintf("Use const x: %s = [ ... ] instead.", cast),
	}
}

func buildReplaceArrayTypeAssertionWithSatisfiesMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceArrayTypeAssertionWithSatisfies",
		Description: fmt.Sprintf("Use const x = [ ... ] satisfies %s instead.", cast),
	}
}

func buildReplaceObjectTypeAssertionWithAnnotationMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceObjectTypeAssertionWithAnnotation",
		Description: fmt.Sprintf("Use const x: %s = { ... } instead.", cast),
	}
}

func buildReplaceObjectTypeAssertionWithSatisfiesMessage(cast string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceObjectTypeAssertionWithSatisfies",
		Description: fmt.Sprintf("Use const x = { ... } satisfies %s instead.", cast),
	}
}

func buildUnexpectedArrayTypeAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedArrayTypeAssertion",
		Description: "Always prefer const x: T[] = [ ... ].",
	}
}

func buildUnexpectedObjectTypeAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedObjectTypeAssertion", 
		Description: "Always prefer const x: T = { ... }.",
	}
}

func isConst(node *ast.Node) bool {
	if !ast.IsTypeReferenceNode(node) {
		return false
	}

	typeRef := node.AsTypeReferenceNode()
	typeName := typeRef.TypeName
	
	return ast.IsIdentifier(typeName) && typeName.Text() == "const"
}

func checkType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindAnyKeyword, ast.KindUnknownKeyword:
		return false
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		// Ignore `as const` and `<const>`
		if isConst(node) {
			// Allow qualified names which have dots between identifiers, `Foo.Bar`
			return ast.IsQualifiedName(typeRef.TypeName)
		}
		return true
	default:
		return true
	}
}

func isAsParameter(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindNewExpression, ast.KindCallExpression, ast.KindThrowStatement:
		return true
	case ast.KindParameter:
		return true
	case ast.KindJsxExpression:
		return true
	case ast.KindTemplateSpan:
		// Check if this is part of a tagged template expression
		templateLiteral := parent.Parent
		if templateLiteral != nil && templateLiteral.Kind == ast.KindTemplateExpression {
			return ast.IsTaggedTemplateExpression(templateLiteral.Parent)
		}
		return false
	default:
		return false
	}
}

func getTypeAnnotationText(ctx rule.RuleContext, node *ast.Node) string {
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}

func getExpressionText(ctx rule.RuleContext, node *ast.Node) string {
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}

func getSuggestions(ctx rule.RuleContext, node *ast.Node, isAsExpression bool, annotationMessageId, satisfiesMessageId string) []rule.RuleSuggestion {
	var suggestions []rule.RuleSuggestion
	var typeAnnotation *ast.Node
	var expression *ast.Node

	if isAsExpression {
		asExpr := node.AsAsExpression()
		typeAnnotation = asExpr.Type
		expression = asExpr.Expression
	} else {
		typeAssertion := node.AsTypeAssertion()
		typeAnnotation = typeAssertion.Type
		expression = typeAssertion.Expression
	}

	cast := getTypeAnnotationText(ctx, typeAnnotation)
	
	// Check if this is a variable declarator that can have type annotation
	parent := node.Parent
	if parent != nil && ast.IsVariableDeclaration(parent) {
		varDecl := parent.AsVariableDeclaration()
		if varDecl.Type == nil {
			// Add annotation suggestion
			suggestions = append(suggestions, rule.RuleSuggestion{
				Message: rule.RuleMessage{
					Id:          annotationMessageId,
					Description: fmt.Sprintf("Use const x: %s = ... instead.", cast),
				},
				FixesArr: []rule.RuleFix{
					rule.RuleFixInsertAfter(varDecl.Name(), fmt.Sprintf(": %s", cast)),
					rule.RuleFixReplace(ctx.SourceFile, node, getExpressionText(ctx, expression)),
				},
			})
		}
	}

	// Always add satisfies suggestion
	suggestions = append(suggestions, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          satisfiesMessageId,
			Description: fmt.Sprintf("Use ... satisfies %s instead.", cast),
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, node, getExpressionText(ctx, expression)),
			rule.RuleFixInsertAfter(node, fmt.Sprintf(" satisfies %s", cast)),
		},
	})

	return suggestions
}

func reportIncorrectAssertionType(ctx rule.RuleContext, node *ast.Node, options Options, isAsExpression bool) {
	var typeAnnotation *ast.Node
	if isAsExpression {
		typeAnnotation = node.AsAsExpression().Type
	} else {
		typeAnnotation = node.AsTypeAssertion().Type
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
	if options.AssertionStyle == "never" ||
		options.ObjectLiteralTypeAssertions == "allow" {
		return
	}

	var expression *ast.Node
	var typeAnnotation *ast.Node
	
	if isAsExpression {
		asExpr := node.AsAsExpression()
		expression = asExpr.Expression
		typeAnnotation = asExpr.Type
	} else {
		typeAssertion := node.AsTypeAssertion()
		expression = typeAssertion.Expression
		typeAnnotation = typeAssertion.Type
	}

	if !ast.IsObjectLiteralExpression(expression) {
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
	if options.AssertionStyle == "never" ||
		options.ArrayLiteralTypeAssertions == "allow" {
		return
	}

	var expression *ast.Node
	var typeAnnotation *ast.Node
	
	if isAsExpression {
		asExpr := node.AsAsExpression()
		expression = asExpr.Expression
		typeAnnotation = asExpr.Type
	} else {
		typeAssertion := node.AsTypeAssertion()
		expression = typeAssertion.Expression
		typeAnnotation = typeAssertion.Type
	}

	if !ast.IsArrayLiteralExpression(expression) {
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
		if options != nil {
			if bytes, err := json.Marshal(options); err == nil {
				json.Unmarshal(bytes, &opts)
			}
		}

		return rule.RuleListeners{
			ast.KindAsExpression: func(node *ast.Node) {
				if opts.AssertionStyle != "as" {
					reportIncorrectAssertionType(ctx, node, opts, true)
					return
				}

				checkExpressionForObjectAssertion(ctx, node, opts, true)
				checkExpressionForArrayAssertion(ctx, node, opts, true)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				if opts.AssertionStyle != "angle-bracket" {
					reportIncorrectAssertionType(ctx, node, opts, false)
					return
				}

				checkExpressionForObjectAssertion(ctx, node, opts, false)
				checkExpressionForArrayAssertion(ctx, node, opts, false)
			},
		}
	},
}