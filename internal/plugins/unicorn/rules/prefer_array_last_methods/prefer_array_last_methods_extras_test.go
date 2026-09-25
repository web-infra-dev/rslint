package prefer_array_last_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_last_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferArrayLastMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_last_methods.PreferArrayLastMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "array.reverse?.().find(logic);"},
			{Code: "array.toReversed?.().reduce(logic);"},
			{Code: "array.reverse().find;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "array.reverse().find();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast();"}}}}},
			{Code: "array.toReversed().indexOf(value, fromIndex);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#lastIndexOf()` over `Array#toReversed().indexOf()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(value, fromIndex);"}}}}},
		},
	)
}
