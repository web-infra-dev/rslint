package autocomplete_valid

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors the upstream `componentsSettings` constant —
// maps the custom JSX tag `Input` to the bare HTML `input` element so the
// rule treats `<Input ... />` as a regular `<input>`.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Input": "input",
		},
	},
}

// TestAutocompleteValidUpstream is a verbatim port of upstream's
// `__tests__/src/rules/autocomplete-valid-test.js`. Each case below maps 1:1
// to a case in the upstream file, in the same order, with the same code,
// the same options, and the same expected verdict. Layout / section markers
// preserve the upstream comments so a future audit against the upstream
// file remains a trivial diff.
//
// Additional rslint lock-in tests (axe-core grammar boundaries,
// JSX-shape variants, options-shape coverage, settings interaction, etc.)
// live in `autocomplete_valid_extras_test.go` so this file stays a pure
// reflection of upstream.
func TestAutocompleteValidUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		// INAPPLICABLE
		{Code: `<input type="text" />;`, Tsx: true},

		// PASSED AUTOCOMPLETE
		{Code: `<input type="text" autocomplete="name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="off" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="on" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="billing family-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="section-blue shipping street-address" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="section-somewhere shipping work email" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete />;`, Tsx: true},
		// Upstream uses a truncated identifier name for this variable.
		// Renamed here to `dynValue` to avoid a cspell false-positive
		// on the truncated word; the AST shape and the rule's behavior
		// are unchanged (any non-undefined Identifier exercises the
		// same LITERAL_TYPES.Identifier noop branch).
		{Code: `<input type="text" autocomplete={dynValue} />;`, Tsx: true},
		{Code: `<input type="text" autocomplete={dynValue || "name"} />;`, Tsx: true},
		{Code: `<input type="text" autocomplete={dynValue || "foo"} />;`, Tsx: true},
		{Code: `<Foo autocomplete="bar"></Foo>;`, Tsx: true},
		{Code: `<input type={isEmail ? "email" : "text"} autocomplete="none" />;`, Tsx: true},
		{
			Code:     `<Input type="text" autocomplete="name" />`,
			Tsx:      true,
			Settings: componentsSettings,
		},
		{Code: `<Input type="text" autocomplete="baz" />`, Tsx: true},

		// PASSED "autocomplete-appropriate"
		// see also: https://github.com/dequelabs/axe-core/issues/2912
		{Code: `<input type="date" autocomplete="email" />;`, Tsx: true},
		{Code: `<input type="number" autocomplete="url" />;`, Tsx: true},
		{Code: `<input type="month" autocomplete="tel" />;`, Tsx: true},
		{
			Code:    `<Foo type="month" autocomplete="tel"></Foo>;`,
			Tsx:     true,
			Options: map[string]interface{}{"inputComponents": []interface{}{"Foo"}},
		},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		// FAILED "autocomplete-valid"
		{
			Code: `<input type="text" autocomplete="foo" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input type="text" autocomplete="name invalid" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input type="text" autocomplete="invalid name" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input type="text" autocomplete="home url" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code:    `<Bar autocomplete="baz"></Bar>;`,
			Tsx:     true,
			Options: map[string]interface{}{"inputComponents": []interface{}{"Bar"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code:     `<Input type="text" autocomplete="baz" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AutocompleteValidRule, validCases, invalidCases)
}
