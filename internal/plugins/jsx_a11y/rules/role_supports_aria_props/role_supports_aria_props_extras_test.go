// cspell:ignore activedescendant atomic checkbox controls describedby disabled
// cspell:ignore dropeffect flowto grabbed haspopup hidden labelledby owns
// cspell:ignore relevant valuemax valuemin valuenow valuetext combobox listbox
// cspell:ignore searchbox spinbutton textbox tabpanel toolbar tooltip alertdialog
// cspell:ignore presentation chcked omponent utton butto

package role_supports_aria_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` jsx-a11y setting —
// `<Foo as="a" href="#" aria-checked />` resolves to elementType "a" before
// the implicit-role lookup. Upstream's own test file does not cover this
// path on `role-supports-aria-props`, but the rule routes through the same
// [GetElementType] chain as every other jsx-a11y rule, so the surface is
// identical.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestRoleSupportsAriaPropsExtras covers cases NOT in upstream's
// `__tests__/src/rules/role-supports-aria-props-test.js` — universal
// edge shapes (Dimension 4 of the port skill), tsgo AST quirks, and
// lock-in tests for each branch in the upstream `JSXOpeningElement`
// listener that the upstream test file leaves uncovered.
func TestRoleSupportsAriaPropsExtras(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &RoleSupportsAriaPropsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Step 3 — non-string / non-literal role values short-circuit.
			// Locks in the `typeof roleValue !== 'string'` arm.
			// ============================================================

			// `<div role={x} />` — Identifier; LITERAL_TYPES.Identifier →
			// null → not a string → skip even with an aria-checked attr.
			{Code: `<div role={x} aria-checked />`, Tsx: true},
			// `<div role={fn()} />` — CallExpression; LITERAL_TYPES.CallExpression
			// noop → null → skip.
			{Code: `<div role={fn()} aria-checked />`, Tsx: true},
			// `<div role={a || b} />` — LogicalExpression; LITERAL_TYPES.LogicalExpression
			// noop → null → skip.
			{Code: `<div role={a || b} aria-checked />`, Tsx: true},
			// `<div role={cond ? a : b} />` — ConditionalExpression; noop → null
			// → skip.
			{Code: `<div role={cond ? a : b} aria-checked />`, Tsx: true},
			// `<div role={true} />` — Literal Boolean; LITERAL_TYPES.Literal
			// returns true (not a string) → skip.
			{Code: `<div role={true} aria-checked />`, Tsx: true},
			// `<div role={null} />` — Literal null; LITERAL_TYPES.Literal
			// returns the magic string "null" — IS a string, but "null"
			// isn't in AriaRolePropsMap → skip via the membership check.
			{Code: `<div role={null} aria-checked />`, Tsx: true},
			// `<div role={undefined} />` — Identifier "undefined" →
			// LITERAL_TYPES.Identifier returns undefined (not a string) → skip.
			{Code: `<div role={undefined} aria-checked />`, Tsx: true},

			// ============================================================
			// Step 5 — case-sensitive membership check on roleValue. Locks
			// in the upstream behavior that `<div role="BUTTON">` is silently
			// not validated (since aria-query keys are lowercase). This is
			// observably DIFFERENT from `role-has-required-aria-props` and
			// `no-redundant-roles`, which lowercase the role value first.
			// ============================================================
			{Code: `<div role="BUTTON" aria-checked />`, Tsx: true},
			{Code: `<div role="Button" aria-checked />`, Tsx: true},

			// ============================================================
			// Step 5 — unknown role names skip via membership check.
			// ============================================================
			{Code: `<div role="not-a-real-role" aria-checked />`, Tsx: true},

			// ============================================================
			// Step 7 — null / undefined attribute values are skipped per
			// upstream `getPropValue(prop) != null`. Lock the explicit-null,
			// explicit-undefined, and TS-cast variants.
			// ============================================================
			{Code: `<a href="#" aria-checked={null} />`, Tsx: true},
			{Code: `<a href="#" aria-checked={undefined} />`, Tsx: true},
			// `as`-wrapped null / undefined — upstream's full extract walks
			// past TSAsExpression, so the inner null/undefined is exposed.
			{Code: `<a href="#" aria-checked={null as any} />`, Tsx: true},
			{Code: `<a href="#" aria-checked={undefined as any} />`, Tsx: true},
			// Parenthesized null — parens are unwrapped in attributeInnerExpression.
			{Code: `<a href="#" aria-checked={(null)} />`, Tsx: true},

			// ============================================================
			// Step 7 — JsxSpreadAttribute is opaque per upstream's
			// `prop.type !== 'JSXSpreadAttribute'` filter, even when the
			// spread argument is a literal object containing aria-* keys.
			// `<a href="#" {...{ "aria-checked": true }} />` should NOT report.
			// ============================================================
			{
				Code: `<a href="#" {...{"aria-checked": true}} />`,
				Tsx:  true,
			},

			// ============================================================
			// Step 7 — mixed-case ARIA prop names are silently NOT validated
			// because upstream's `propName(prop)` returns the raw name and
			// the `invalidAriaPropsForRole` set is built from lowercase
			// aria.keys(). `<a href="#" Aria-Checked />` therefore passes.
			// ============================================================
			{Code: `<a href="#" Aria-Checked />`, Tsx: true},

			// ============================================================
			// Step 7 — the rule walks ALL attributes; each gets independently
			// classified. Mixing valid (aria-label) and skipped-because-aria-
			// not-recognized (data-foo, onClick) attributes alongside a
			// supported ARIA prop should pass.
			// ============================================================
			{
				Code: `<a href="#" aria-label="x" data-foo="y" onClick={fn} />`,
				Tsx:  true,
			},

			// ============================================================
			// `<a>` without href — no implicit role → skip. Ensures the
			// implicit-role gating is consistent across all aria-* attrs.
			// ============================================================
			{Code: `<a aria-checked />`, Tsx: true},
			{Code: `<area aria-checked />`, Tsx: true},
			{Code: `<link aria-checked />`, Tsx: true},

			// ============================================================
			// `<img alt="" />` — empty alt suppresses the implicit "img"
			// role; aria-checked therefore passes.
			// ============================================================
			{Code: `<img alt="" aria-checked />`, Tsx: true},
			// `<img src="foo.svg" />` — svg src suppresses the implicit
			// "img" role.
			{Code: `<img src="foo.svg" aria-checked />`, Tsx: true},
			// `<img src={someVar} />` — non-literal src; LITERAL_TYPES.Identifier
			// → null → optional-chain `?.includes` short-circuits → still 'img'.
			// aria-checked is invalid on img → expected to FAIL the rule.
			// (See invalid section below.)

			// ============================================================
			// `<select multiple />` → implicit role "listbox" — aria-multiselectable
			// is supported on listbox.
			// ============================================================
			{Code: `<select multiple aria-multiselectable />`, Tsx: true},
			{Code: `<select size="3" aria-multiselectable />`, Tsx: true},

			// ============================================================
			// `<select size="0x10" />` → 16 > 1 → implicit role "listbox".
			// LiteralPropJSNumber recognizes JS-style hex prefixes.
			// ============================================================
			{Code: `<select size="0x10" aria-multiselectable />`, Tsx: true},

			// ============================================================
			// Polymorphic `as` — `<Foo as="a" href="#" aria-expanded />`
			// resolves to "a" → implicit "link" → aria-expanded is supported
			// on link.
			// ============================================================
			{
				Code:     `<Foo as="a" href="#" aria-expanded />`,
				Tsx:      true,
				Settings: polymorphicSettings,
			},

			// ============================================================
			// Custom component without componentsSettings — elementType is
			// "Foo", which has no implicit role table entry → skip even when
			// the user adds clearly-invalid aria props.
			// ============================================================
			{Code: `<Foo aria-checked />`, Tsx: true},

			// ============================================================
			// Empty role attribute string — splits into "", which is not in
			// AriaRolePropsMap → skip.
			// ============================================================
			{Code: `<div role="" aria-checked />`, Tsx: true},

			// ============================================================
			// Paired (non-self-closing) JSX element — both KindJsxOpeningElement
			// and KindJsxSelfClosingElement listeners are registered, so paired
			// `<a href="#" aria-expanded></a>` triggers via the opening element.
			// ============================================================
			{Code: `<a href="#" aria-expanded></a>`, Tsx: true},

			// ============================================================
			// Nested elements — the inner element's role check is independent
			// of the outer. `<div><a href="#" aria-expanded /></div>` exercises
			// the per-listener invocation; both elements pass.
			// ============================================================
			{Code: `<div><a href="#" aria-expanded /></div>`, Tsx: true},

			// ============================================================
			// JsxExpression-wrapped string literal role — same handling as
			// direct string. `<div role={"button"} aria-pressed />` → button
			// supports aria-pressed.
			// ============================================================
			{Code: `<div role={"button"} aria-pressed />`, Tsx: true},

			// ============================================================
			// HTML entity-decoded role string — `<div role="&#98;utton" ...>`
			// decodes to "button" via directAttributeStringValue. button
			// supports aria-pressed.
			// ============================================================
			{Code: `<div role="&#98;utton" aria-pressed />`, Tsx: true},

			// ============================================================
			// Abstract role — `role="command"` is a valid key in
			// AriaRolePropsMap (we include abstract roles per upstream's
			// `roles.get(roleValue)`). The `command` props set includes
			// aria-label, so this passes. Locks in the include-abstract behavior.
			// ============================================================
			{Code: `<div role="command" aria-label="x" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Lock in: `<img src={someVar} aria-checked />` — non-literal src
			// keeps the implicit "img" role; aria-checked is NOT supported
			// on img.
			// ============================================================
			{
				Code: `<img src={someVar} aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "img", "img", true),
					Line:      1, Column: 1,
				}},
			},

			// ============================================================
			// Polymorphic `as` — `<Foo as="a" href="#" aria-checked />`
			// resolves to "a" → implicit "link" → aria-checked is NOT
			// supported on link.
			// ============================================================
			{
				Code:     `<Foo as="a" href="#" aria-checked />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},

			// ============================================================
			// Paired form — listener fires on the opening element. The
			// reported diagnostic should land at the opening element, not
			// the closing. Locks in the JsxOpeningElement listener arm.
			// ============================================================
			{
				Code: `<a href="#" aria-checked></a>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},

			// ============================================================
			// Multi-line JSX — verify the position assertion uses the
			// opening element's position (line 2, col 5 — the `<` is at
			// column 5 of line 2 after leading spaces).
			// ============================================================
			{
				Code: `
    <a
      href="#"
      aria-checked
    />`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      2, Column: 5,
				}},
			},

			// ============================================================
			// Multiple invalid aria props on the same element — each emits
			// an independent diagnostic. Upstream uses `forEach`, so per-prop
			// reporting is the spec.
			// ============================================================
			{
				Code: `<a href="#" aria-checked aria-pressed />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "role-supports-aria-props",
						Message:   errorMessage("aria-checked", "link", "a", true),
						Line:      1, Column: 1,
					},
					{
						MessageId: "role-supports-aria-props",
						Message:   errorMessage("aria-pressed", "link", "a", true),
						Line:      1, Column: 1,
					},
				},
			},

			// ============================================================
			// Explicit role with no implicit fallback — `<div role="link" />`
			// has explicit role "link", isImplicit = false, so the error
			// message uses the non-implicit phrasing (no trailing "This role
			// is implicit on the element X." sentence). Locks in the
			// isImplicit branch.
			// ============================================================
			{
				Code: `<div role="link" aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "div", false),
					Line:      1, Column: 1,
				}},
			},

			// ============================================================
			// Explicit role on a custom component — elementType is "Foo",
			// the rule doesn't gate on IsDOMElement, so `<Foo role="link"
			// aria-checked />` reports with the custom tag name in the error
			// message. (isImplicit = false because role attribute is set.)
			// ============================================================
			{
				Code: `<Foo role="link" aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "Foo", false),
					Line:      1, Column: 1,
				}},
			},
		},
	)
}

// polymorphicAllowList exercises the `polymorphicAllowList` setting — when
// set, only listed raw element names participate in the polymorphicProp
// substitution. Locks in the upstream `!polymorphicAllowList || includes(...)`
// gate; without this case the difference between "allow-listed" and "not
// allow-listed" elements would only show up in real-world use.
var polymorphicAllowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"polymorphicAllowList": []interface{}{
			"Box",
		},
	},
}

// componentsSettingsLink mirrors the upstream Link → a mapping under a
// different name so the robust suite doesn't share the upstream-test
// fixture (the two suites must be readable independently).
var componentsSettingsLink = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyLink":  "a",
			"MyImage": "img",
		},
	},
}

// componentsAndPolymorphic combines `components` AND `polymorphicPropName`.
// Mirrors a real-world design-system pattern where custom primitives both
// expose a polymorphic `as` prop AND map to their default DOM element via
// the components map.
var componentsAndPolymorphic = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"components": map[string]interface{}{
			"Box":    "div",
			"Anchor": "a",
		},
	},
}

// TestRoleSupportsAriaPropsRobust extends [TestRoleSupportsAriaPropsExtras]
// with real-world component / hook / HOC patterns, listener-boundary cases
// across nested elements, JS-style coercion edge cases, settings-shape
// robustness, and tsgo-vs-ESTree AST differences. Designed to catch silent
// regressions when the rule or its shared helpers are refactored.
//
// Each group lives behind a comment marker `============= GROUP X =============`;
// individual cases name the upstream branch / scenario / quirk they protect.
func TestRoleSupportsAriaPropsRobust(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &RoleSupportsAriaPropsRule,
		[]rule_tester.ValidTestCase{
			// =============================================================
			// GROUP A — Real-world component patterns. Designed around
			// idioms a maintainer is likely to grep for when expanding the
			// rule (forwardRef wrappers, JSX namespaces, JSX member access,
			// fragment children).
			// =============================================================

			// `<Foo.Bar role="link" aria-expanded />` — JSX member expression
			// resolves to "Foo.Bar" via getJsxElementTypeString. The rule
			// doesn't gate on IsDOMElement, so it proceeds with explicit role
			// "link" and accepts aria-expanded.
			{Code: `<Foo.Bar role="link" aria-expanded />`, Tsx: true},
			// `<svg:circle role="img" aria-label="x" />` — JSX namespace.
			// "svg:circle" has no implicit role, but explicit role "img"
			// is recognized and aria-label is supported.
			{Code: `<svg:circle role="img" aria-label="x" />`, Tsx: true},
			// React.forwardRef-wrapped JSX usage — the inner JSX of the
			// forwardRef callback is what the listener fires on.
			{
				Code: `
const ForwardedAnchor = React.forwardRef((props, ref) => (
  <a ref={ref} href={props.to} aria-expanded={props.expanded} />
));
				`,
				Tsx: true,
			},
			// React.memo-wrapped component — outer JSX in render is fine,
			// inner usage of the memoized output triggers no rule violation.
			{
				Code: `
const Heading = React.memo(({ level, children }) => (
  <div role="heading" aria-level={level}>{children}</div>
));
				`,
				Tsx: true,
			},
			// Fragment children — each child is independently classified.
			// All four `<a>` children are valid (link supports aria-expanded).
			{
				Code: `
<>
  <a href="/a" aria-expanded />
  <a href="/b" aria-expanded />
  <a href="/c" aria-expanded />
  <a href="/d" aria-expanded />
</>
				`,
				Tsx: true,
			},
			// Conditional rendering pattern — the rule fires on every JSX
			// opening element produced by the source, regardless of the
			// surrounding ternary / logical expression.
			{
				Code: `cond ? <a href="/" aria-expanded /> : <button aria-pressed />`,
				Tsx:  true,
			},

			// =============================================================
			// GROUP B — TypeScript wrapper handling. Each TS wrapper category
			// has a distinct interaction with `getLiteralPropValue` (= our
			// LiteralPropStringValue) vs. `getPropValue` (= our
			// PropValueIsNullish-via-staticEval). These cases lock the
			// extractor wiring against silent regressions.
			// =============================================================

			// `as` wrapper around the role value — upstream's
			// LITERAL_TYPES.TSAsExpression is noop (returns null), so
			// `<div role={"button" as string}>` skips the rule (typeof
			// roleValue !== 'string').
			{Code: `<div role={"button" as string} aria-checked />`, Tsx: true},
			// Non-null assertion on the role value — upstream synthesizes
			// `"button!"` which doesn't match any role → skip.
			{Code: `<div role={someRole!} aria-checked />`, Tsx: true},
			// `satisfies` wrapper — LITERAL_TYPES has no entry, falls
			// through to noop → null → skip.
			{Code: `<div role={"button" satisfies string} aria-checked />`, Tsx: true},
			// Generic component — `<MyGeneric<string> aria-checked />` is
			// elementType "MyGeneric" with no implicit role → skip.
			{Code: `<MyGeneric<string> aria-checked />`, Tsx: true},
			// Parens around the role value — stripped by attributeInnerExpression.
			// "button" supports aria-pressed.
			{Code: `<div role={("button")} aria-pressed />`, Tsx: true},
			// Multiple parens layers — same as single.
			{Code: `<div role={(("button"))} aria-pressed />`, Tsx: true},
			// `as any` around the aria-prop value — unwrapped by staticEval
			// for nullish detection. `aria-checked={null as any}` is nullish
			// and skipped.
			{Code: `<a href="/" aria-checked={null as any} />`, Tsx: true},
			// Non-null assertion on a nullish aria-prop value — synthesizes
			// "null!" or "undefined!"; both are non-empty strings, NOT
			// nullish, so the attr proceeds. But link doesn't support
			// aria-checked, so this should REPORT — covered in invalid below.

			// =============================================================
			// GROUP C — ARIA prop value type coverage. Probes upstream's
			// `getPropValue(prop) != null` filter for every literal value
			// kind that JS can produce. Upstream rejects ONLY null and
			// undefined; everything else (including false, 0, NaN,
			// Infinity, "", empty array) proceeds to the membership check.
			// =============================================================

			// All five "definitely nullish" forms on a link element. None
			// should report aria-checked — the value-filter discards the attr
			// before the membership check.
			{Code: `<a href="/" aria-checked={null} />`, Tsx: true},
			{Code: `<a href="/" aria-checked={undefined} />`, Tsx: true},
			{Code: `<a href="/" aria-checked={(null)} />`, Tsx: true},
			{Code: `<a href="/" aria-checked={(undefined)} />`, Tsx: true},
			// Empty JsxExpression — tsgo synthesizes for malformed input;
			// upstream's TYPES has no entry → null fallback → skip.
			{Code: `<a href="/" aria-checked={} />`, Tsx: true},

			// =============================================================
			// GROUP D — Spread + literal interactions. Upstream `getProp`
			// walks literal-spread objects to find `role`, but the
			// per-attribute filter `prop.type !== 'JSXSpreadAttribute'`
			// means aria-* props inside spreads are NOT validated.
			// =============================================================

			// Literal spread with `role` — getProp walks in and finds
			// "button". aria-pressed IS in button's supported set.
			{
				Code: `<div {...{role: "button"}} aria-pressed />`,
				Tsx:  true,
			},
			// Spread of an Identifier — opaque. The role lookup falls back
			// to the implicit-role table; div has no implicit role → skip.
			// aria-checked is therefore silently accepted.
			{Code: `<div {...rest} aria-checked />`, Tsx: true},
			// Literal spread containing an aria-* key — the spread is opaque
			// for the per-attribute filter, so aria-checked inside the spread
			// is NOT validated. With explicit role="link", aria-checked from
			// outside would fail — but here it's INSIDE the spread, so it
			// passes.
			{
				Code: `<a href="/" role="link" {...{"aria-checked": true}} />`,
				Tsx:  true,
			},
			// Multiple spreads — first defines a non-role identifier-only
			// spread (opaque); rule walks both spreads looking for role
			// per upstream's `getProp` semantics, finds none, and falls back
			// to implicit role. div has none → skip → no report.
			{Code: `<div {...a} {...b} aria-checked />`, Tsx: true},
			// Spread followed by explicit role — explicit wins (getProp
			// returns the first match in attribute order).
			{
				Code: `<div {...rest} role="button" aria-pressed />`,
				Tsx:  true,
			},
			// Explicit role followed by spread — same.
			{
				Code: `<div role="button" {...rest} aria-pressed />`,
				Tsx:  true,
			},

			// =============================================================
			// GROUP E — Components map + polymorphic combinations. Locks in
			// the GetElementType resolution chain for the patterns design
			// systems actually ship: a polymorphic `as` prop layered on top
			// of a default-element components map.
			// =============================================================

			// Components map: MyLink → a. <MyLink href="/" aria-expanded />
			// resolves to elementType "a" → implicit "link" → aria-expanded
			// is supported.
			{
				Code:     `<MyLink href="/" aria-expanded />`,
				Tsx:      true,
				Settings: componentsSettingsLink,
			},
			// Components map: MyImage → img. Without alt empty, implicit
			// role is "img" which supports aria-busy.
			{
				Code:     `<MyImage src="/x.png" aria-busy />`,
				Tsx:      true,
				Settings: componentsSettingsLink,
			},
			// PolymorphicAllowList: only "Box" gets `as`-prop substitution.
			// `<Container as="a" href="/" aria-checked />` does NOT substitute
			// (Container not in allowList) → "Container" has no implicit role
			// → skip.
			{
				Code:     `<Container as="a" href="/" aria-checked />`,
				Tsx:      true,
				Settings: polymorphicAllowListSettings,
			},
			// Components + polymorphic: Box → div by default; with as="a"
			// the polymorphic substitution wins over the components mapping
			// (substitution runs FIRST, then components map; if "a" isn't a
			// components key, it's used directly).
			{
				Code:     `<Box as="a" href="/" aria-expanded />`,
				Tsx:      true,
				Settings: componentsAndPolymorphic,
			},
			// Components + polymorphic: bare Box (no `as`) → div via
			// components map → no implicit role → skip.
			{
				Code:     `<Box aria-checked />`,
				Tsx:      true,
				Settings: componentsAndPolymorphic,
			},

			// =============================================================
			// GROUP F — Multi-line / formatted JSX. The listener fires on
			// the opening element regardless of layout; the diagnostic
			// position is the opening element's position.
			// =============================================================

			// Attributes on separate lines — every aria-* attribute is on
			// link, all valid.
			{
				Code: `
<a
  href="/"
  aria-expanded
  aria-busy
  aria-controls="x"
/>
				`,
				Tsx: true,
			},
			// Comments between attributes — common in real source.
			{
				Code: `
<a
  // primary nav
  href="/"
  /* expandable submenu */
  aria-expanded
/>
				`,
				Tsx: true,
			},

			// =============================================================
			// GROUP G — Listener boundary tests across nested elements.
			// Each opening element is independently checked; outer-element
			// state must not bleed into inner.
			// =============================================================

			// Outer div has no implicit role; inner anchor has implicit link.
			// Both attributes (outer aria-controls on div, inner aria-expanded
			// on a) are valid via different paths.
			{
				Code: `
<div aria-controls="panel">
  <a href="/" aria-expanded />
</div>
				`,
				Tsx: true,
			},
			// Same role, nested — both valid; locks "no listener bleed".
			{
				Code: `
<div role="heading" aria-level={2}>
  <span role="heading" aria-level={3} />
</div>
				`,
				Tsx: true,
			},

			// =============================================================
			// GROUP H — Role value extraction edge cases.
			// =============================================================

			// `<div role="presentation" aria-checked />` — explicit role
			// "presentation" supports aria-checked? Per aria-query, no —
			// presentation has very limited supported props. Wait: that's
			// invalid then. Move to invalid section.

			// (Removed: `<div role="none" aria-busy />` — `none` actually
			// has an EMPTY supported-props set per aria-query, so EVERY
			// aria-* attribute is invalid on it. Lock the empty-set
			// behavior in the invalid section instead.)
			// Trailing whitespace in role — aria-query keys have no trailing
			// whitespace, so `roles.get("button ")` is undefined → skip the
			// rule entirely. aria-checked is therefore silently accepted.
			{Code: `<div role="button " aria-checked />`, Tsx: true},
			// Leading whitespace — same.
			{Code: `<div role=" button" aria-checked />`, Tsx: true},
			// HTML-entity-decoded role with aria-pressed (button supports it).
			{Code: `<div role="butto&#110;" aria-pressed />`, Tsx: true},
			// NoSubstitutionTemplateLiteral as role.
			{Code: "<div role={`button`} aria-pressed />", Tsx: true},
			// TemplateExpression with substitution — synthesizes a placeholder
			// string that doesn't match any role → skip.
			{Code: "<div role={`but${ton}`} aria-checked />", Tsx: true},

			// =============================================================
			// GROUP I — Implicit-role table edge cases not in upstream tests.
			// =============================================================

			// `<a>` boolean-form href → href is present (boolean) → implicit
			// link role applies. aria-expanded is supported.
			{Code: `<a href aria-expanded />`, Tsx: true},
			// `<a href={someVar}>` — href present (any value) → implicit link
			// applies. aria-expanded supported.
			{Code: `<a href={someVar} aria-expanded />`, Tsx: true},
			// `<a href={null}>` — null is still presence; implicit link applies.
			{Code: `<a href={null} aria-expanded />`, Tsx: true},
			// `<img>` without src or alt — implicit role "img" applies.
			// aria-busy IS supported on img.
			{Code: `<img aria-busy />`, Tsx: true},
			// `<img src="foo.SVG">` — uppercase .SVG is NOT matched by
			// upstream (case-sensitive `.svg` substring). Implicit role
			// stays "img"; aria-busy still valid.
			{Code: `<img src="foo.SVG" aria-busy />`, Tsx: true},
			// `<select multiple={false}>` — getLiteralPropValue returns false
			// (boolean), `false && ...` short-circuits → falls to size check
			// → no size → combobox. aria-expanded supported on combobox.
			{Code: `<select multiple={false} aria-expanded />`, Tsx: true},
			// `<select size={1}>` — 1 > 1 false → combobox.
			{Code: `<select size={1} aria-expanded />`, Tsx: true},
			// `<select size={2}>` — 2 > 1 true → listbox. aria-multiselectable
			// is in listbox's supported set.
			{Code: `<select size={2} aria-multiselectable />`, Tsx: true},
		},

		[]rule_tester.InvalidTestCase{
			// =============================================================
			// GROUP B-INVERSE — Non-null assertion on nullish aria-prop value
			// synthesizes a non-empty string ("null!" / "undefined!"), which
			// is NOT nullish, so the attr proceeds to the membership check
			// and reports.
			// =============================================================
			{
				Code: `<a href="/" aria-checked={null!} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},

			// =============================================================
			// GROUP C-INVERSE — Non-nullish primitive values. Each must
			// trigger the report; locks in that the value-filter discards
			// ONLY null/undefined, not the JS-falsy values.
			// =============================================================

			// Boolean false on a link element.
			{
				Code: `<a href="/" aria-checked={false} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// Empty string.
			{
				Code: `<a href="/" aria-checked="" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// Numeric zero.
			{
				Code: `<a href="/" aria-checked={0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// NaN — upstream getPropValue returns NaN; NaN != null is true.
			{
				Code: `<a href="/" aria-checked={NaN} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// Identifier — non-undefined identifiers stringify to their
			// name → non-null → proceeds. aria-checked invalid for link.
			{
				Code: `<a href="/" aria-checked={someVar} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// CallExpression — getPropValue returns a synthesized string
			// → non-null → proceeds → REPORT.
			{
				Code: `<a href="/" aria-checked={fn()} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},

			// =============================================================
			// GROUP D-INVERSE — Spread interactions where the report comes
			// from an outside-the-spread aria-* prop.
			// =============================================================

			// Literal spread with role + outside aria — outside attr is
			// validated against the spread-supplied role.
			{
				Code: `<div {...{role: "link"}} aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "div", false),
					Line:      1, Column: 1,
				}},
			},
			// Spread + multiple outside aria props — each reports independently.
			// aria-checked and aria-selected are both invalid on button; aria-expanded
			// IS in button's supported set so cannot be the second offender.
			{
				Code: `<div {...rest} role="button" aria-checked aria-selected />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "role-supports-aria-props",
						Message:   errorMessage("aria-checked", "button", "div", false),
						Line:      1, Column: 1,
					},
					{
						MessageId: "role-supports-aria-props",
						Message:   errorMessage("aria-selected", "button", "div", false),
						Line:      1, Column: 1,
					},
				},
			},

			// =============================================================
			// GROUP E-INVERSE — Components map + polymorphic interactions
			// triggering reports.
			// =============================================================

			// MyImage → img; aria-checked is invalid on img.
			{
				Code:     `<MyImage src="/x.png" aria-checked />`,
				Tsx:      true,
				Settings: componentsSettingsLink,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "img", "img", true),
					Line:      1, Column: 1,
				}},
			},
			// `<Box as="a" href="/" aria-checked />` — Box+as substitution →
			// "a" → implicit "link" → aria-checked invalid.
			{
				Code:     `<Box as="a" href="/" aria-checked />`,
				Tsx:      true,
				Settings: componentsAndPolymorphic,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// Anchor (mapped to "a") with no `as` — Anchor → a via components
			// map → implicit "link" → aria-checked invalid.
			{
				Code:     `<Anchor href="/" aria-checked />`,
				Tsx:      true,
				Settings: componentsAndPolymorphic,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},

			// =============================================================
			// GROUP G-INVERSE — Nested elements where ONLY the inner reports.
			// Locks in that the listener boundary for each opening element
			// independently determines its own report state.
			// =============================================================

			// Outer div is innocuous; inner anchor reports.
			{
				Code: `
<div>
  <a href="/" aria-checked />
</div>
				`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      3, Column: 3,
				}},
			},

			// =============================================================
			// GROUP H-INVERSE — `presentation` role doesn't accept most
			// aria-* props. Lock in the explicit-role + restricted-props
			// path.
			// =============================================================

			// `presentation` role supports only the global ARIA props inherited
			// from "roletype" (aria-atomic, aria-busy, …). aria-checked is
			// not in that set.
			{
				Code: `<div role="presentation" aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "presentation", "div", false),
					Line:      1, Column: 1,
				}},
			},

			// `none` role has an EMPTY supported-props set per aria-query.
			// EVERY aria-* attribute reports — even the otherwise-global
			// ones like aria-busy and aria-atomic. Locks in the empty-set
			// branch (other roles inherit the global set; none does not).
			{
				Code: `<div role="none" aria-busy />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-busy", "none", "div", false),
					Line:      1, Column: 1,
				}},
			},

			// =============================================================
			// GROUP I-INVERSE — Implicit-role table inverse cases.
			// =============================================================

			// `<select multiple aria-checked />` → listbox; aria-checked
			// invalid for listbox.
			{
				Code: `<select multiple aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "listbox", "select", true),
					Line:      1, Column: 1,
				}},
			},
			// `<select size={5} aria-checked />` — size > 1 → listbox; same.
			{
				Code: `<select size={5} aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "listbox", "select", true),
					Line:      1, Column: 1,
				}},
			},
			// `<img>` no alt no src — implicit "img"; aria-checked invalid.
			{
				Code: `<img aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "img", "img", true),
					Line:      1, Column: 1,
				}},
			},
			// `<button>` always carries implicit role "button"; aria-checked
			// is invalid on button.
			{
				Code: `<button aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "button", "button", true),
					Line:      1, Column: 1,
				}},
			},
			// `<datalist aria-checked />` — implicit "listbox" doesn't accept
			// aria-checked.
			{
				Code: `<datalist aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "listbox", "datalist", true),
					Line:      1, Column: 1,
				}},
			},
			// `<details aria-checked />` — implicit "group"; aria-checked invalid.
			{
				Code: `<details aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "group", "details", true),
					Line:      1, Column: 1,
				}},
			},

			// =============================================================
			// GROUP J — Position assertions for multi-attribute / multi-line
			// edge cases. Locks in that the diagnostic always lands on the
			// opening element, not the attribute.
			// =============================================================

			// Self-closing element with the offending attr at the end.
			{
				Code: `<a href="/" data-x="y" onClick={fn} aria-checked />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      1, Column: 1,
				}},
			},
			// Multi-line where the offending attr is on a later line — the
			// report still lands on column 1 of the opening element's first
			// line.
			{
				Code: `
<a
  href="/"
  data-x="y"
  onClick={fn}
  aria-checked
/>
				`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage("aria-checked", "link", "a", true),
					Line:      2, Column: 1,
				}},
			},
		},
	)
}
