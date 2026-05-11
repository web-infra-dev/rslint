package no_distracting_elements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoDistractingElementsExtras covers behavior the upstream test suite
// leaves unaudited — Dimension 1-4 universal edge shapes, options coverage
// (default vs explicit empty array vs custom list), settings coverage
// (components map / polymorphic prop / polymorphicAllowList), exact position
// assertions across the JSX surface, and the listener boundary between
// nested elements.
func TestNoDistractingElementsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDistractingElementsRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Dimension 1: AST node type / tag-shape variants
		// ============================================================

		// PropertyAccess tag — full type is "UI.Marquee", not "marquee".
		{Code: `<UI.Marquee />`, Tsx: true},
		// Namespaced tag — full type is "svg:marquee", not "marquee". Locks in
		// that namespaced names match the COMPOSITE form, not just the local
		// name.
		{Code: `<svg:marquee />`, Tsx: true},
		// Case-sensitive comparison: uppercase / mixed-case never match.
		{Code: `<MARQUEE />`, Tsx: true},
		{Code: `<Marquee />`, Tsx: true},
		{Code: `<MarQuee />`, Tsx: true},
		{Code: `<BLINK />`, Tsx: true},
		{Code: `<Blink />`, Tsx: true},

		// ---- Element kind survey: rule is a no-op for every non-listed tag ----
		{Code: `<a />`, Tsx: true},
		{Code: `<input />`, Tsx: true},
		{Code: `<Component />`, Tsx: true},

		// ---- Multi-segment PropertyAccess — full type is "A.B.C", not any
		// segment in isolation. Locks in that the joined dotted form is
		// what's compared.
		{Code: `<A.B.C />`, Tsx: true},
		// `<this.Foo />` — type is "this.Foo".
		{Code: `<this.Foo />`, Tsx: true},
		// PropertyAccess where the FINAL segment is lowercased "marquee" —
		// upstream's getElementType returns "UI.marquee" (the dotted whole),
		// which does not strict-equal "marquee". Differentially verified
		// against eslint-plugin-jsx-a11y v6.10.2 — no report.
		{Code: `<UI.marquee />`, Tsx: true},

		// ============================================================
		// Options: explicitly empty `elements` array disables every check
		// ============================================================
		// Upstream's `options.elements || DEFAULT_ELEMENTS` only falls back
		// when the LHS is JS-falsy. An empty array is JS-truthy, so it
		// replaces the default and the rule becomes a no-op.
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{}},
		},
		{
			Code:    `<blink />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{}},
		},

		// ============================================================
		// Options: custom `elements` list — types not in the list pass
		// ============================================================
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"blink"}},
		},
		{
			Code:    `<blink />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"marquee"}},
		},

		// ============================================================
		// Settings: components map without a matching entry leaves rawType
		// alone; `<Blink />` doesn't match the default elements list.
		// ============================================================
		{
			Code: `<Blink />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"OtherTag": "marquee"},
				},
			},
		},

		// ============================================================
		// Settings: polymorphic prop with allow-list NOT containing the tag —
		// rawType remains the React component name, no match.
		// ============================================================
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Bar"},
				},
			},
		},

		// ============================================================
		// Empty options object — equivalent to the default case
		// ============================================================
		{
			Code:    `<div />`,
			Tsx:     true,
			Options: map[string]interface{}{},
		},

		// ============================================================
		// `options.elements` is NOT a list — fallback to defaults; `<div />`
		// still passes.
		// ============================================================
		{
			Code:    `<div />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": "marquee"}, // wrong type, ignored
		},

		// `options.elements` is null — StringSliceOption returns nil →
		// fallback to default. `<div />` is still valid.
		{
			Code:    `<div />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": nil},
		},

		// rslint extension: an `elements` array containing only non-string
		// entries — StringSliceOption filters them all out, leaving an empty
		// `[]string`, which silences the rule. ESLint's JSON schema rejects
		// this at config-load time so it never reaches rule logic upstream;
		// rslint accepts the input and we lock the silenced behavior.
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{123, true}},
		},

		// ============================================================
		// Components map: map-to-non-distracting / non-string value
		// ============================================================
		// `<Foo />` aliased to a non-distracting tag — no report.
		{
			Code: `<Foo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": "div"},
				},
			},
		},
		// **Reverse aliasing**: a literal `<marquee />` mapped to "div" —
		// upstream's getElementType replaces "marquee" with "div" before the
		// lookup, so the rule does NOT report. Locks in that components-map
		// supports per-codebase opt-out, not just opt-in. Differentially
		// verified against eslint-plugin-jsx-a11y v6.10.2.
		{
			Code: `<marquee />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"marquee": "div"},
				},
			},
		},
		// Components map with a non-string value — upstream silently
		// preserves the original rawType (the type assertion fails).
		// `<Foo />` stays "Foo", not "marquee".
		{
			Code: `<Foo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": 123},
				},
			},
		},

		// ============================================================
		// Polymorphic prop edge cases (verified differentially)
		// ============================================================
		// `polymorphicPropName` set, but the `as` prop is missing on the
		// element — getElementType keeps rawType. `<Foo />` stays "Foo".
		{
			Code: `<Foo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		// `as={x}` (Identifier) — literalPropValue maps Identifier → null,
		// polymorphic doesn't replace, rawType stays "Foo".
		{
			Code: `<Foo as={x} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		// `as=""` — empty string is JS-falsy, polymorphic doesn't replace.
		{
			Code: `<Foo as="" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		// `as={cond ? "marquee" : "blink"}` — Conditional is noop in
		// LITERAL_TYPES → null → polymorphic doesn't replace.
		{
			Code: `<Foo as={cond ? "marquee" : "blink"} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		// Empty `polymorphicAllowList` — zero entries means no rawType
		// passes the allow check → polymorphic doesn't replace anything.
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{},
				},
			},
		},

		// Polymorphic REVERSE exemption: an intrinsic distracting tag
		// remapped to a non-distracting one via the `as` prop is silenced.
		// Mirrors the components-map reverse alias (`marquee: 'div'`) but
		// at the polymorphic layer — `<marquee as="div" />` resolves to
		// "div", which is not in the elements list. Differentially
		// verified against eslint-plugin-jsx-a11y v6.10.2.
		{
			Code: `<marquee as="div" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		{
			Code: `<blink as="span" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},

		// Polymorphic replaces rawType with a non-distracting intermediate
		// component name AND no components-map entry rebinds it back —
		// rule does not report. Locks in that the polymorphic step alone
		// is sufficient to silence; we don't need a components-map entry
		// for the replaced name to "exist".
		{
			Code: `<Foo as="Bar" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},

		// ============================================================
		// Defensive type handling: settings shapes that ARE truthy but
		// don't match expected types. All differentially verified against
		// eslint-plugin-jsx-a11y v6.10.2 — same observable behavior (no
		// report) on all the inputs that upstream doesn't crash on.
		// ============================================================
		// `settings['jsx-a11y']` is a string (not a map) — type assertion
		// in getJsxA11ySettings fails → no polymorphic / components.
		{
			Code:     `<Foo />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": "invalid"},
		},
		// `settings['jsx-a11y']` is null — same path.
		{
			Code:     `<Foo />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": nil},
		},
		// `components` is a string instead of a map — the inner type
		// assertion fails → components block is a no-op, rawType stays.
		{
			Code: `<Foo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"components": "invalid"},
			},
		},
		// `polymorphicPropName` is a number — upstream THROWS TypeError
		// here (`getProp(attrs, 123)` calls `123.toUpperCase()`). rslint
		// can't reproduce the throw under Go's static typing — the
		// `.(string)` assertion fails, polymorphicPropName becomes "",
		// the polymorphic block is skipped silently. This is a Go-natural
		// robustness divergence (no report rather than crashing the lint
		// run), not a rule-semantics divergence.
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": 123},
			},
		},
		// `polymorphicAllowList` containing only non-strings — every entry
		// fails the `.(string)` assertion, the for-loop never finds rawType
		// in the allow list, polymorphic doesn't replace.
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{123},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Dimension 1: paired element form — listener fires on the opening
		// tag. Position covers only the opening tag (`<marquee>`), not the
		// entire JsxElement. `<marquee>` is 9 characters → EndColumn 10.
		// ============================================================
		{
			Code: `<marquee>scrolling</marquee>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "distractingElement",
				Message:   "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
			}},
		},

		// Empty body: still reports.
		{
			Code:   `<marquee></marquee>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Position assertions on a self-closing element. `<marquee />` is
		// 11 characters → EndColumn 12 (exclusive).
		// ============================================================
		{
			Code: `<marquee />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "distractingElement",
				Message:   "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 12,
			}},
		},

		// Multi-line element — position must span the entire opening
		// self-closing tag.
		{
			Code: "<marquee\n  lang=\"en\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "distractingElement",
				Message:   "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
				Line:      1, Column: 1, EndLine: 3, EndColumn: 3,
			}},
		},

		// ============================================================
		// Dimension 2: same-kind nesting — both outer and inner report,
		// listener doesn't dedupe.
		// ============================================================
		{
			Code: `<marquee><marquee /></marquee>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
			},
		},

		// ============================================================
		// Boolean / numeric / spread attribute forms — none affect the
		// type comparison, but exercise the listener for shape coverage.
		// ============================================================
		{
			Code:   `<marquee draggable />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<marquee className="x" id={n} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Options: custom `elements` list — `<custom />` reports when
		// "custom" is in the list.
		// ============================================================
		{
			Code:    `<custom />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"custom"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "distractingElement",
				Message:   "Do not use <custom> elements as they can create visual accessibility issues and are deprecated.",
			}},
		},
		// Custom list with only one of the defaults — only that one fires.
		{
			Code:    `<blink />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"blink"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},

		// ============================================================
		// Options array shape — the rule_tester accepts both bare-map
		// and array-wrapped option shapes; cover the array form to lock
		// the JSON path through GetOptionsMap.
		// ============================================================
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"elements": []interface{}{"marquee"}}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Settings: components map with multiple aliases.
		// ============================================================
		{
			Code: `<CustomMarquee />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"CustomMarquee": "marquee"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Settings: polymorphic prop without an allow-list — every truthy
		// `as` value replaces rawType, so `<Foo as="marquee" />` reports.
		// ============================================================
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// Polymorphic with an allow-list that DOES include the rawType.
		{
			Code: `<Foo as="blink" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},

		// ============================================================
		// Listener boundary: marquee inside non-distracting wrapper, and
		// blink inside marquee — each fires independently.
		// ============================================================
		{
			Code: `<div><marquee /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "distractingElement",
				Message:   "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
				Line:      1, Column: 6, EndLine: 1, EndColumn: 17,
			}},
		},
		{
			Code: `<marquee><blink /></marquee>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				{MessageId: "distractingElement", Message: "Do not use <blink> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
			},
		},

		// ============================================================
		// Real-world component patterns — locks in that the listener fires
		// inside common React shapes.
		// ============================================================
		{
			Code:   `function Banner() { return <marquee>News</marquee>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `const x = items.map(item => <blink key={item.id}>{item.text}</blink>)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},
		{
			Code:   `const x = cond ? <marquee /> : <div />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code: `class C { render() { return <div><marquee /><blink /></div>; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedMarqueeError,
				expectedBlinkError,
			},
		},

		// ============================================================
		// Deep nesting (3 levels): each opening tag fires independently.
		// Position assertions lock in tsgo column counting on adjacent
		// JSX elements: outer `<marquee>` cols 1-9, middle 10-18, inner
		// `<marquee />` cols 19-29.
		// ============================================================
		{
			Code: `<marquee><marquee><marquee /></marquee></marquee>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				{MessageId: "distractingElement", Message: "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
					Line: 1, Column: 19, EndLine: 1, EndColumn: 30},
			},
		},

		// ============================================================
		// JsxFragment surrounding a distracting element. The Fragment
		// itself is NOT a JsxOpeningElement / JsxSelfClosingElement and
		// is silently skipped, but the inner marquee still fires.
		// ============================================================
		{
			Code:   `<><marquee /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// JSX inside an expression container (children, attribute value,
		// logical short-circuit, optional chain) — listener fires
		// regardless of where the JSX appears.
		// ============================================================
		{
			Code:   `<div>{<marquee />}</div>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<div>{cond && <marquee />}</div>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<div content={<marquee />} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<Provider value={data}>{<marquee />}</Provider>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Components map: map-to-distracting and map-to-self
		// ============================================================
		// `<Foo />` aliased to "marquee" — reports.
		{
			Code: `<Foo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": "marquee"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// `<marquee />` aliased to "marquee" (self-map) — still reports.
		{
			Code: `<marquee />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"marquee": "marquee"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Polymorphic + components combination: polymorphic replaces FIRST,
		// then components looks up the (already-replaced) rawType. So
		// `<Foo as="marquee" />` with `components: { Foo: 'div' }` still
		// reports — polymorphic turns rawType into "marquee", and there's
		// no "marquee" key in components, so the lookup is a no-op.
		// Differentially verified against eslint-plugin-jsx-a11y v6.10.2.
		// ============================================================
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Foo": "div"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// Same shape, components mapping `Foo: 'blink'` — also reports
		// `marquee` (NOT blink), because polymorphic already changed
		// rawType to "marquee" before components ran.
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Foo": "blink"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// Polymorphic + components CHAIN: rawType "Foo" → polymorphic
		// replaces with "Bar" → components map rebinds "Bar" to "marquee".
		// Locks in that the components step looks up the POST-polymorphic
		// rawType, not the original tag name. Differentially verified
		// against eslint-plugin-jsx-a11y v6.10.2.
		{
			Code: `<Foo as="Bar" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Bar": "marquee"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Polymorphic prop value forms — JsxExpression-wrapped string
		// literal, no-substitution template literal — both extract through
		// LITERAL_TYPES and replace rawType.
		// ============================================================
		{
			Code: `<Foo as={"marquee"} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code: "<Foo as={`marquee`} />",
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Real-world patterns (extended): HOC, filter+map chain, array
		// literal, function arg, object property, forwardRef pattern.
		// ============================================================
		// HOC wrapping returning a distracting element.
		{
			Code:   `const Banner = withTheme(() => <marquee />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// Iterator-chain: `.filter().map(...)` returning JSX.
		{
			Code:   `const x = items.filter(x => x).map(x => <marquee key={x.id} />)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// Array literal with multiple distracting elements.
		{
			Code: `const items = [<marquee key="1" />, <blink key="2" />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedMarqueeError,
				expectedBlinkError,
			},
		},
		// JSX as a function argument.
		{
			Code:   `wrap(<marquee />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// JSX as an object-literal property value.
		{
			Code:   `const x = { content: <marquee /> };`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// React.forwardRef + spread props — listener still fires on the
		// inner JSX.
		{
			Code:   `const Marquee = React.forwardRef((props, ref) => <marquee ref={ref} {...props} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// Defensive type handling (invalid side): mixed-type allowList —
		// non-string entries are silently dropped, "Foo" is honored, so
		// polymorphic replaces rawType and the rule reports.
		// Differentially verified against eslint-plugin-jsx-a11y v6.10.2.
		// ============================================================
		{
			Code: `<Foo as="marquee" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{123, "Foo"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},

		// ============================================================
		// rslint-specific options coverage. ESLint's JSON schema
		// (`enumArraySchema(DEFAULT_ELEMENTS)`) rejects any of these at
		// config-load time, so they never reach the rule logic upstream.
		// rslint does not perform JSON-schema validation, so these inputs
		// reach the rule's `find` and we lock in the literal-comparison
		// behavior.
		// ============================================================
		// Mixed-type array — non-strings are dropped by StringSliceOption,
		// "marquee" still matches.
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"marquee", 123}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// (Note: "all non-strings → silenced" is a Valid case, see the
		// valid block above for `[123, true]` coverage location.)
		// Empty-string entry alongside marquee — empty never matches a
		// real tag name, marquee still does.
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"", "marquee"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		// Duplicate entry — `find` short-circuits on the first hit, only
		// one report emitted. (Schema rejects upstream as
		// uniqueItems-violation.)
		{
			Code:    `<marquee />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"marquee", "marquee"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
	})
}
