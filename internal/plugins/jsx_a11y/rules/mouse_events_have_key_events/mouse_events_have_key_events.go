// Package mouse_events_have_key_events ports eslint-plugin-jsx-a11y's
// `mouse-events-have-key-events` rule. Enforces that any HTML DOM element
// carrying a configured "hover-in" handler (default: onMouseOver) also
// carries a non-null `onFocus`, and that any "hover-out" handler (default:
// onMouseOut) is paired with `onBlur`. Keyboard-only users rely on
// focus / blur events to get the same affordance the mouse handlers
// expose.
//
// Upstream signature:
//
//	options: {
//	  hoverInHandlers?:  string[]  // default ['onMouseOver']
//	  hoverOutHandlers?: string[]  // default ['onMouseOut']
//	}
//
// Per-element flow (upstream JSXOpeningElement listener):
//
//  1. Resolve `name` via `node.name.name` — NOTE: upstream reads the raw
//     tag name directly, NOT the `getElementType(context)(node)` helper
//     other jsx-a11y rules use. So `settings['jsx-a11y'].components` and
//     `polymorphicPropName` do NOT apply here — `<Foo as="div" />` and
//     `settings.components.Footer = 'footer'` are both treated as custom
//     components. We mirror by passing the raw `GetJsxElementTypeString`
//     output through `IsDOMElement`.
//  2. `dom.get(name)` must be truthy → tag is in aria-query's HTML DOM
//     map. Custom components (`<MyElement>`) and member-expression /
//     namespaced tags (`<Foo.Bar>`, `<svg:path>`) are skipped — the rule
//     doesn't know what low-level element they render.
//  3. For each configured `hoverInHandlers` entry, find the FIRST one
//     whose `getProp` exists AND whose `getPropValue` is `!= null` (loose:
//     excludes literal `null` / `undefined`, but accepts the boolean
//     attribute form `<div onMouseOver />` which extracts to JS `true`).
//     If none found → no hover-in pairing required.
//  4. If hover-in handler found → check `onFocus` presence + non-null
//     value via the same `getProp` + `getPropValue != null` semantics.
//     If missing → report on the hover-in attribute.
//  5. Same flow for `hoverOutHandlers` + `onBlur`.
//
// A single element can produce both diagnostics independently when both
// pairings are missing.
package mouse_events_have_key_events

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// defaultHoverInHandlers / defaultHoverOutHandlers mirror upstream's
// `DEFAULT_HOVER_IN_HANDLERS` / `DEFAULT_HOVER_OUT_HANDLERS`. Used when
// the option key is absent or explicitly `undefined` per upstream's
// `options[0]?.hoverInHandlers ?? DEFAULT_*` semantics. An EXPLICIT
// empty array (`{ hoverInHandlers: [] }`) is a deliberate "check nothing"
// signal and is NOT replaced with the default — upstream's `??` only
// catches nullish values.
var (
	defaultHoverInHandlers  = []string{"onMouseOver"}
	defaultHoverOutHandlers = []string{"onMouseOut"}
)

type options struct {
	HoverInHandlers  []string
	HoverOutHandlers []string
}

func parseOptions(raw any) options {
	opts := options{
		HoverInHandlers:  defaultHoverInHandlers,
		HoverOutHandlers: defaultHoverOutHandlers,
	}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	// Explicit empty array stays empty; a non-array / missing key falls
	// back to the upstream defaults via `??`. StringSliceOption returns
	// nil for non-`[]interface{}` inputs and a non-nil zero-length slice
	// for `[]`, so the `nil` check below distinguishes the two cases.
	if v, ok := optsMap["hoverInHandlers"]; ok {
		parsed := jsxa11yutil.StringSliceOption(v)
		if parsed != nil {
			opts.HoverInHandlers = parsed
		}
	}
	if v, ok := optsMap["hoverOutHandlers"]; ok {
		parsed := jsxa11yutil.StringSliceOption(v)
		if parsed != nil {
			opts.HoverOutHandlers = parsed
		}
	}
	return opts
}

// firstHoverHandlerWithValue mirrors upstream's
//
//	hoverInHandlers.find((handler) => {
//	  const prop = getProp(attributes, handler);
//	  const propValue = getPropValue(prop);
//	  return propValue != null;
//	});
//
// Returns the handler name and its attribute node (used for the report
// position), or ("", nil) when no configured handler is present with a
// non-null value.
func firstHoverHandlerWithValue(attrs []*ast.Node, handlers []string) (string, *ast.Node) {
	for _, handler := range handlers {
		prop := jsxa11yutil.FindAttributeByName(attrs, handler)
		if prop == nil {
			continue
		}
		// `getPropValue(prop) != null` (loose inequality).
		// PropValueIsNullish encodes the inverse: true when the
		// extracted value is null / undefined / unresolvable.
		if jsxa11yutil.PropValueIsNullish(prop) {
			continue
		}
		return handler, prop
	}
	return "", nil
}

// pairAttributeIsMissing mirrors upstream's
//
//	const hasOn{Focus,Blur} = getProp(attributes, 'on{Focus,Blur}');
//	const on{...}Value = getPropValue(hasOn{...});
//	hasOn{...} === false || on{...}Value === null || on{...}Value === undefined
//
// `hasOn === false` is never reached in practice (jsx-ast-utils returns
// `undefined` when missing), so the gate simplifies to "absent or value
// is nullish".
func pairAttributeIsMissing(attrs []*ast.Node, pairName string) bool {
	prop := jsxa11yutil.FindAttributeByName(attrs, pairName)
	if prop == nil {
		return true
	}
	return jsxa11yutil.PropValueIsNullish(prop)
}

var MouseEventsHaveKeyEventsRule = rule.Rule{
	Name: "jsx-a11y/mouse-events-have-key-events",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.UnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		check := func(node *ast.Node) {
			// Upstream reads `node.name.name` — the raw JSX tag identifier
			// — directly, without applying `getElementType(context)`. We
			// mirror by passing the raw string to IsDOMElement, which
			// short-circuits non-Identifier tags (member-expression,
			// namespaced) since their stringified forms aren't in
			// aria-query's `dom` map.
			name := reactutil.GetJsxElementTypeString(node)
			if !jsxa11yutil.IsDOMElement(name) {
				return
			}

			attrs := reactutil.GetJsxElementAttributes(node)

			if hoverIn, hoverInProp := firstHoverHandlerWithValue(attrs, opts.HoverInHandlers); hoverInProp != nil {
				if pairAttributeIsMissing(attrs, "onFocus") {
					ctx.ReportNode(hoverInProp, rule.RuleMessage{
						Id:          "mouseOver",
						Description: hoverIn + " must be accompanied by onFocus for accessibility.",
					})
				}
			}

			if hoverOut, hoverOutProp := firstHoverHandlerWithValue(attrs, opts.HoverOutHandlers); hoverOutProp != nil {
				if pairAttributeIsMissing(attrs, "onBlur") {
					ctx.ReportNode(hoverOutProp, rule.RuleMessage{
						Id:          "mouseOut",
						Description: hoverOut + " must be accompanied by onBlur for accessibility.",
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
