package no_autofocus

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's expected error shape — every invalid case
// emits the same single message. Centralized so a future text tweak lives in
// one place. Shared with no_autofocus_extras_test.go.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noAutoFocus",
	Message:   errorMessage,
}

// ignoreNonDOMOption mirrors upstream's `[{ ignoreNonDOM: true }]` schema —
// passed via the JSON path (not a typed struct) so the GetOptionsMap
// integration is exercised end-to-end. Shared with extras tests.
var ignoreNonDOMOption = []interface{}{
	map[string]interface{}{
		"ignoreNonDOM": true,
	},
}

// componentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Button: 'button' } } }
//
// Used to verify GetElementType honors the components map for
// `ignoreNonDOM` resolution.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Button": "button",
		},
	},
}

// TestNoAutofocusUpstream covers the full valid/invalid suite migrated 1:1
// from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/no-autofocus-test.js`. Order and grouping mirror the
// upstream file so a future audit can grep across both side-by-side.
//
// rslint-specific lock-ins (TS wrappers, spread literals, listener boundary
// repeats, position assertions, the upstream getPropValue branches that
// upstream never tests directly) live in no_autofocus_extras_test.go.
func TestNoAutofocusUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAutofocusRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid ----
		{Code: `<div />;`, Tsx: true},
		// Lowercase `autofocus` is the HTML DOM attribute, not the React prop;
		// the rule only matches the camelCased `autoFocus`.
		{Code: `<div autofocus />;`, Tsx: true},
		{Code: `<input autofocus="true" />;`, Tsx: true},
		// Component without the autoFocus prop — nothing to inspect.
		{Code: `<Foo bar />`, Tsx: true},
		// Explicit boolean false → !== false fails → no report.
		{Code: `<div autoFocus={false} />`, Tsx: true},
		// String "false" coerces to boolean false via jsxAstUtilsLiteralCoerce
		// → !== false fails → no report.
		{Code: `<div autoFocus="false" />`, Tsx: true},
		// ignoreNonDOM: true — `Foo` is not in the dom set, so the rule skips
		// even though autoFocus is truthy.
		{Code: `<Foo autoFocus />`, Tsx: true, Options: ignoreNonDOMOption},
		// ignoreNonDOM: true on a nested div with the lowercase HTML attribute
		// — the inner div has `autofocus` (lowercase, not matched) so even
		// without ignoreNonDOM there is nothing to report; with it, the same
		// holds.
		{Code: `<div><div autofocus /></div>`, Tsx: true, Options: ignoreNonDOMOption},
		// components map promotes `Button` → `button`, but no autoFocus is
		// declared so nothing to report.
		{Code: `<Button />`, Tsx: true, Settings: componentsSettings},
		// Same shape, with ignoreNonDOM: true (still no autoFocus).
		{Code: `<Button />`, Tsx: true, Options: ignoreNonDOMOption, Settings: componentsSettings},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid ----
		// Boolean attribute form → extractValue null-attr-value → JS true.
		{Code: `<div autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Explicit boolean true.
		{Code: `<div autoFocus={true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `undefined` !== false AND !== "false" → reports. Locks in upstream's
		// "anything other than literal false" semantic — even an explicit
		// undefined trips the rule.
		{Code: `<div autoFocus={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String "true" → coerces to boolean true via jsxAstUtilsLiteralCoerce
		// → !== false → reports.
		{Code: `<div autoFocus="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Same boolean-form on a different intrinsic element — listener fires
		// regardless of element kind when ignoreNonDOM is unset.
		{Code: `<input autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Custom component — without ignoreNonDOM, the rule fires on all
		// elements equally.
		{Code: `<Foo autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// components map promotes `Button` → `button`, then the rule reports.
		{Code: `<Button autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}, Settings: componentsSettings},
		// Same shape with ignoreNonDOM: true — the resolved name is `button`
		// (in the dom set), so the skip does NOT apply and the rule still
		// reports. Locks in the components-map → ignoreNonDOM interaction.
		{Code: `<Button autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
