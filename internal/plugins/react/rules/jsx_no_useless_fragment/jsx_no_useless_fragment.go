package jsx_no_useless_fragment

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed jsx_no_useless_fragment.schema.json
var schemaJSON []byte

const (
	needsMoreChildrenDescription  = "Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all."
	childOfHtmlElementDescription = "Passing a fragment to an HTML element is useless."
)

var JsxNoUselessFragmentRule = rule.Rule{
	Name:   "react/jsx-no-useless-fragment",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowExpressions := parseAllowExpressions(options)
		reactPragma := reactutil.GetReactPragmaFromContext(ctx)
		fragmentPragma := reactutil.GetReactFragmentPragma(ctx.Settings)
		sourceText := ctx.SourceFile.Text()

		// rawText returns a JsxText child's source spelling — upstream reads
		// `node.raw`, so entity references and the exact whitespace run must
		// come from the source, not from tsgo's decoded `.Text` field.
		rawText := func(node *ast.Node) string {
			if node == nil {
				return ""
			}
			return sourceText[node.Pos():node.End()]
		}

		// isPaddingSpaces mirrors upstream's helper of the same name: a
		// whitespace-only JSX text run that the React runtime strips, which
		// upstream detects by the presence of a literal "\n" (not any
		// LineTerminator — `arrayIncludes(node.raw, '\n')` compares
		// characters).
		isPaddingSpaces := func(node *ast.Node) bool {
			if node == nil || node.Kind != ast.KindJsxText {
				return false
			}
			raw := rawText(node)
			return ecmascript.IsBlank(raw) && strings.Contains(raw, "\n")
		}

		// nonPaddingChildSummary returns the first relevant child and a count
		// capped at two. Every downstream question distinguishes only zero, one,
		// or multiple children, so retaining a slice would allocate for each
		// visited fragment with children without carrying extra information.
		nonPaddingChildSummary := func(node *ast.Node) (int, *ast.Node) {
			children := reactutil.GetJsxChildren(node)
			var first *ast.Node
			count := 0
			for _, child := range children {
				if isPaddingSpaces(child) {
					continue
				}
				if count == 0 {
					first = child
				}
				count++
				if count == 2 {
					break
				}
			}
			return count, first
		}

		isChildOfHtmlElement := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindJsxElement {
				return false
			}
			tagName := reactutil.GetJsxTagName(parent.AsJsxElement().OpeningElement)
			if tagName == nil {
				return false
			}
			var name string
			switch tagName.Kind {
			case ast.KindIdentifier:
				name = tagName.AsIdentifier().Text
			case ast.KindThisKeyword:
				// ESTree represents the legal `<this>` tag as JSXIdentifier("this").
				name = "this"
			default:
				return false
			}
			return isAllLowercaseLetters(name)
		}

		isChildOfComponentElement := func(node *ast.Node) bool {
			parent := node.Parent
			return parent != nil && parent.Kind == ast.KindJsxElement &&
				!isChildOfHtmlElement(node) &&
				!reactutil.IsFragmentTag(parent, reactPragma, fragmentPragma)
		}

		isNonSpaceJsxTextOrJsxCurly := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			if node.Kind == ast.KindJsxText {
				return !ecmascript.IsBlank(rawText(node))
			}
			return isExpressionContainer(node)
		}

		canFix := func(node *ast.Node) bool {
			parent := node.Parent
			parentIsJsxContainer := parent != nil &&
				(parent.Kind == ast.KindJsxElement || parent.Kind == ast.KindJsxFragment)

			// Not safe to fix fragments without a jsx parent.
			if !parentIsJsxContainer {
				children := reactutil.GetJsxChildren(node)

				// const a = <></>
				if len(children) == 0 {
					return false
				}

				// const a = <>cat {meow}</>
				for _, child := range children {
					if isNonSpaceJsxTextOrJsxCurly(child) {
						return false
					}
				}
			}

			// Not safe to fix `<Eeee><>foo</></Eeee>` because `Eeee` might
			// require its children be a ReactElement.
			if isChildOfComponentElement(node) {
				return false
			}

			// Upstream guards the "old TS parser" shape where a JSXFragment
			// carries no opening / closing fragment token. tsgo always parses
			// both, so this arm is defensive only.
			if node.Kind == ast.KindJsxFragment {
				fragment := node.AsJsxFragment()
				if fragment.OpeningFragment == nil || fragment.ClosingFragment == nil {
					return false
				}
			}

			return true
		}

		buildFix := func(node *ast.Node) []rule.RuleFix {
			if !canFix(node) {
				return nil
			}

			var childrenStart, childrenEnd int
			switch node.Kind {
			case ast.KindJsxFragment:
				fragment := node.AsJsxFragment()
				childrenStart = fragment.OpeningFragment.End()
				childrenEnd = fragment.ClosingFragment.Pos()
			case ast.KindJsxElement:
				element := node.AsJsxElement()
				if element.OpeningElement == nil || element.ClosingElement == nil {
					return nil
				}
				childrenStart = element.OpeningElement.End()
				childrenEnd = element.ClosingElement.Pos()
			case ast.KindJsxSelfClosingElement:
				// Upstream's `opener.selfClosing ? '' : ...` arm — a
				// self-closing fragment element has no children text.
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "")}
			default:
				return nil
			}

			return []rule.RuleFix{
				rule.RuleFixReplace(ctx.SourceFile, node, trimLikeReact(sourceText[childrenStart:childrenEnd])),
			}
		}

		checkNode := func(node *ast.Node) {
			if isKeyedElement(node) {
				return
			}

			nonPaddingChildCount, firstNonPaddingChild := nonPaddingChildSummary(node)
			allowedSingleExpression := allowExpressions &&
				nonPaddingChildCount == 1 &&
				isExpressionContainer(firstNonPaddingChild)
			if nonPaddingChildCount < 2 &&
				!containsCallExpression(firstNonPaddingChild) &&
				!isFragmentWithOnlyTextAndIsNotChild(node) &&
				!allowedSingleExpression {
				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          "NeedsMoreChildren",
					Description: needsMoreChildrenDescription,
				}, func() []rule.RuleFix { return buildFix(node) })
			}

			if isChildOfHtmlElement(node) {
				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          "ChildOfHtmlElement",
					Description: childOfHtmlElementDescription,
				}, func() []rule.RuleFix { return buildFix(node) })
			}
		}

		checkFragmentElement := func(node *ast.Node) {
			if reactutil.IsFragmentTag(node, reactPragma, fragmentPragma) {
				checkNode(node)
			}
		}

		return rule.RuleListeners{
			// ESTree models `<Fragment>…</Fragment>` and `<Fragment />` as one
			// JSXElement kind; tsgo splits them, so both listeners are needed
			// to cover upstream's single `JSXElement` visitor.
			ast.KindJsxElement:            checkFragmentElement,
			ast.KindJsxSelfClosingElement: checkFragmentElement,
			ast.KindJsxFragment:           checkNode,
		}
	},
}

