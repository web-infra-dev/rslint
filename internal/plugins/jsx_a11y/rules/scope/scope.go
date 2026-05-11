// Package scope ports eslint-plugin-jsx-a11y's `scope` rule. The `scope`
// HTML attribute is only valid on `<th>` elements (per the HTML spec / WCAG
// 1.3.1 / axe-core's `scope-attr-valid` check); using it on any other element
// is a no-op for assistive technology and signals a structural mistake.
//
// Upstream signature: no options.
//
// Trigger: a JsxAttribute whose name is "scope" (case-insensitive per
// upstream's `name.toUpperCase() !== 'SCOPE'`) on a parent element whose
// resolved tag name is in aria-query's `dom` map AND is not "th"
// (case-insensitive). The two case-sensitivity asymmetries are intentional
// and inherited from upstream:
//
//   - The DOM-set membership check (`dom.has(tagName)`) is case-sensitive
//     against aria-query's lowercase keys, so `<TH scope />` is silently
//     skipped (resolved tag "TH" is NOT in the map). Mirror exactly via
//     IsDOMElement, which also uses case-sensitive lookup.
//   - The "is th" exemption is case-insensitive (`.toUpperCase() === 'TH'`),
//     so once we're past the dom-set gate, both `<th>` and any case-variant
//     that somehow survived would be exempt.
//
// `getElementType` honors `polymorphicPropName` and the `components` map from
// `settings['jsx-a11y']`, so `<TableHeader scope="row" />` with
// `components: { TableHeader: 'th' }` resolves to "th" and skips the report,
// while `<Foo scope="bar" />` with `components: { Foo: 'div' }` resolves to
// "div" and reports.
//
// Namespaced attribute names like `<th xml:scope />` are NOT matched —
// upstream's `propName` returns the composite "xml:scope" string, which
// uppercases to "XML:SCOPE" ≠ "SCOPE". `reactutil.GetJsxPropName` mirrors
// this composite shape.
package scope

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "The scope prop can only be used on <th> elements."

var ScopeRule = rule.Rule{
	Name: "jsx-a11y/scope",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Upstream: `if (name && name.toUpperCase() !== 'SCOPE') return;`
				// — the `name && ...` short-circuit means an empty name string
				// FALLS THROUGH to the parent check (defensive for malformed
				// AST upstream; tsgo produces a non-empty Identifier or
				// JsxNamespacedName for every legal JsxAttribute). Mirror the
				// quirk literally so we don't accidentally skip cases upstream
				// would still process.
				name := reactutil.GetJsxPropName(attr)
				if name != "" && !strings.EqualFold(name, "scope") {
					return
				}
				// Resolve the parent element via the JsxAttributes container
				// — JsxAttribute.Parent is the JsxAttributes node, whose
				// Parent is the JsxOpeningElement / JsxSelfClosingElement.
				parent := reactutil.GetJsxParentElement(attr)
				if parent == nil {
					return
				}
				// `getElementType(parent)` honors polymorphicPropName + the
				// components map. Custom components, member-expression tags
				// (`<Foo.Bar>`), and namespaced tags (`<svg:circle>`) all
				// resolve to non-DOM strings and skip via the IsDOMElement
				// gate below.
				tagName := jsxa11yutil.GetElementType(parent, ctx.Settings)
				// Mirror upstream's `if (!dom.has(tagName)) return;` —
				// aria-query's `dom` map is keyed by lowercase HTML element
				// names, so the lookup is case-sensitive. `<TH scope />`
				// resolves to "TH", which is NOT in the map → skipped, even
				// though TH is semantically a valid scope target. Locked in
				// by an upstream-faithful test below.
				if !jsxa11yutil.IsDOMElement(tagName) {
					return
				}
				// Upstream: `if (tagName && tagName.toUpperCase() === 'TH') return;`
				// — case-insensitive exemption for the legal target. Combined
				// with the case-sensitive IsDOMElement gate above, only
				// lowercase "th" will reach this branch in practice; the
				// EqualFold is upstream-faithful but functionally equivalent
				// to a direct `tagName == "th"` check.
				if strings.EqualFold(tagName, "th") {
					return
				}
				ctx.ReportNode(attr, rule.RuleMessage{
					Id:          "scopeOnTh",
					Description: errorMessage,
				})
			},
		}
	},
}
