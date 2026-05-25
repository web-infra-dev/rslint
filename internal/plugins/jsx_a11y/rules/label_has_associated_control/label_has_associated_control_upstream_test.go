// label_has_associated_control_upstream_test.go is a 1:1 port of upstream's
//
//	__tests__/src/rules/label-has-associated-control-test.js
//
// It preserves upstream's source organization (`alwaysValid`, `htmlForValid`,
// `nestingValid`, `bothValid`, `htmlForInvalid`, `nestingInvalid`,
// `neverValid`) and its "run the same case list once per assert mode"
// structure (4 `ruleTester.run` blocks → 4 `Test*_Assert*` Go tests).
//
// Lock-in tests for branches not directly hit by the upstream suite live in
// `label_has_associated_control_extras_test.go`; they share the variables
// declared here.
package label_has_associated_control

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedAccessibleLabel matches the `accessibleLabel` diagnostic (always
// reported first when the label has no accessible text — regardless of the
// configured `assert` mode).
var expectedAccessibleLabel = rule_tester.InvalidTestCaseError{
	MessageId: "accessibleLabel",
	Message:   msgAccessibleLabel,
}

var expectedHtmlFor = rule_tester.InvalidTestCaseError{
	MessageId: "htmlFor",
	Message:   msgHtmlFor,
}

var expectedNesting = rule_tester.InvalidTestCaseError{
	MessageId: "nesting",
	Message:   msgNesting,
}

var expectedBoth = rule_tester.InvalidTestCaseError{
	MessageId: "both",
	Message:   msgBoth,
}

var expectedEither = rule_tester.InvalidTestCaseError{
	MessageId: "either",
	Message:   msgEither,
}

// componentsSettings mirrors upstream `componentsSettings` —
// `CustomInput` → `input`, `CustomLabel` → `label`.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"CustomInput": "input",
			"CustomLabel": "label",
		},
	},
}

// attributesSettings mirrors upstream `attributesSettings` —
// `settings['jsx-a11y'].attributes.for = ['htmlFor', 'for']` enables `for=…`
// as an htmlFor-equivalent attribute alongside the React-canonical `htmlFor`.
var attributesSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"attributes": map[string]interface{}{
			"for": []interface{}{"htmlFor", "for"},
		},
	},
}

// caseWithOptions packages a code/options/settings tuple — mirrors upstream's
// `htmlForValid`/`nestingValid`/`neverValid`/`bothValid` arrays.
type caseWithOptions struct {
	code     string
	options  map[string]interface{}
	settings map[string]interface{}
}

// neverValidCase extends caseWithOptions with a per-row error toggle: when
// `useAccessibleLabel` is true the row asserts `accessibleLabel` (no text
// content); otherwise it asserts the `assert`-specific error. Mirrors
// upstream's `neverValid(assertType)` factory shape.
type neverValidCase struct {
	code               string
	options            map[string]interface{}
	settings           map[string]interface{}
	useAccessibleLabel bool
}

func validWithAssert(assert string, cases ...caseWithOptions) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, 0, len(cases))
	for _, c := range cases {
		opts := mergeOptions(c.options, assert)
		out = append(out, rule_tester.ValidTestCase{
			Code:     c.code,
			Tsx:      true,
			Options:  opts,
			Settings: c.settings,
		})
	}
	return out
}

func invalidWithAssertAndError(assert string, expected rule_tester.InvalidTestCaseError, cases ...caseWithOptions) []rule_tester.InvalidTestCase {
	out := make([]rule_tester.InvalidTestCase, 0, len(cases))
	for _, c := range cases {
		opts := mergeOptions(c.options, assert)
		out = append(out, rule_tester.InvalidTestCase{
			Code:     c.code,
			Tsx:      true,
			Options:  opts,
			Settings: c.settings,
			Errors:   []rule_tester.InvalidTestCaseError{expected},
		})
	}
	return out
}

