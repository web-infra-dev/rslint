package consistent_type_assertions

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_type_assertions.schema.json
var schemaJSON []byte

type AssertionStyle string

const (
	AssertionStyleAs           AssertionStyle = "as"
	AssertionStyleAngleBracket AssertionStyle = "angle-bracket"
	AssertionStyleNever        AssertionStyle = "never"
)

type LiteralAssertion string

const (
	LiteralAssertionAllow        LiteralAssertion = "allow"
	LiteralAssertionNever        LiteralAssertion = "never"
	LiteralAssertionAllowAsParam LiteralAssertion = "allow-as-parameter"
)

type ConsistentTypeAssertionsOptions struct {
	AssertionStyle              AssertionStyle   `json:"assertionStyle"`
	ObjectLiteralTypeAssertions LiteralAssertion `json:"objectLiteralTypeAssertions"`
	ArrayLiteralTypeAssertions  LiteralAssertion `json:"arrayLiteralTypeAssertions"`
}

const (
	objectAnnotationSuggestionID       = "replaceObjectTypeAssertionWithAnnotation"
	objectAnnotationSuggestionTemplate = "Use const x: {{cast}} = { ... } instead."
	objectSatisfiesSuggestionID        = "replaceObjectTypeAssertionWithSatisfies"
	objectSatisfiesSuggestionTemplate  = "Use const x = { ... } satisfies {{cast}} instead."
	arrayAnnotationSuggestionID        = "replaceArrayTypeAssertionWithAnnotation"
	arrayAnnotationSuggestionTemplate  = "Use const x: {{cast}} = [ ... ] instead."
	arraySatisfiesSuggestionID         = "replaceArrayTypeAssertionWithSatisfies"
	arraySatisfiesSuggestionTemplate   = "Use const x = [ ... ] satisfies {{cast}} instead."
)

type literalAssertionMessages struct {
	diagnosticID                     string
	annotationOrSatisfiesDescription string
	satisfiesDescription             string
	annotationSuggestionID           string
	annotationSuggestionTemplate     string
	satisfiesSuggestionID            string
	satisfiesSuggestionTemplate      string
}

var (
	objectLiteralAssertionMessages = literalAssertionMessages{
		diagnosticID:                     "unexpectedObjectTypeAssertion",
		annotationOrSatisfiesDescription: "Use a type annotation or the `satisfies` operator instead of a type assertion for object literals.",
		satisfiesDescription:             "Use the `satisfies` operator instead of a type assertion for object literals.",
		annotationSuggestionID:           objectAnnotationSuggestionID,
		annotationSuggestionTemplate:     objectAnnotationSuggestionTemplate,
		satisfiesSuggestionID:            objectSatisfiesSuggestionID,
		satisfiesSuggestionTemplate:      objectSatisfiesSuggestionTemplate,
	}
	arrayLiteralAssertionMessages = literalAssertionMessages{
		diagnosticID:                     "unexpectedArrayTypeAssertion",
		annotationOrSatisfiesDescription: "Use a type annotation or the `satisfies` operator instead of a type assertion for array literals.",
		satisfiesDescription:             "Use the `satisfies` operator instead of a type assertion for array literals.",
		annotationSuggestionID:           arrayAnnotationSuggestionID,
		annotationSuggestionTemplate:     arrayAnnotationSuggestionTemplate,
		satisfiesSuggestionID:            arraySatisfiesSuggestionID,
		satisfiesSuggestionTemplate:      arraySatisfiesSuggestionTemplate,
	}
)

var asExpressionPrecedence = ast.GetOperatorPrecedence(
	ast.KindAsExpression,
	ast.KindUnknown,
	ast.OperatorPrecedenceFlagsNone,
)

type consistentTypeAssertionsEdits struct {
	sourceFile *ast.SourceFile
}

// skipParensUp returns the first non-parenthesis ancestor of node. ESTree has
// no parenthesis nodes, so upstream's node.parent is exactly this.
func skipParensUp(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// isParameterPosition mirrors upstream isAsParameter after skipParensUp has
// reconciled tsgo's explicit parenthesis nodes with ESTree.
func isParameterPosition(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindNewExpression, ast.KindCallExpression, ast.KindThrowStatement,
		ast.KindJsxExpression, ast.KindParameter, ast.KindBindingElement,
		ast.KindShorthandPropertyAssignment:
		// Parameter / BindingElement / ShorthandPropertyAssignment cover the
		// ESTree AssignmentPattern default-value cases; the assertion can only
		// be the default/initializer in those nodes.
		return true
	case ast.KindTemplateSpan:
		// Only tagged templates count: the substitution's TemplateExpression
		// must itself be the argument of a TaggedTemplateExpression.
		return parent.Parent != nil && parent.Parent.Parent != nil &&
			parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	}
	return false
}

// annotationTarget returns the binding after which upstream can insert a type
// annotation. Keeping this decision outside the deferred builder lets the
// diagnostic text describe exactly the suggestions that can be produced.
func annotationTarget(parent *ast.Node) *ast.Node {
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := parent.AsVariableDeclaration()
	if declaration == nil || declaration.Type != nil {
		return nil
	}
	return declaration.Name()
}

