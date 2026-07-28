package consistent_type_assertions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
	annotationID string,
	annotationTemplate string,
	satisfiesID string,
	satisfiesTemplate string,
) []rule.RuleSuggestion {
	typeText := e.text(typeNode)
	expressionText := e.textWithParentheses(expression)
	suggestions := make([]rule.RuleSuggestion, 0, 2)
	if effectiveParent := skipParensUp(node); effectiveParent != nil && effectiveParent.Kind == ast.KindVariableDeclaration {
		if declaration := effectiveParent.AsVariableDeclaration(); declaration.Type == nil {
			if name := declaration.Name(); name != nil {
				suggestions = append(suggestions, rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          annotationID,
						Description: substituteCast(annotationTemplate, typeText),
						Data:        map[string]string{"cast": typeText},
					},
					FixesArr: []rule.RuleFix{
						rule.RuleFixInsertAfter(name, ": "+typeText),
						rule.RuleFixReplace(e.sourceFile, node, expressionText),
					},
				})
			}
		}
	}
	suggestions = append(suggestions, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          satisfiesID,
			Description: substituteCast(satisfiesTemplate, typeText),
			Data:        map[string]string{"cast": typeText},
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(e.sourceFile, node, expressionText),
			rule.RuleFixInsertAfter(node, " satisfies "+typeText),
		},
	})
	return suggestions
}

func (e consistentTypeAssertionsEdits) objectLiteralSuggestions(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
) []rule.RuleSuggestion {
	return e.literalSuggestions(
		node,
		expression,
		typeNode,
		objectAnnotationSuggestionID,
		objectAnnotationSuggestionTemplate,
		objectSatisfiesSuggestionID,
		objectSatisfiesSuggestionTemplate,
	)
}

func (e consistentTypeAssertionsEdits) arrayLiteralSuggestions(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
) []rule.RuleSuggestion {
	return e.literalSuggestions(
		node,
		expression,
		typeNode,
		arrayAnnotationSuggestionID,
		arrayAnnotationSuggestionTemplate,
		arraySatisfiesSuggestionID,
		arraySatisfiesSuggestionTemplate,
	)
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
// observably: the same diagnostics (message ids + texts), the same
// `annotation`/`satisfies` suggestions, and the same angle-bracket→as autofix.
//
// AST note: upstream runs on ESTree, where parentheses are not nodes, so
// `node.expression` / `node.parent` transparently see through them. tsgo
// preserves `ParenthesizedExpression`, so the literal check skips parens on the
// expression (`ast.SkipParentheses`) and the parameter / variable-declarator
// lookups skip parens on the way up (`skipParensUp`). This is what makes
// `({ ... }) as T` — the parenthesized object literal — get detected, matching
// upstream.
var ConsistentTypeAssertionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-assertions",
	Run:  run,
})