func invalidNeverValid(assert string, assertError rule_tester.InvalidTestCaseError, cases ...neverValidCase) []rule_tester.InvalidTestCase {
	out := make([]rule_tester.InvalidTestCase, 0, len(cases))
	for _, c := range cases {
		opts := mergeOptions(c.options, assert)
		errOut := assertError
		if c.useAccessibleLabel {
			errOut = expectedAccessibleLabel
		}
		out = append(out, rule_tester.InvalidTestCase{
			Code:     c.code,
			Tsx:      true,
			Options:  opts,
			Settings: c.settings,
			Errors:   []rule_tester.InvalidTestCaseError{errOut},
		})
	}
	return out
}

func mergeOptions(base map[string]interface{}, assert string) []interface{} {
	merged := map[string]interface{}{}
	for k, v := range base {
		merged[k] = v
	}
	merged["assert"] = assert
	return []interface{}{merged}
}

// ----------------------------------------------------------------------------
// Upstream source arrays — keep order and content byte-aligned with
// __tests__/src/rules/label-has-associated-control-test.js so future audits
// can diff line-for-line.
// ----------------------------------------------------------------------------

// alwaysValid (upstream `alwaysValid`) — not a label / not a custom-label-like
// shape; rule never triggers regardless of `assert`.
var alwaysValid = []caseWithOptions{
	{code: `<div />`},
	{code: `<CustomElement />`},
	{code: `<input type="hidden" />`},
}

// htmlForValid (upstream `htmlForValid`) — labels that pass the `htmlFor`
// mode (have a valid htmlFor + accessible text).
var htmlForValid = []caseWithOptions{
	{code: `<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}},
	{code: `<label htmlFor="js_id" aria-label="A label" />`},
	{code: `<label htmlFor="js_id" aria-labelledby="A label" />`},
	{code: `<div><label htmlFor="js_id">A label</label><input id="js_id" /></div>`},
	{code: `<label for="js_id"><span><span><span>A label</span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}, settings: attributesSettings},
	{code: `<label for="js_id" aria-label="A label" />`, settings: attributesSettings},
	{code: `<label for="js_id" aria-labelledby="A label" />`, settings: attributesSettings},
	{code: `<div><label for="js_id">A label</label><input id="js_id" /></div>`, settings: attributesSettings},
	// Custom label component.
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, options: map[string]interface{}{"labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, settings: componentsSettings},
	{code: `<MUILabel htmlFor="js_id" aria-label="A label" />`, options: map[string]interface{}{"labelComponents": []interface{}{"*Label"}}},
	{code: `<LabelCustom htmlFor="js_id" label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"Label*"}}},
	// Custom label attributes.
	{code: `<label htmlFor="js_id" label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}}},
	// Glob support for controlComponents option.
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, options: map[string]interface{}{"controlComponents": []interface{}{"Custom*"}}},
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, options: map[string]interface{}{"controlComponents": []interface{}{"*Label"}}},
	// Rule does not error if presence of accessible label cannot be determined.
	{code: `<div><label htmlFor="js_id"><CustomText /></label><input id="js_id" /></div>`},
}

// nestingValid (upstream `nestingValid`) — labels that pass the `nesting`
// mode (have an accessible-text descendant + a nested form control).
var nestingValid = []caseWithOptions{
	{code: `<label>A label<input /></label>`},
	{code: `<label>A label<textarea /></label>`},
	{code: `<label><img alt="A label" /><input /></label>`},
	{code: `<label><img aria-label="A label" /><input /></label>`},
	{code: `<label><span>A label<input /></span></label>`},
	{code: `<label><span><span>A label<input /></span></span></label>`, options: map[string]interface{}{"depth": float64(3)}},
	{code: `<label><span><span><span>A label<input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}},
	{code: `<label><span><span><span><span>A label</span><input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	{code: `<label><span><span><span><span aria-label="A label" /><input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	{code: `<label><span><span><span><input aria-label="A label" /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	// Other controls.
	{code: `<label>foo<meter /></label>`},
	{code: `<label>foo<output /></label>`},
	{code: `<label>foo<progress /></label>`},
	{code: `<label>foo<textarea /></label>`},
	// Custom controlComponents.
	{code: `<label>A label<CustomInput /></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}}},
	{code: `<label><span>A label<CustomInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}}},
	{code: `<label><span>A label<CustomInput /></span></label>`, settings: componentsSettings},
	{code: `<CustomLabel><span>A label<CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel><span label="A label"><CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}, "labelAttributes": []interface{}{"label"}}},
	// Glob support for controlComponents.
	{code: `<label><span>A label<CustomInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"Custom*"}}},
	{code: `<label><span>A label<CustomInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"*Input"}}},
	{code: `<label><span>A label<TextInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"????Input"}}},
	// Rule does not error if presence of accessible label cannot be determined.
	{code: `<label><CustomText /><input /></label>`},
}

