package no_meaningless_void_operator

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoMeaninglessVoidOperatorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMeaninglessVoidOperatorRule, []rule_tester.ValidTestCase{
		{Code: `
(() => {})();

function foo() {}
foo(); // nothing to discard

function bar(x: number) {
  void x;
  return 2;
}
void bar(); // discarding a number
    `},
		{Code: `
function bar(x: never) {
  void x;
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "void (() => {})();",
			Output: []string{" (() => {})();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo() {}
void foo();
      `,
			Output: []string{`
function foo() {}
 foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `
function bar(x: never) {
  void x;
}
      `,
			Options: NoMeaninglessVoidOperatorOptions{CheckNever: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeVoid",
							Output: `
function bar(x: never) {
   x;
}
      `,
						},
					},
				},
			},
		},
	})
}
