// cspell:ignore busys checke describeby haspup hdiden labeledby qxqq readyonly rowin rowind thisisaveryverylongbogusattributename

package aria_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestAriaPropsExtras locks in behavior NOT covered by upstream's own test
// file. Each block is a single semantic category — see comments for the
// upstream branch or real-user scenario each case exercises.
//
// Shared helpers (`bareMessage`, `withSuggestions`) live in
// aria_props_upstream_test.go at package scope.
func TestAriaPropsExtras(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		// ============================================================
		// Case-sensitivity of the `aria-` prefix
		// ============================================================
		// Upstream uses `String.prototype.indexOf` (case-sensitive). Uppercase
		// or mixed-case prefixes early-return BEFORE the validation branch.
		{Code: `<div ARIA-HIDDEN="true" />`, Tsx: true},
		{Code: `<div Aria-Hidden="true" />`, Tsx: true},

		// ============================================================
		// Attribute name shapes that should NOT enter validation
		// ============================================================
		// JsxNamespacedName attribute — `aria:hidden` resolves via
		// reactutil.GetJsxPropName to "aria:hidden" (colon, not hyphen),
		// which fails the `aria-` prefix check.
		{Code: `<svg aria:hidden="true" />`, Tsx: true},
		// `data-*` is not aria-*.
		{Code: `<div data-aria-hidden="true" />`, Tsx: true},
		// JsxSpreadAttribute — KindJsxSpreadAttribute ≠ KindJsxAttribute,
		// so the listener never fires on spreads. Spread payload is NOT
		// inspected, so an invalid aria-* key embedded in the spread
		// object is silently allowed (upstream matches).
		{Code: `<div {...{'aria-labeledby': 'x'}} />`, Tsx: true},

		// ============================================================
		// Value-side shapes — rule never reads the value
		// ============================================================
		// Boolean form (no `={...}`) — propName still returns the bare name.
		{Code: `<div aria-hidden />`, Tsx: true},
		// TS type-assertion wrappers on value side.
		{Code: `<div aria-hidden={true as boolean} />`, Tsx: true},
		{Code: `<div aria-label={"x" as const} />`, Tsx: true},
		// Expression / optional-chain / template / arithmetic values.
		{Code: `<div aria-level={depth + 1} aria-rowindex={index * 2} />`, Tsx: true},
		{Code: `<div aria-label={user?.name} />`, Tsx: true},
		{Code: "<div aria-label={`hello ${name}`} />", Tsx: true},

		// ============================================================
		// JSX layout / formatting shapes
		// ============================================================
		// Multi-line attribute layout.
		{Code: "<div\n  aria-label=\"hi\"\n  aria-hidden\n/>", Tsx: true},
		{Code: "<div\n  aria-label=\"x\"\n  aria-hidden\n  aria-disabled\n/>", Tsx: true},
		// Self-closing without trailing space.
		{Code: `<input aria-label="x"/>`, Tsx: true},
		// Comments interleaved with attributes.
		{Code: `<div /* lead */ aria-label="x" /* mid */ aria-hidden /* trail */ />`, Tsx: true},
		// Repeated aria attribute names — JSX syntactically accepts; the
		// rule does NOT deduplicate, both names are looked up, both are
		// recognized, neither reports.
		{Code: `<div aria-hidden="false" aria-hidden="true" />`, Tsx: true},
		// Multiple spreads interleaved with named attrs.
		{Code: `<div {...a} aria-hidden {...b} aria-label="y" {...c} />`, Tsx: true},

		// ============================================================
		// Tag shapes — rule applies regardless of element type
		// ============================================================
		{Code: `<Custom aria-hidden />`, Tsx: true},
		{Code: `<Foo.Bar aria-hidden />`, Tsx: true},
		{Code: `<my-element aria-hidden />`, Tsx: true},
		{Code: `<Button aria-pressed="true" aria-label="Save" />`, Tsx: true},

		// ============================================================
		// Realistic compound elements — recommended ARIA matrix
		// ============================================================
		{Code: `<button type="button" aria-pressed="true" aria-label="Save" aria-disabled="false" />`, Tsx: true},
		{Code: `<input type="checkbox" aria-checked="mixed" aria-required="true" />`, Tsx: true},
		{Code: `<ul role="listbox" aria-multiselectable="true" aria-activedescendant="opt-1" />`, Tsx: true},

		// ============================================================
		// React / library wrappers — listener fires through every layer
		// ============================================================
		{Code: `function L({xs}) { return xs.map(x => <div aria-hidden key={x} />); }`, Tsx: true},
		{Code: `class C { render() { return <div aria-label="x" />; } }`, Tsx: true},
		{Code: `function f<T>(x: T) { return <div aria-hidden />; }`, Tsx: true},
		{Code: `<>{cond && <div aria-hidden>x</div>}</>`, Tsx: true},
		{Code: `const Btn = React.forwardRef<HTMLButtonElement, {}>((props, ref) => <button ref={ref} aria-pressed="true" />);`, Tsx: true},
		{Code: `const M = React.memo(function M() { return <div aria-label="x" />; });`, Tsx: true},
		{Code: `function F() { const v = React.useMemo(() => <div aria-hidden>x</div>, []); return v; }`, Tsx: true},
		{Code: `<Suspense fallback={<div aria-busy="true">loading</div>}><Page aria-label="x" /></Suspense>`, Tsx: true},

		// ============================================================
		// JSX positions — non-statement positions
		// ============================================================
		{Code: `const renderMap = { ok: <div aria-live="polite" /> };`, Tsx: true},
		{Code: `const arr = [<div key="1" aria-hidden />, <div key="2" aria-label="y" />];`, Tsx: true},
		{Code: `function F({ child = <div aria-hidden /> }: { child?: any }) { return child; }`, Tsx: true},
		{Code: `async function F() { return <div aria-busy="true" />; }`, Tsx: true},
		{Code: `const f = () => <div aria-hidden />;`, Tsx: true},
		{Code: `export default <div aria-label="x" />;`, Tsx: true},
		{Code: `function F() { try { return <div aria-hidden />; } catch { return null; } }`, Tsx: true},
		{Code: `<section aria-label="outer"><article aria-hidden><h1 aria-level="2">x</h1></article></section>`, Tsx: true},
		{Code: `<div>{<span aria-hidden /> as React.ReactElement}</div>`, Tsx: true},
		{Code: `<Outer renderIcon={<span aria-hidden />} aria-label="x" />`, Tsx: true},

		// ============================================================
		// Settings interaction — rule does NOT consult settings
		// ============================================================
		// `components` map / `polymorphicPropName` are honored by other
		// jsx-a11y rules but NOT by aria-props (which looks only at the
		// attribute name). Providing them must not change the verdict.
		{
			Code: `<Foo aria-hidden />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components":          map[string]interface{}{"Foo": "div"},
					"polymorphicPropName": "as",
				},
			},
		},
		// Empty settings — must not panic.
		{
			Code:     `<div aria-hidden />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{}},
		},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		// ============================================================
		// Case-sensitivity quirk: lowercase prefix + uppercase suffix
		// ============================================================
		// `aria.has` is case-sensitive (map lookup), so `aria-LABEL` misses
		// and reports. `getSuggestion` upper-cases both sides before
		// measuring distance, so the canonical lowercase form surfaces as
		// the suggestion. Locks the upstream quirk that case-variants of
		// canonical names ARE reported (and their canonical form suggested).
		{
			Code: `<div aria-LABEL="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				// Distance 0 vs `aria-label`; distance 2 vs `aria-level`
				// (subs A↔E, B↔V) — both within threshold, limit = 2.
				Message: withSuggestions("aria-LABEL", "aria-label", "aria-level"),
			}},
		},
		{
			Code: `<div aria-Hidden />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-Hidden", "aria-hidden"),
			}},
		},

		// ============================================================
		// Diagnostic position (single-line and multi-line)
		// ============================================================
		// Single-line: span = whole JsxAttribute incl. `={...}`.
		{
			Code: `<div aria-="foobar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-"),
				Line:      1, Column: 6, EndLine: 1, EndColumn: 20,
			}},
		},
		{
			Code: `<div aria-labeledby="foobar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-labeledby", "aria-labelledby"),
				Line:      1, Column: 6, EndLine: 1, EndColumn: 29,
			}},
		},
		{
			Code: `<div aria-fake="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-fake"),
				Line:      1, Column: 6, EndLine: 1, EndColumn: 21,
			}},
		},
		// Multi-line: column is on the attribute's start line, not the
		// element's opening `<`.
		{
			Code: "<div\n  aria-typo=\"x\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
				Line:      2, Column: 3, EndLine: 2, EndColumn: 16,
			}},
		},
		// Child JSX element — column shifts to inner element.
		{
			Code: `<div><span aria-bogus="x" /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-bogus"),
				Line:      1, Column: 12, EndLine: 1, EndColumn: 26,
			}},
		},

		// ============================================================
		// Attribute-value shape independence (rule never reads value)
		// ============================================================
		// Boolean form of invalid aria-*.
		{
			Code: `<div aria-foobar />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-foobar"),
			}},
		},
		// TS type-assertion on value of invalid attribute.
		{
			Code: `<div aria-typo={"x" as const} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// Self-closing without trailing space.
		{
			Code: `<div aria-foo="x"/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-foo"),
			}},
		},

		// ============================================================
		// Multi-attribute combinations on the same element
		// ============================================================
		// Two close typos — each reports its own suggestion.
		{
			Code: `<div aria-labeledby="x" aria-describeby="y" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: withSuggestions("aria-labeledby", "aria-labelledby")},
				{MessageId: "invalidAriaProp", Message: withSuggestions("aria-describeby", "aria-describedby")},
			},
		},
		// Valid + invalid interleaved — only invalid reports.
		{
			Code: `<div aria-hidden aria-foo="x" aria-label="y" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-foo"),
			}},
		},
		// Stress: 5 invalid + 2 valid attrs — order matches source.
		{
			Code: `<div aria-foo="1" aria-hidden aria-bar="2" aria-label="x" aria-baz="3" aria-qux="4" aria-quux="5" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-foo")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-bar")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-baz")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-qux")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-quux")},
			},
		},
		// Same invalid name written twice on one element — JSX syntactically
		// accepts duplicates (React warns at runtime). The listener fires
		// per JsxAttribute, so the diagnostic count equals the source-level
		// occurrence count. Locks no-dedup behavior, matching upstream.
		{
			Code: `<div aria-typo="a" aria-typo="b" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-typo")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-typo")},
			},
		},

		// ============================================================
		// Spread interleaving — spread never inspected
		// ============================================================
		// Mixed spread + invalid named — spread is skipped, named reports.
		{
			Code: `<div {...a} aria-typo="x" {...b} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// Spread carrying a recognized aria name BUT the named attr is
		// invalid — only the named reports (spread payload not inspected).
		{
			Code: `<div {...{'aria-hidden': true}} aria-typo />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},

		// ============================================================
		// Real-user typos — close enough for a suggestion
		// ============================================================
		// Distance 1 / 2 typos producing real suggestions. These are the
		// concrete cases users actually file as "why doesn't my aria
		// attribute work?" — locking them ensures the suggestion path
		// stays useful, not just non-zero.
		{
			Code: `<div aria-describeby="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-describeby", "aria-describedby"),
			}},
		},
		{
			Code: `<div aria-readyonly="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-readyonly", "aria-readonly"),
			}},
		},
		{
			Code: `<div aria-busys="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-busys", "aria-busy"),
			}},
		},
		{
			Code: `<div aria-checke="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-checke", "aria-checked"),
			}},
		},
		{
			Code: `<div aria-haspup="menu" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-haspup", "aria-haspopup"),
			}},
		},

		// ============================================================
		// Real-user confusion: HTML / React attrs spelled with aria-*
		// ============================================================
		// `aria-tabindex` — tabIndex is a native HTML attribute, not ARIA.
		{
			Code: `<div aria-tabindex="0" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-tabindex"),
			}},
		},
		// `aria-onclick` — onClick is a React synthetic event.
		{
			Code: `<div aria-onclick={handler} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-onclick"),
			}},
		},
		// `aria-class` — user confused with className / class.
		{
			Code: `<div aria-class="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-class"),
			}},
		},
		// `aria-id` / `aria-style` — user confused with HTML attributes.
		{
			Code: `<div aria-id="x" aria-style="color: red" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-id")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-style")},
			},
		},

		// ============================================================
		// AST shape edge cases — listener must descend through wrappers
		// ============================================================
		// JSX expression child containing JSX with invalid attr.
		{
			Code: `<div>{<span aria-typo="x" />}</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// JSX as a prop value (render-prop pattern).
		{
			Code: `<Outer icon={<span aria-typo="x" />} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// JSX as array literal element.
		{
			Code: `const xs = [<div key="a" aria-typo="x" />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// JSX as object-literal property value.
		{
			Code: `const m = { a: <div aria-typo="x" /> };`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// JSX as arrow body (no braces).
		{
			Code: `const f = () => <div aria-typo="x" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// JSX inside library wrappers.
		{
			Code: `const X = React.forwardRef((p, r) => <div ref={r} aria-typo="x" />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `function F() { return React.useMemo(() => <div aria-typo="x" />, []); }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `<Suspense fallback={<div aria-typo="x" />}><Page /></Suspense>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// Capitalized component / map / ternary / call argument /
		// fragment — listener fires regardless of surrounding shape.
		{
			Code: `<MyComponent aria-typo="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `function L({xs}) { return xs.map(x => <div aria-typo="y" key={x} />); }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `function F({c}: {c: boolean}) { return c ? <div aria-typo /> : null; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `render(<div aria-typo="x" />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		{
			Code: `<>{cond && <div aria-typo="y">x</div>}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},

		// ============================================================
		// Multi-element nesting — listener fires at every depth
		// ============================================================
		// Single-deep nested invalid.
		{
			Code: `<section><article><header><h1 aria-typo="x" /></header></article></section>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},
		// Outer + inner — no bleed across boundaries.
		{
			Code: `<div aria-zzzz="o"><span aria-qxqq="i" /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-zzzz")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-qxqq")},
			},
		},
		// Multi-level depth with invalid at multiple depths.
		{
			Code: `<main><section><article><h2 aria-t1="a"><span aria-t2="b" /></h2></article></section></main>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-t1")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-t2")},
			},
		},
		// Outer JSX + inner JSX via prop value.
		{
			Code: `<div aria-outer="x" content={<span aria-inner="y" />}>k</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-outer")},
				{MessageId: "invalidAriaProp", Message: bareMessage("aria-inner")},
			},
		},
		// Component receives invalid aria-* alongside otherwise valid
		// custom props.
		{
			Code: `<MyButton onClick={fn} disabled aria-typo title="x">label</MyButton>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},

		// ============================================================
		// Settings interaction — rule does NOT consult settings
		// ============================================================
		// `components` map cannot rescue an invalid aria-* — the rule
		// doesn't resolve element types.
		{
			Code: `<Foo aria-typo="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": "div"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-typo"),
			}},
		},

		// ============================================================
		// Algorithm boundary conditions
		// ============================================================
		// Distance 2 (upper edge of THRESHOLD) — suggestion fires.
		{
			Code: `<div aria-rowind="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-rowind", "aria-rowindex"),
			}},
		},
		// Distance 3 (just past THRESHOLD) — no suggestion.
		{
			Code: `<div aria-rowin="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-rowin"),
			}},
		},
		// Adjacent transposition — OSA distance 1 (plain Levenshtein
		// would give 2). Locks the transposition branch of OSA.
		{
			Code: `<div aria-hdiden />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-hdiden", "aria-hidden"),
			}},
		},
		// Very long name — length delta alone exceeds threshold.
		{
			Code: `<div aria-thisisaveryverylongbogusattributename="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-thisisaveryverylongbogusattributename"),
			}},
		},
		// Very short name `aria-a` — canonical entries are all ≥ 9 chars.
		{
			Code: `<div aria-a="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-a"),
			}},
		},
		// Digits in the attribute name.
		{
			Code: `<div aria-foo123="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-foo123"),
			}},
		},
		// Trailing hyphen — distance 1 from `aria-label`.
		{
			Code: `<div aria-label- />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-label-", "aria-label"),
			}},
		},
		// Double hyphen after prefix — distance 1 from `aria-label`.
		{
			Code: `<div aria--label />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria--label", "aria-label"),
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaPropsRule, validCases, invalidCases)
}
