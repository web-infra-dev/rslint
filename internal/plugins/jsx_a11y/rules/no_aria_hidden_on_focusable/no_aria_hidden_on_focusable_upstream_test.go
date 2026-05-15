package no_aria_hidden_on_focusable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's `expectedError` object — every invalid
// case carries the same single diagnostic. Shared with the extras test.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noAriaHiddenOnFocusable",
	Message:   errorMessage,
}

// TestNoAriaHiddenOnFocusableUpstream migrates the full upstream suite from
// eslint-plugin-jsx-a11y's `__tests__/src/rules/no-aria-hidden-on-focusable-test.js`
// 1:1, preserving case order so a future audit can grep across both
// side-by-side. rslint-specific lock-ins (value-extraction branches, TS
// wrappers, position assertions, etc.) live in
// no_aria_hidden_on_focusable_extras_test.go.
func TestNoAriaHiddenOnFocusableUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAriaHiddenOnFocusableRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid ----
		// div is not interactive; no tabIndex → tabIndex >= 0 is false → not focusable.
		{Code: `<div aria-hidden="true" />;`, Tsx: true},
		// div with onClick — onClick alone does NOT make it interactive per
		// upstream's `isInteractiveElement` (which looks at tagName + aria-query
		// schemas, not event handlers). Still not focusable.
		{Code: `<div onClick={() => void 0} aria-hidden="true" />;`, Tsx: true},
		// img is in the dom set but does not match any interactive role schema
		// → isInteractiveElement returns false → tabIndex >= 0 is false → safe.
		{Code: `<img aria-hidden="true" />`, Tsx: true},
		// aria-hidden="false" → upstream `getPropValue(...) === true` is false,
		// rule short-circuits before checking focus.
		{Code: `<a aria-hidden="false" href="#" />`, Tsx: true},
		// button is interactive; tabIndex="-1" → tabIndex < 0 → !focusable.
		{Code: `<button aria-hidden="true" tabIndex="-1" />`, Tsx: true},
		// No aria-hidden at all — the IsAriaHiddenTrue gate fails immediately.
		{Code: `<button />`, Tsx: true},
		{Code: `<a href="/" />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid ----
		// Non-interactive div made focusable by tabIndex="0" — tabIndex >= 0 → focusable.
		{Code: `<div aria-hidden="true" tabIndex="0" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// input is interactive; missing tabIndex → undefined === undefined → focusable.
		{Code: `<input aria-hidden="true" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<a>` with `href` matches the interactive-anchor schema → interactive
		// → no tabIndex → focusable.
		{Code: `<a href="/" aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// button: inherently interactive, no tabIndex → focusable.
		{Code: `<button aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// textarea: same — interactive, no tabIndex.
		{Code: `<textarea aria-hidden="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Lowercase `tabindex` — jsx-ast-utils' `getProp` is case-insensitive
		// by default, so this still resolves to tabIndex=0. p is non-interactive,
		// `0 >= 0` → focusable.
		{Code: `<p tabindex="0" aria-hidden="true">text</p>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
