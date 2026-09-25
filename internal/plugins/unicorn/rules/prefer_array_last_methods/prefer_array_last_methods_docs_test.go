package prefer_array_last_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_last_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferArrayLastMethodsDocs(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_last_methods.PreferArrayLastMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "const result = array.findLast(isUnicorn);"},
			{Code: "const result = array.reduceRight(reducer, initialValue);"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const result = array.reverse().find(isUnicorn);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "const result = array.findLast(isUnicorn);"}}}}},
			{Code: "const result = array.toReversed().reduce(reducer, initialValue);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "const result = array.reduceRight(reducer, initialValue);"}}}}},
		},
	)
}
