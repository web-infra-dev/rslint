package tabindex_no_positive

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError is the single error shape every invalid case produces.
// Centralized so a future error-text tweak touches one place. Shared with
// tabindex_no_positive_extras_test.go via package scope.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "tabIndexNoPositive",
	Message:   errorMessage,
}

// TestTabindexNoPositiveUpstream mirrors the full valid/invalid suite from
// upstream's __tests__/src/rules/tabindex-no-positive-test.js, 1:1 and in
// upstream order so a future audit can grep across both side-by-side.
//
// Anything NOT in upstream's test file — full ToNumber-coercion survey,
// case-variant matches, options matrix (empty), spread / TS-wrapper / non-
// DOM-element coverage, real-world misuse patterns — lives in
// tabindex_no_positive_extras_test.go.
func TestTabindexNoPositiveUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &TabindexNoPositiveRule, []rule_tester.ValidTestCase{
		// No tabIndex prop — listener early-returns on the `propName` filter.
		{Code: `<div />;`, Tsx: true},
		// JsxSpreadAttribute — listener fires on KindJsxAttribute only, so
		// spread shapes are never visited.
		{Code: `<div {...props} />`, Tsx: true},
		// Non-tabIndex prop.
		{Code: `<div id="main" />`, Tsx: true},
		// `tabIndex={undefined}` — JSXExpressionContainer.Identifier "undefined"
		// → getLiteralPropValue → undefined → Number(undefined) = NaN → skip.
		{Code: `<div tabIndex={undefined} />`, Tsx: true},
		// Template `${undefined}` — extractValue's TemplateLiteral arm renders
		// the Identifier-undefined substitution as the bare word "undefined",
		// joined to ""; Number("undefined") = NaN → skip.
		{Code: "<div tabIndex={`${undefined}`} />", Tsx: true},
		// Two consecutive `${undefined}` substitutions concatenate to
		// "undefinedundefined" → NaN → skip.
		{Code: "<div tabIndex={`${undefined}${undefined}`} />", Tsx: true},
		// tabIndex = 0 — Number(0) = 0, `0 <= 0` short-circuits.
		{Code: `<div tabIndex={0} />`, Tsx: true},
		// tabIndex = -1 — Number(-1) = -1, `-1 <= 0` short-circuits.
		{Code: `<div tabIndex={-1} />`, Tsx: true},
		// `tabIndex={null}` — getLiteralPropValue special-cases null literal to
		// the string "null"; Number("null") = NaN → skip. (Distinct from
		// non-undefined Identifier null which would be Number(null) = 0 →
		// `0 <= 0` skip — same observable result.)
		{Code: `<div tabIndex={null} />`, Tsx: true},
		// CallExpression — LITERAL_TYPES.CallExpression is noop → null →
		// Number(null) = 0 → skip.
		{Code: `<div tabIndex={bar()} />`, Tsx: true},
		// Identifier (non-undefined) — LITERAL_TYPES.Identifier = () => null →
		// Number(null) = 0 → skip.
		{Code: `<div tabIndex={bar} />`, Tsx: true},
		// Non-numeric string literal inside JsxExpressionContainer — passes
		// through StringLiteral extractor (no "true"/"false" coercion since
		// "foobar" isn't true/false); Number("foobar") = NaN → skip.
		{Code: `<div tabIndex={"foobar"} />`, Tsx: true},
		// String "0" — Number("0") = 0 → `0 <= 0` skip.
		{Code: `<div tabIndex="0" />`, Tsx: true},
		// String "-1" — Number("-1") = -1 → skip.
		{Code: `<div tabIndex="-1" />`, Tsx: true},
		// String "-5" — same.
		{Code: `<div tabIndex="-5" />`, Tsx: true},
		// String "-5.5" — Number trims/parses as float; -5.5 <= 0 → skip.
		// (Note: this differs from no-noninteractive-tabindex's getTabIndex
		// which requires integers — tabindex-no-positive does NOT, so the
		// invalid pair `{1.589}` below DOES report.)
		{Code: `<div tabIndex="-5.5" />`, Tsx: true},
		// Numeric -5.5 → -5.5 <= 0 → skip.
		{Code: `<div tabIndex={-5.5} />`, Tsx: true},
		// Numeric -5 → skip.
		{Code: `<div tabIndex={-5} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// String "1" → Number("1") = 1 → report.
		{Code: `<div tabIndex="1" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Numeric 1 → report.
		{Code: `<div tabIndex={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// String "1" inside JsxExpressionContainer — StringLiteral extractor
		// returns "1" (no true/false coercion); Number("1") = 1 → report.
		{Code: `<div tabIndex={"1"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// NoSubstitutionTemplateLiteral `\`1\`` — extractValue's TemplateLiteral
		// arm returns the raw quasi "1"; Number("1") = 1 → report.
		// IMPORTANT: NoSubstitutionTemplateLiteral does NOT go through the
		// `"true"/"false" → boolean` coercion that StringLiteral does, so
		// `\`false\`` would NOT route through this path's report; see
		// extras tests for that boundary.
		{Code: "<div tabIndex={`1`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Non-integer numeric — Number(1.589) = 1.589 > 0 → report. This is
		// the key divergence from no-noninteractive-tabindex's getTabIndex
		// which would reject non-integers in step 1.
		{Code: `<div tabIndex={1.589} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