func isConstType(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}
	typeReference := typeNode.AsTypeReferenceNode()
	return typeReference != nil &&
		typeReference.TypeName != nil &&
		typeReference.TypeName.Kind == ast.KindIdentifier &&
		typeReference.TypeName.AsIdentifier().Text == "const"
}

// shouldReportLiteralType mirrors upstream checkType: only bare any/unknown
// keywords and const assertions are exempt.
func shouldReportLiteralType(typeNode *ast.Node) bool {
	switch typeNode.Kind {
	case ast.KindAnyKeyword, ast.KindUnknownKeyword:
		return false
	case ast.KindTypeReference:
		return !isConstType(typeNode)
	default:
		return true
	}
}

func substituteCast(template, cast string) string {
	return strings.ReplaceAll(template, "{{cast}}", cast)
}

func wrappedCode(text string, nodePrecedence, parentPrecedence ast.OperatorPrecedence) string {
	if nodePrecedence > parentPrecedence {
		return text
	}
	return "(" + text + ")"
}

func (e consistentTypeAssertionsEdits) text(node *ast.Node) string {
	return utils.TrimmedNodeText(e.sourceFile, node)
}

// textWithParentheses mirrors upstream getTextWithParentheses: the text of the
// paren-stripped expression, wrapped in one pair of parentheses when it was
// parenthesized in source.
func (e consistentTypeAssertionsEdits) textWithParentheses(expression *ast.Node) string {
	inner := ast.SkipParentheses(expression)
	text := e.text(inner)
	if inner != expression {
		return "(" + text + ")"
	}
	return text
}

func (e consistentTypeAssertionsEdits) literalSuggestions(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
	annotationTarget *ast.Node,
	messages *literalAssertionMessages,
) []rule.RuleSuggestion {
	typeText := e.text(typeNode)
	expressionText := e.textWithParentheses(expression)
	suggestions := make([]rule.RuleSuggestion, 0, 2)
	if annotationTarget != nil {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          messages.annotationSuggestionID,
				Description: substituteCast(messages.annotationSuggestionTemplate, typeText),
				Data:        map[string]string{"cast": typeText},
			},
			FixesArr: []rule.RuleFix{
				rule.RuleFixInsertAfter(annotationTarget, ": "+typeText),
				rule.RuleFixReplace(e.sourceFile, node, expressionText),
			},
		})
	}
	suggestions = append(suggestions, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          messages.satisfiesSuggestionID,
			Description: substituteCast(messages.satisfiesSuggestionTemplate, typeText),
			Data:        map[string]string{"cast": typeText},
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(e.sourceFile, node, expressionText),
			rule.RuleFixInsertAfter(node, " satisfies "+typeText),
		},
	})
	return suggestions
}

// angleToAsFix converts an angle-bracket assertion into an as assertion,
// preserving the same precedence wrapping as the upstream fixer.
func (e consistentTypeAssertionsEdits) angleToAsFix(
	node *ast.Node,
	expression *ast.Node,
	typeText string,
) []rule.RuleFix {
	innerExpression := ast.SkipParentheses(expression)
	expressionText := wrappedCode(
		e.text(innerExpression),
		ast.GetExpressionPrecedence(innerExpression),
		asExpressionPrecedence,
	)
	text := expressionText + " as " + typeText

	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		return []rule.RuleFix{rule.RuleFixReplace(e.sourceFile, node, text)}
	}

	parentKind := ast.KindUnknown
	operatorKind := ast.KindUnknown
	flags := ast.OperatorPrecedenceFlagsNone
	if parent != nil {
		parentKind = parent.Kind
		switch parent.Kind {
		case ast.KindBinaryExpression:
			if operator := parent.AsBinaryExpression().OperatorToken; operator != nil {
				operatorKind = operator.Kind
			}
		case ast.KindNewExpression:
			newExpression := parent.AsNewExpression()
			if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) == 0 {
				flags = ast.OperatorPrecedenceFlagsNewWithoutArguments
			}
		}
	}
	parentPrecedence := ast.GetOperatorPrecedence(parentKind, operatorKind, flags)
	return []rule.RuleFix{
		rule.RuleFixReplace(
			e.sourceFile,
			node,
			wrappedCode(text, asExpressionPrecedence, parentPrecedence),
		),
	}
}

