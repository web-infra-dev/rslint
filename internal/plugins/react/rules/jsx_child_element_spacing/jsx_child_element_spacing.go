package jsx_child_element_spacing

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// inlineElements mirrors upstream's INLINE_ELEMENTS set
// (https://developer.mozilla.org/en-US/docs/Web/HTML/Inline_elements). `br` is
// intentionally excluded because surrounding whitespace has no rendered effect.
var inlineElements = map[string]struct{}{
	"a": {}, "abbr": {}, "acronym": {}, "b": {}, "bdo": {}, "big": {},
	"button": {}, "cite": {}, "code": {}, "dfn": {}, "em": {}, "i": {},
	"img": {}, "input": {}, "kbd": {}, "label": {}, "map": {}, "object": {},
	"q": {}, "samp": {}, "script": {}, "select": {}, "small": {}, "span": {},
	"strong": {}, "sub": {}, "sup": {}, "textarea": {}, "tt": {}, "var": {},
}

var (
	textFollowingElementPattern = regexp.MustCompile(`^\s*\n\s*\S`)
	textPrecedingElementPattern = regexp.MustCompile(`\S\s*\n\s*$`)
)

// elementName returns the simple-identifier tag name of a JSX element node, or
// "" when the node is not an element, has a non-Identifier tag (e.g. `Foo.Bar`,
// `<this>`, namespaced like `<a:b>`), or has no resolvable opening element.
//
// NOTE: Unlike ESLint, tsgo splits paired elements (`<a>...</a>` →
// KindJsxElement) and self-closing elements (`<a/>` → KindJsxSelfClosingElement)
// into two distinct kinds, while ESTree models both as JSXElement. The upstream
// rule's `isInlineElement` matches both shapes via `node.type === 'JSXElement'`,
// so we MUST recognize self-closing elements too — otherwise inline tags that
// are commonly self-closing (notably `<img/>` and `<input/>` from the inline
// set) would silently fall outside the inline check, causing false negatives.
func elementName(node *ast.Node) string {
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

func isInlineElement(node *ast.Node) bool {
	name := elementName(node)
	if name == "" {
		return false
	}
	_, ok := inlineElements[name]
	return ok
}

// handleJSX walks the children with a 3-element sliding window
// (lastChild, child, nextChild), matching the upstream
// `node.children.concat([null]).forEach((nextChild) => {...})` shape verbatim.
// The synthetic trailing nil drives the final iteration where `child` is the
// last real child and `nextChild` is nil — needed for spacingAfterPrev to fire
// when an inline element is the penultimate child.
func handleJSX(ctx rule.RuleContext, node *ast.Node) {
	children := reactutil.GetJsxChildren(node)
	if len(children) == 0 {
		return
	}
	var lastChild, child *ast.Node
	for i := 0; i <= len(children); i++ {
		var nextChild *ast.Node
		if i < len(children) {
			nextChild = children[i]
		}
		if (lastChild != nil || nextChild != nil) &&
			(lastChild == nil || isInlineElement(lastChild)) &&
			(child != nil && child.Kind == ast.KindJsxText) &&
			(nextChild == nil || isInlineElement(nextChild)) {
			text := child.AsJsxText().Text
			if lastChild != nil && textFollowingElementPattern.MatchString(text) {
				pos := lastChild.End()
				ctx.ReportRange(core.NewTextRange(pos, pos), rule.RuleMessage{
					Id:          "spacingAfterPrev",
					Description: "Ambiguous spacing after previous element " + elementName(lastChild),
				})
			} else if nextChild != nil && textPrecedingElementPattern.MatchString(text) {
				pos := nextChild.Pos()
				ctx.ReportRange(core.NewTextRange(pos, pos), rule.RuleMessage{
					Id:          "spacingBeforeNext",
					Description: "Ambiguous spacing before next element " + elementName(nextChild),
				})
			}
		}
		lastChild = child
		child = nextChild
	}
}

var JsxChildElementSpacingRule = rule.Rule{
	Name: "react/jsx-child-element-spacing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxElement:  func(node *ast.Node) { handleJSX(ctx, node) },
			ast.KindJsxFragment: func(node *ast.Node) { handleJSX(ctx, node) },
		}
	},
}
