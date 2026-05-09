package aria_unsupported_elements

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// allDomElements mirrors aria-query's `dom.keys()` in order. Used to drive
// the procedural test generation upstream relies on
// (`domElements.map(...)` in `aria-unsupported-elements-test.js`).
var allDomElements = []string{
	"a", "abbr", "acronym", "address", "applet", "area", "article", "aside",
	"audio", "b", "base", "bdi", "bdo", "big", "blink", "blockquote", "body",
	"br", "button", "canvas", "caption", "center", "cite", "code", "col",
	"colgroup", "content", "data", "datalist", "dd", "del", "details", "dfn",
	"dialog", "dir", "div", "dl", "dt", "em", "embed", "fieldset", "figcaption",
	"figure", "font", "footer", "form", "frame", "frameset", "h1", "h2", "h3",
	"h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "i", "iframe",
	"img", "input", "ins", "kbd", "keygen", "label", "legend", "li", "link",
	"main", "map", "mark", "marquee", "menu", "menuitem", "meta", "meter",
	"nav", "noembed", "noscript", "object", "ol", "optgroup", "option",
	"output", "p", "param", "picture", "pre", "progress", "q", "rp", "rt",
	"rtc", "ruby", "s", "samp", "script", "section", "select", "small",
	"source", "spacer", "span", "strike", "strong", "style", "sub", "summary",
	"sup", "table", "tbody", "td", "textarea", "tfoot", "th", "thead", "time",
	"title", "tr", "track", "tt", "u", "ul", "var", "video", "wbr", "xmp",
}

// metaSettings mirrors upstream's `settings: { 'jsx-a11y': { components: {
// Meta: 'meta' } } }` — used to test that the components map is honored when
// resolving the element type for reserved-element matching.
var metaSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Meta": "meta",
		},
	},
}