func run(ctx rule.RuleContext, _options []any) rule.RuleListeners {
	options := rule.LegacyUnwrapOptions(_options)
	opts := ConsistentTypeAssertionsOptions{
		AssertionStyle:              AssertionStyleAs,
		ObjectLiteralTypeAssertions: LiteralAssertionAllow,
		ArrayLiteralTypeAssertions:  LiteralAssertionAllow,
	}
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optsMap, ok := optArray[0].(map[string]interface{}); ok {
				parseOptionsMap(optsMap, &opts)
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			parseOptionsMap(optsMap, &opts)
		}
	}

	edits := consistentTypeAssertionsEdits{sourceFile: ctx.SourceFile}

	// isConst reports whether the assertion target is `const` (the `as const` /
	// `<const>` assertion). Mirrors upstream `isConst`.
	isConst := func(typeNode *ast.Node) bool {
		if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
			return false
		}
		tr := typeNode.AsTypeReferenceNode()
		if tr == nil || tr.TypeName == nil || tr.TypeName.Kind != ast.KindIdentifier {
			return false
		}
		return tr.TypeName.AsIdentifier().Text == "const"
	}

	// checkType reports whether a literal assertion to `typeNode` should be
	// flagged. Mirrors upstream `checkType`: only the *bare* `any` / `unknown`
	// keywords are exempt (a union such as `any | string` is NOT), and `as
	// const` is exempt.
	checkType := func(typeNode *ast.Node) bool {
		switch typeNode.Kind {
		case ast.KindAnyKeyword, ast.KindUnknownKeyword:
			return false
		case ast.KindTypeReference:
			return !isConst(typeNode)
		default:
			return true
		}
	}

	// isAsParameter mirrors upstream `isAsParameter`: the assertion is in a
	// position where an object/array literal is "passed" rather than assigned to
	// a typeable binding (call / new / throw arguments, JSX expression
	// containers, default values, tagged-template substitutions).
	isAsParameter := func(node *ast.Node) bool {
		p := skipParensUp(node)
		if p == nil {
			return false
		}
		switch p.Kind {
		case ast.KindNewExpression, ast.KindCallExpression, ast.KindThrowStatement,
			ast.KindJsxExpression, ast.KindParameter, ast.KindBindingElement,
			ast.KindShorthandPropertyAssignment:
			// Parameter / BindingElement / ShorthandPropertyAssignment cover the
			// ESTree `AssignmentPattern` default-value cases; the assertion can
			// only be the default/initializer in those nodes.
			return true
		case ast.KindTemplateSpan:
			// Only tagged templates count: the substitution's TemplateExpression
			// must itself be the argument of a TaggedTemplateExpression.
			return p.Parent != nil && p.Parent.Parent != nil &&
				p.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
		}
		return false
	}

	// checkObject / checkArray mirror upstream's checkExpressionForObjectAssertion
	// / checkExpressionForArrayAssertion. `exprRaw` is the assertion's raw
	// expression; the literal check uses the paren-stripped form.
	checkObject := func(node, exprRaw, typeNode *ast.Node) {
		expr := ast.SkipParentheses(exprRaw)
		if opts.AssertionStyle == AssertionStyleNever ||
			opts.ObjectLiteralTypeAssertions == LiteralAssertionAllow ||
			expr.Kind != ast.KindObjectLiteralExpression {
			return
		}
		if opts.ObjectLiteralTypeAssertions == LiteralAssertionAllowAsParam && isAsParameter(node) {
			return
		}
		if checkType(typeNode) {
			ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
				Id:          "unexpectedObjectTypeAssertion",
				Description: "Always prefer const x: T = { ... }.",
			}, func() []rule.RuleSuggestion {
				return edits.objectLiteralSuggestions(node, exprRaw, typeNode)
			})
		}
	}

	checkArray := func(node, exprRaw, typeNode *ast.Node) {
		expr := ast.SkipParentheses(exprRaw)
		if opts.AssertionStyle == AssertionStyleNever ||
			opts.ArrayLiteralTypeAssertions == LiteralAssertionAllow ||
			expr.Kind != ast.KindArrayLiteralExpression {
			return
		}
		if opts.ArrayLiteralTypeAssertions == LiteralAssertionAllowAsParam && isAsParameter(node) {
			return
		}
		if checkType(typeNode) {
			ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{
				Id:          "unexpectedArrayTypeAssertion",
				Description: "Always prefer const x: T[] = [ ... ].",
			}, func() []rule.RuleSuggestion {
				return edits.arrayLiteralSuggestions(node, exprRaw, typeNode)
			})
		}
	}

	// reportIncorrectAssertionType handles the assertion-style mismatch reports
	// (`as` / `angle-bracket` / `never`). Mirrors upstream
	// `reportIncorrectAssertionType`.
	reportIncorrectAssertionType := func(node, exprRaw, typeNode *ast.Node) {
		style := opts.AssertionStyle
		// `as const` / `<const>` is never reported under `never`.
		if style == AssertionStyleNever && isConst(typeNode) {
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
			checkObject(node, asExpr.Expression, asExpr.Type)
			checkArray(node, asExpr.Expression, asExpr.Type)
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
			checkObject(node, typeAssertion.Expression, typeAssertion.Type)
			checkArray(node, typeAssertion.Expression, typeAssertion.Type)
		},
	}
}

func parseOptionsMap(optsMap map[string]interface{}, opts *ConsistentTypeAssertionsOptions) {
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
}
