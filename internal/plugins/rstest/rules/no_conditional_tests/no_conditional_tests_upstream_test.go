// TestNoConditionalTestsUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 no-conditional-tests suite. Rstest-only edge
// cases, the source matrix and the branch lock-ins that upstream's fixed
// four-level parent check cannot express live in
// no_conditional_tests_extras_test.go.
package no_conditional_tests

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalTestsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConditionalTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("shows error", () => {});`},
			{Code: `it("foo", function () {})`},
			{Code: `it('foo', () => {}); function myTest() { if ('bar') {} }`},
			{Code: `function myFunc(str: string) {
    return str;
  }
  describe("myTest", () => {
     it("convert shortened equal filter", () => {
      expect(
      myFunc("5")
      ).toEqual("5");
     });
    });`},
			{Code: `describe("shows error", () => {
     if (1 === 2) {
      myFunc();
     }
     expect(true).toBe(false);
    });`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe("shows error", () => {
    if(true) {
     test("shows error", () => {
      expect(true).toBe(true);
     })
    }
   })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      3,
					Column:    6,
					EndLine:   3,
					EndColumn: 10,
				}},
			},
			{
				Code: `
   describe("shows error", () => {
    if(true) {
     it("shows error", () => {
      expect(true).toBe(true);
      })
     }
   })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      4,
					Column:    6,
					EndLine:   4,
					EndColumn: 8,
				}},
			},
			{
				Code: `describe("errors", () => {
    if (Math.random() > 0) {
     test("test2", () => {
     expect(true).toBeTruthy();
    });
    }
   });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      3,
					Column:    6,
					EndLine:   3,
					EndColumn: 10,
				}},
			},
		},
	)
}
