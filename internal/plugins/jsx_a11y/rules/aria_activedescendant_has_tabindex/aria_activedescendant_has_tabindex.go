// Package aria_activedescendant_has_tabindex ports eslint-plugin-jsx-a11y's
// `aria-activedescendant-has-tabindex` rule. The rule enforces that any DOM
// element managing focus via `aria-activedescendant` is tabbable — i.e.
// either inherently focusable (input, button, ...) without a custom
// `tabIndex`, or carrying an explicit `tabIndex >= -1`. A negative value
// other than `-1` removes the element from the focus tree, defeating the
// activedescendant pattern (WAI-ARIA Authoring Practices §3.10).
//
// Upstream signature: no options — schema is `generateObjSchema()` (an
// empty object).
//
// Trigger: a JsxOpeningElement / JsxSelfClosingElement that
//
//  1. carries an `aria-activedescendant` prop (case-insensitive,
//     boolean form / `={undefined}` / `={null}` all count — `getProp`
//     returns the attribute regardless of value),
//  2. resolves to a DOM element name per aria-query's `dom` map
//     (custom components / `Foo.Bar` / namespaced `svg:path` are skipped),
//  3. is NOT both inherently interactive (input/button/option/...) AND
//     missing a `tabIndex` (those rely on the browser-native tab order),
//  4. has `tabIndex < -1` or no resolvable `tabIndex` value.
//
// The diagnostic is reported on the JSX opening element node itself
// (matches upstream `context.report({ node, ... })` where `node` is the
// JSXOpeningElement).
package aria_activedescendant_has_tabindex

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "An element that manages focus with `aria-activedescendant` must have a tabindex"

var AriaActivedescendantHasTabindexRule = rule.Rule{
	Name:   "jsx-a11y/aria-activedescendant-has-tabindex",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		// sourceText is required by GetTabIndexEx for raw-text template
		// literal extraction (NoSubstitutionTemplate has no RawText field).
		sourceText := ctx.SourceFile.Text()
		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)

			// gate-1: `getProp(attributes, 'aria-activedescendant') === undefined`.
			// FindAttributeByName mirrors jsx-ast-utils' `getProp` with default
			// `ignoreCase: true`, so `Aria-ActiveDescendant` matches the same
			// as `aria-activedescendant`. Boolean attribute form
			// (`<div aria-activedescendant />`) and explicit-undefined value
			// (`<div aria-activedescendant={undefined} />`) BOTH return a
			// non-nil node here — `getProp` returns the attribute regardless
			// of value, so gate-1 only short-circuits when the prop is truly
			// absent.
			if jsxa11yutil.FindAttributeByName(attrs, "aria-activedescendant") == nil {
				return
			}

			// gate-2: `!dom.has(type)`. Resolve via `getElementType` so the
			// `polymorphicPropName` / `components` map settings are honored
			// (e.g. `<CustomComponent />` mapped to `div` participates in the
			// rule). Custom components, dotted-namespace tags (`Foo.Bar`),
			// and SVG-namespaced tags (`svg:path`) all fall outside the
			// aria-query DOM set and skip silently.
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if !jsxa11yutil.IsDOMElement(elementType) {
				return
			}

			// gate-3: `isInteractiveElement(type, attributes) && tabIndex === undefined`.
			// upstream's `tabIndex === undefined` is strict equality, so it
			// matches only when getTabIndex's step-1 short-circuited (boolean
			// form, empty string, NaN, missing prop). The step-2 fallback
			// returns `null` for unrecognized expression types, and `null
			// === undefined` is false, so gate-3 must NOT skip in that case.
			// GetTabIndexEx surfaces the distinction via nullLike.
			tabIndexProp := jsxa11yutil.FindAttributeByName(attrs, "tabIndex")
			tabIndex, hasTabIndex, nullLike := jsxa11yutil.GetTabIndexEx(tabIndexProp, sourceText)
			if !hasTabIndex && !nullLike && jsxa11yutil.IsInteractiveElement(elementType, attrs) {
				return
			}

			// gate-4: `tabIndex >= -1`. Two arms:
			//   - resolved Number: direct numeric `>= -1`.
			//   - upstream-null (nullLike): JS `null >= -1` ToNumber-coerces
			//     to `0 >= -1` = true → skip. This arm applies to
			//     TSSatisfiesExpression, AwaitExpression, YieldExpression,
			//     ImportExpression, and any future ESTree expression type
			//     jsx-ast-utils' TYPES table doesn't recognize. Without
			//     this arm rslint would over-report relative to upstream;
			//     with it the contract is byte-for-byte aligned.
			// undefined-like (`!hasTabIndex && !nullLike`) falls through to
			// the report — `undefined >= -1` is false in JS.
			if hasTabIndex && tabIndex >= -1 {
				return
			}
			if nullLike {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "tabIndexRequired",
				Description: errorMessage,
			})
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