func parseAllowExpressions(options []any) bool {
	if len(options) == 0 {
		return false
	}
	config, ok := options[0].(map[string]any)
	if !ok {
		return false
	}
	allow, _ := config["allowExpressions"].(bool)
	return allow
}

// isExpressionContainer reports whether a JSX child is an ESTree
// JSXExpressionContainer — `{expr}`. tsgo parses a JSX spread child
// (`{...expr}`) into the same KindJsxExpression, distinguished only by
// DotDotDotToken, whereas ESTree gives it a separate JSXSpreadChild type that
// upstream's `type === 'JSXExpressionContainer'` tests all reject. Every
// container-specific check therefore has to exclude the spread form.
func isExpressionContainer(node *ast.Node) bool {
	return node != nil &&
		node.Kind == ast.KindJsxExpression &&
		node.AsJsxExpression().DotDotDotToken == nil
}

// containsCallExpression mirrors upstream's helper: a JSX expression container
// whose expression is a plain CallExpression. Parentheses are transparent
// because ESTree has no ParenthesizedExpression node, while three call-shaped
// forms are NOT matches, because ESTree gives each of them its own node type
// that upstream's `type === 'CallExpression'` check rejects:
// an optional call (`{foo?.()}`, a ChainExpression), a dynamic import
// (`{import("x")}`, an ImportExpression), and a spread child (`{...make()}`,
// a JSXSpreadChild). tsgo collapses the first two into KindCallExpression and
// the third into KindJsxExpression, so each needs an explicit exclusion.
func containsCallExpression(node *ast.Node) bool {
	if !isExpressionContainer(node) {
		return false
	}
	// `{}` — an empty JSX expression container — has no expression at all.
	if node.AsJsxExpression().Expression == nil {
		return false
	}
	expression := ast.SkipParentheses(node.AsJsxExpression().Expression)
	if expression == nil {
		return false
	}
	return expression.Kind == ast.KindCallExpression &&
		!ast.IsOptionalChain(expression) &&
		!ast.IsImportCall(expression)
}

