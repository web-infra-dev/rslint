// TestNoConditionalTestsExtrasClassTiming locks in which class-member
// expressions run while a class is defined and which run later. General
// Rstest-only edge shapes live in no_conditional_tests_extras_test.go.
package no_conditional_tests

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalTestsExtrasClassTiming(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConditionalTestsRule,
		[]rule_tester.ValidTestCase{
			// Method and constructor bodies run only when called.
			{Code: `if (x) { class C { m() { test("a", () => {}); } } }`},
			{Code: `if (x) { class C { constructor() { test("a", () => {}); } } }`},
			// Parameter defaults and binding patterns run only when called.
			{Code: `if (x) { function f(t = test("a", () => {})) {} }`},
			{Code: `if (x) { class C { m({ p = test("a", () => {}) } = {}) {} } }`},
			// An instance field initializer runs once per construction.
			{Code: `if (x) { class C { p = test("a", () => {}); } }`},
		},
		[]rule_tester.InvalidTestCase{
			// Static blocks and static initializers run while defining the class.
			{
				Code: `if (x) { class C { static { test("a", () => {}); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code: `if (x) { class C { static p = test("a", () => {}); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 35,
				}},
			},
			// Computed names run while defining both methods and fields.
			{
				Code: `if (x) { class C { [test("a", () => {})]() {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			{
				Code: `if (x) { class C { [test("a", () => {})] = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			// Member and parameter decorators run while defining the class.
			{
				Code: `if (x) { class C { @dec(test("a", () => {})) m() {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code: `if (x) { class C { @dec(test("a", () => {})) p = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code: `if (x) { class C { m(@dec(test("a", () => {})) value: unknown) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    27,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
		},
	)
}