// ConsistentTypeAssertionsRule is the rslint port of upstream
// `@typescript-eslint/consistent-type-assertions`. It mirrors upstream v8
// observably: the same diagnostic locations and message ids, the same
// `annotation`/`satisfies` suggestions, and the same angle-bracket→as autofix.
// Literal-assertion descriptions intentionally name only replacements that
// are valid in the assertion's context; upstream always recommends a const
// declaration even when the assertion is not a variable initializer.
//
// AST note: upstream runs on ESTree, where parentheses are not nodes, so
// `node.expression` / `node.parent` transparently see through them. tsgo
// preserves `ParenthesizedExpression`, so the literal check skips parens on the
// expression (`ast.SkipParentheses`) and the parameter / variable-declarator
// lookups skip parens on the way up (`skipParensUp`). This is what makes
// `({ ... }) as T` — the parenthesized object literal — get detected, matching
// upstream.
var ConsistentTypeAssertionsRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-type-assertions",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)

	edits := consistentTypeAssertionsEdits{sourceFile: ctx.SourceFile}

	// Upstream implements object and array literal checks separately. Selecting
	// the literal once avoids duplicate parenthesis walks and keeps each
	// diagnostic coupled to the exact suggestions its context permits.
	checkLiteral := func(node, expression, typeNode *ast.Node) {
		if opts.AssertionStyle == AssertionStyleNever ||
			(opts.ObjectLiteralTypeAssertions == LiteralAssertionAllow &&
				opts.ArrayLiteralTypeAssertions == LiteralAssertionAllow) {
			return
		}

		var literalOption LiteralAssertion
		var messages *literalAssertionMessages
		switch ast.SkipParentheses(expression).Kind {
		case ast.KindObjectLiteralExpression:
			literalOption = opts.ObjectLiteralTypeAssertions
			messages = &objectLiteralAssertionMessages
		case ast.KindArrayLiteralExpression:
			literalOption = opts.ArrayLiteralTypeAssertions
			messages = &arrayLiteralAssertionMessages
		default:
			return
		}
		if literalOption == LiteralAssertionAllow {
			return
		}

		parent := skipParensUp(node)
		if literalOption == LiteralAssertionAllowAsParam && isParameterPosition(parent) {
			return
		}
		if !shouldReportLiteralType(typeNode) {
			return
		}

		annotationNode := annotationTarget(parent)
		description := messages.satisfiesDescription
		if annotationNode != nil {
			description = messages.annotationOrSatisfiesDescription
		}
		ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
			Id:          messages.diagnosticID,
			Description: description,
		}, func() []rule.RuleSuggestion {
			return edits.literalSuggestions(node, expression, typeNode, annotationNode, messages)
		})
	}

	// reportIncorrectAssertionType handles the assertion-style mismatch reports
	// (`as` / `angle-bracket` / `never`). Mirrors upstream
	// `reportIncorrectAssertionType`.
	reportIncorrectAssertionType := func(node, exprRaw, typeNode *ast.Node) {
		style := opts.AssertionStyle
		// `as const` / `<const>` is never reported under `never`.
		if style == AssertionStyleNever && isConstType(typeNode) {
			return
		}
		switch style {
		case AssertionStyleAs:
			cast := edits.text(typeNode)
			ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
				Id:          "as",
				Description: substituteCast("Use 'as {{cast}}' instead of '<{{cast}}>'.", cast),
				Data:        map[string]string{"cast": cast},
			}, func() []rule.RuleFix {
				return edits.angleToAsFix(node, exprRaw, cast)
			})
		case AssertionStyleAngleBracket:
			cast := edits.text(typeNode)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "angle-bracket",
				Description: substituteCast("Use '<{{cast}}>' instead of 'as {{cast}}'.", cast),
				Data:        map[string]string{"cast": cast},
			})
		case AssertionStyleNever:
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "never",
				Description: "Do not use any type assertions.",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindAsExpression: func(node *ast.Node) {
			asExpr := node.AsAsExpression()
			if asExpr == nil {
				return
			}
			if opts.AssertionStyle != AssertionStyleAs {
				reportIncorrectAssertionType(node, asExpr.Expression, asExpr.Type)
				return
			}
			checkLiteral(node, asExpr.Expression, asExpr.Type)
		},
		ast.KindTypeAssertionExpression: func(node *ast.Node) {
			typeAssertion := node.AsTypeAssertion()
			if typeAssertion == nil {
				return
			}
			if opts.AssertionStyle != AssertionStyleAngleBracket {
				reportIncorrectAssertionType(node, typeAssertion.Expression, typeAssertion.Type)
				return
			}
			checkLiteral(node, typeAssertion.Expression, typeAssertion.Type)
		},
	}
}

func parseOptions(options []any) ConsistentTypeAssertionsOptions {
	opts := ConsistentTypeAssertionsOptions{
		AssertionStyle:              AssertionStyleAs,
		ObjectLiteralTypeAssertions: LiteralAssertionAllow,
		ArrayLiteralTypeAssertions:  LiteralAssertionAllow,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})

	if v, exists := optsMap["assertionStyle"]; exists {
		if str, ok := v.(string); ok {
			opts.AssertionStyle = AssertionStyle(str)
		}
	}

	if v, exists := optsMap["objectLiteralTypeAssertions"]; exists {
		if str, ok := v.(string); ok {
			opts.ObjectLiteralTypeAssertions = LiteralAssertion(str)
		}
	}

	if v, exists := optsMap["arrayLiteralTypeAssertions"]; exists {
		if str, ok := v.(string); ok {
			opts.ArrayLiteralTypeAssertions = LiteralAssertion(str)
		}
	}

	return opts
}
