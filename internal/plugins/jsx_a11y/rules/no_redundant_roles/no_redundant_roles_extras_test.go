package no_redundant_roles

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This file holds every test case that is NOT a 1:1 mirror of upstream's
// own test file. Upstream-parity cases live in
// `no_redundant_roles_upstream_test.go` so it stays trivially comparable
// against future upstream updates via diff.
//
// Two top-level test functions split the surface by intent:
//
//   - [TestNoRedundantRolesExtras]: AST / Dimension 1–4 edge shapes,
//     plus exhaustive coverage of the implicitRoles table arms that
//     upstream's tests skip. Each case names the specific branch or
//     AST quirk it locks in.
//   - [TestNoRedundantRolesRobust]: real-world component / hook / HOC
//     patterns, listener-boundary cases across nested elements, JS-style
//     numeric coercion edge cases, options-shape robustness, and tsgo-vs-
//     ESTree AST differences. Designed to catch silent regressions when
//     the rule or its shared helpers are refactored.

// polymorphicSettings exercises the `polymorphicPropName` settings —
// `<Foo as="button" role="button" />` resolves to nodeType "button" and
// trips the redundant-role check. Upstream's own test file doesn't
// cover this path.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestNoRedundantRolesExtras locks in branches that upstream's test
// file doesn't exercise but are reachable through the rule's listener
// gate. Cases are grouped by the Dimension 1–4 / semantic-walk axis
// they cover; the inline comments name the specific upstream branch /
// AST edge they protect.
func TestNoRedundantRolesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantRolesRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: explicit-role attribute shapes that DON'T
			//              extract to a literal-string role → rule
			//              short-circuits before any comparison.
			// ============================================================

			// ---- Identifier value (`role={x}`) — getLiteralPropValue
			//      returns null → no comparison. ----
			{Code: `<button role={someRole} />`, Tsx: true},
			// ---- CallExpression / MemberExpression value — null under
			//      getLiteralPropValue (TYPES.CallExpression / .MemberExpression
			//      → null). ----
			{Code: `<button role={getRole()} />`, Tsx: true},
			{Code: `<button role={config.role} />`, Tsx: true},
			// ---- Conditional / Logical — non-literal under LITERAL_TYPES. ----
			{Code: `<button role={cond ? "button" : "main"} />`, Tsx: true},
			{Code: `<button role={r || "button"} />`, Tsx: true},
			// ---- Boolean form `<button role />` — extractValue's null-attr-value
			//      path returns true (boolean), but literalPropValue mirror sees
			//      the JsxAttribute with no Initializer and routes to the
			//      "boolean attr" path. We treat boolean form as having no
			//      literal string value → no comparison.
			//      Locks the rule's behavior on degenerate input. ----
			{Code: `<button role />`, Tsx: true},
			// ---- Empty string role — "" extracts to "" → toLowerCase = "" →
			//      `rolesMap.has("")` is false → explicit role is null → no
			//      comparison. ----
			{Code: `<button role="" />`, Tsx: true},
			// ---- Role string that isn't an ARIA role — rolesMap.has filter
			//      drops it → no comparison. ----
			{Code: `<button role="not-a-real-role" />`, Tsx: true},
			// ---- Template literal with substitution — placeholder text is
			//      not a real ARIA role → null → no comparison. ----
			{Code: "<button role={`role-${suffix}`} />", Tsx: true},

			// ============================================================
			// Dimension 1 / Upstream semantic walk — element types that
			//                                       have no implicit role
			//                                       (default branch).
			// ============================================================

			// ---- div, span, p, etc. — not in implicitRoles → null → no
			//      comparison. Locks the default switch arm. ----
			{Code: `<div role="document" />`, Tsx: true},
			{Code: `<span role="button" />`, Tsx: true},
			{Code: `<p role="paragraph" />`, Tsx: true},
			{Code: `<svg role="img" />`, Tsx: true},

			// ============================================================
			// Upstream semantic walk: implicitRoleForImg branches.
			// ============================================================

			// ---- alt="" arm — empty alt suppresses 'img' implicit role.
			//      `<img alt="" role="img" />` is upstream-valid because the
			//      `alt && getLiteralPropValue(alt) === ''` branch returns ''. ----
			{Code: `<img alt="" role="img" />`, Tsx: true},
			// ---- src with `.svg` substring — SVG arm returns '' → no role. ----
			{Code: `<img src="logo.svg" role="img" />`, Tsx: true},
			{Code: `<img src="/assets/icons/x.svg" role="img" />`, Tsx: true},
			// ---- Boolean alt form `<img alt />` extracts to JS true → NOT
			//      empty string → falls through to src check → img stays.
			//      But role="img" matches → REPORT (covered in invalid below).
			//      Lock the "empty alt is special, boolean alt is not" branch.

			// ============================================================
			// Upstream semantic walk: implicitRoleForInput branches.
			// ============================================================

			// ---- type=button → button. role="button" matches → REPORT
			//      (in invalid block). Here lock the NEGATIVE cases. ----
			{Code: `<input type="text" role="button" />`, Tsx: true},
			// ---- type=hidden → upstream's switch falls through to default
			//      'textbox' → role="textbox" matches → REPORT. The
			//      corresponding INVALID is in the invalid block. ----
			{Code: `<input type="hidden" role="button" />`, Tsx: true},
			// ---- No type attribute — defaults to textbox; explicit
			//      role="button" doesn't match. ----
			{Code: `<input role="button" />`, Tsx: true},
			// ---- Non-literal type → default textbox. role="button" doesn't
			//      match → no report. ----
			{Code: `<input type={someType} role="button" />`, Tsx: true},
			// ---- Range → slider. role="slider" matches (in invalid block).
			//      Here: role="button" doesn't match range's slider. ----
			{Code: `<input type="range" role="button" />`, Tsx: true},

			// ============================================================
			// Upstream semantic walk: implicitRoleForSelect branches.
			// ============================================================

			// ---- multiple+listbox via {true} → listbox implicit, but
			//      role="combobox" doesn't match. ----
			{Code: `<select multiple role="combobox" />`, Tsx: true},
			{Code: `<select multiple={true} role="combobox" />`, Tsx: true},
			{Code: `<select size={5} role="combobox" />`, Tsx: true},
			// ---- multiple={null} — literalPropValue maps to "null"
			//      (truthy) → listbox; role="combobox" mismatch → valid.
			//      Locks the upstream `null` → "null" special-case. ----
			{Code: `<select multiple={null} role="combobox" />`, Tsx: true},
			// ---- size value just at boundary — size={1} → 1 > 1 false →
			//      combobox; role="listbox" mismatch → valid. ----
			{Code: `<select size={1} role="listbox" />`, Tsx: true},
			{Code: `<select size="1" role="listbox" />`, Tsx: true},
			// ---- size=NaN-ish (non-numeric string) → NaN > 1 false → combobox. ----
			{Code: `<select size="abc" role="listbox" />`, Tsx: true},
			{Code: `<select size={someN} role="listbox" />`, Tsx: true},
			// ---- Boolean form `<select size />` → upstream's getLiteralPropValue
			//      sees an attr with no Initializer; our LiteralPropTruthy
			//      path treats boolean form as truthy true → 1, 1 > 1 false →
			//      combobox. role="listbox" mismatch → valid. ----
			{Code: `<select size role="listbox" />`, Tsx: true},
			// ---- size={undefined} → undef → falsy → combobox. ----
			{Code: `<select size={undefined} role="listbox" />`, Tsx: true},
			// ---- size hex/oct/bin: 0x01 / 0o1 / 0b1 → all = 1, 1 > 1 false →
			//      combobox; role="listbox" mismatch → valid. ----
			{Code: `<select size="0x01" role="listbox" />`, Tsx: true},
			{Code: `<select size={0x01} role="listbox" />`, Tsx: true},
			{Code: `<select size="0o01" role="listbox" />`, Tsx: true},
			{Code: `<select size="0b01" role="listbox" />`, Tsx: true},
			// ---- size with leading/trailing whitespace — JS Number(" 1 ") = 1 → combobox. ----
			{Code: `<select size=" 1 " role="listbox" />`, Tsx: true},
			// ---- size fractional > 1 BUT testing NOT-listbox role: a fractional
			//      ≤ 1 maps to combobox. ----
			{Code: `<select size="0.5" role="listbox" />`, Tsx: true},
			{Code: `<select size={0.5} role="listbox" />`, Tsx: true},

			// ============================================================
			// Upstream semantic walk: a/area/link href branch.
			// ============================================================

			// ---- No href → no implicit role → role="link" doesn't match. ----
			{Code: `<a role="link" />`, Tsx: true},
			{Code: `<area role="link" />`, Tsx: true},
			{Code: `<link role="link" />`, Tsx: true},
			// ---- href present (any value) → implicit 'link'. role mismatch
			//      → valid. ----
			{Code: `<a href="/x" role="button" />`, Tsx: true},

			// ============================================================
			// Upstream semantic walk: menu / menuitem type branches.
			// ============================================================

			// ---- menu type=toolbar → toolbar; role="menu" mismatch. ----
			{Code: `<menu type="toolbar" role="menu" />`, Tsx: true},
			// ---- menu without type → '' implicit → no comparison. ----
			{Code: `<menu role="menu" />`, Tsx: true},
			// ---- menuitem with no/unknown type → '' → no comparison. ----
			{Code: `<menuitem role="menuitem" />`, Tsx: true},
			{Code: `<menuitem type="other" role="menuitem" />`, Tsx: true},
			// ---- menuitem type=command → menuitem; but role mismatch
			//      `role="checkbox"`. ----
			{Code: `<menuitem type="command" role="checkbox" />`, Tsx: true},
			// ---- menuitem type=radio → menuitemradio; role mismatch. ----
			{Code: `<menuitem type="radio" role="menuitem" />`, Tsx: true},

			// ============================================================
			// Dimension 2: Scoping & nesting — listener fires per element.
			// ============================================================

			// ---- Nested elements — only matching ones report (none here). ----
			{Code: `<div><div /><span /></div>`, Tsx: true},
			// ---- Self-closing inside JsxFragment. ----
			{Code: `<><button role="main" /></>`, Tsx: true},
			// ---- Mixed scope: outer element has redundant role, inner
			//      doesn't (or vice versa) — handled per element. ----
			{Code: `<nav><a href="/x" role="button" /></nav>`, Tsx: true},

			// ============================================================
			// Dimension 4: case-insensitive role attribute name (jsx-ast-utils
			//              `getProp` default `ignoreCase: true`).
			// ============================================================

			{Code: `<button ROLE="main" />`, Tsx: true},
			{Code: `<button Role="main" />`, Tsx: true},
			{Code: `<button rOlE="main" />`, Tsx: true},

			// ============================================================
			// Dimension 4: TS expression wrappers around role value.
			//              jsx-ast-utils' literalPropValue does NOT strip
			//              TSAsExpression / TSNonNullExpression — they hit
			//              the `noop → null` arm → no comparison.
			// ============================================================

			// ---- TS `as` wrapper — literalPropValue returns null → no
			//      comparison → valid even though "button" is inside. ----
			{Code: `<button role={"button" as string} />`, Tsx: true},
			// ---- TS `!` non-null assertion — same as `as`. ----
			{Code: `<button role={"button"!} />`, Tsx: true},
			// ---- TS `satisfies` — same. ----
			{Code: `<button role={"button" satisfies any} />`, Tsx: true},

			// ============================================================
			// Dimension 4: parenthesized role value (single + multi-level).
			// ============================================================

			{Code: `<button role={("main")} />`, Tsx: true},
			{Code: `<button role={(("main"))} />`, Tsx: true},

			// ============================================================
			// Dimension 4: element-type forms — capitalization, member access,
			//              namespaced names. Custom components / non-DOM types
			//              skip the implicit-role lookup.
			// ============================================================

			// ---- Capitalized HTML names — case-sensitive table miss. ----
			{Code: `<BODY role="document" />`, Tsx: true},
			{Code: `<Button role="button" />`, Tsx: true},
			// ---- Member-access tag — type "Foo.Button" → not in table. ----
			{Code: `<Foo.Button role="button" />`, Tsx: true},
			// ---- Namespaced JSX (XML-like) — type "svg:rect" → not in table. ----
			{Code: `<svg:rect role="img" />`, Tsx: true},

			// ============================================================
			// componentsSettings + non-matching role → valid.
			// ============================================================

			// `<Button>` remaps to "button" via settings. role="main" ≠ button
			// → valid.
			{Code: `<Button role="main" />`, Tsx: true, Settings: componentsSettings},

			// ============================================================
			// polymorphicPropName settings (rslint extra coverage).
			// ============================================================

			// `<Foo as="div" role="button" />` resolves to div → no implicit
			// role → no comparison.
			{Code: `<Foo as="div" role="button" />`, Tsx: true, Settings: polymorphicSettings},

			// ============================================================
			// Literal-spread role lookup — jsx-ast-utils' getProp walks
			// literal ObjectLiteral spreads (Phase 1 Dimension 4: literal
			// spreads).
			// ============================================================

			// ---- Spread-only attributes with no role match — opaque
			//      non-literal spread → role not found → no comparison. ----
			{Code: `<button {...props} />`, Tsx: true},
			// ---- Literal-spread with non-matching role. ----
			{Code: `<button {...{role: "main"}} />`, Tsx: true},

			// ============================================================
			// Multi-line / formatted JSX — line/column must still resolve.
			// ============================================================

			{Code: "<button\n  role=\"main\"\n/>", Tsx: true},

			// ============================================================
			// Allow-list option permutations.
			// ============================================================

			// ---- Custom allow-list: button is explicitly allowed for the
			//      'button' implicit role. ----
			{
				Code:    `<button role="button" />`,
				Tsx:     true,
				Options: map[string]interface{}{"button": []interface{}{"button"}},
			},
			// ---- Allow-list with different element doesn't affect the
			//      current element. (Locked previously by upstream tests
			//      but re-confirmed for clarity.) ----
			{
				Code:    `<nav role="navigation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"nav": []interface{}{"navigation"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Upstream semantic walk: img — boolean alt form does NOT
			// suppress 'img' (only literal-empty-string alt does).
			// ============================================================

			// ---- `<img alt />` — boolean form extracts to true, NOT '' →
			//      img stays. role="img" matches → REPORT. Upstream's tests
			//      don't cover this. ----
			{
				Code:   `<img alt role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("img", "img")},
			},
			// ---- `<img alt={undefined} />` — undefined → null under
			//      literalPropValue → falsy → not '' → img stays → REPORT. ----
			{
				Code:   `<img alt={undefined} role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("img", "img")},
			},
			// ---- `<img alt={x} src="x.png" />` — non-literal alt → not ''
			//      → img stays. src has no .svg → img stays → REPORT. ----
			{
				Code:   `<img alt={x} src="logo.png" role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("img", "img")},
			},

			// ============================================================
			// Upstream semantic walk: input type → ARIA role mapping.
			// ============================================================

			// ---- type=button → button implicit. role="button" matches → REPORT. ----
			{
				Code:   `<input type="button" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "button")},
			},
			// ---- type=submit → button. ----
			{
				Code:   `<input type="submit" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "button")},
			},
			// ---- type=reset → button. ----
			{
				Code:   `<input type="reset" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "button")},
			},
			// ---- type=image → button. ----
			{
				Code:   `<input type="image" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "button")},
			},
			// ---- type=checkbox → checkbox. ----
			{
				Code:   `<input type="checkbox" role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "checkbox")},
			},
			// ---- type=radio → radio. ----
			{
				Code:   `<input type="radio" role="radio" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "radio")},
			},
			// ---- type=range → slider. ----
			{
				Code:   `<input type="range" role="slider" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "slider")},
			},
			// ---- No type → textbox. role="textbox" matches → REPORT. ----
			{
				Code:   `<input role="textbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "textbox")},
			},
			// ---- type="text" → falls through to default (textbox). ----
			{
				Code:   `<input type="text" role="textbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "textbox")},
			},
			// ---- type="hidden" → also falls through to default (textbox)
			//      because upstream's switch only special-cases button /
			//      checkbox / radio / range. ----
			{
				Code:   `<input type="hidden" role="textbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "textbox")},
			},

			// ============================================================
			// JS-style number coercion for select `size`: hex / oct / bin
			// string prefixes coerce per JS Number() — `<select size="0x10" />`
			// → 16 > 1 → listbox. Locks the upstream behavior that a
			// hand-rolled strconv.ParseFloat would miss.
			// ============================================================

			// ---- size="0x10" → Number("0x10") = 16 → 16 > 1 → listbox. ----
			{
				Code:   `<select size="0x10" role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- size="0o10" → 8 → listbox. ----
			{
				Code:   `<select size="0o10" role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- size="0b10" → 2 → listbox. ----
			{
				Code:   `<select size="0b10" role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- size={1.5} → 1.5 > 1 → listbox. ----
			{
				Code:   `<select size={1.5} role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- size="1.5" → JS Number("1.5") = 1.5 → > 1 → listbox. ----
			{
				Code:   `<select size="1.5" role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- size with leading/trailing whitespace " 3 " → Number trims → 3 → listbox. ----
			{
				Code:   `<select size=" 3 " role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- multiple={null} — literalPropValue maps null → "null"
			//      string → !!truthy → listbox. role="listbox" matches. ----
			{
				Code:   `<select multiple={null} role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- multiple="anything" — non-empty string → truthy → listbox. ----
			{
				Code:   `<select multiple="any" role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- type case-insensitive (upstream toUpperCase). ----
			{
				Code:   `<input type="BUTTON" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("input", "button")},
			},
			// ---- Role case-insensitive — explicit role is lower-cased
			//      before comparison. ----
			{
				Code:   `<button role="BUTTON" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},
			{
				Code:   `<button role="Button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},

			// ============================================================
			// Upstream semantic walk: every implicit-role table entry that
			// upstream's own tests don't exercise — locks each table arm.
			// ============================================================

			// ---- article. ----
			{
				Code:   `<article role="article" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("article", "article")},
			},
			// ---- aside. ----
			{
				Code:   `<aside role="complementary" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("aside", "complementary")},
			},
			// ---- datalist. ----
			{
				Code:   `<datalist role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("datalist", "listbox")},
			},
			// ---- details. ----
			{
				Code:   `<details role="group" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("details", "group")},
			},
			// ---- dialog. ----
			{
				Code:   `<dialog role="dialog" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("dialog", "dialog")},
			},
			// ---- form. ----
			{
				Code:   `<form role="form" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("form", "form")},
			},
			// ---- h1-h6 → heading. ----
			{
				Code:   `<h1 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h1", "heading")},
			},
			{
				Code:   `<h2 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h2", "heading")},
			},
			{
				Code:   `<h3 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h3", "heading")},
			},
			{
				Code:   `<h4 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h4", "heading")},
			},
			{
				Code:   `<h5 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h5", "heading")},
			},
			{
				Code:   `<h6 role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("h6", "heading")},
			},
			// ---- hr → separator. ----
			{
				Code:   `<hr role="separator" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("hr", "separator")},
			},
			// ---- li → listitem. ----
			{
				Code:   `<li role="listitem" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("li", "listitem")},
			},
			// ---- meter, progress → progressbar. ----
			{
				Code:   `<meter role="progressbar" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("meter", "progressbar")},
			},
			{
				Code:   `<progress role="progressbar" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("progress", "progressbar")},
			},
			// ---- option → option. ----
			{
				Code:   `<option role="option" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("option", "option")},
			},
			// ---- output → status. ----
			{
				Code:   `<output role="status" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("output", "status")},
			},
			// ---- section → region. ----
			{
				Code:   `<section role="region" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("section", "region")},
			},
			// ---- tbody/tfoot/thead → rowgroup. ----
			{
				Code:   `<tbody role="rowgroup" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("tbody", "rowgroup")},
			},
			{
				Code:   `<tfoot role="rowgroup" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("tfoot", "rowgroup")},
			},
			{
				Code:   `<thead role="rowgroup" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("thead", "rowgroup")},
			},
			// ---- textarea → textbox. ----
			{
				Code:   `<textarea role="textbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("textarea", "textbox")},
			},

			// ============================================================
			// Upstream semantic walk: a/area/link with href → link role.
			// ============================================================

			// ---- a href present → link implicit. ----
			{
				Code:   `<a href="/x" role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("a", "link")},
			},
			{
				Code:   `<area href="#x" role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("area", "link")},
			},
			{
				Code:   `<link href="/x" role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("link", "link")},
			},
			// ---- href with non-literal value — getProp returns the attr
			//      regardless of value, so implicit is still 'link'. ----
			{
				Code:   `<a href={url} role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("a", "link")},
			},

			// ============================================================
			// Upstream semantic walk: menu / menuitem type branches.
			// ============================================================

			// ---- menu type=toolbar → toolbar implicit. ----
			{
				Code:   `<menu type="toolbar" role="toolbar" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("menu", "toolbar")},
			},
			// ---- menuitem type=command → menuitem. ----
			{
				Code:   `<menuitem type="command" role="menuitem" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("menuitem", "menuitem")},
			},
			// ---- menuitem type=checkbox → menuitemcheckbox. ----
			{
				Code:   `<menuitem type="checkbox" role="menuitemcheckbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("menuitem", "menuitemcheckbox")},
			},
			// ---- menuitem type=radio → menuitemradio. ----
			{
				Code:   `<menuitem type="radio" role="menuitemradio" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("menuitem", "menuitemradio")},
			},

			// ============================================================
			// polymorphicPropName settings — `<Foo as="button">` resolves
			// to "button" → redundant role triggers.
			// ============================================================

			{
				Code:     `<Foo as="button" role="button" />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},
			{
				Code:     `<Foo as="body" role="document" />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{invalidErr("body", "document")},
			},

			// ============================================================
			// Allow-list options — explicit empty array suppresses the
			// default nav allowance.
			// ============================================================

			// ---- options-array shape: `[{ nav: ['navigation'] }]` is the
			//      shape rule_tester sends via JSON; the same shape covers
			//      the array-format CLI invocation. Both pass the
			//      `additionalProperties` schema. ----
			{
				Code:    `<button role="button" />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{}}, // empty allow-list per element
				Errors:  []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},

			// ============================================================
			// Literal-spread role attribute — getProp walks ObjectLiteral
			// spreads. Locks Dimension 4 universal edge: literal spread.
			// ============================================================

			{
				Code:   `<button {...{role: "button"}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},

			// ============================================================
			// Multi-line opening tag — column 1 of the JsxOpeningElement.
			// ============================================================

			{
				Code: "<button\n  role=\"button\"\n/>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noRedundantRoles", Message: errorMessage("button", "button"), Line: 1, Column: 1},
				},
			},

			// ============================================================
			// Position assertions — multiple elements, line/col offsets.
			// ============================================================

			{
				Code: `<>
  <div />
  <button role="button" />
</>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noRedundantRoles", Message: errorMessage("button", "button"), Line: 3, Column: 3},
				},
			},

			// ============================================================
			// Paired (non-self-closing) form — listener fires on the opening tag.
			// ============================================================

			{
				Code:   `<button role="button">Click</button>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},
		},
	)
}
// TestNoRedundantRolesRobust covers two surfaces that the upstream test
// file and our Dimension-1–4 extras don't reach:
//
//  1. Real-world component / hook / HOC patterns where the rule must still
//     fire (or, more often, stay silent) — these protect against regressions
//     from listener / scope-walk refactors.
//
//  2. Branches that only diverge between tsgo and ESTree at the AST level —
//     template-with-substitutions, parens around the receiver vs the role,
//     TS wrapper transparency vs opacity, JsxExpression empty containers,
//     SkipParentheses depth, NumericLiteral text-normalization (separators,
//     hex), and Identifier "undefined" / "Infinity" / "NaN" reserved
//     handling.
//
// Each case carries an inline comment naming the SPECIFIC upstream branch
// or AST quirk it locks in. Cases are organized so a top-to-bottom read
// gives a mental model of the rule's full behavior surface.
func TestNoRedundantRolesRobust(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantRolesRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Real-world component patterns
			// ============================================================

			// ---- React.forwardRef wrapper — the inner JSX still fires per
			//      element. Wrapping the component should not affect the
			//      diagnostic surface; here the wrapped element doesn't
			//      duplicate its implicit role, so no report. ----
			{
				Code: `const Btn = React.forwardRef((props, ref) => <button ref={ref} {...props}>OK</button>);`,
				Tsx:  true,
			},
			// ---- React.memo wrapper with non-redundant role. ----
			{
				Code: `const Card = React.memo(({ id }) => <button id={id}>{id}</button>);`,
				Tsx:  true,
			},
			// ---- HOC composition — the inner JSX is what the listener walks. ----
			{
				Code: `const Enhanced = withTracking(({ v }) => <button data-v={v}>OK</button>);`,
				Tsx:  true,
			},
			// ---- props.children render prop returning JSX. ----
			{
				Code: `function Wrapper({ render }) { return render({ size: 2 }); }`,
				Tsx:  true,
			},
			// ---- Array.map over items rendering non-redundant elements. ----
			{
				Code: `const list = items.map(x => <li key={x.id}>{x.label}</li>);`,
				Tsx:  true,
			},
			// ---- Functional component with defaulted destructured props. ----
			{
				Code: `function Heading({ level = 2, children }) { return <h2>{children}</h2>; }`,
				Tsx:  true,
			},
			// ---- Class component render. ----
			{
				Code: `class Item extends React.Component { render() { return <li>item</li>; } }`,
				Tsx:  true,
			},
			// ---- Async function component — the listener walks all JSX
			//      regardless of containing function flavor. ----
			{
				Code: `async function Async() { return <button>OK</button>; }`,
				Tsx:  true,
			},
			// ---- Generator function with yielded JSX. ----
			{
				Code: `function* gen() { yield <li>1</li>; yield <li>2</li>; }`,
				Tsx:  true,
			},
			// ---- IIFE returning JSX. ----
			{
				Code: `const x = (() => <button>OK</button>)();`,
				Tsx:  true,
			},
			// ---- Conditional rendering `cond && <X />`. ----
			{
				Code: `const x = <>{cond && <button>OK</button>}</>;`,
				Tsx:  true,
			},
			// ---- Ternary in a child position. ----
			{
				Code: `const x = <div>{cond ? <button>Y</button> : <a href="/x">N</a>}</div>;`,
				Tsx:  true,
			},
			// ---- TypeScript generic on a CUSTOM component — Custom<T> is
			//      not in the implicitRoles table, so no comparison runs. ----
			{
				Code: `<List<string> role="list" />`,
				Tsx:  true,
			},
			{
				Code: `<Cell<{a: number}> role="cell" />`,
				Tsx:  true,
			},
			// ---- this.Foo / Foo.Bar.Baz / namespace member tags — not in
			//      implicitRoles → no comparison. ----
			{Code: `<this.Foo role="button" />`, Tsx: true},
			{Code: `<Foo.Bar.Baz role="button" />`, Tsx: true},
			// ---- Custom-element naming (`<my-element>`) — tsgo carries
			//      the hyphenated tag name verbatim; not in implicitRoles
			//      table. ----
			{Code: `<my-element role="button" />`, Tsx: true},

			// ============================================================
			// Reserved / non-implicit-role elements — explicit `role` is
			// not redundant because the element has no implicit role.
			// ============================================================

			{Code: `<div role="navigation" />`, Tsx: true},
			{Code: `<div role="main" />`, Tsx: true},
			{Code: `<span role="region" />`, Tsx: true},
			// ---- main element — surprisingly NOT in upstream implicitRoles.
			//      Locks the "table-arm exhaustion" against future drift. ----
			{Code: `<main role="main" />`, Tsx: true},
			// ---- header / footer — also not in upstream's table. ----
			{Code: `<header role="banner" />`, Tsx: true},
			{Code: `<footer role="contentinfo" />`, Tsx: true},
			// ---- search — landmark in HTML5 but not in the implicit map. ----
			{Code: `<search role="search" />`, Tsx: true},
			// ---- table — surprisingly NOT in upstream's implicitRoles
			//      (only the row-group children are). ----
			{Code: `<table role="table" />`, Tsx: true},
			{Code: `<tr role="row" />`, Tsx: true},
			{Code: `<td role="cell" />`, Tsx: true},
			{Code: `<th role="columnheader" />`, Tsx: true},
			// ---- Media elements — no implicit role in the table. ----
			{Code: `<audio role="application" />`, Tsx: true},
			{Code: `<video role="application" />`, Tsx: true},
			{Code: `<canvas role="img" />`, Tsx: true},
			{Code: `<iframe role="application" />`, Tsx: true},
			{Code: `<object role="application" />`, Tsx: true},
			{Code: `<embed role="application" />`, Tsx: true},
			// ---- Form structure elements — not in implicit table. ----
			{Code: `<label role="generic" />`, Tsx: true},
			{Code: `<fieldset role="group" />`, Tsx: true},
			{Code: `<legend role="region" />`, Tsx: true},

			// ============================================================
			// Role attribute value edge shapes
			// ============================================================

			// ---- role with surrounding whitespace — upstream's toLowerCase
			//      preserves whitespace, "  button  " is not in rolesMap →
			//      no comparison → valid. ----
			{Code: `<button role="  button  " />`, Tsx: true},
			// ---- role with leading whitespace only. ----
			{Code: `<button role=" button" />`, Tsx: true},
			// ---- role with multiple space-separated values — first token rule
			//      doesn't apply here; upstream uses the entire string for
			//      rolesMap.has, "button main" is not a role → valid. ----
			{Code: `<button role="button main" />`, Tsx: true},
			// ---- role with hyphenated invalid value. ----
			{Code: `<button role="custom-button" />`, Tsx: true},
			// ---- role with numbers — not in rolesMap. ----
			{Code: `<button role="button1" />`, Tsx: true},
			// ---- role with dot — DPub roles use `doc-*` prefix; "button.foo"
			//      is NOT in rolesMap → valid. ----
			{Code: `<button role="button.foo" />`, Tsx: true},
			// ---- role with single quote inside double-quoted string. ----
			{Code: `<button role="button's" />`, Tsx: true},
			// ---- role with empty space — fully whitespace. ----
			{Code: `<button role="   " />`, Tsx: true},
			// ---- role with role-like but wrong-case substring. ----
			{Code: `<button role="iAmAButton" />`, Tsx: true},

			// ============================================================
			// JsxExpression edge shapes
			// ============================================================

			// ---- Empty JsxExpression `role={}` — tsgo synthesizes for
			//      malformed source. literalPropValue returns null → no
			//      comparison. ----
			{Code: `<button role={} />`, Tsx: true},
			// ---- Comment-only JsxExpression `role={/* x */}` — same as
			//      empty container. ----
			{Code: `<button role={/* comment */} />`, Tsx: true},
			// ---- JsxExpression with multi-line comment inside literal. ----
			{Code: `<button role={/* picked */ "main"} />`, Tsx: true},

			// ============================================================
			// Template literal substitution shapes (tsgo vs ESTree differ
			// on placeholder synthesis; we mirror upstream via static_eval).
			// ============================================================

			// ---- ${a}${b} multi-substitution — the placeholder synthesis
			//      yields a non-role string → valid. ----
			{Code: "<button role={`${a}${b}`} />", Tsx: true},
			// ---- Trailing static text after substitution. ----
			{Code: "<button role={`${a}suffix`} />", Tsx: true},
			// ---- Pure substitution with NO static text. ----
			{Code: "<button role={`${a}`} />", Tsx: true},
			// ---- NoSubstitutionTemplate with a real role name should still
			//      match because jsx-ast-utils' TemplateLiteral.js returns the
			//      raw text (no boolean coercion on this path). "button" ==
			//      "button" → REPORT — but ONLY when wrapped in a template,
			//      not a string literal. See INVALID block below. ----

			// ============================================================
			// Parenthesized receivers / role values — multi-level.
			// ============================================================

			// ---- Parens around the element type itself — not relevant here
			//      because the element type is the tagName, not an
			//      expression. Just ensure no panic. ----
			{Code: `<button role={("main")} />`, Tsx: true},
			{Code: `<button role={(("main"))} />`, Tsx: true},
			{Code: `<button role={((("main")))} />`, Tsx: true},
			// ---- Parens nested with TS wrappers. ----
			{Code: `<button role={("main" as string)} />`, Tsx: true},
			{Code: `<button role={(("main") as string)} />`, Tsx: true},
			{Code: `<button role={(("main")! as string)} />`, Tsx: true},

			// ============================================================
			// Numeric literal forms in select size — ES2021 separators,
			// scientific notation, BigInt.
			// ============================================================

			// ---- Numeric separator `1_000` (ES2021) — must normalize to
			//      number 1000 → > 1 → listbox. Here paired with
			//      role="combobox" so it DOES NOT match → valid. ----
			{Code: `<select size={1_000} role="combobox" />`, Tsx: true},
			// ---- Scientific notation `1e2` = 100 → listbox; with
			//      role="combobox" → no match → valid. ----
			{Code: `<select size={1e2} role="combobox" />`, Tsx: true},
			// ---- Negative size — `-3 > 1` → false → combobox. ----
			{Code: `<select size={-3} role="listbox" />`, Tsx: true},
			// ---- Zero size — combobox. ----
			{Code: `<select size={0} role="listbox" />`, Tsx: true},
			// ---- Tiny fractional — 0.99 > 1 false → combobox. ----
			{Code: `<select size={0.99} role="listbox" />`, Tsx: true},

			// ============================================================
			// Listener boundary: outer and inner element scope independence
			// ============================================================

			// ---- Inner button has redundant role, outer doesn't — only
			//      inner reports. Locked in INVALID block; here lock the
			//      reverse (outer has, inner doesn't) — handled in invalid. ----

			// ---- Mixed forms in same source — no report when none redundant. ----
			{Code: `function App() { return <><div /><button role="main" /><nav><a href="/x" /></nav></>; }`, Tsx: true},

			// ============================================================
			// Spread shapes — opaque vs literal walk
			// ============================================================

			// ---- Non-literal spread is OPAQUE — `getProp` returns nothing
			//      for `role` → no comparison. ----
			{Code: `<button {...props} />`, Tsx: true},
			{Code: `<button {...x.y} />`, Tsx: true},
			{Code: `<button {...obj["k"]} />`, Tsx: true},
			{Code: `<button {...fn()} />`, Tsx: true},
			// ---- Multiple spreads — same opacity rule. ----
			{Code: `<button {...a} {...b} />`, Tsx: true},
			// ---- Spread containing non-role keys — role not found → no comparison. ----
			{Code: `<button {...{className: "x"}} />`, Tsx: true},
			// ---- Spread with non-matching role string — comparison runs but
			//      "main" != "button" → no report. ----
			{Code: `<button {...{role: "main"}} />`, Tsx: true},
			// ---- Shorthand role inside spread — `{...{ role }}` carries the
			//      bound Identifier value, which extracts to null via
			//      LITERAL_TYPES.Identifier (non-undefined) → no comparison. ----
			{Code: `<button {...{ role }} />`, Tsx: true},

			// ============================================================
			// Options option-shape robustness
			// ============================================================

			// ---- Empty options array `[]` — no options object reaches
			//      parseOptions; GetOptionsMap returns nil → defaults apply.
			//      `<nav role="navigation" />` stays valid. ----
			{Code: `<nav role="navigation" />`, Tsx: true, Options: []interface{}{}},
			// ---- Options with `{}` empty object — parseOptions yields
			//      empty map → hasOwn false for every key → defaults apply
			//      → `<nav role="navigation" />` stays valid. ----
			{Code: `<nav role="navigation" />`, Tsx: true, Options: map[string]interface{}{}},
			// ---- Custom element with own allow-list. ----
			{
				Code:    `<button role="button" />`,
				Tsx:     true,
				Options: map[string]interface{}{"button": []interface{}{"button"}},
			},
			// ---- Multi-element allow-list. ----
			{
				Code:    `<ul role="list" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"list"}, "ol": []interface{}{"list"}, "nav": []interface{}{"navigation"}},
			},
			// ---- Allow-list with non-string entries — StringSliceOption
			//      drops them silently, leaving the valid strings. ----
			{
				Code:    `<button role="button" />`,
				Tsx:     true,
				Options: map[string]interface{}{"button": []interface{}{"button", 42, nil, true}},
			},
			// ---- Malformed value type (not an array) — StringSliceOption
			//      returns nil; hasOwn still true → allow-list is the nil
			//      slice → no entries match → REPORT (in invalid block). ----

			// ============================================================
			// Polymorphic + components map interaction (rslint additional
			// surface).
			// ============================================================

			// ---- polymorphicAllowList — `<Other as="div" role="main" />`
			//      where `Other` is NOT in the allow-list → as-swap skipped
			//      → resolved = "Other" → not in implicitRoles → valid. ----
			{
				Code: `<Other as="div" role="main" />`,
				Tsx:  true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "as",
						"polymorphicAllowList": []interface{}{"Box"},
					},
				},
			},
			// ---- polymorphic + components — both layers apply. Component
			//      "Custom" maps to "iframe", which has no implicit role. ----
			{
				Code: `<Custom role="application" />`,
				Tsx:  true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{
							"Custom": "iframe",
						},
					},
				},
			},

			// ============================================================
			// Comments inside the JSX attribute area — must not interfere.
			// ============================================================

			{Code: `<button /* a */ role="main" /* b */ />`, Tsx: true},
			{Code: `<button role={/* x */ "main"} />`, Tsx: true},
			{Code: "<button\n  // line comment\n  role=\"main\"\n/>", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Real-world business patterns (must REPORT)
			// ============================================================

			// ---- Array.map rendering with redundant role on every item. ----
			{
				Code: `const list = items.map(x => <li key={x.id} role="listitem">{x.name}</li>);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("li", "listitem", 1),
				},
			},
			// ---- Conditional render — listener walks into the branch JSX. ----
			{
				Code: `function Foo({ on }) { return on ? <button role="button">x</button> : null; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- React.forwardRef body — inner JSX still walked. ----
			{
				Code: `const Btn = React.forwardRef((p, ref) => <button ref={ref} role="button" {...p} />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- React.memo — same. ----
			{
				Code: `const Card = React.memo(({ id }) => <section id={id} role="region" />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("section", "region", 1),
				},
			},
			// ---- HOC with redundant role. ----
			{
				Code: `const E = withT(({ v }) => <button data-v={v} role="button" />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Generator function yielding redundant elements. ----
			{
				Code: `function* g() { yield <li role="listitem">a</li>; yield <li role="listitem">b</li>; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("li", "listitem", 1),
					invalidErrLine("li", "listitem", 1),
				},
			},
			// ---- Class component render. ----
			{
				Code: `class C extends React.Component { render() { return <button role="button">x</button>; } }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Async function component. ----
			{
				Code: `async function A() { return <button role="button">x</button>; }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},

			// ============================================================
			// Listener boundary — nested same-kind elements
			// ============================================================

			// ---- Outer + inner both have redundant role; listener fires on
			//      each opening tag independently → two reports. Inner column
			//      depends on tsgo's pos() conventions; line is what matters. ----
			{
				Code: `<button role="button"><button role="button" /></button>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Different elements nested — each independent. ----
			{
				Code: `<article role="article"><section role="region"><h1 role="heading">x</h1></section></article>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("article", "article", 1),
					invalidErrLine("section", "region", 1),
					invalidErrLine("h1", "heading", 1),
				},
			},
			// ---- Sibling redundant elements — multi-error in one file. ----
			{
				Code: `function App() { return (<><button role="button" /><h1 role="heading">t</h1><nav role="navigation"></nav></>); }`,
				Tsx:  true,
				Options: map[string]interface{}{"nav": []interface{}{}}, // disable nav default
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
					invalidErrLine("h1", "heading", 1),
					invalidErrLine("nav", "navigation", 1),
				},
			},
			// ---- Mixed redundant + non-redundant siblings. ----
			{
				Code: `<><button role="button" /><button role="main" /><h1 role="heading" /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
					invalidErrLine("h1", "heading", 1),
				},
			},
			// ---- Redundant role inside JsxExpression child. ----
			{
				Code: `<div>{<button role="button" />}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Redundant role inside arrow render prop. ----
			{
				Code: `<X render={() => <button role="button" />} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Redundant role inside Fragment. ----
			{
				Code: `<><button role="button" /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
				},
			},
			// ---- Multi-file-like: multiple top-level statements (TSX
			//      module with several JSX exports). Each on its own line. ----
			{
				Code: `function A() { return <button role="button" />; }
function B() { return <h1 role="heading">x</h1>; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("button", "button", 1),
					invalidErrLine("h1", "heading", 2),
				},
			},

			// ============================================================
			// AST-shape edge cases that tsgo vs ESTree differ on.
			// ============================================================

			// ---- Parenthesized role value with TS-non-null. literalPropValue
			//      keeps TS wrappers OPAQUE for the literal path (returns null),
			//      so a wrapper around the role REJECTS the report. The cases
			//      below DO report because the role value extracts to a literal
			//      string in non-wrapped form — these are sanity checks of the
			//      paired form vs self-closing form interplay. ----
			{
				Code: `<button role="button"></button>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- Paired form with whitespace-only children. ----
			{
				Code: `<button role="button">   </button>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- Paired form with newline children. ----
			{
				Code: `<button role="button">
</button>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- NoSubstitutionTemplateLiteral as role value — upstream's
			//      TemplateLiteral extractor returns the raw text WITHOUT
			//      boolean coercion, so `` role={`button`} `` extracts to the
			//      string "button" → match. tsgo's literalPropValue for
			//      KindNoSubstitutionTemplateLiteral returns jvString "button"
			//      directly. ----
			{
				Code: "<button role={`button`} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			{
				Code: "<h1 role={`heading`} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("h1", "heading"),
				},
			},
			// ---- NoSubstitutionTemplateLiteral with case mismatch — still
			//      lowercased before rolesMap.has → matches. ----
			{
				Code: "<button role={`BUTTON`} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},

			// ============================================================
			// Comment / whitespace interference around the trigger.
			// ============================================================

			// ---- Block comment between opening tag and the role attr. ----
			{
				Code: `<button /* explain */ role="button" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- Comment inside the JsxExpression. ----
			{
				Code: `<button role={/* lock */ "button"} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- Multi-line opening tag with comments. ----
			{
				Code: `<button
  /* before */ role="button" /* after */
/>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noRedundantRoles", Message: errorMessage("button", "button"), Line: 1, Column: 1},
				},
			},

			// ============================================================
			// Spread literal with role mapped to redundant value
			// ============================================================

			// ---- Literal spread carrying the role attribute — walks like
			//      a direct JsxAttribute. ----
			{
				Code: `<button {...{role: "button"}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},
			// ---- Multiple spreads — non-literal spread is opaque; literal
			//      spread still matches. Direct JsxAttribute wins over spread
			//      if both present (upstream's `getProp` returns the first
			//      match in attrs order). ----
			{
				Code: `<button {...props} {...{role: "button"}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErr("button", "button"),
				},
			},

			// ============================================================
			// select size numeric edge cases
			// ============================================================

			// ---- Numeric separator `1_000` (ES2021) — normalized to 1000 → listbox. ----
			{
				Code:   `<select size={1_000} role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- Scientific notation `1e2` = 100 → listbox. ----
			{
				Code:   `<select size={1e2} role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},
			// ---- Hex number literal `0x10` = 16 → listbox. ----
			{
				Code:   `<select size={0x10} role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")},
			},

			// ============================================================
			// Options edge cases
			// ============================================================

			// ---- Empty allow-list explicitly set to `[]` — disables ALL
			//      default exceptions for that key. nav must report. ----
			{
				Code:    `<nav role="navigation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"nav": []interface{}{}},
				Errors:  []rule_tester.InvalidTestCaseError{invalidErr("nav", "navigation")},
			},
			// ---- Allow-list with non-matching role — `["main"]` doesn't
			//      include "navigation", so nav reports. ----
			{
				Code:    `<nav role="navigation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"nav": []interface{}{"main"}},
				Errors:  []rule_tester.InvalidTestCaseError{invalidErr("nav", "navigation")},
			},
			// ---- Allow-list value is non-array (malformed) — StringSliceOption
			//      drops it to nil. hasOwn=true → allowed=nil → no matches →
			//      REPORT. Locks the defensive coercion. ----
			{
				Code:    `<nav role="navigation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"nav": "not-an-array"},
				Errors:  []rule_tester.InvalidTestCaseError{invalidErr("nav", "navigation")},
			},
			// ---- Allow-list keyed differently — `{ button: [...] }` covers
			//      `button` but not `nav`, so nav still gets default. ----
			{
				Code:    `<button role="button" />`,
				Tsx:     true,
				Options: map[string]interface{}{"button": []interface{}{"main"}}, // doesn't allow button
				Errors:  []rule_tester.InvalidTestCaseError{invalidErr("button", "button")},
			},

			// ============================================================
			// Multi-element a11y misuse patterns
			// ============================================================

			// ---- A common misuse: form structure with redundant roles. Both
			//      inner elements are `<button>` (not `<input>`), each with
			//      redundant role="button". ----
			{
				Code: `<form role="form">
  <button type="button" role="button">Submit</button>
  <button type="submit" role="button">Save</button>
</form>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrAt("form", "form", 1, 1),
					invalidErrLine("button", "button", 2),
					invalidErrLine("button", "button", 3),
				},
			},
			// ---- Heading hierarchy with redundant roles — three same-line
			//      siblings inside a Fragment, listener fires three times. ----
			{
				Code: `<><h1 role="heading">A</h1><h2 role="heading">B</h2><h3 role="heading">C</h3></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("h1", "heading", 1),
					invalidErrLine("h2", "heading", 1),
					invalidErrLine("h3", "heading", 1),
				},
			},
			// ---- A list of links — anchor href→link redundant; two anchors
			//      on the same line. ----
			{
				Code: `<ul><li><a href="/a" role="link">A</a></li><li><a href="/b" role="link">B</a></li></ul>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidErrLine("a", "link", 1),
					invalidErrLine("a", "link", 1),
				},
			},
		},
	)
}
