package no_distracting_elements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedMarqueeError / expectedBlinkError mirror upstream's
// `expectedError(element)` helper. Both messages are interpolated forms of
// the same template, but it is more readable to spell them out.
var expectedMarqueeError = rule_tester.InvalidTestCaseError{
	MessageId: "distractingElement",
	Message:   "Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.",
}

var expectedBlinkError = rule_tester.InvalidTestCaseError{
	MessageId: "distractingElement",
	Message:   "Do not use <blink> elements as they can create visual accessibility issues and are deprecated.",
}

// blinkComponentSettings mirrors upstream's `componentsSettings` constant —
// the components-map entry that aliases `<Blink />` to the intrinsic `blink`
// tag, exercising the GetElementType components-map path.
var blinkComponentSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Blink": "blink",
		},
	},
}

// TestNoDistractingElementsUpstream covers the full valid / invalid suite
// migrated 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/no-distracting-elements-test.js`. Order and grouping
// mirror the upstream file so a future audit can grep across both
// side-by-side.
//
// rslint-specific lock-ins (Dimension 1-4 edge shapes, options coverage,
// position assertions, polymorphic prop, listener boundary) live in
// no_distracting_elements_extras_test.go.
func TestNoDistractingElementsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDistractingElementsRule, []rule_tester.ValidTestCase{
		{Code: `<div />;`, Tsx: true},
		// React component name is case-sensitive — `<Marquee />` is a user
		// component, NOT the intrinsic `marquee`. Locks in the strict-equality
		// comparison.
		{Code: `<Marquee />`, Tsx: true},
		// `marquee` as an attribute on a `<div>` doesn't trigger — the rule
		// inspects the tag name only.
		{Code: `<div marquee />`, Tsx: true},
		// Same for `Blink`: capital-B component name doesn't match without a
		// components-map alias.
		{Code: `<Blink />`, Tsx: true},
		{Code: `<div blink />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `<marquee />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<marquee {...props} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<marquee lang={undefined} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedMarqueeError},
		},
		{
			Code:   `<blink />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},
		{
			Code:   `<blink {...props} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},
		{
			Code:   `<blink foo={undefined} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},
		// `<Blink />` aliased to `blink` via the components-map setting —
		// jsx-ast-utils' getElementType replaces "Blink" with "blink" before
		// the lookup, so the rule reports.
		{
			Code:     `<Blink />`,
			Tsx:      true,
			Settings: blinkComponentSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedBlinkError},
		},
	})
}