// TestAriaUnsupportedElements covers the upstream valid/invalid suite from
// `__tests__/src/rules/aria-unsupported-elements-test.js` plus rslint
// edge-case lock-ins (case-insensitive matching, namespaced names,
// JSXSpreadAttribute exclusion, polymorphicPropName, etc.).
func TestAriaUnsupportedElements(t *testing.T) {
	var validCases []rule_tester.ValidTestCase
	var invalidCases []rule_tester.InvalidTestCase

	// ---- Upstream `roleValidityTests` ----
	// For each DOM element: non-reserved gets a `role` attribute and
	// reserved gets nothing. Both are valid because the rule only fires on
	// reserved elements that carry an `aria-*` / `role` attribute.
	for _, el := range allDomElements {
		role := "role"
		if _, isReserved := reservedElements[el]; isReserved {
			role = ""
		}
		validCases = append(validCases, rule_tester.ValidTestCase{
			Code: fmt.Sprintf(`<%s %s />`, el, role),
			Tsx:  true,
		})
	}

	// ---- Upstream `ariaValidityTests` ----
	// Same shape as above but with `aria-hidden` instead of `role`. Plus the
	// upstream `<fake aria-hidden />` lockdown — `fake` is not in the dom
	// map, so isReservedNodeType is false and the rule must not fire even
	// though the attribute is on the invalid list.
	for _, el := range allDomElements {
		attr := "aria-hidden"
		if _, isReserved := reservedElements[el]; isReserved {
			attr = ""
		}
		validCases = append(validCases, rule_tester.ValidTestCase{
			Code: fmt.Sprintf(`<%s %s />`, el, attr),
			Tsx:  true,
		})
	}
	validCases = append(validCases, rule_tester.ValidTestCase{
		Code: `<fake aria-hidden />`,
		Tsx:  true,
	})

	// ---- Extra valid cases (Dimension 4 + upstream-walk lockdowns) ----
	validCases = append(validCases,
		// Custom (capitalized) component without a components-map entry: the
		// raw type is "Custom", not in the dom set → not reserved → no
		// report even with aria-hidden.
		rule_tester.ValidTestCase{Code: `<Custom aria-hidden />`, Tsx: true},
		// Pure spread on a reserved element. JsxSpreadAttribute is skipped
		// before the name lookup, so this must not throw and must not report.
		rule_tester.ValidTestCase{Code: `<base {...props} />`, Tsx: true},
		// Reserved element with a non-aria/non-role attribute: data-* and
		// arbitrary names are not on the invalid list and must pass through.
		rule_tester.ValidTestCase{Code: `<base data-foo="x" />`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<meta charset="UTF-8" />`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<link rel="stylesheet" href="x" />`, Tsx: true},
		// Non-reserved element with role/aria — must remain valid.
		rule_tester.ValidTestCase{Code: `<div role="button" aria-pressed="true" />`, Tsx: true},
		// Namespaced attribute on a reserved element: tsgo represents
		// `aria:hidden` as a JsxNamespacedName; GetJsxPropName returns
		// "aria:hidden" (with colon, not hyphen), which is NOT in
		// invalidAttributes. Locks the namespaced-name branch.
		rule_tester.ValidTestCase{Code: `<base aria:hidden />`, Tsx: true},
		// Reserved element where the components-map points elsewhere: a
		// JSX tag named `Title` mapped to `div` is no longer reserved.
		rule_tester.ValidTestCase{
			Code: `<Title aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Title": "div",
					},
				},
			},
		},
	)

	// ---- Upstream `invalidRoleValidityTests` ----
	// Each reserved element with `role {...props}` reports `role`. The
	// JsxSpreadAttribute is silently skipped — only the named `role`
	// attribute triggers the diagnostic.
	for _, el := range allDomElements {
		if _, isReserved := reservedElements[el]; !isReserved {
			continue
		}
		invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
			Code: fmt.Sprintf(`<%s role {...props} />`, el),
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("role"),
			}},
		})
	}
	// `<Meta aria-hidden />` with components: { Meta: 'meta' } — components
	// map is honored before the reserved-element check.
	invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
		Code:     `<Meta aria-hidden />`,
		Tsx:      true,
		Settings: metaSettings,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "unsupported",
			Message:   errorMessage("aria-hidden"),
		}},
	})

	// ---- Upstream `invalidAriaValidityTests` ----
	// Each reserved element with `aria-hidden aria-role="none" {...props}`
	// reports exactly one diagnostic for `aria-hidden`. `aria-role` is NOT
	// a real ARIA attribute (it's a typo), so it is NOT on the invalid list
	// and must NOT report — locks the "match by exact lowercase name"
	// branch.
	for _, el := range allDomElements {
		if _, isReserved := reservedElements[el]; !isReserved {
			continue
		}
		invalidCases = append(invalidCases, rule_tester.InvalidTestCase{
			Code: fmt.Sprintf(`<%s aria-hidden aria-role="none" {...props} />`, el),
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		})
	}

	// ---- Extra invalid lockdowns ----
	invalidCases = append(invalidCases,
		// Case-insensitive name matching: upstream lowercases via
		// `propName(prop).toLowerCase()`. Locks the lowercase normalization.
		rule_tester.InvalidTestCase{
			Code: `<base ROLE />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("role"),
				Line:      1, Column: 1, EndLine: 1, EndColumn: 14,
			}},
		},
		rule_tester.InvalidTestCase{
			Code: `<base ARIA-HIDDEN />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Multiple invalid attributes on the same element: upstream's
		// `forEach` reports each one separately.
		rule_tester.InvalidTestCase{
			Code: `<base role aria-hidden aria-label="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsupported", Message: errorMessage("role")},
				{MessageId: "unsupported", Message: errorMessage("aria-hidden")},
				{MessageId: "unsupported", Message: errorMessage("aria-label")},
			},
		},
		// Diagnostic position: upstream reports on the JSXOpeningElement
		// node, which spans `<` through `>` of the opening tag. tsgo's
		// JsxSelfClosingElement spans the whole self-closing tag including
		// the trailing `/>`. Locks the position to detect listener bleed.
		rule_tester.InvalidTestCase{
			Code: `<base aria-hidden />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
				Line:      1, Column: 1, EndLine: 1, EndColumn: 21,
			}},
		},
		// Paired opening/closing form (not self-closing): the
		// JsxOpeningElement listener fires on `<style>`, span ends at `>`.
		rule_tester.InvalidTestCase{
			Code: `<style aria-hidden></style>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
				Line:      1, Column: 1, EndLine: 1, EndColumn: 20,
			}},
		},
		// polymorphicPropName: `<Box as="base" aria-hidden />` resolves to
		// `base` via the polymorphic prop, so the reserved-element check
		// must fire even though the source tag is `Box`.
		rule_tester.InvalidTestCase{
			Code: `<Box as="base" aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Components-map miss + polymorphicPropName combined: rawType is
		// `Foo`, the polymorphic prop maps it to `meta`, then the components
		// map (which keys on the post-poly value) is consulted in the
		// upstream order. Locks that the chain ends at `meta` and the
		// reserved-element check fires.
		rule_tester.InvalidTestCase{
			Code: `<Foo as="meta" role="none" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("role"),
			}},
		},
		// Components map points at a reserved element: the JSX tag `Foo`
		// becomes `meta`, and `meta` is reserved → report.
		rule_tester.InvalidTestCase{
			Code: `<Foo aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Foo": "meta",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Reserved within non-reserved: outer `div` is not reserved (no
		// report on it), inner `base` is reserved with aria-hidden (report).
		// Locks that the listener doesn't bleed across nesting boundaries.
		rule_tester.InvalidTestCase{
			Code: `<div><base aria-hidden /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
				Line:      1, Column: 6, EndLine: 1, EndColumn: 26,
			}},
		},
		// Reserved within reserved: both fire independently. The opening
		// `<head>` listener and the inner `<base />` listener each report.
		rule_tester.InvalidTestCase{
			Code: `<head role="x"><base aria-hidden /></head>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsupported", Message: errorMessage("role")},
				{MessageId: "unsupported", Message: errorMessage("aria-hidden")},
			},
		},
		// Reserved element with mixed valid + invalid attributes — only the
		// invalid ones report, valid HTML attributes pass through silently.
		rule_tester.InvalidTestCase{
			Code: `<base href="x" target="_blank" aria-hidden />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Spread interleaved with named attributes on a reserved element:
		// spreads in any position must be silently skipped.
		rule_tester.InvalidTestCase{
			Code: `<base {...a} role {...b} aria-hidden {...c} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsupported", Message: errorMessage("role")},
				{MessageId: "unsupported", Message: errorMessage("aria-hidden")},
			},
		},
		// Multi-line attributes: report position is on the opening element
		// span, which extends across lines. Locks the multi-line position.
		rule_tester.InvalidTestCase{
			Code: "<base\n  aria-hidden\n  role=\"none\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsupported",
					Message:   errorMessage("aria-hidden"),
					Line:      1, Column: 1, EndLine: 4, EndColumn: 3,
				},
				{
					MessageId: "unsupported",
					Message:   errorMessage("role"),
					Line:      1, Column: 1, EndLine: 4, EndColumn: 3,
				},
			},
		},
		// Boolean form of role on a reserved element — upstream's
		// propName treats this identically to `role="x"`. Locks the
		// boolean-attribute branch.
		rule_tester.InvalidTestCase{
			Code: `<base role />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("role"),
			}},
		},
		// Reserved fragment-style element (svg foreign tag with `<title>`):
		// the bare `<title>` is reserved per aria-query (it's an HTML head
		// element). Locks that the listener fires regardless of the
		// surrounding JSX context.
		rule_tester.InvalidTestCase{
			Code: `<>{cond && <title aria-hidden>x</title>}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Boolean form of an aria-* attribute (no value at all). Upstream's
		// propName resolves to the bare name regardless of the value form.
		rule_tester.InvalidTestCase{
			Code: `<base aria-label />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-label"),
			}},
		},
		// Reserved tag with children (semantically illegal HTML for void
		// elements like base/meta/link, but JSX accepts the syntax). Locks
		// that body presence does not change reporting.
		rule_tester.InvalidTestCase{
			Code: `<base aria-hidden>x</base>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Self-closing without space before `/>`. tsgo tokenizes the same
		// way; locks the parsing-shape independence.
		rule_tester.InvalidTestCase{
			Code: `<base aria-hidden/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// JSX inside Array.map callback — common React pattern. Each
		// iteration produces a separate JSX element; the listener must fire
		// on every one (and the diagnostics list collapses by source position
		// — only one source location, so one report).
		rule_tester.InvalidTestCase{
			Code: `function L({xs}) { return xs.map(x => <base aria-hidden key={x} />); }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// JSX inside class component render method.
		rule_tester.InvalidTestCase{
			Code: `class C { render() { return <link aria-hidden />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// Conditional render via && — the JSX expression child wraps the
		// reserved element. Listener fires through the JsxExpression child.
		rule_tester.InvalidTestCase{
			Code: `function F() { return <div>{cond && <meta aria-hidden />}</div>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// TypeScript type assertion on attribute value. We never read the
		// value, so the assertion is irrelevant — but locks that the wrapper
		// doesn't somehow disrupt name extraction.
		rule_tester.InvalidTestCase{
			Code: `<base aria-hidden={true as boolean} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// `as const` assertion on attribute value.
		rule_tester.InvalidTestCase{
			Code: `<base aria-label={"x" as const} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-label"),
			}},
		},
		// JSX inside a generic function — locks that surrounding TS
		// constructs don't affect listener firing.
		rule_tester.InvalidTestCase{
			Code: `function f<T>(x: T) { return <base aria-hidden />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
		// polymorphicAllowList kicks in: `Box` IS in the allow list, so
		// `as="meta"` substitutes rawType to "meta" → reserved → report.
		rule_tester.InvalidTestCase{
			Code: `<Box as="meta" aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Box"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsupported",
				Message:   errorMessage("aria-hidden"),
			}},
		},
	)

	// Extra valid lockdowns for shapes that should NOT report:
	validCases = append(validCases,
		// Member-expression tag — `Foo.base` is NOT the same string as
		// "base" in jsx-ast-utils' elementType, so it's not in the dom set.
		rule_tester.ValidTestCase{Code: `<Foo.base aria-hidden />`, Tsx: true},
		// Capitalized member chain — same reasoning.
		rule_tester.ValidTestCase{Code: `<lib.Title aria-hidden />`, Tsx: true},
		// JSX namespaced tag name — `svg:title` is not in dom (only the
		// bare `title` is). Locks that namespaced tags don't accidentally
		// match the bare name.
		rule_tester.ValidTestCase{Code: `<svg:title aria-hidden />`, Tsx: true},
		// Custom-element tag name (with hyphens) — `my-element` is not in
		// the dom set; even custom elements that look HTML-ish must not
		// false-positive.
		rule_tester.ValidTestCase{Code: `<my-element aria-hidden />`, Tsx: true},
		// SVG element name (`image` is SVG, not HTML) — not in the dom set
		// of aria-query; should NOT report.
		rule_tester.ValidTestCase{Code: `<image aria-hidden />`, Tsx: true},
		// Comments around / between attributes — JSX expression body can
		// contain comments. Must not interfere with name extraction.
		rule_tester.ValidTestCase{
			Code: `<base /* leading */ data-foo="x" /* trailing */ />`,
			Tsx:  true,
		},
		// Spread of a literal containing an aria-* key. Upstream's rule
		// skips JSXSpreadAttribute entirely (does NOT inspect the spread
		// argument), so `{...{ 'aria-hidden': true }}` is not reported.
		// rslint matches that limitation deliberately. Locks the
		// no-deep-spread-inspection contract.
		rule_tester.ValidTestCase{
			Code: `<base {...{'aria-hidden': true}} />`,
			Tsx:  true,
		},
		// PolymorphicPropName with non-literal value — upstream's
		// `getLiteralPropValue` returns null/undefined, so the substitution
		// is skipped and rawType stays "Box" (not reserved).
		rule_tester.ValidTestCase{
			Code: `<Box as={dynamicTag} aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
		},
		// PolymorphicPropName with empty string — upstream's truthiness
		// check (`!!getLiteralPropValue(prop)`) is false, so the
		// substitution is skipped and rawType stays "Box".
		rule_tester.ValidTestCase{
			Code: `<Box as="" aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
		},
		// PolymorphicPropName with `as={false}` — boolean false is falsy,
		// substitution skipped, rawType stays "Box".
		rule_tester.ValidTestCase{
			Code: `<Box as={false} aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
		},
		// polymorphicAllowList excludes the source tag — substitution
		// skipped, rawType stays "Box" (not reserved).
		rule_tester.ValidTestCase{
			Code: `<Box as="meta" aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"OtherComponent"},
				},
			},
		},
		// Empty jsx-a11y settings object — should not panic, behaves like
		// no settings.
		rule_tester.ValidTestCase{
			Code:     `<Foo aria-hidden />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{}},
		},
		// Nil settings — falls through cleanly.
		rule_tester.ValidTestCase{
			Code: `<Foo aria-hidden />`,
			Tsx:  true,
		},
		// Components map with non-string value — must be ignored, not crash.
		rule_tester.ValidTestCase{
			Code: `<Foo aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": 123},
				},
			},
		},
		// Components map missing target key — falls through, rawType
		// stays the JSX tag string ("Foo" → not reserved).
		rule_tester.ValidTestCase{
			Code: `<Foo aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Bar": "meta"},
				},
			},
		},
		// JSX inside a hook callback — common React pattern.
		rule_tester.ValidTestCase{
			Code: `function F() { useEffect(() => { renderer(<div role="button" />); }); }`,
			Tsx:  true,
		},
		// Reserved element BUT no aria/role — must remain valid even after
		// the rule sees it as reserved. Verifies that the early-return on
		// "no invalid attribute found" is correctly placed.
		rule_tester.ValidTestCase{
			Code: `<base href="https://example.com" target="_blank" />`,
			Tsx:  true,
		},
		// Reserved element with a `key` prop (React internal). `key` is
		// not on the invalid list — must pass.
		rule_tester.ValidTestCase{Code: `<base key="x" />`, Tsx: true},
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaUnsupportedElementsRule, validCases, invalidCases)
}
