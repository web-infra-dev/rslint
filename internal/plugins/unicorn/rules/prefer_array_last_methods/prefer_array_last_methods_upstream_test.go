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
			{Code: "array.reverse().find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic);"}}}}},
			{Code: "array.reverse().findIndex(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLastIndex()` over `Array#reverse().findIndex()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic);"}}}}},
			{Code: "array.reverse().indexOf(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#lastIndexOf()` over `Array#reverse().indexOf()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(logic);"}}}}},
			{Code: "array.reverse().reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#reduceRight()` over `Array#reverse().reduce()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic);"}}}}},
			{Code: "array.toReversed().find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#toReversed().find()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic);"}}}}},
			{Code: "array.toReversed().findIndex(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLastIndex()` over `Array#toReversed().findIndex()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic);"}}}}},
			{Code: "array.toReversed().indexOf(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#lastIndexOf()` over `Array#toReversed().indexOf()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(logic);"}}}}},
			{Code: "array.toReversed().reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic);"}}}}},
			{Code: "array.reverse().find(logic, thisArgument);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast(logic, thisArgument);"}}}}},
			{Code: "array.toReversed().findIndex(logic, thisArgument);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLastIndex()` over `Array#toReversed().findIndex()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLastIndex(logic, thisArgument);"}}}}},
			{Code: "array.reverse().reduce(logic, initialValue);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#reduceRight()` over `Array#reverse().reduce()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic, initialValue);"}}}}},
			{Code: "array.toReversed().reduce(logic, initialValue);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.reduceRight(logic, initialValue);"}}}}},
			{Code: "(array.reverse()).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "(array).findLast(logic);"}}}}},
			{Code: "(array.toReversed()).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#toReversed().find()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "(array).findLast(logic);"}}}}},
			{Code: "array.reverse(/* comment */).find(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34}}},
			{Code: "array.toReversed() /* comment */ .reduce(logic);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 41}}},
		},
	)
}
