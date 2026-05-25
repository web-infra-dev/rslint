package lang

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors upstream's `componentsSettings` used by the lang
// test file — a `Foo` componentMap remap to `html`, with `polymorphicPropName`
// set to `as`. Both upstream's valid and invalid suites reuse this shape; keep
// it as a single source so future updates flow.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"components": map[string]interface{}{
			"Foo": "html",
		},
	},
}

// TestLangUpstream covers the full valid/invalid suite migrated 1:1 from
// upstream eslint-plugin-jsx-a11y's `__tests__/src/rules/lang-test.js`. The
// 14 valid + 5 invalid cases here mirror the upstream file byte-for-byte —
// do not add new cases to this file. rslint-specific lock-ins (semantic-walk
// branches, Dimension 4 universal edge shapes, tsgo AST quirks, BCP-47
// validation edge cases) belong in `lang_extras_test.go` so the upstream-
// parity surface stays trivially comparable against future upstream updates.
func TestLangUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LangRule, []rule_tester.ValidTestCase{
		// ---- Non-lang attributes / non-html parents: listener gate skips ----
		{Code: `<div />;`, Tsx: true},
		{Code: `<div foo="bar" />;`, Tsx: true},
		{Code: `<div lang="foo" />;`, Tsx: true},

		// ---- html with valid BCP-47 tags ----
		{Code: `<html lang="en" />`, Tsx: true},
		{Code: `<html lang="en-US" />`, Tsx: true},
		{Code: `<html lang="zh-Hans" />`, Tsx: true},
		{Code: `<html lang="zh-Hant-HK" />`, Tsx: true},
		{Code: `<html lang="zh-yue-Hant" />`, Tsx: true},
		{Code: `<html lang="ja-Latn" />`, Tsx: true},

		// ---- Identifier lang value — upstream getLiteralPropValue returns
		//      null for non-undefined Identifier → skip. ----
		{Code: `<html lang={foo} />`, Tsx: true},

		// ---- Capitalized custom component "HTML" — type "HTML" is truthy
		//      AND not "html", so the listener short-circuits. No remap
		//      configured here; the bare custom tag is left unchecked. ----
		{Code: `<HTML lang="foo" />`, Tsx: true},

		// ---- Custom component Foo without componentMap — type "Foo" != "html"
		//      → listener short-circuits, undefined lang doesn't trip. ----
		{Code: `<Foo lang={undefined} />`, Tsx: true},

		// ---- componentMap remap `Foo: 'html'` + valid lang ----
		{
			Code:     `<Foo lang="en" />`,
			Tsx:      true,
			Settings: componentsSettings,
		},

		// ---- polymorphicPropName "as" — `<Box as="html" lang="en" />`
		//      resolves to "html", lang is valid → no report. ----
		{
			Code:     `<Box as="html" lang="en"  />`,
			Tsx:      true,
			Settings: componentsSettings,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- html with invalid BCP-47 tags ----
		{
			Code: `<html lang="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="zz-LL" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Explicit `lang={undefined}` — upstream `value === undefined`
		//      → REPORT. ----
		{
			Code: `<html lang={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- componentMap remap `Foo: 'html'` + explicit undefined. ----
		{
			Code:     `<Foo lang={undefined} />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 6},
			},
		},

		// ---- polymorphicPropName remap `<Box as="html">` + invalid BCP-47. ----
		{
			Code:     `<Box as="html" lang="foo" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 16},
			},
		},
	})
}
