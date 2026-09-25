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
		},
		[]rule_tester.InvalidTestCase{
			{Code: "array.reverse().find();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast();"}}}}},
			{Code: "array.toReversed().indexOf(value, fromIndex);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(value, fromIndex);"}}}}},
		},
	)
}
