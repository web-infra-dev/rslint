package no_redundant_roles

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors upstream's `componentsSettings`, where
// `<Button>` is remapped to the HTML `button` element. Drives the
// `<Button role="button" />` invalid case and the
// `<Button role={`${foo}button`} />` valid case.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Button": "button",
		},
	},
}

// invalidErr matches upstream's `expectedError(element, implicitRole)` —
// the error message text is the rule's deterministic output, and the
// rule reports on the JSXOpeningElement (column 1 for top-level JSX
// expressions in our tests).
func invalidErr(element, implicitRole string) rule_tester.InvalidTestCaseError {
	return invalidErrAt(element, implicitRole, 1, 1)
}

// invalidErrAt is the explicit-position variant of [invalidErr] —
// required for cases where the JsxOpeningElement isn't at line 1,
// column 1 (nested JSX, multi-line tests, declarations preceded by a
// `function`/`const`/`class` prefix).
func invalidErrAt(element, implicitRole string, line, col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noRedundantRoles",
		Message:   errorMessage(element, implicitRole),
		Line:      line,
		Column:    col,
	}
}

// invalidErrLine asserts only the diagnostic's line (not the column).
// Useful for declaration-wrapped JSX (`function`, `class`, `const`,
// `async`, `function*`, etc.) where the JsxOpeningElement's column
// depends on the surrounding prefix. The exact column-1 surface is
// already locked in by [TestNoRedundantRolesUpstream*] and
// [TestNoRedundantRolesExtras]; this helper keeps the real-world
// pattern suite focused on logic coverage without re-computing offsets.
func invalidErrLine(element, implicitRole string, line int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noRedundantRoles",
		Message:   errorMessage(element, implicitRole),
		Line:      line,
	}
}

