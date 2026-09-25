package prefer_array_last_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_last_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferArrayLastMethodsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_last_methods.PreferArrayLastMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "array.findLast(logic);"},
			{Code: "array.findLastIndex(logic);"},
			{Code: "array.lastIndexOf(logic);"},
			{Code: "array.reduceRight(logic);"},
			{Code: "array.reverse();"},
			{Code: "array.toReversed();"},
			{Code: "array.reverse(logic).find(logic);"},
			{Code: "array.toReversed(logic).find(logic);"},
			{Code: "array.reverse?.().find(logic);"},
			{Code: "array.toReversed?.().find(logic);"},
			{Code: "array?.reverse().find(logic);"},
			{Code: "array?.toReversed().find(logic);"},
			{Code: "array[reverse]().find(logic);"},
			{Code: "array.reverse()[find](logic);"},
			{Code: "array.reverse().map(logic);"},
			{Code: "array.toReversed().map(logic);"},
			{Code: "const reversed = array.reverse(); reversed.find(logic);"},
			{Code: "array.reverse().find;"},
			{Code: "array.toReversed().find;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "array.reverse().find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic);"}}}}},
			{Code: "array.reverse().findIndex(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic);"}}}}},
			{Code: "array.reverse().indexOf(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(logic);"}}}}},
			{Code: "array.reverse().reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic);"}}}}},
			{Code: "array.toReversed().find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic);"}}}}},
			{Code: "array.toReversed().findIndex(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic);"}}}}},
			{Code: "array.toReversed().indexOf(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(logic);"}}}}},
			{Code: "array.toReversed().reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic);"}}}}},
			{Code: "array.reverse().find(logic, thisArgument);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic, thisArgument);"}}}}},
			{Code: "array.toReversed().findIndex(logic, thisArgument);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic, thisArgument);"}}}}},
			{Code: "array.reverse().reduce(logic, initialValue);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic, initialValue);"}}}}},
			{Code: "array.toReversed().reduce(logic, initialValue);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic, initialValue);"}}}}},
			{Code: "(array.reverse()).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "(array).findLast(logic);"}}}}},
			{Code: "(array.toReversed()).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "(array).findLast(logic);"}}}}},
			{Code: "array.reverse(/* comment */).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods"}}},
			{Code: "array.toReversed() /* comment */ .reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods"}}},
		},
	)
}