// isFragmentWithOnlyTextAndIsNotChild mirrors upstream: a fragment like
// `<Foo content={<>ee eeee eeee …</>} />` is genuinely useful, so a fragment
// holding a single text child outside of any JSX parent is left alone.
func isFragmentWithOnlyTextAndIsNotChild(node *ast.Node) bool {
	children := reactutil.GetJsxChildren(node)
	if len(children) != 1 || children[0].Kind != ast.KindJsxText {
		return false
	}
	parent := node.Parent
	return parent == nil ||
		(parent.Kind != ast.KindJsxElement && parent.Kind != ast.KindJsxFragment)
}

// isKeyedElement mirrors upstream's `<Fragment key={_}>_</Fragment>` test — a
// keyed fragment is meaningful and never reported.
func isKeyedElement(node *ast.Node) bool {
	if node.Kind != ast.KindJsxElement && node.Kind != ast.KindJsxSelfClosingElement {
		return false
	}
	opening := node
	if node.Kind == ast.KindJsxElement {
		opening = node.AsJsxElement().OpeningElement
	}
	for _, attribute := range reactutil.GetJsxElementAttributes(opening) {
		if attribute.Kind == ast.KindJsxAttribute && reactutil.GetJsxPropName(attribute) == "key" {
			return true
		}
	}
	return false
}

// isAllLowercaseLetters ports upstream's `/^[a-z]+$/` tag-name test used to
// decide whether a fragment's parent is an HTML element. JavaScript's `$`
// anchors at the very end of the string (no trailing-newline allowance), so a
// byte scan is exactly equivalent.
func isAllLowercaseLetters(name string) bool {
	if name == "" {
		return false
	}
	for i := range len(name) {
		if name[i] < 'a' || name[i] > 'z' {
			return false
		}
	}
	return true
}

// trimLikeReact ports upstream's helper: a leading or trailing whitespace run
// is dropped only when it contains a literal "\n", which is how the React
// runtime treats JSX indentation. Note upstream tests the run for the "\n"
// character specifically, not for any ECMAScript LineTerminator.
func trimLikeReact(text string) string {
	leadingEnd := ecmascript.SkipLeadingWhitespace(text, 0, len(text))
	trailingStart := ecmascript.SkipTrailingWhitespace(text, 0, len(text))

	start := 0
	if strings.Contains(text[:leadingEnd], "\n") {
		start = leadingEnd
	}
	end := len(text)
	if strings.Contains(text[trailingStart:], "\n") {
		end = trailingStart
	}

	// Mirrors JavaScript's `String.prototype.slice`, which yields "" when the
	// start index runs past the end index (an all-whitespace run hits this).
	if start >= end {
		return ""
	}
	return text[start:end]
}