// TestNoRedundantRolesUpstreamRecommendedDefault mirrors upstream's first
// `ruleTester.run(..., 'recommended', rule, {valid: [alwaysValid, nav], invalid: neverValid})`
// suite — the default-options surface (no rule-options override). The
// default `nav: ['navigation']` exception is the lone behavioral
// difference from the "noNavExceptions" suite below.
func TestNoRedundantRolesUpstreamRecommendedDefault(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantRolesRule,
		[]rule_tester.ValidTestCase{
			// ---- alwaysValid ----
			{Code: `<div />;`, Tsx: true},
			{Code: `<button role="main" />`, Tsx: true},
			{Code: `<MyComponent role="button" />`, Tsx: true},
			{Code: "<button role={`${foo}button`} />", Tsx: true},
			{Code: "<Button role={`${foo}button`} />", Tsx: true, Settings: componentsSettings},
			{Code: `<select role="menu"><option>1</option><option>2</option></select>`, Tsx: true},
			{Code: `<select role="menu" size={2}><option>1</option><option>2</option></select>`, Tsx: true},
			{Code: `<select role="menu" multiple><option>1</option><option>2</option></select>`, Tsx: true},
			// ---- default-exception: nav with the implicit `navigation` role
			//      is the W3C-recommended pattern; the default options
			//      table `DEFAULT_ROLE_EXCEPTIONS = { nav: ['navigation'] }`
			//      lets this pass. ----
			{Code: `<nav role="navigation" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- neverValid ----
			// body — DOCUMENT (any case) matches implicit 'document'.
			{Code: `<body role="DOCUMENT" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("body", "document")}},
			// button — implicit 'button'.
			{Code: `<button role="button" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")}},
			// componentsMap: <Button> → 'button' via settings.
			{Code: `<Button role="button" />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")}},
			// select default (no size, no multiple) — implicit 'combobox'.
			{Code: `<select role="combobox"><option>1</option><option>2</option></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size="" — "" coerces to 0, 0 > 1 false → combobox.
			{Code: `<select role="combobox" size="" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size={1} — 1 > 1 false → combobox.
			{Code: `<select role="combobox" size={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size="1" — "1" coerces to 1, 1 > 1 false → combobox.
			{Code: `<select role="combobox" size="1" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size={null} — upstream `getLiteralPropValue` maps
			// the `null` literal to the magic string "null"; parseFloat("null")
			// → NaN; NaN > 1 → false → combobox.
			{Code: `<select role="combobox" size={null}></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size={undefined} — upstream returns undefined → NaN > 1 → false → combobox.
			{Code: `<select role="combobox" size={undefined}></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with multiple={undefined} — falsy → falls through to size (absent) → combobox.
			{Code: `<select role="combobox" multiple={undefined}></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with multiple={false} — falsy → falls through → combobox.
			{Code: `<select role="combobox" multiple={false}></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with multiple="" — empty string coerces falsy → combobox.
			{Code: `<select role="combobox" multiple=""></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			// select with size="3" — > 1 → listbox.
			{Code: `<select role="listbox" size="3" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			// select with size={2} — > 1 → listbox.
			{Code: `<select role="listbox" size={2} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			// select with multiple (boolean form) — listbox.
			{Code: `<select role="listbox" multiple><option>1</option><option>2</option></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			// select with multiple={true} — listbox.
			{Code: `<select role="listbox" multiple={true}></select>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
		},
	)
}

// noNavExceptionsOptions disables the default `nav: ['navigation']`
// exception. With an explicit empty entry under `nav`, the
// `hasOwn(allowedRedundantRoles, type)` branch returns true and uses
// the empty list, suppressing the default fallback.
var noNavExceptionsOptions = map[string]interface{}{
	"nav": []interface{}{},
}

// TestNoRedundantRolesUpstreamNoNavExceptions mirrors upstream's second
// suite — `{ nav: [] }` options disables the default nav allowance, so
// `<nav role="navigation" />` becomes a violation. All the alwaysValid
// cases remain valid (the options object is keyed by element, so
// non-nav elements ignore it).
func TestNoRedundantRolesUpstreamNoNavExceptions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantRolesRule,
		[]rule_tester.ValidTestCase{
			// alwaysValid — options unchanged.
			{Code: `<div />;`, Tsx: true, Options: noNavExceptionsOptions},
			{Code: `<button role="main" />`, Tsx: true, Options: noNavExceptionsOptions},
			{Code: `<MyComponent role="button" />`, Tsx: true, Options: noNavExceptionsOptions},
			{Code: "<button role={`${foo}button`} />", Tsx: true, Options: noNavExceptionsOptions},
			{Code: "<Button role={`${foo}button`} />", Tsx: true, Settings: componentsSettings, Options: noNavExceptionsOptions},
			{Code: `<select role="menu"><option>1</option><option>2</option></select>`, Tsx: true, Options: noNavExceptionsOptions},
			{Code: `<select role="menu" size={2}><option>1</option><option>2</option></select>`, Tsx: true, Options: noNavExceptionsOptions},
			{Code: `<select role="menu" multiple><option>1</option><option>2</option></select>`, Tsx: true, Options: noNavExceptionsOptions},
		},
		[]rule_tester.InvalidTestCase{
			// neverValid (same as default suite) — options change doesn't
			// affect non-nav elements.
			{Code: `<body role="DOCUMENT" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("body", "document")}},
			{Code: `<button role="button" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")}},
			{Code: `<Button role="button" />`, Tsx: true, Settings: componentsSettings, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("button", "button")}},
			{Code: `<select role="combobox"><option>1</option><option>2</option></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" size="" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" size={1} />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" size="1" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" size={null}></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" size={undefined}></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" multiple={undefined}></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" multiple={false}></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="combobox" multiple=""></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "combobox")}},
			{Code: `<select role="listbox" size="3" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			{Code: `<select role="listbox" size={2} />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			{Code: `<select role="listbox" multiple><option>1</option><option>2</option></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			{Code: `<select role="listbox" multiple={true}></select>`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("select", "listbox")}},
			// With nav exception disabled, <nav role="navigation" /> reports.
			{Code: `<nav role="navigation" />`, Tsx: true, Options: noNavExceptionsOptions, Errors: []rule_tester.InvalidTestCaseError{invalidErr("nav", "navigation")}},
		},
	)
}

// listExceptionOptions adds list-role allowances for `ul` and `ol`.
// Replaces (does not augment) the default exceptions table for those
// keys — `nav` is unaffected and retains its default `['navigation']`
// allowance only if the table also includes a `nav` entry. Upstream's
// test does NOT include `nav` in the override, so `<nav role="navigation" />`
// is NOT in the valid/invalid mix here.
var listExceptionOptions = map[string]interface{}{
	"ul": []interface{}{"list"},
	"ol": []interface{}{"list"},
}

// TestNoRedundantRolesUpstreamListOverride mirrors upstream's third
// suite — `{ ul: ['list'], ol: ['list'] }` allows `role="list"` on
// `<ul>` / `<ol>`, but `<img role="img" />` still reports because the
// options table doesn't override `img`'s default (which is the absence
// of any allowance — img is reported as redundant when `role="img"` is
// set alongside no `alt=""`/no `.svg` src).
//
// `<dl role="list" />` is valid because `dl` doesn't have an implicit
// role in upstream's implicitRoles table → getImplicitRole returns
// null → no comparison.
//
// `<img src="example.svg" role="img" />` is valid because the SVG src
// arm of `implicitRoleForImg` returns '' → null → no comparison.
//
// `<svg role="img" />` is valid because `svg` isn't in the implicitRoles
// table at all → null → no comparison.
//
// `<img src={someVariable} role="img" />` is INVALID because the SVG src
// arm requires a LITERAL src whose value includes ".svg"; a non-literal
// `src={someVariable}` extracts to null, the optional-chain
// `null?.includes('.svg')` short-circuits, and implicit role stays 'img'.
func TestNoRedundantRolesUpstreamListOverride(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantRolesRule,
		[]rule_tester.ValidTestCase{
			{Code: `<ul role="list" />`, Tsx: true, Options: listExceptionOptions},
			{Code: `<ol role="list" />`, Tsx: true, Options: listExceptionOptions},
			// dl has no implicit role in upstream → no comparison happens.
			// Note that the option is irrelevant for this case; included as in
			// the upstream test for parity.
			{Code: `<dl role="list" />`, Tsx: true, Options: listExceptionOptions},
			// img with literal .svg src → SVG arm returns '' → no implicit
			// role → no comparison.
			{Code: `<img src="example.svg" role="img" />`, Tsx: true, Options: listExceptionOptions},
			// svg is not in upstream's implicitRoles → no comparison.
			{Code: `<svg role="img" />`, Tsx: true, Options: listExceptionOptions},
		},
		[]rule_tester.InvalidTestCase{
			// ul without the list-override would be valid, but upstream's
			// 3rd invalid suite passes NO options-mapper — so these cases
			// re-test with no allowance and confirm the report. Options
			// field is omitted to match upstream's parsersMap behavior.
			{Code: `<ul role="list" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("ul", "list")}},
			{Code: `<ol role="list" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("ol", "list")}},
			// img with no alt and no src → implicit 'img', matches → REPORT.
			{Code: `<img role="img" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("img", "img")}},
			// img with non-literal src — SVG arm doesn't trigger, implicit
			// stays 'img', matches → REPORT.
			{Code: `<img src={someVariable} role="img" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidErr("img", "img")}},
		},
	)
}
