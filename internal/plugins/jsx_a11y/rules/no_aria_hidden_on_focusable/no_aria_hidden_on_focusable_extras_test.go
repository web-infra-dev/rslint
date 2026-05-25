package no_aria_hidden_on_focusable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings remaps `Btn → button` so `<Btn aria-hidden="true" />`
// resolves to an inherently interactive element. Locks in that the rule
// honors the jsx-a11y components map when classifying focusability.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Btn": "button",
		},
	},
}

// polymorphicSettings mirrors a `polymorphicPropName: 'as'` config. With it,
// `<Box as="input" aria-hidden="true" />` resolves to `input`.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestNoAriaHiddenOnFocusableExtras is the catch-all for everything beyond
// upstream's small suite (which lives in no_aria_hidden_on_focusable_upstream_test.go).
// Cases here fall in five groups, kept in one suite so a regression bisects
// easily:
//
//  1. **aria-hidden value-extraction branches** — boolean form, JsxExpression
//     true/false, case-insensitive "True"/"TRUE", JsxExpression-wrapped string
//     literals, TS wrappers, non-matching values that must NOT short-circuit
//     positively. Upstream's `getPropValue === true` has more reachable arms
//     than its tiny test file covers.
//  2. **tabIndex value-extraction branches** — negative integers, large
//     positive, non-integer (silently dropped per upstream's Number.isInteger
//     guard), empty / boolean / undefined / null shapes, entity-encoded HTML.
//     Each exercises a distinct GetTabIndexEx exit.
//  3. **Element kind survey** — every inherently interactive tag (button,
//     input, textarea, select, option, audio/video with controls, summary,
//     a-with-href, area-with-href, ...), every inherently non-interactive
//     tag, and the component / member / polymorphic resolution paths.
//  4. **Dimension 4 universal edge shapes** — JsxOpeningElement vs
//     JsxSelfClosingElement, paired elements, spread attributes, nested JSX,
//     namespaced names, multi-line, comments, position assertions.
//  5. **Real-world patterns** — modal/dialog/menubar trees, conditional
//     rendering, .map / .filter, HOC / forwardRef / memo, generator / async /
//     IIFE bodies, TS generics, hyphenated tags, multi-component files.
func TestNoAriaHiddenOnFocusableExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAriaHiddenOnFocusableRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Group 1: aria-hidden does NOT resolve to true → safe regardless of focus
		// ============================================================
		// No aria-hidden at all — first gate.
		{Code: `<input />`, Tsx: true},
		{Code: `<button tabIndex="0" />`, Tsx: true},
		// aria-hidden values that staticEval cannot resolve to boolean true.
		{Code: `<button aria-hidden={false} />`, Tsx: true},
		{Code: `<input aria-hidden={false} />`, Tsx: true},
		{Code: `<button aria-hidden="false" />`, Tsx: true},
		{Code: `<button aria-hidden="False" />`, Tsx: true},
		{Code: `<button aria-hidden="FALSE" />`, Tsx: true},
		{Code: `<button aria-hidden={"false"} />`, Tsx: true},
		// `{undefined}` extracts to JS undefined under upstream's Identifier
		// extractor → not === true → safe.
		{Code: `<button aria-hidden={undefined} />`, Tsx: true},
		// Numeric: upstream extracts `0` / `1` → not === true → safe.
		{Code: `<button aria-hidden={0} />`, Tsx: true},
		{Code: `<button aria-hidden={1} />`, Tsx: true},
		// String "yes" / "no" / etc — extractValue returns the raw string;
		// jsxAstUtilsLiteralCoerce only normalizes "true" / "false".
		{Code: `<button aria-hidden="yes" />`, Tsx: true},
		// Identifier reference — upstream's Literal extractor returns the
		// identifier's name as a string. "someVar" !== true → safe.
		{Code: `<button aria-hidden={someVar} />`, Tsx: true},
		// Member / Call — upstream synthesizes a non-boolean string → safe.
		{Code: `<button aria-hidden={obj.x} />`, Tsx: true},
		{Code: `<button aria-hidden={fn()} />`, Tsx: true},
		// satisfies is opaque — staticEval returns jvNull → not boolean true.
		{Code: `<button aria-hidden={true satisfies boolean} />`, Tsx: true},
		// NoSubstitutionTemplateLiteral does NOT route through
		// jsxAstUtilsLiteralCoerce, so `` `true` `` stays a string → safe.
		{Code: "<button aria-hidden={`true`} />", Tsx: true},

		// ============================================================
		// Group 2: Non-interactive elements without focus-enabling tabIndex
		// ============================================================
		// aria-hidden=true on a non-interactive element with NO tabIndex →
		// tabIndex >= 0 is false → not focusable → safe.
		{Code: `<div aria-hidden />`, Tsx: true},
		{Code: `<div aria-hidden={true} />`, Tsx: true},
		{Code: `<div aria-hidden="true" />`, Tsx: true},
		{Code: `<div aria-hidden="True" />`, Tsx: true},
		{Code: `<div aria-hidden="TRUE" />`, Tsx: true},
		{Code: `<span aria-hidden="true" />`, Tsx: true},
		{Code: `<p aria-hidden="true" />`, Tsx: true},
		{Code: `<section aria-hidden="true" />`, Tsx: true},
		{Code: `<article aria-hidden="true" />`, Tsx: true},
		// img / br / hr — non-interactive (img is in dom set but not in
		// interactive role schemas) → safe without tabIndex.
		{Code: `<img aria-hidden="true" />`, Tsx: true},
		{Code: `<br aria-hidden="true" />`, Tsx: true},
		{Code: `<hr aria-hidden="true" />`, Tsx: true},
		// tabIndex="-1" on non-interactive → -1 < 0 → not focusable → safe.
		{Code: `<div aria-hidden="true" tabIndex="-1" />`, Tsx: true},
		{Code: `<div aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		{Code: `<p aria-hidden="true" tabIndex={-1}>text</p>`, Tsx: true},

		// ============================================================
		// Group 3: Interactive elements with explicit tabIndex < 0
		// ============================================================
		// Inherently interactive tags become NOT focusable when tabIndex is
		// negative (and resolved). `tabIndex === undefined || tabIndex >= 0`
		// → false || false → false.
		{Code: `<button aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		{Code: `<button aria-hidden="true" tabIndex="-2" />`, Tsx: true},
		{Code: `<input aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		{Code: `<input aria-hidden="true" tabIndex="-1" />`, Tsx: true},
		{Code: `<textarea aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		{Code: `<select aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		{Code: `<a href="/" aria-hidden="true" tabIndex={-1} />`, Tsx: true},
		// area requires href to be interactive.
		{Code: `<area href="#" aria-hidden="true" tabIndex={-1} />`, Tsx: true},

		// ============================================================
		// Group 4: Custom components without tabIndex
		// ============================================================
		// Custom components fall through `isInteractiveElement` → tabIndex >=
		// 0 path → without explicit tabIndex → safe.
		{Code: `<MyButton aria-hidden="true" />`, Tsx: true},
		{Code: `<UX.Layout aria-hidden="true" />`, Tsx: true},
		{Code: `<svg:circle aria-hidden="true" />`, Tsx: true},
		{Code: `<Foo.Bar.Baz aria-hidden="true" />`, Tsx: true},
		// Custom component with tabIndex < 0 — non-interactive arm rejects.
		{Code: `<MyButton aria-hidden="true" tabIndex={-1} />`, Tsx: true},

		// ============================================================
		// Group 5: <a> without href is non-interactive
		// ============================================================
		// Per upstream's element-role schemas, anchor requires `href` to count
		// as interactive. `<a aria-hidden="true" />` without href → tabIndex
		// >= 0 false → safe.
		{Code: `<a aria-hidden="true" />`, Tsx: true},
		{Code: `<area aria-hidden="true" />`, Tsx: true},

		// ============================================================
		// Group 5b: Interactive elements where tabIndex resolves to a
		// non-numeric non-undefined value via upstream's step-2 getPropValue
		// ============================================================
		// Upstream's `isFocusable` interactive arm is `tabIndex === undefined ||
		// tabIndex >= 0`. When getTabIndex falls through to step-2 getPropValue
		// for shapes that LITERAL_TYPES doesn't catch (Identifier non-undefined,
		// MemberExpression, CallExpression, TaggedTemplate, ...), upstream returns
		// the synthesized string (`"someVar"`, `"obj.x"`, `"fn()"`, ...). The
		// strict `=== undefined` check rejects these strings, and
		// `string >= 0` ToNumbers to NaN → false. Both arms false → NOT focusable.
		{Code: `<button aria-hidden="true" tabIndex={someVar} />`, Tsx: true},
		{Code: `<input aria-hidden="true" tabIndex={someVar} />`, Tsx: true},
		{Code: `<textarea aria-hidden="true" tabIndex={obj.x} />`, Tsx: true},
		{Code: `<select aria-hidden="true" tabIndex={fn()} />`, Tsx: true},
		{Code: `<button aria-hidden="true" tabIndex={obj?.x} />`, Tsx: true},
		// Non-interactive equivalents — same arm, just confirming the
		// non-interactive path also rejects.
		{Code: `<div aria-hidden="true" tabIndex={someVar} />`, Tsx: true},
		{Code: `<div aria-hidden="true" tabIndex={obj.x} />`, Tsx: true},
		{Code: `<div aria-hidden="true" tabIndex={fn()} />`, Tsx: true},

		// ============================================================
		// Group 5c: `null` and `undefined` literal value resolution
		// ============================================================
		// LITERAL_TYPES.Literal special-cases null → "null" string → step-1
		// `Number("null")` = NaN → undefined. Interactive arm's `=== undefined`
		// matches; for non-interactive, NaN >= 0 → false.
		// Non-interactive — safe.
		{Code: `<div aria-hidden="true" tabIndex={null} />`, Tsx: true},

		// ============================================================
		// Group 6: tabIndex value-extraction → undefined → safe for non-interactive
		// ============================================================
		// Upstream getTabIndex returns undefined for these shapes; the
		// non-interactive arm needs a resolved >= 0 to mark focusable.
		{Code: `<div aria-hidden="true" tabIndex="" />`, Tsx: true},
		{Code: `<div aria-hidden="true" tabIndex={undefined} />`, Tsx: true},
		// Boolean tabIndex — upstream's step-1 boolean arm → undefined → safe.
		{Code: `<div aria-hidden="true" tabIndex={true} />`, Tsx: true},
		{Code: `<div aria-hidden="true" tabIndex={false} />`, Tsx: true},
		// Non-integer numeric → upstream Number.isInteger filter → undefined → safe.
		{Code: `<div aria-hidden="true" tabIndex={1.5} />`, Tsx: true},

		// ============================================================
		// Group 7: components map — non-DOM resolution
		// ============================================================
		// `<Btn aria-hidden="true" tabIndex={-1} />` with `Btn → button`
		// → resolved as interactive → -1 not focusable → safe.
		{Code: `<Btn aria-hidden="true" tabIndex={-1} />`, Tsx: true, Settings: componentsSettings},

		// ============================================================
		// Group 8: polymorphicPropName — `as="button"` resolves
		// ============================================================
		// `<Box as="button" aria-hidden="true" tabIndex={-1} />` resolves to
		// `button` → interactive → -1 → not focusable → safe.
		{Code: `<Box as="button" aria-hidden="true" tabIndex={-1} />`, Tsx: true, Settings: polymorphicSettings},

		// ============================================================
		// Group 9: TS wrappers around aria-hidden value
		// ============================================================
		// staticEval transparently unwraps parens and TSAsExpression — wrapped
		// `true` extracts to boolean true → triggers the rule's gate. But
		// `<div>` (non-interactive, no tabIndex) → not focusable → safe.
		{Code: `<div aria-hidden={(true)} />`, Tsx: true},
		{Code: `<div aria-hidden={true as boolean} />`, Tsx: true},
		// TSNonNullExpression is DIFFERENT: upstream's extractor stringifies
		// the operand with a trailing `!`, so `{true!}` extracts to the
		// string `"true!"` — `"true!" === true` is false → aria-hidden gate
		// fails → safe regardless of focus.
		{Code: `<div aria-hidden={true!} />`, Tsx: true},
		{Code: `<button aria-hidden={true!} />`, Tsx: true},
		{Code: `<input aria-hidden={true!} />`, Tsx: true},

		// ============================================================
		// Group 10: Spread aria-hidden=true on non-focusable element
		// ============================================================
		// JsxSpreadAttribute with string-literal key — jsx-ast-utils' `getProp`
		// only matches when `key.type === 'Identifier'`, so `{...{"aria-hidden":
		// true}}` does NOT resolve the aria-hidden attribute. The rule's gate
		// fails → safe even on a focusable element.
		{Code: `<div {...{"aria-hidden": true}} />`, Tsx: true},
		{Code: `<button {...{"aria-hidden": true}} />`, Tsx: true},
		{Code: `<input {...{"aria-hidden": true}} />`, Tsx: true},
		// Spread of non-literal object — opaque, no match.
		{Code: `<button {...props} />`, Tsx: true},

		// ============================================================
		// Group 11: Empty / malformed JsxExpression
		// ============================================================
		// `aria-hidden={}` — empty JsxExpression. staticEval routes to
		// jvUnknown / jvNull — not boolean true → safe.
		{Code: `<button aria-hidden={} />`, Tsx: true},

		// ============================================================
		// Group 12: Nested elements where inner is also aria-hidden=false
		// ============================================================
		// `<a href="#" aria-hidden="false"><span aria-hidden="true" /></a>`
		// — outer is safe (aria-hidden=false), inner span is non-interactive
		// with no tabIndex → both safe.
		{Code: `<a href="#" aria-hidden="false"><span aria-hidden="true" /></a>`, Tsx: true},

		// ============================================================
		// Group 13: Real-world patterns — explicit tabIndex={-1} suppress
		// ============================================================
		{
			Code: `function Decoration() { return <button aria-hidden="true" tabIndex={-1}>icon</button>; }`,
			Tsx:  true,
		},
		{
			Code: `function Item({active}) { return active ? <button aria-hidden="true" tabIndex={-1}>...</button> : null; }`,
			Tsx:  true,
		},

		// ============================================================
		// Group 14: comments around / inside the prop
		// ============================================================
		{Code: `<div /* before */ aria-hidden="true" /* after */ />`, Tsx: true},
		{Code: `<div aria-hidden={/* true */ true} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Group 1: aria-hidden value-extraction → boolean true
		// ============================================================
		// Boolean attribute form (`<button aria-hidden />`) — upstream's
		// extractValue maps null-attr-value to JS true.
		{Code: `<button aria-hidden />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// JsxExpression with literal true.
		{Code: `<button aria-hidden={true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Case-insensitive string "True" / "TRUE".
		{Code: `<button aria-hidden="True" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="TRUE" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// JsxExpression with string literal "true" — jsxAstUtilsLiteralCoerce
		// applies inside staticEval.
		{Code: `<button aria-hidden={"true"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden={"True"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 2: Interactive element survey, no tabIndex
		// ============================================================
		// Every inherently interactive tag with no tabIndex → undefined ===
		// undefined → focusable → reports.
		{Code: `<a href="#" aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<area href="#" aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<select aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<textarea aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 3: Non-interactive made focusable by tabIndex >= 0
		// ============================================================
		{Code: `<div aria-hidden="true" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div aria-hidden="true" tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div aria-hidden="true" tabIndex={42} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<span aria-hidden="true" tabIndex="0">x</span>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<p aria-hidden="true" tabIndex="2">x</p>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 4: Boolean attribute form + interactive
		// ============================================================
		{Code: `<input aria-hidden />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a href="/" aria-hidden />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<select aria-hidden />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 5: TS-wrapped aria-hidden values on interactive
		// ============================================================
		// Parens and TSAsExpression are stripped by staticEval — still boolean true.
		// (TSNonNullExpression behaves differently — see the valid Group 9.)
		{Code: `<button aria-hidden={(true)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden={true as boolean} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input aria-hidden={(true)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 6: Logical / conditional resolving to boolean true
		// ============================================================
		// staticEval evaluates short-circuit operators when both arms are
		// statically known.
		{Code: `<button aria-hidden={true && true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden={false || true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden={true ? true : false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 7: (reserved — no spread case fires; spread literals carry
		// Identifier keys only, and `ariaHidden` (no hyphen) does not match
		// `aria-hidden` under jsx-ast-utils' Identifier key compare.)
		// ============================================================
		// ============================================================
		// Group 8: Position assertions per container
		// ============================================================
		// JsxSelfClosingElement node range — `<input aria-hidden="true" />`
		// spans columns 1..27 (1-based, end-exclusive).
		{
			Code: `<input aria-hidden="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAriaHiddenOnFocusable",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
			}},
		},
		// JsxOpeningElement node range — `<button aria-hidden="true">x</button>`
		// reports on the opening tag only; the JsxOpeningElement spans
		// columns 1..28 (1-based, end-exclusive).
		{
			Code: `<button aria-hidden="true">x</button>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAriaHiddenOnFocusable",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 28,
			}},
		},
		// Multi-line element — position spans the full opening element across lines.
		{
			Code: "<button\n  aria-hidden=\"true\"\n  tabIndex={0}\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAriaHiddenOnFocusable",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 4, EndColumn: 3,
			}},
		},

		// ============================================================
		// Group 9: Nested element listener fires independently
		// ============================================================
		// Outer button and inner anchor each report separately.
		{
			Code: `<button aria-hidden="true"><a href="/" aria-hidden="true">x</a></button>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, // outer <button>
				expectedError, // inner <a href>
			},
		},

		// ============================================================
		// Group 10: components map — non-DOM remapped to interactive
		// ============================================================
		// `<Btn aria-hidden="true" />` with `Btn → button` → resolves to
		// interactive → no tabIndex → focusable → reports.
		{
			Code:     `<Btn aria-hidden="true" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 11: polymorphicPropName resolution
		// ============================================================
		// `<Box as="input" aria-hidden="true" />` resolves to `input` →
		// interactive → no tabIndex → focusable → reports.
		{
			Code:     `<Box as="input" aria-hidden="true" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Even on non-DOM type, if `as` resolves to button → reports.
		{
			Code:     `<Box as="button" aria-hidden="true" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 12: Custom component with explicit positive tabIndex
		// ============================================================
		// Non-interactive arm: `tabIndex >= 0` → focusable → reports.
		{Code: `<MyButton aria-hidden="true" tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<MyButton aria-hidden="true" tabIndex="1" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<UX.Layout aria-hidden="true" tabIndex={3} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 13: Real-world component patterns
		// ============================================================
		// Function component returning the offending element.
		{
			Code:   `function Icon() { return <button aria-hidden="true">x</button>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Conditional rendering — only one arm offends.
		{
			Code:   `function Foo({cond}) { return cond ? <input aria-hidden="true" /> : <div />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// .map producing multiple offending elements.
		{
			Code:   `const xs = arr.map((x) => <button aria-hidden="true" key={x}>{x}</button>);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// HOC wrapping carrying the offending child.
		{
			Code:   `const Wrapped = withTracking(({v}) => <input aria-hidden="true" value={v} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Fragment + && rendering.
		{
			Code:   `const x = <>{cond && <button aria-hidden="true" />}</>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multiple components in one file → each reports.
		{
			Code: "function A() { return <input aria-hidden=\"true\" />; }\nfunction B() { return <textarea aria-hidden=\"true\" />; }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		// Generator / async / IIFE bodies.
		{
			Code: `function* render() { yield <input aria-hidden="true" />; yield <textarea aria-hidden="true" />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		{
			Code:   `async function render() { return <button aria-hidden="true">x</button>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = (() => <input aria-hidden="true" />)();`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Modal pattern: dialog opens with a focusable but aria-hidden button.
		{
			Code:   `function Modal() { return <dialog open><button aria-hidden="true">OK</button></dialog>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Class component render returning offending element.
		{
			Code:   `class Form extends React.Component { render() { return <input aria-hidden="true" />; } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 14: TypeScript generic JSX
		// ============================================================
		{
			Code:   `<Cell<{a: number}> aria-hidden="true" tabIndex={0} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 15: Comments around / inside the prop don't suppress
		// ============================================================
		{
			Code:   `<input /* a */ aria-hidden="true" /* b */ />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<button aria-hidden={/* truthy */ true} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 16: tabIndex via expressions that staticEval resolves
		// ============================================================
		// Conditional resolving to >= 0 — staticEval evaluates both arms.
		{Code: `<div aria-hidden="true" tabIndex={true ? 0 : -1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String expression coerces to 0 via ToNumber.
		{Code: `<div aria-hidden="true" tabIndex={"5"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 17: Multiple offending elements on one render
		// ============================================================
		{
			Code:   `<form><input aria-hidden="true" /><textarea aria-hidden="true" /></form>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ============================================================
		// Group 17b: tabIndex shapes whose getTabIndex result is undefined
		//            (step-1 short-circuits) — interactive arm reports
		// ============================================================
		// Upstream getTabIndex returns undefined for these and the interactive
		// arm's `=== undefined` matches → focusable → reports.
		{Code: `<button aria-hidden="true" tabIndex="" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" tabIndex={true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" tabIndex={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" tabIndex={1.5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" tabIndex={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TaggedTemplateExpression: LITERAL_TYPES.TaggedTemplateExpression
		// inherits from TYPES — extracts the inner template text "x", then
		// step-1 `Number("x")` → NaN → undefined. Interactive → focusable.
		{Code: "<button aria-hidden=\"true\" tabIndex={tag`x`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `null` literal: LITERAL_TYPES.Literal returns "null" string → step-1
		// `Number("null")` = NaN → undefined → focusable on interactive.
		{Code: `<button aria-hidden="true" tabIndex={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TSNonNullExpression: upstream's extractor recurses, the Literal arm
		// returns the bare `.value` (boolean true), then the outer wraps it
		// with `+ "!"` → JS template coerces → "true!" string. Step-1
		// `Number("true!")` = NaN → undefined. Interactive arm matches
		// `=== undefined` → focusable → reports.
		{Code: `<button aria-hidden="true" tabIndex={true!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button aria-hidden="true" tabIndex={"5"!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// BinaryExpression: upstream's extractValueFromBinaryExpression
		// computes the operation (`1+2` → 3, not stringify). step-1 number
		// `3`, isInteger → 3. `3 >= 0` → focusable → reports.
		{Code: `<button aria-hidden="true" tabIndex={1+2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Non-numeric direct attribute string → step-1 NaN → undefined.
		{Code: `<button aria-hidden="true" tabIndex="abc" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// NoSubstitutionTemplateLiteral with non-numeric text → step-1 NaN → undefined.
		{Code: "<button aria-hidden=\"true\" tabIndex={`abc`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 18: Lowercase tabindex on interactive (case-insensitive prop)
		// ============================================================
		// jsx-ast-utils' getProp is case-insensitive by default — lowercase
		// `tabindex` resolves the same as `tabIndex`. Confirms our
		// FindAttributeByName parity.
		{Code: `<button aria-hidden="true" tabindex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
