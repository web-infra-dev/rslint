package no_adjacent_inline_elements

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// inlineNames mirrors eslint-plugin-react's list of HTML inline elements.
var inlineNames = map[string]struct{}{
	"a": {}, "abbr": {}, "acronym": {}, "b": {}, "bdo": {}, "big": {},
	"br": {}, "button": {}, "cite": {}, "code": {}, "dfn": {}, "em": {},
	"i": {}, "img": {}, "input": {}, "kbd": {}, "label": {}, "map": {},
	"object": {}, "q": {}, "samp": {}, "script": {}, "select": {},
	"small": {}, "span": {}, "strong": {}, "sub": {}, "sup": {},
	"textarea": {}, "time": {}, "tt": {}, "var": {},
}

const inlineElementMessage = "Child elements which render as inline HTML elements should be separated by a space or wrapped in block level elements."

func hasBoundaryWhitespace(value string) bool {
	if value == "" {
		return false
	}
	var first, last rune
	haveFirst := false
	for _, r := range value {
		if !haveFirst {
			first = r
			haveFirst = true
		}
		last = r
	}
	return ecmascript.IsWhiteSpaceOrLineTerminator(first) || ecmascript.IsWhiteSpaceOrLineTerminator(last)
}

// jsxElementName returns the simple tag name accepted by upstream's
// `openingElement.name.name` lookup. Member and namespaced JSX tags therefore
// do not count as inline HTML elements.
func jsxElementName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	var tagName *ast.Node
	switch node.Kind {
	case ast.KindJsxElement:
		opening := node.AsJsxElement().OpeningElement
		if opening == nil || opening.Kind != ast.KindJsxOpeningElement {
			return ""
		}
		tagName = opening.AsJsxOpeningElement().TagName
	case ast.KindJsxSelfClosingElement:
		tagName = node.AsJsxSelfClosingElement().TagName
	default:
		return ""
	}
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return ""
	}
	return tagName.AsIdentifier().Text
}

func isInline(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return !hasBoundaryWhitespace(node.AsStringLiteral().Text)
	case ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		// ESTree represents all of these as Literal nodes. Upstream's
		// whitespace check coerces their values to strings, which never
		// produces boundary whitespace for these literal kinds.
		return true
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		_, ok := inlineNames[jsxElementName(node)]
		return ok
	case ast.KindCallExpression:
		// typescript-eslint wraps optional calls and optional-member calls in a
		// ChainExpression, so eslint-plugin-react's astUtil.isCallExpression
		// does not classify them as inline children. ts-go keeps them as
		// KindCallExpression with an optional-chain flag.
		if ast.IsOptionalChain(node) {
			return false
		}
		call := node.AsCallExpression()
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			// Upstream's `node.arguments[0].value` assumes a first argument.
			// A malformed or user-written zero-argument child must not crash the
			// linter; it is simply not an inline element.
			return false
		}
		first := ast.SkipParentheses(call.Arguments.Nodes[0])
		if first == nil || first.Kind != ast.KindStringLiteral {
			return false
		}
		_, ok := inlineNames[first.AsStringLiteral().Text]
		return ok
	default:
		return false
	}
}

func validate(ctx rule.RuleContext, node *ast.Node, children []*ast.Node) {
	if len(children) == 0 {
		return
	}
	previousIsInline := false
	for _, child := range children {
		currentIsInline := isInline(child)
		if previousIsInline && currentIsInline {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "inlineElement",
				Description: inlineElementMessage,
			})
			return
		}
		previousIsInline = currentIsInline
	}
}

var NoAdjacentInlineElementsRule = rule.Rule{
	Name:   "react/no-adjacent-inline-elements",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		return rule.RuleListeners{
			ast.KindJsxElement: func(node *ast.Node) {
				jsx := node.AsJsxElement()
				if jsx.Children == nil {
					return
				}
				validate(ctx, node, jsx.Children.Nodes)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if ctx.TypeChecker == nil && callee != nil && callee.Kind == ast.KindIdentifier {
					// The syntax-only import fallback cannot distinguish a bare
					// createElement binding from a same-named local shadow.
					return
				}
				if !reactutil.IsCreateElementCallWithChecker(call.Expression, pragma, ctx.TypeChecker) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) < 3 {
					return
				}
				children := ast.SkipParentheses(call.Arguments.Nodes[2])
				if children == nil || children.Kind != ast.KindArrayLiteralExpression {
					return
				}
				array := children.AsArrayLiteralExpression()
				if array.Elements == nil {
					return
				}
				validate(ctx, node, array.Elements.Nodes)
			},
		}
	},
}
