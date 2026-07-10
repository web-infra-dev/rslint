// Package click_events_have_key_events ports eslint-plugin-jsx-a11y's
// `click-events-have-key-events` rule. Enforces that any visible,
// non-interactive DOM element carrying an `onClick` handler also carries at
// least one of `onKeyDown` / `onKeyUp` / `onKeyPress`, so keyboard-only
// users can trigger the same interaction.
//
// Upstream signature: no options. The rule receives every JSXOpeningElement
// (paired and self-closing) and runs five short-circuits in order:
//
//  1. no `onClick` prop → return (case-insensitive; spreads opaque under
//     `getProp` default `ignoreCase: true, spreadStrict: false` — but
//     `<div {...{onClick}} />`-style literal spreads ARE walked, matching
//     upstream's `getProp` default behavior).
//  2. element type not in aria-query's `dom` map → return (custom
//     components / unknown tag names skip; we don't second-guess what
//     low-level DOM the component renders).
//  3. element is hidden from screen readers OR has role "presentation" /
//     "none" → return.
//  4. element is inherently interactive (per elementRoles /
//     elementAXObjects, e.g. <button>, <a href>, <input type="text">) →
//     return.
//  5. element declares any of onKeyDown / onKeyUp / onKeyPress (case-
//     insensitive, spread attrs opaque per upstream's `hasAnyProp` default
//     `spreadStrict: true`) → return.
//
// Otherwise, report at the JSX opening element. The single message id is
// `clickEventsHaveKeyEvents`, mirroring upstream's single `errorMessage`.
package click_events_have_key_events

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "Visible, non-interactive elements with click handlers must have at least one keyboard listener."

var ClickEventsHaveKeyEventsRule = rule.Rule{
	Name: "jsx-a11y/click-events-have-key-events",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		getElementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)

			// Upstream `getProp(props, 'onclick')` — case-insensitive, walks
			// literal-spread (default `spreadStrict: false`). Our
			// FindAttributeByName matches both behaviors.
			if jsxa11yutil.FindAttributeByName(attrs, "onClick") == nil {
				return
			}

			elementType := getElementType(node)
			if !jsxa11yutil.IsDOMElement(elementType) {
				return
			}

			if jsxa11yutil.IsHiddenFromScreenReader(node, getElementType) ||
				jsxa11yutil.IsPresentationRole(attrs) {
				return
			}

			if jsxa11yutil.IsInteractiveElement(elementType, attrs) {
				return
			}

			// Upstream `hasAnyProp(props, requiredProps)` — default
			// `spreadStrict: true`, so `<div onClick={...} {...props} />`
			// cannot be saved by a possibly-keyboard handler inside `props`.
			if jsxa11yutil.HasAnyJsxPropStrict(attrs, "onKeyDown", "onKeyUp", "onKeyPress") {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "clickEventsHaveKeyEvents",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
