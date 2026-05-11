package aria_activedescendant_has_tabindex

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError is the single error shape every invalid case produces.
// Centralized so a future error-text tweak touches one place. Shared with
// aria_activedescendant_has_tabindex_extras_test.go via package scope.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "tabIndexRequired",
	Message:   errorMessage,
	Line:      1,
	Column:    1,
}

// componentsCustomDiv mirrors the upstream test setting
// `{ 'jsx-a11y': { components: { CustomComponent: 'div' } } }` used by two
// upstream cases (one valid, one invalid) to verify that `getElementType`
// honors the components map before the dom.has() / interactive checks run.
var componentsCustomDiv = map[string]any{
	"jsx-a11y": map[string]any{
		"components": map[string]any{
			"CustomComponent": "div",
		},
	},
}

// TestAriaActivedescendantHasTabindexUpstream mirrors the full valid/invalid
// suite from upstream's __tests__/src/rules/aria-activedescendant-has-tabindex-test.js,
// 1:1 and in upstream order so a future audit can grep across both
// side-by-side.
//
// Anything NOT in upstream's test file — case-variant prop matching, literal
// spread, boolean-form / explicit-undefined value, paired element form,
// gate-3 boundary cases, polymorphicPropName, NaN-coercing tabIndex strings —
// lives in aria_activedescendant_has_tabindex_extras_test.go.
func TestAriaActivedescendantHasTabindexUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaActivedescendantHasTabindexRule, []rule_tester.ValidTestCase{
		// CustomComponent — gate-2 (`!dom.has('CustomComponent')`) skips.
		{Code: `<CustomComponent />;`, Tsx: true},
		// CustomComponent + aria-activedescendant — same gate-2 skip; the
		// activedescendant prop's presence is irrelevant for non-DOM tags.
		{Code: `<CustomComponent aria-activedescendant={someID} />;`, Tsx: true},
		// CustomComponent + aria-activedescendant + tabIndex={0} — gate-2 skip
		// (tabIndex value never inspected for non-DOM tags).
		{Code: `<CustomComponent aria-activedescendant={someID} tabIndex={0} />;`, Tsx: true},
		// CustomComponent + aria-activedescendant + tabIndex={-1} — gate-2 skip.
		{Code: `<CustomComponent aria-activedescendant={someID} tabIndex={-1} />;`, Tsx: true},
		// CustomComponent mapped to "div" via settings, with tabIndex={0} —
		// gate-2 passes (resolved to div which IS in dom map); gate-3 fails
		// (div is NOT interactive); gate-4 0 >= -1 → skip.
		{Code: `<CustomComponent aria-activedescendant={someID} tabIndex={0} />;`, Tsx: true, Settings: componentsCustomDiv},
		// Bare div / input — no aria-activedescendant → gate-1 skip.
		{Code: `<div />;`, Tsx: true},
		{Code: `<input />;`, Tsx: true},
		// div with tabIndex but no aria-activedescendant — gate-1 skip.
		{Code: `<div tabIndex={0} />;`, Tsx: true},
		// div + activedescendant + tabIndex={0} — gate-4 skip (0 >= -1).
		{Code: `<div aria-activedescendant={someID} tabIndex={0} />;`, Tsx: true},
		// div + activedescendant + tabIndex="0" — string-numeric "0" coerces
		// to 0 via GetTabIndex's StringToNumber path; gate-4 skip.
		{Code: `<div aria-activedescendant={someID} tabIndex="0" />;`, Tsx: true},
		// div + activedescendant + tabIndex={1} — gate-4 skip (1 >= -1).
		{Code: `<div aria-activedescendant={someID} tabIndex={1} />;`, Tsx: true},
		// input + activedescendant, no tabIndex — gate-3 skip (input is
		// inherently interactive AND tabIndex is undefined).
		{Code: `<input aria-activedescendant={someID} />;`, Tsx: true},
		// input + activedescendant + tabIndex={1} — gate-3 fails (tabIndex
		// resolves, not undefined); gate-4 1 >= -1 → skip.
		{Code: `<input aria-activedescendant={someID} tabIndex={1} />;`, Tsx: true},
		// input + activedescendant + tabIndex={0} — same gate-3/4 path.
		{Code: `<input aria-activedescendant={someID} tabIndex={0} />;`, Tsx: true},
		// input + activedescendant + tabIndex={-1} — gate-4 -1 >= -1 → skip.
		{Code: `<input aria-activedescendant={someID} tabIndex={-1} />;`, Tsx: true},
		// div + activedescendant + tabIndex={-1} — gate-4 skip (-1 >= -1).
		{Code: `<div aria-activedescendant={someID} tabIndex={-1} />;`, Tsx: true},
		// div + activedescendant + tabIndex="-1" — string "-1" coerces to -1
		// via GetTabIndex; gate-4 skip.
		{Code: `<div aria-activedescendant={someID} tabIndex="-1" />;`, Tsx: true},
		// Repeat of input + activedescendant + tabIndex={-1} — kept verbatim
		// because upstream's test file lists it twice; preserving the count
		// matches a future grep / line-count audit against upstream.
		{Code: `<input aria-activedescendant={someID} tabIndex={-1} />;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// div + activedescendant, no tabIndex — gate-3 fails (div is not
		// interactive); gate-4 fails (undefined >= -1 is false in JS).
		// Reports on the JsxSelfClosingElement node, column 1.
		{Code: `<div aria-activedescendant={someID} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// CustomComponent mapped to "div" via settings — same flow as above
		// after the gate-2 components-map resolution.
		{Code: `<CustomComponent aria-activedescendant={someID} />;`, Tsx: true, Settings: componentsCustomDiv, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
