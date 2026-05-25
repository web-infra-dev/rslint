package html_has_lang

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// htmlTopSettings mirrors upstream's settings used to remap `HTMLTop` to the
// canonical `html` element name. Both upstream's valid and invalid suites use
// this exact settings shape — keep it as a single source so updates flow.
var htmlTopSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"HTMLTop": "html",
		},
	},
}

// TestHtmlHasLangUpstream covers the full valid/invalid suite migrated 1:1
// from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/html-has-lang-test.js`. The 7 valid + 4 invalid cases
// here mirror the upstream file byte-for-byte — do not add new cases to this
// file. rslint-specific lock-ins (semantic-walk branches, Dimension 4
// universal edge shapes, tsgo AST quirks, real-world production patterns)
// belong in `html_has_lang_extras_test.go` so the upstream-parity surface
// stays trivially comparable against future upstream updates.
func TestHtmlHasLangUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HtmlHasLangRule, []rule_tester.ValidTestCase{
		// ---- Non-html element — listener gate skips ----
		{Code: `<div />;`, Tsx: true},

		// ---- html with truthy lang values ----
		{Code: `<html lang="en" />`, Tsx: true},
		{Code: `<html lang="en-US" />`, Tsx: true},
		{Code: `<html lang={foo} />`, Tsx: true},

		// ---- Boolean-attribute form (`<html lang />`) — getPropValue
		//      returns boolean true → truthy → no report ----
		{Code: `<html lang />`, Tsx: true},

		// ---- Capitalized custom component "HTML" — type "HTML" is
		//      truthy AND not "html", so the listener short-circuits.
		//      No componentMap remap is configured here; the bare custom
		//      tag is left unchecked. ----
		{Code: `<HTML />`, Tsx: true},

		// ---- componentMap remap with truthy lang. Upstream's test file
		//      has a stray `errors` field on this entry, but the entry
		//      sits in the `valid` array; RuleTester ignores `errors`
		//      for valid cases. We treat it as valid here. ----
		{
			Code:     `<HTMLTop lang="en" />`,
			Tsx:      true,
			Settings: htmlTopSettings,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Bare html with no lang ----
		{
			Code: `<html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Spread-only attributes — upstream's `getProp` walks LITERAL
		//      object spreads, but a non-literal `{...props}` is opaque:
		//      no lang prop is found → falsy → REPORT. ----
		{
			Code: `<html {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Explicit `lang={undefined}` — getPropValue returns the
		//      actual undefined value (Identifier `undefined` is the only
		//      special-cased identifier in jsx-ast-utils' extractor). ----
		{
			Code: `<html lang={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- componentMap remap without lang. ----
		{
			Code:     `<HTMLTop />`,
			Tsx:      true,
			Settings: htmlTopSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
