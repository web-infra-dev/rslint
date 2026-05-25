package heading_has_content

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsRuleOpts mirrors upstream's `components: ['Heading', 'Title']`
// rule-options test fixture. Names are appended to the typeCheck list AS-IS
// (not run through getElementType's componentMap).
var componentsRuleOpts = map[string]interface{}{
	"components": []interface{}{"Heading", "Title"},
}

// componentsSettings mirrors upstream's settings-based custom-component
// remapping. Note that `CustomInput → input` is here so the
// `<h1><CustomInput type="hidden" /></h1>` case can resolve CustomInput
// to "input" and trigger isHiddenFromScreenReader's `type=hidden` arm.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"CustomInput": "input",
			"Title":       "h1",
			"Heading":     "h2",
		},
	},
}

// TestHeadingHasContentUpstream covers the full valid/invalid suite migrated
// 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/heading-has-content-test.js`. rslint-specific lock-ins
// (semantic-walk branches, Dimension 4 universal edge shapes, tsgo AST quirks)
// live in heading_has_content_extras_test.go.
func TestHeadingHasContentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HeadingHasContentRule, []rule_tester.ValidTestCase{
		// ---- DEFAULT ELEMENT TESTS ----
		{Code: `<div />;`, Tsx: true},
		{Code: `<h1>Foo</h1>`, Tsx: true},
		{Code: `<h2>Foo</h2>`, Tsx: true},
		{Code: `<h3>Foo</h3>`, Tsx: true},
		{Code: `<h4>Foo</h4>`, Tsx: true},
		{Code: `<h5>Foo</h5>`, Tsx: true},
		{Code: `<h6>Foo</h6>`, Tsx: true},
		{Code: `<h6>123</h6>`, Tsx: true},
		{Code: `<h1><Bar /></h1>`, Tsx: true},
		{Code: `<h1>{foo}</h1>`, Tsx: true},
		{Code: `<h1>{foo.bar}</h1>`, Tsx: true},
		{Code: `<h1 dangerouslySetInnerHTML={{ __html: "foo" }} />`, Tsx: true},
		{Code: `<h1 children={children} />`, Tsx: true},

		// ---- CUSTOM ELEMENT TESTS FOR COMPONENTS OPTION ----
		{Code: `<Heading>Foo</Heading>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Title>Foo</Title>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Heading><Bar /></Heading>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Heading>{foo}</Heading>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Heading>{foo.bar}</Heading>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Heading dangerouslySetInnerHTML={{ __html: "foo" }} />`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Heading children={children} />`, Tsx: true, Options: componentsRuleOpts},

		// ---- aria-hidden on the heading itself → exempt (not announced) ----
		{Code: `<h1 aria-hidden />`, Tsx: true},

		// ---- CUSTOM ELEMENT TESTS FOR COMPONENTS SETTINGS ----
		{Code: `<Heading>Foo</Heading>`, Tsx: true, Settings: componentsSettings},
		// Native `<h1>` containing a `<CustomInput type="hidden" />` — the
		// CustomInput is remapped to "input", and `type="hidden"` makes it
		// hidden, so the inner element doesn't count as accessible.
		// However, the `componentsSettings` here mirrors upstream's exact
		// settings, where CustomInput remaps to "input". Upstream's test
		// list places this under "valid" because the entire test suite
		// uses settings: componentsSettings — but reading the upstream test
		// again reveals this case is in the VALID list WITHOUT settings:
		// componentsSettings, which means CustomInput stays as "CustomInput",
		// not remapped, and is a custom component (not hidden by
		// isHiddenFromScreenReader's input-type=hidden branch). The h1
		// therefore has a non-hidden child → valid.
		{Code: `<h1><CustomInput type="hidden" /></h1>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- DEFAULT ELEMENT TESTS ----
		{
			Code: `<h1 />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1><Bar aria-hidden /></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1>{undefined}</h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1><input type="hidden" /></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- CUSTOM ELEMENT TESTS FOR COMPONENTS OPTION ----
		{
			Code:    `<Heading />`,
			Tsx:     true,
			Options: componentsRuleOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Heading><Bar aria-hidden /></Heading>`,
			Tsx:     true,
			Options: componentsRuleOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Heading>{undefined}</Heading>`,
			Tsx:     true,
			Options: componentsRuleOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- CUSTOM ELEMENT TESTS FOR COMPONENTS SETTINGS ----
		{
			Code:     `<Heading />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Locks in upstream's settings-based remap reaching the
		// isHiddenFromScreenReader input-type-hidden arm: with
		// `componentsSettings`, CustomInput is remapped to "input", so the
		// hidden-typed input is recognized as hidden → no accessible child →
		// h1 is reported. Without the remap (the valid case above), CustomInput
		// stays a custom component and is non-hidden.
		{
			Code:     `<h1><CustomInput type="hidden" /></h1>`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