// bothValid (upstream `bothValid`) — labels that pass the `both` mode (have
// BOTH an htmlFor attribute AND a nested form control AND accessible text).
var bothValid = []caseWithOptions{
	{code: `<label htmlFor="js_id"><span><span><span>A label<input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}},
	{code: `<label htmlFor="js_id" aria-label="A label"><input /></label>`},
	{code: `<label htmlFor="js_id" aria-labelledby="A label"><input /></label>`},
	{code: `<label htmlFor="js_id" aria-labelledby="A label"><textarea /></label>`},
	// Custom label component.
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label"><input /></CustomLabel>`, options: map[string]interface{}{"labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" label="A label"><input /></CustomLabel>`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" label="A label"><input /></CustomLabel>`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"*Label"}}},
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label"><input /></CustomLabel>`, settings: componentsSettings},
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label"><CustomInput /></CustomLabel>`, settings: componentsSettings},
	// Custom label attributes.
	{code: `<label htmlFor="js_id" label="A label"><input /></label>`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}}},
	{code: `<label htmlFor="selectInput">Some text<select id="selectInput" /></label>`},
}

// htmlForInvalid (upstream `htmlForInvalid(assertType)`) — labels that PASS
// the `htmlFor` mode (have a valid htmlFor attribute) but FAIL stricter modes
// (`nesting` / `both`). For `assert: 'htmlFor'` these are valid; the upstream
// generator only re-runs them in non-`htmlFor` modes.
var htmlForInvalid = []caseWithOptions{
	{code: `<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}},
	{code: `<label htmlFor="js_id" aria-label="A label" />`},
	{code: `<label htmlFor="js_id" aria-labelledby="A label" />`},
	// Custom label component.
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, options: map[string]interface{}{"labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel htmlFor="js_id" aria-label="A label" />`, settings: componentsSettings},
	// Custom label attributes.
	{code: `<label htmlFor="js_id" label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}}},
}

// nestingInvalid (upstream `nestingInvalid(assertType)`) — labels that pass
// `nesting` mode (have a nested control + text) but fail `htmlFor` / `both`.
var nestingInvalid = []caseWithOptions{
	{code: `<label>A label<input /></label>`},
	{code: `<label>A label<textarea /></label>`},
	{code: `<label><img alt="A label" /><input /></label>`},
	{code: `<label><img aria-label="A label" /><input /></label>`},
	{code: `<label><span>A label<input /></span></label>`},
	{code: `<label><span><span>A label<input /></span></span></label>`, options: map[string]interface{}{"depth": float64(3)}},
	{code: `<label><span><span><span>A label<input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(4)}},
	{code: `<label><span><span><span><span>A label</span><input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	{code: `<label><span><span><span><span aria-label="A label" /><input /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	{code: `<label><span><span><span><input aria-label="A label" /></span></span></span></label>`, options: map[string]interface{}{"depth": float64(5)}},
	// Custom controlComponents.
	{code: `<label>A label<OtherCustomInput /></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}}},
	{code: `<label><span>A label<CustomInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}}},
	{code: `<CustomLabel><span>A label<CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel><span label="A label"><CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}, "labelAttributes": []interface{}{"label"}}},
	{code: `<label><span>A label<CustomInput /></span></label>`, settings: componentsSettings},
	{code: `<CustomLabel><span>A label<CustomInput /></span></CustomLabel>`, settings: componentsSettings},
}

// neverValid (upstream `neverValid(assertType)`) — labels that fail in EVERY
// mode. The first few entries always emit `accessibleLabel` (no text);
// remaining entries emit the assert-specific error.
var neverValid = []neverValidCase{
	{code: `<label htmlFor="js_id" />`, useAccessibleLabel: true},
	{code: `<label htmlFor="js_id"><input /></label>`, useAccessibleLabel: true},
	{code: `<label htmlFor="js_id"><textarea /></label>`, useAccessibleLabel: true},
	{code: `<label></label>`, useAccessibleLabel: true},
	{code: `<label>A label</label>`},
	{code: `<div><label /><input /></div>`, useAccessibleLabel: true},
	{code: `<div><label>A label</label><input /></div>`},
	// Custom label component.
	{code: `<CustomLabel aria-label="A label" />`, options: map[string]interface{}{"labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<MUILabel aria-label="A label" />`, options: map[string]interface{}{"labelComponents": []interface{}{"???Label"}}},
	{code: `<CustomLabel label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}, "labelComponents": []interface{}{"CustomLabel"}}},
	{code: `<CustomLabel aria-label="A label" />`, settings: componentsSettings},
	// Custom label attributes.
	{code: `<label label="A label" />`, options: map[string]interface{}{"labelAttributes": []interface{}{"label"}}},
	// Custom controlComponents.
	{code: `<label><span><CustomInput /></span></label>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}}, useAccessibleLabel: true},
	{code: `<CustomLabel><span><CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}}, useAccessibleLabel: true},
	{code: `<CustomLabel><span><CustomInput /></span></CustomLabel>`, options: map[string]interface{}{"controlComponents": []interface{}{"CustomInput"}, "labelComponents": []interface{}{"CustomLabel"}, "labelAttributes": []interface{}{"label"}}, useAccessibleLabel: true},
	{code: `<label><span><CustomInput /></span></label>`, settings: componentsSettings, useAccessibleLabel: true},
	{code: `<CustomLabel><span><CustomInput /></span></CustomLabel>`, settings: componentsSettings, useAccessibleLabel: true},
}

// ----------------------------------------------------------------------------
// One Go test per upstream `ruleTester.run` block — matches upstream's
// "run the same case list once per assert mode" pattern. The four blocks
// share the source arrays above with different per-mode dispositions.
// ----------------------------------------------------------------------------

func TestLabelHasAssociatedControl_AssertHtmlFor(t *testing.T) {
	// Upstream: valid = alwaysValid ++ htmlForValid;
	// invalid = neverValid('htmlFor') ++ nestingInvalid('htmlFor').
	valid := append(validWithAssert("htmlFor", alwaysValid...), validWithAssert("htmlFor", htmlForValid...)...)

	invalid := append(invalidNeverValid("htmlFor", expectedHtmlFor, neverValid...),
		invalidWithAssertAndError("htmlFor", expectedHtmlFor, nestingInvalid...)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule, valid, invalid)
}

func TestLabelHasAssociatedControl_AssertNesting(t *testing.T) {
	// Upstream: valid = alwaysValid ++ nestingValid;
	// invalid = neverValid('nesting') ++ htmlForInvalid('nesting').
	valid := append(validWithAssert("nesting", alwaysValid...), validWithAssert("nesting", nestingValid...)...)

	invalid := append(invalidNeverValid("nesting", expectedNesting, neverValid...),
		invalidWithAssertAndError("nesting", expectedNesting, htmlForInvalid...)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule, valid, invalid)
}

func TestLabelHasAssociatedControl_AssertEither(t *testing.T) {
	// Upstream: valid = alwaysValid ++ htmlForValid ++ nestingValid;
	// invalid = neverValid('either').
	valid := append(validWithAssert("either", alwaysValid...), validWithAssert("either", htmlForValid...)...)
	valid = append(valid, validWithAssert("either", nestingValid...)...)

	invalid := invalidNeverValid("either", expectedEither, neverValid...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule, valid, invalid)
}

func TestLabelHasAssociatedControl_AssertBoth(t *testing.T) {
	// Upstream: valid = alwaysValid ++ bothValid;
	// invalid = neverValid('both') ++ htmlForInvalid('both') ++ nestingInvalid('both').
	valid := append(validWithAssert("both", alwaysValid...), validWithAssert("both", bothValid...)...)

	invalid := append(invalidNeverValid("both", expectedBoth, neverValid...),
		invalidWithAssertAndError("both", expectedBoth, htmlForInvalid...)...)
	invalid = append(invalid, invalidWithAssertAndError("both", expectedBoth, nestingInvalid...)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule, valid, invalid)
}
