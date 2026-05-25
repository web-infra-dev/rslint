// Package no_static_element_interactions ports eslint-plugin-jsx-a11y's
// `no-static-element-interactions` rule. The rule flags HTML elements that
// carry an interactive event handler (default: focus + keyboard + mouse
// handlers) but lack an interactive ARIA role — visible static elements
// dispatching to mouse / keyboard listeners without a role are invisible
// to assistive technology.
//
// Upstream signature:
//
//	options: {
//	  handlers?:              string[]  (default: focus + keyboard + mouse handlers)
//	  allowExpressionValues?: boolean   (default: undefined / false-ish)
//	}
//
// Trigger sequence — each predicate is checked in order against the JSX
// opening element. Bail-outs return without reporting:
//
//  1. Type isn't an aria-query DOM element → bail (custom components).
//  2. No interactive event handler attached with a non-null value (mirrors
//     `hasProp(attrs, h) && getPropValue(getProp(attrs, h)) != null` for the
//     `handlers` list) → bail.
//  3. Element is hidden from screen readers (`aria-hidden={true}` or
//     `<input type="hidden">`) → bail.
//  4. Element role resolves to `presentation` / `none` → bail.
//  5. Element is inherently interactive, OR carries an interactive role,
//     OR is inherently non-interactive, OR carries a non-interactive role,
//     OR carries an abstract role → bail.
//  6. `allowExpressionValues` is true AND `role` is non-literal → bail.
//     Upstream's "special case for ternary with literals on both side"
//     inside the same if-block returns regardless of which arm matches,
//     so the observable behavior collapses to "any non-literal role under
//     allowExpressionValues=true is exempt"; we mirror only the observable.
//  7. Otherwise → REPORT on the JSX opening element node.
//
// Diagnostic text mirrors upstream verbatim:
//
//	"Avoid non-native interactive elements. If using native HTML is not
//	possible, add an appropriate role and support for tabbing, mouse,
//	keyboard, and touch inputs to an interactive content element."
package no_static_element_interactions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Avoid non-native interactive elements. If using native HTML is not possible, add an appropriate role and support for tabbing, mouse, keyboard, and touch inputs to an interactive content element."

// options holds the parsed configuration. `Handlers` is nil iff the user
// did not provide the option at all — the listener then falls back to
// [jsxa11yutil.DefaultStaticInteractionHandlers]. An explicit empty array
// (`handlers: []`) is preserved as a non-nil empty slice so the listener
// can mirror upstream's `[].some(...)` → false short-circuit (no handler
// matches → never reports).
type options struct {
	Handlers              []string
	HandlersProvided      bool
	AllowExpressionValues bool
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["handlers"]; ok {
		opts.HandlersProvided = true
		opts.Handlers = jsxa11yutil.StringSliceOption(v)
		if opts.Handlers == nil {
			// StringSliceOption returns nil for non-array shapes; preserve
			// "user supplied something" so the fallback to the default
			// list does not silently kick back in.
			opts.Handlers = []string{}
		}
	}
	if v, ok := m["allowExpressionValues"].(bool); ok {
		opts.AllowExpressionValues = v
	}
	return opts
}

// hasNonNullInteractiveHandler mirrors upstream's
//
//	handlers.some((prop) =>
//	  hasProp(attributes, prop)
//	  && getPropValue(getProp(attributes, prop)) != null
//	)
//
// Iteration order mirrors upstream byte-for-byte: outer loop over `handlers`,
// inner loop over `attrs` short-circuiting on the FIRST matching JsxAttribute
// (jsx-ast-utils' `getProp` is `Array.prototype.find`, so duplicate JSX
// attributes — invalid source, but tsgo still parses them — are ignored past
// the first occurrence). `!= null` is JS loose equality and excludes both
// `null` and `undefined`, so `<div onClick={null} />` and
// `<div onClick={undefined} />` do NOT count as interactive.
//
// `hasProp` / `getProp` default to spreadStrict=true, so JsxSpreadAttribute
// nodes are opaque — only direct JsxAttribute children participate. Name
// comparison is case-insensitive to mirror upstream's `ignoreCase: true`.
func hasNonNullInteractiveHandler(attrs []*ast.Node, handlers []string) bool {
	for _, h := range handlers {
		for _, attr := range attrs {
			if attr.Kind != ast.KindJsxAttribute {
				continue
			}
			if !strings.EqualFold(reactutil.GetJsxPropName(attr), h) {
				continue
			}
			// First direct attribute with a matching name — `getProp`'s
			// `Array.find` returns this one. If its value is non-null
			// the handler counts; otherwise upstream moves to the next
			// handler without inspecting subsequent duplicate attributes.
			if !jsxa11yutil.PropValueIsNullish(attr) {
				return true
			}
			break
		}
	}
	return false
}

var NoStaticElementInteractionsRule = rule.Rule{
	Name: "jsx-a11y/no-static-element-interactions",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		handlers := opts.Handlers
		if !opts.HandlersProvided {
			handlers = jsxa11yutil.DefaultStaticInteractionHandlers
		}

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)

			// Step 1: type must be a known HTML element (aria-query's `dom`).
			if !jsxa11yutil.IsDOMElement(elementType) {
				return
			}

			// Step 2: at least one configured handler must be attached with
			// a non-null value. Upstream collapses `!hasInteractiveProps` /
			// `isHiddenFromScreenReader` / `isPresentationRole` into one
			// boolean OR; we keep them as discrete returns for readability
			// — same observable behavior.
			if !hasNonNullInteractiveHandler(attrs, handlers) {
				return
			}

			getElementType := func(child *ast.Node) string {
				return jsxa11yutil.GetElementType(child, ctx.Settings)
			}
			if jsxa11yutil.IsHiddenFromScreenReader(node, getElementType) {
				return
			}

			if jsxa11yutil.IsPresentationRole(attrs) {
				return
			}

			// Step 3: any of the five "already accessible / abstract /
			// non-interactive" classifications exempts the element.
			if jsxa11yutil.IsInteractiveElement(elementType, attrs) ||
				jsxa11yutil.IsInteractiveRole(elementType, attrs) ||
				jsxa11yutil.IsNonInteractiveElement(elementType, attrs) ||
				jsxa11yutil.IsNonInteractiveRole(elementType, attrs) ||
				jsxa11yutil.IsAbstractRole(elementType, attrs) {
				return
			}

			// Step 4: allowExpressionValues + non-literal `role` → skip.
			// Upstream's inner "ConditionalExpression with two Literal arms"
			// branch returns regardless of which arm matches, so the
			// surrounding if-block reduces to an unconditional skip
			// whenever `IsNonLiteralProperty` matches. Mirror the
			// observable behavior; the inner check is dead.
			if opts.AllowExpressionValues && jsxa11yutil.IsNonLiteralProperty(attrs, "role") {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noStaticElementInteractions",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
