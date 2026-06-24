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

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
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

	src := ctx.SourceFile
	getText := func(n *ast.Node) string {
		return utils.TrimmedNodeText(src, n)
	}

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

	// skipParensUp returns the first non-parenthesis ancestor of `node`. ESTree
	// has no parenthesis nodes, so upstream's `node.parent` is exactly this.
	skipParensUp := func(node *ast.Node) *ast.Node {
		p := node.Parent
		for p != nil && p.Kind == ast.KindParenthesizedExpression {
			p = p.Parent
		}
		return p
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

	// asPrecedence is the precedence of an `x as T` expression — the pivot the
	// angle-bracket→as autofix uses to decide where wrapping parens are needed.
	asPrecedence := ast.GetOperatorPrecedence(ast.KindAsExpression, ast.KindUnknown, ast.OperatorPrecedenceFlagsNone)

	// getWrappedCode mirrors upstream's helper of the same name: wrap `text` in
	// parens unless its precedence is strictly higher than the surrounding
	// context's.
	getWrappedCode := func(text string, nodePrec, parentPrec ast.OperatorPrecedence) string {
		if nodePrec > parentPrec {
			return text
		}
		return "(" + text + ")"
	}

	subst := func(tmpl, cast string) string {
		return strings.ReplaceAll(tmpl, "{{cast}}", cast)
	}

	// textWithParentheses mirrors upstream `getTextWithParentheses`: the text of
	// the (paren-stripped) expression, wrapped in a single pair of parens when
	// it was parenthesized in source. ESTree exposes only the inner node, so the
	// wrapping is always exactly one level regardless of nesting depth.
	textWithParentheses := func(exprRaw *ast.Node) string {
		inner := ast.SkipParentheses(exprRaw)
		t := getText(inner)
		if inner != exprRaw {
			return "(" + t + ")"
		}
		return t
	}

	// buildSuggestions mirrors upstream `getSuggestions`: an annotation
	// suggestion (only when the assertion initializes a variable with no type
	// annotation) plus a `satisfies` suggestion (always). `exprRaw` is the raw
	// expression (parens preserved) — upstream uses `getTextWithParentheses`.
	buildSuggestions := func(node, exprRaw, typeNode *ast.Node, annID, annTmpl, satID, satTmpl string) []rule.RuleSuggestion {
		typeText := getText(typeNode)
		exprText := textWithParentheses(exprRaw)
		suggestions := make([]rule.RuleSuggestion, 0, 2)
		if eff := skipParensUp(node); eff != nil && eff.Kind == ast.KindVariableDeclaration {
			if vd := eff.AsVariableDeclaration(); vd.Type == nil {
				if name := vd.Name(); name != nil {
					suggestions = append(suggestions, rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:          annID,
							Description: subst(annTmpl, typeText),
							Data:        map[string]string{"cast": typeText},
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixInsertAfter(name, ": "+typeText),
							rule.RuleFixReplace(src, node, exprText),
						},
					})
				}
			}
		}
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          satID,
				Description: subst(satTmpl, typeText),
				Data:        map[string]string{"cast": typeText},
			},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplace(src, node, exprText),
				rule.RuleFixInsertAfter(node, " satisfies "+typeText),
			},
		})
		return suggestions
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
			ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
				Id:          "unexpectedObjectTypeAssertion",
				Description: "Always prefer const x: T = { ... }.",
			}, buildSuggestions(node, exprRaw, typeNode,
				"replaceObjectTypeAssertionWithAnnotation", "Use const x: {{cast}} = { ... } instead.",
				"replaceObjectTypeAssertionWithSatisfies", "Use const x = { ... } satisfies {{cast}} instead.")...)
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
			ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
				Id:          "unexpectedArrayTypeAssertion",
				Description: "Always prefer const x: T[] = [ ... ].",
			}, buildSuggestions(node, exprRaw, typeNode,
				"replaceArrayTypeAssertionWithAnnotation", "Use const x: {{cast}} = [ ... ] instead.",
				"replaceArrayTypeAssertionWithSatisfies", "Use const x = [ ... ] satisfies {{cast}} instead.")...)
		}
	}

	// buildAngleToAsFix converts an angle-bracket assertion `<T>expr` into
	// `expr as T`, wrapping the expression and/or the whole assertion in parens
	// exactly where precedence requires. Mirrors upstream's `as` fixer.
	buildAngleToAsFix := func(node, exprRaw, typeNode *ast.Node) rule.RuleFix {
		skipExpr := ast.SkipParentheses(exprRaw)
		exprText := getWrappedCode(getText(skipExpr), ast.GetExpressionPrecedence(skipExpr), asPrecedence)
		text := exprText + " as " + getText(typeNode)

		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			// Already parenthesized in source — the existing parens suffice.
			return rule.RuleFixReplace(src, node, text)
		}

		parentKind := ast.KindUnknown
		operatorKind := ast.KindUnknown
		flags := ast.OperatorPrecedenceFlagsNone
		if parent != nil {
			parentKind = parent.Kind
			switch parent.Kind {
			case ast.KindBinaryExpression:
				if op := parent.AsBinaryExpression().OperatorToken; op != nil {
					operatorKind = op.Kind
				}
			case ast.KindNewExpression:
				ne := parent.AsNewExpression()
				if ne.Arguments == nil || len(ne.Arguments.Nodes) == 0 {
					flags = ast.OperatorPrecedenceFlagsNewWithoutArguments
				}
			}
		}
		parentPrec := ast.GetOperatorPrecedence(parentKind, operatorKind, flags)
		return rule.RuleFixReplace(src, node, getWrappedCode(text, asPrecedence, parentPrec))
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
			cast := getText(typeNode)
			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "as",
				Description: subst("Use 'as {{cast}}' instead of '<{{cast}}>'.", cast),
				Data:        map[string]string{"cast": cast},
			}, buildAngleToAsFix(node, exprRaw, typeNode))
		case AssertionStyleAngleBracket:
			cast := getText(typeNode)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "angle-bracket",
				Description: subst("Use '<{{cast}}>' instead of 'as {{cast}}'.", cast),
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
