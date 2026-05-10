package iframe_has_title

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors upstream's settings used to remap `FooComponent`
// to the canonical `iframe` element name. Both upstream's valid and invalid
// suites use this exact settings shape — keep it as a single source so
// updates flow.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"FooComponent": "iframe",
		},
	},
}

// TestIframeHasTitleUpstream covers the full valid/invalid suite migrated 1:1
// from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/iframe-has-title-test.js`. The 5 valid + 11 invalid
// cases here mirror the upstream file byte-for-byte — do not add new cases
// to this file. rslint-specific lock-ins (semantic-walk branches, Dimension
// 4 universal edge shapes, tsgo AST quirks, real-world production patterns)
// belong in `iframe_has_title_extras_test.go` so the upstream-parity surface
// stays trivially comparable against future upstream updates.
func TestIframeHasTitleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IframeHasTitleRule, []rule_tester.ValidTestCase{
		// ---- Non-iframe element — listener gate skips ----
		{Code: `<div />;`, Tsx: true},

		// ---- iframe with truthy STRING title values ----
		{Code: `<iframe title="Unique title" />`, Tsx: true},

		// ---- Identifier title — upstream's getPropValue extractor returns
		//      the bare identifier name as a string ("foo") → typeof string
		//      → no report. ----
		{Code: `<iframe title={foo} />`, Tsx: true},

		// ---- Custom component "FooComponent" — type "FooComponent" is
		//      truthy AND not "iframe", so the listener short-circuits.
		//      No componentMap remap is configured here; the bare custom
		//      tag is left unchecked. ----
		{Code: `<FooComponent />`, Tsx: true},

		// ---- componentMap remap with truthy string title. ----
		{
			Code:     `<FooComponent title="Unique title" />`,
			Tsx:      true,
			Settings: componentsSettings,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Bare iframe with no title ----
		{
			Code: `<iframe />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Spread-only attributes — upstream's `getProp` walks LITERAL
		//      object spreads, but a non-literal `{...props}` is opaque:
		//      no title prop is found → falsy → REPORT. ----
		{
			Code: `<iframe {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Explicit `title={undefined}` — getPropValue returns the
		//      actual undefined value (Identifier `undefined` is the only
		//      special-cased identifier in jsx-ast-utils' extractor). ----
		{
			Code: `<iframe title={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty string title → falsy. ----
		{
			Code: `<iframe title="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- title={false} → boolean, not string → REPORT. Locks in the
		//      typeof-string gate that distinguishes iframe-has-title from
		//      html-has-lang's truthy-only check. ----
		{
			Code: `<iframe title={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- title={true} → boolean true is truthy BUT typeof "boolean"
		//      → REPORT. The defining iframe-has-title case: truthy alone
		//      isn't enough. ----
		{
			Code: `<iframe title={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty single-quoted string in JsxExpression. ----
		{
			Code: `<iframe title={''} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty no-substitution template literal → empty string → REPORT. ----
		{
			Code: "<iframe title={``} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty double-quoted string in JsxExpression. ----
		{
			Code: `<iframe title={""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- title={42} → number, typeof "number" → REPORT. Locks in the
		//      typeof-string gate for non-zero numerics (truthy but not
		//      string). ----
		{
			Code: `<iframe title={42} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- componentMap remap without title. ----
		{
			Code:     `<FooComponent />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
