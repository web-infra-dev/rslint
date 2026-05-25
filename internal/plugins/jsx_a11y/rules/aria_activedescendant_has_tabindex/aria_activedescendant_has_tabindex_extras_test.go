package aria_activedescendant_has_tabindex

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings mirrors anchor_is_valid's `polymorphicSettings` —
// `polymorphicPropName: 'as'` makes `getElementType` rewrite the tag name
// from the literal `as=` value before the dom.has() / interactive checks
// run, so `<Box as="div" />` resolves to `div`.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestAriaActivedescendantHasTabindexExtras locks in behavior on inputs the
// upstream test suite does NOT exercise. Each block is annotated with the
// upstream branch / jsx-ast-utils helper it covers so future audits can
// trace each case to a specific arm.
func TestAriaActivedescendantHasTabindexExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaActivedescendantHasTabindexRule, []rule_tester.ValidTestCase{
		// ---- Case-insensitive prop match (jsx-ast-utils `getProp` ignoreCase: true) ----
		// `Aria-ActiveDescendant` matches the same as `aria-activedescendant`.
		// Tabbable (tabIndex={0}) so the rule doesn't fire — locks in only
		// the case-folding path; case-insensitive INVALID is in the invalid
		// section below.
		{Code: `<div Aria-ActiveDescendant={x} tabIndex={0} />;`, Tsx: true},

		// ---- gate-3: interactive elements skip when tabIndex is undefined ----
		// `<a href>` is interactive (interactiveElementRoleSchemas requires
		// `href` for `a`); no tabIndex → gate-3 takes the skip arm. Upstream
		// covers <input> but not <a href>; this case locks the schema entry.
		{Code: `<a href="#" aria-activedescendant={x} />;`, Tsx: true},
		// `<button>` is unconditionally interactive — interactiveElementRoleSchemas
		// has a no-attribute entry for it. Locks the bare-button arm.
		{Code: `<button aria-activedescendant={x} />;`, Tsx: true},
		// `<select size={1}>` is interactive via the size-attribute schema
		// entry. Locks an attribute-conditional interactive arm.
		{Code: `<select size={1} aria-activedescendant={x} />;`, Tsx: true},

		// ---- gate-2: non-DOM tag forms ----
		// SVG-namespaced tag — tsgo exposes the colon in the tag name, which
		// is not in aria-query's dom map → gate-2 skip.
		{Code: `<svg:path aria-activedescendant={x} />;`, Tsx: true},
		// Dotted-namespace tag — `Foo.Bar` isn't in dom map → gate-2 skip.
		// Locks that GetElementType doesn't accidentally split on the dot.
		{Code: `<Foo.Bar aria-activedescendant={x} />;`, Tsx: true},

		// ---- gate-2: components map / polymorphic resolution ----
		// `<Box as="span" />` → resolved tag is "span", which IS in dom map.
		// Span is non-interactive but tabIndex={0} → gate-4 skips.
		{Code: `<Box as="span" aria-activedescendant={x} tabIndex={0} />;`, Tsx: true, Settings: polymorphicSettings},

		// ---- gate-4: tabIndex exactly at boundary ----
		// `tabIndex={-1}` is the minimum valid value (programmatic-only focus).
		{Code: `<div aria-activedescendant={x} tabIndex={-1} />;`, Tsx: true},
		// String "1" — string-numeric coerces via GetTabIndex's
		// StringToNumber path; gate-4 1 >= -1 → skip.
		{Code: `<div aria-activedescendant={x} tabIndex="1" />;`, Tsx: true},

		// ---- Nested JSX — only inner element matches when outer doesn't ----
		// Outer span has no aria-activedescendant; inner div has activedescendant
		// + tabIndex={0} → both arms skip independently.
		{Code: `<span><div aria-activedescendant={x} tabIndex={0} /></span>;`, Tsx: true},

		// ---- Spread without literal `aria-activedescendant` ----
		// Spread payload is opaque (not a literal ObjectExpression); upstream's
		// `getProp` returns undefined → gate-1 skip. Same as `<div />`.
		{Code: `<div {...props} />;`, Tsx: true},

		// ---- Literal-spread tabIndex tracking ----
		// jsx-ast-utils' `getProp` walks literal-object spreads when the key
		// is an Identifier (`tabIndex`). FindAttributeByName mirrors this; the
		// resolved PropertyAssignment routes through GetTabIndex's literal
		// path identically to a direct JsxAttribute. Locks the spread→prop
		// resolution chain end-to-end.
		{Code: `<div aria-activedescendant={x} {...{tabIndex: 0}} />;`, Tsx: true},

		// ---- TS-wrapper unwrapping in tabIndex value (paren / as / !) ----
		// GetTabIndex routes through staticEval which strips parens / `as` /
		// non-null via OEKParentheses|OEKTypeAssertions|OEKNonNullAssertions.
		// Each wrapper variant must resolve to 0 (>= -1 → skip). `satisfies`
		// is handled separately (see Differences from ESLint in the .md and
		// the matching invalid lock-ins below).
		{Code: `<div aria-activedescendant={x} tabIndex={(0)} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={(0 as number)} />;`, Tsx: true},
		// Note: TSNonNullExpression `0!` / `(0)!` is INVALID per upstream —
		// jsx-ast-utils' TSNonNullExpression extractor stringifies to "0!"
		// → Number("0!") = NaN → undefined → aria gate-4 fails → REPORT.
		// Lock-in lives in the invalid section.

		// ---- NoSubstitutionTemplateLiteral tabIndex ----
		// `` `-1` `` reaches the template-literal extractor (NOT the
		// "true"/"false" → bool coercion), yielding the string "-1" → -1
		// via GetTabIndex's StringToNumber. -1 >= -1 → skip.
		{Code: "<div aria-activedescendant={x} tabIndex={`-1`} />;", Tsx: true},
		{Code: "<div aria-activedescendant={x} tabIndex={`0`} />;", Tsx: true},

		// ---- Opaque expression types resolve to upstream `null` ----
		// jsx-ast-utils' getPropValue returns `null` for any ESTree expression
		// type its TYPES table doesn't recognize (TSSatisfiesExpression,
		// AwaitExpression, YieldExpression, ...). Upstream's downstream
		// `tabIndex >= -1` then ToNumber-coerces null to 0, evaluates
		// `0 >= -1` true, and skips. rslint's GetTabIndexEx surfaces this as
		// `nullLike`, and the rule's gate-4 takes the matching skip arm.
		// These cases lock the contract aligned with upstream byte-for-byte.
		{Code: `<div aria-activedescendant={x} tabIndex={0 satisfies number} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={(-2) satisfies number} />;`, Tsx: true},
		{Code: `async function f() { return <div aria-activedescendant={x} tabIndex={await p} />; }`, Tsx: true},
		{Code: `function* g() { yield <div aria-activedescendant={x} tabIndex={yield 0} />; }`, Tsx: true},

		// ---- Comments around the attribute don't suppress detection ----
		// Trivia is whitespace-equivalent for the listener; the attribute
		// resolves identically. Locks against accidental trivia-driven
		// regressions in FindAttributeByName / GetTabIndex.
		{Code: `<div /* before */ aria-activedescendant={x} /* mid */ tabIndex={0} /* after */ />;`, Tsx: true},

		// ============================================================
		// Expression-form coverage for tabIndex value (valid arms)
		// ============================================================
		// ConditionalExpression — both branches valid (>= -1). Upstream's
		// ConditionalExpression extractor evaluates the test arm; for unknown
		// identifiers truthy-default, so whenTrue is taken.
		{Code: `<div aria-activedescendant={x} tabIndex={cond ? 0 : -1} />;`, Tsx: true},
		// Statically-evaluable test arm: `true` → whenTrue.
		{Code: `<div aria-activedescendant={x} tabIndex={true ? -1 : -5} />;`, Tsx: true},
		// Statically-evaluable falsy test arm: `false` → whenFalse.
		{Code: `<div aria-activedescendant={x} tabIndex={false ? -5 : 0} />;`, Tsx: true},

		// LogicalExpression `??` — null on left, right wins (0).
		{Code: `<div aria-activedescendant={x} tabIndex={null ?? 0} />;`, Tsx: true},

		// LogicalExpression `||` — number on left wins (1).
		{Code: `<div aria-activedescendant={x} tabIndex={1 || 5} />;`, Tsx: true},

		// LogicalExpression `&&` — truthy left → right wins (0).
		{Code: `<div aria-activedescendant={x} tabIndex={true && 0} />;`, Tsx: true},

		// BinaryExpression arithmetic — both operands numeric.
		{Code: `<div aria-activedescendant={x} tabIndex={1 - 1} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={1 - 2} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={2 * 0} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={10 / 10} />;`, Tsx: true},

		// ArrayLiteralExpression — Array.join with single numeric element.
		{Code: `<div aria-activedescendant={x} tabIndex={[0]} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={[-1]} />;`, Tsx: true},
		// Array with null element — Array.join special-cases null to "".
		{Code: `<div aria-activedescendant={x} tabIndex={[null]} />;`, Tsx: true},
		// Empty array → "" → Number("") = 0.
		{Code: `<div aria-activedescendant={x} tabIndex={[]} />;`, Tsx: true},

		// Numeric literal radix variants — all 0.
		{Code: `<div aria-activedescendant={x} tabIndex={0x0} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={0o0} />;`, Tsx: true},
		{Code: `<div aria-activedescendant={x} tabIndex={0b0} />;`, Tsx: true},
		// Scientific notation (1e0 = 1).
		{Code: `<div aria-activedescendant={x} tabIndex={1e0} />;`, Tsx: true},
		// Numeric separators (NumericLiteral text "1_000").
		{Code: `<div aria-activedescendant={x} tabIndex={1_000} />;`, Tsx: true},

		// JS-RESERVED `Infinity` Identifier resolves to +Inf, +Inf >= -1 → skip.
		// Upstream: getLiteralPropValue → null (Identifier→null per LITERAL_TYPES);
		// step-2 getPropValue → +Inf via JS_RESERVED extractor; +Inf >= -1 → true.
		{Code: `<div aria-activedescendant={x} tabIndex={Infinity} />;`, Tsx: true},
		// Negative-zero — JS `-0 >= -1` is true.
		{Code: `<div aria-activedescendant={x} tabIndex={-0} />;`, Tsx: true},
		// Double-unary minus — `-(-2)` = 2.
		{Code: `<div aria-activedescendant={x} tabIndex={-(-2)} />;`, Tsx: true},

		// Multi-level parentheses — paren-stripping must be transparent.
		{Code: `<div aria-activedescendant={x} tabIndex={(((-1)))} />;`, Tsx: true},

		// ============================================================
		// aria-activedescendant value-form coverage (gate-1 only checks
		// presence, but the value form must not crash the listener)
		// ============================================================
		// Direct StringLiteral initializer (HTML-style attribute value).
		// gate-1 sees the prop; div + no tabIndex would be invalid — so we
		// keep this valid by adding tabIndex={0}.
		{Code: `<div aria-activedescendant="staticID" tabIndex={0} />;`, Tsx: true},
		// NoSubstitutionTemplate as activedescendant value.
		{Code: "<div aria-activedescendant={`staticID`} tabIndex={0} />;", Tsx: true},
		// MemberExpression as activedescendant value.
		{Code: `<div aria-activedescendant={state.focusedId} tabIndex={0} />;`, Tsx: true},
		// ConditionalExpression as activedescendant value.
		{Code: `<div aria-activedescendant={cond ? a : b} tabIndex={0} />;`, Tsx: true},
		// CallExpression as activedescendant value.
		{Code: `<div aria-activedescendant={getId()} tabIndex={0} />;`, Tsx: true},

		// ============================================================
		// Real-world a11y patterns
		// ============================================================
		// combobox pattern: input is interactive, no tabIndex needed.
		{Code: `<input role="combobox" aria-activedescendant={x} />;`, Tsx: true},
		// listbox pattern: ul with explicit tabIndex={0}.
		{Code: `<ul role="listbox" tabIndex={0} aria-activedescendant={focusedId}><li>x</li></ul>;`, Tsx: true},
		// Dynamic href on anchor — href attribute presence (any value)
		// matches the interactive a-with-href schema.
		{Code: `<a href={url} aria-activedescendant={x} />;`, Tsx: true},
		// Map render of accessible items.
		{Code: `function L() { return items.map(i => <li tabIndex={-1} aria-activedescendant={i.id} key={i.id}>x</li>); }`, Tsx: true},

		// ============================================================
		// Tag-name case sensitivity — uppercase 'DIV' is treated as a
		// component reference per JSX semantics; not in lowercase dom map.
		// ============================================================
		{Code: `<DIV aria-activedescendant={x} />;`, Tsx: true},

	}, []rule_tester.InvalidTestCase{
		// TSNonNullExpression on tabIndex — jsx-ast-utils' TSNonNullExpression
		// extractor stringifies (`0!` → "0!", `(0)!` → "0!", `(5)!` → "5!").
		// Number(string) = NaN → step-1 undefined → aria gate-4 `undefined >=
		// -1` false → REPORT. Lock-in for Cluster A.
		{Code: `<div aria-activedescendant={x} tabIndex={0!} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div aria-activedescendant={x} tabIndex={(0)!} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- gate-1: boolean attribute form is "defined" per getProp ----
		// `<div aria-activedescendant />` — upstream's `getProp` returns the
		// attribute regardless of value, so gate-1 does NOT skip. div is not
		// interactive, no tabIndex → report.
		{Code: `<div aria-activedescendant />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- gate-1: explicit-undefined value is "defined" per getProp ----
		// jsx-ast-utils' `getProp` returns the attribute node even when the
		// value evaluates to undefined; only a MISSING prop trips gate-1.
		{Code: `<div aria-activedescendant={undefined} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Case-insensitive prop match — INVALID arm ----
		// Same case-folding path as the valid arm above, but without a
		// tabIndex → reports. Locks the matching against accidental
		// case-sensitive regressions in `FindAttributeByName`.
		{Code: `<div Aria-ActiveDescendant={x} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Paired element form (upstream tests use only self-closing) ----
		// JsxElement contains a JsxOpeningElement; the listener fires on
		// the opening element. Position is still column 1 (the `<` of the
		// opening element). Locks that the listener doesn't accidentally
		// double-fire on JsxElement and JsxOpeningElement both.
		{Code: `<div aria-activedescendant={x}></div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- gate-3 boundary: interactive + tabIndex defined → DON'T skip ----
		// `<input>` is interactive, but `tabIndex={-2}` makes hasTabIndex true,
		// so gate-3's `!hasTabIndex && IsInteractive` is false (gate-3 only
		// short-circuits when tabIndex is undefined). gate-4: -2 >= -1 is
		// false → report. Locks that gate-3 isn't accidentally widened to
		// "interactive → always skip".
		{Code: `<input aria-activedescendant={x} tabIndex={-2} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- a without href is NOT interactive ----
		// interactiveElementRoleSchemas requires `href` on `a`; absent it,
		// `a` falls through to the non-interactive schemas
		// (`{Name: "a"}`). gate-3 doesn't skip; no tabIndex → report.
		{Code: `<a aria-activedescendant={x} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- gate-4 boundary: tabIndex < -1 ----
		// `tabIndex={-2}` on a non-interactive element. -2 >= -1 is false →
		// gate-4 doesn't skip → report.
		{Code: `<div aria-activedescendant={x} tabIndex={-2} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- gate-4: NaN-coercing tabIndex string is treated as undefined ----
		// `tabIndex="abc"` → GetTabIndex returns (_, false). hasTabIndex is
		// false, so gate-4 falls through to report (mirrors upstream
		// `Number(getLiteralPropValue) → NaN → undefined` flow).
		{Code: `<div aria-activedescendant={x} tabIndex="abc" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- polymorphicPropName resolution ----
		// `<Box as="div" />` resolves to div via the polymorphic prop;
		// without a tabIndex → report. Locks that the polymorphic resolution
		// applies before the dom.has() / interactive checks.
		{Code: `<Box as="div" aria-activedescendant={x} />;`, Tsx: true, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Nested JSX — listener fires per element independently ----
		// Outer span has no activedescendant (gate-1 skip); inner div has
		// activedescendant but no tabIndex (report). Position assertion
		// covers the inner element's column (after `<span>`).
		{Code: `<span><div aria-activedescendant={x} /></span>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 7}}},

		// ---- Multi-line case for position assertion ----
		// Locks line/column accounting on a non-trivial source layout. The
		// `<div` token starts at column 3 of line 2.
		{
			Code: `
  <div
    aria-activedescendant={x}
  />;`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 2, Column: 3},
			},
		},

		// ---- TS-wrapper tabIndex value resolving to invalid range ----
		// `(-2 as number)` / `(-2)!` should still resolve to -2 via staticEval's
		// wrapper unwrap, then gate-4 fails. Locks that the TS-wrapper unwrap
		// doesn't accidentally lose the negative sign.
		{Code: `<div aria-activedescendant={x} tabIndex={(-2 as number)} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div aria-activedescendant={x} tabIndex={(-2)!} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Multiple invalid elements in one source ----
		// Listener fires per element independently — three sibling self-closing
		// elements each lacking tabIndex produce three reports. Locks that the
		// listener doesn't share state across siblings.
		// Column accounting (1-based): `<>` cols 1-2; first `<div...` cols 3-35;
		// `<span...` cols 36-69; `<p...` cols 70-100.
		{
			Code: `<><div aria-activedescendant={x} /><span aria-activedescendant={y} /><p aria-activedescendant={z} /></>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 36},
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 70},
			},
		},

		// ---- Real-world component patterns ----
		// Verifies the listener fires inside arrow / function / class /
		// fragment-wrapped JSX trees identically to top-level placement.
		{
			Code:   `const Item = () => <li aria-activedescendant={x} />;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 20}},
		},
		{
			Code:   `function R() { return <div aria-activedescendant={x} />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 23}},
		},

		// ============================================================
		// Expression-form coverage for tabIndex value (invalid arms)
		// ============================================================
		// ConditionalExpression — both branches < -1.
		{Code: `<div aria-activedescendant={x} tabIndex={cond ? -2 : -3} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Statically-evaluable falsy test, falsy branch is invalid.
		{Code: `<div aria-activedescendant={x} tabIndex={false ? 0 : -2} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// LogicalExpression `||` with truthy left → left wins (string), NaN.
		{Code: `<div aria-activedescendant={x} tabIndex={x || 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// LogicalExpression `??` — left non-null/undef wins (Identifier name → NaN).
		{Code: `<div aria-activedescendant={x} tabIndex={x ?? 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// BinaryExpression — `1 - 5` = -4.
		{Code: `<div aria-activedescendant={x} tabIndex={1 - 5} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// ArrayLiteralExpression — single negative element.
		{Code: `<div aria-activedescendant={x} tabIndex={[-2]} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Array with two numeric elements — Array.join "5,6" → NaN.
		{Code: `<div aria-activedescendant={x} tabIndex={[5,6]} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Boolean literal as tabIndex — upstream's getTabIndex step-1 boolean
		// arm returns undefined; downstream `undefined >= -1` is false.
		{Code: `<div aria-activedescendant={x} tabIndex={true} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div aria-activedescendant={x} tabIndex={false} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Null literal — upstream LITERAL_TYPES.Literal special-cases null
		// to the string "null"; Number("null") = NaN → undefined → REPORT.
		{Code: `<div aria-activedescendant={x} tabIndex={null} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Negative Infinity — -Inf < -1 → REPORT.
		{Code: `<div aria-activedescendant={x} tabIndex={-Infinity} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Template literal with NumericLiteral substitution. jsx-ast-utils'
		// TemplateLiteral.js does NOT recurse into substitution literals;
		// NumericLiteral falls to the `otherwise → ""` arm. Template value
		// becomes the empty string, step-1 short-circuits to undefined,
		// gate-4 NaN >= -1 → false → REPORT. Both rslint and ESLint behave
		// the same way — locks the upstream substitution-literal blind spot.
		{Code: "<div aria-activedescendant={x} tabIndex={`${0}`} />;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Template literal with Identifier substitution renders as `{name}`
		// (single curly). `Number("{cond}")` = NaN → REPORT.
		{Code: "<div aria-activedescendant={x} tabIndex={`${cond}`} />;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Real-world a11y misuse patterns
		// ============================================================
		// listbox without tabIndex — ul is non-interactive, missing focus
		// reachability. The most common real-world failure mode.
		{
			Code:   `<ul aria-activedescendant={x}><li>x</li></ul>;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 1}},
		},
		// Custom focusable wrapper with negative-2 tabIndex.
		{
			Code:   `<section aria-activedescendant={x} tabIndex={-2}>content</section>;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 1}},
		},
		// Map render produces multiple offending elements.
		{
			Code:   `function L() { return items.map(i => <li aria-activedescendant={i.id}>x</li>); }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tabIndexRequired", Message: errorMessage, Line: 1, Column: 38}},
		},

		// ============================================================
		// Multi-line attribute value position assertion
		// ============================================================
		// tabIndex on its own line — the rule reports on the JSX opening
		// element (column 3 of line 2), not on the attribute or value.
		{
			Code: `
  <div
    aria-activedescendant={x}
    tabIndex={
      -2
    }
  />;`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 2, Column: 3},
			},
		},

		// ============================================================
		// Mixed nested tree — outer + inner each independently invalid
		// ============================================================
		{
			Code: `function App() {
  return (
    <ul aria-activedescendant={a}>
      <li aria-activedescendant={b}>x</li>
    </ul>
  );
}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 3, Column: 5},
				{MessageId: "tabIndexRequired", Message: errorMessage, Line: 4, Column: 7},
			},
		},
	})
}
