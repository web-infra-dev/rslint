// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-array-sort.js
package no_array_sort_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_sort"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoArraySortUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_sort.NoArraySortRule, []rule_tester.ValidTestCase{
		{Code: "function f(foo: Set<number>) { foo.sort(); }", FileName: "case.ts", Options: []any{}},
		{Code: "sorted =[...array].toSorted()", FileName: "case.js", Options: []any{}},
		{Code: "sorted =array.toSorted()", FileName: "case.js", Options: []any{}},
		{Code: "sorted =[...array].sort", FileName: "case.js", Options: []any{}},
		{Code: "sorted =[...array].sort?.()", FileName: "case.js", Options: []any{}},
		{Code: "array.sort()", FileName: "case.js", Options: []any{}},
		{Code: "array.sort?.()", FileName: "case.js", Options: []any{}},
		{Code: "array?.sort()", FileName: "case.js", Options: []any{}},
		{Code: "if (true) array.sort()", FileName: "case.js", Options: []any{}},
		{Code: "sorted = array.sort(...[])", FileName: "case.js", Options: []any{}},
		{Code: "sorted = array.sort(...[compareFn])", FileName: "case.js", Options: []any{}},
		{Code: "sorted = array.sort(compareFn, extraArgument)", FileName: "case.js", Options: []any{}},
		{Code: "sorted = collection.sort({field: 1})", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort(\"field\")", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort(1)", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort(-1)", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort(+1)", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort(`field`)", FileName: "case.js", Options: []any{}},
		{Code: "sorted = query.sort([criteria])", FileName: "case.js", Options: []any{}},
		{Code: "const docs = collection.find({id}).sort({expireAt: -1}).limit(1).toArray()", FileName: "case.js", Options: []any{}},
		{Code: "[...array].sort({field: 1})", FileName: "case.js", Options: []any{}},
		{Code: "collection.sort({field: 1})", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}},
		{Code: "const sorted = array.toSorted();", FileName: "case.js", Options: []any{}},
		{Code: "const sorted = [...iterable].toSorted();", FileName: "case.js", Options: []any{}},
		{Code: "const sorted = array.toSorted((a, b) => a - b);", FileName: "case.js", Options: []any{}},
		{Code: "const sortedArray = array.toSorted();", FileName: "case.js", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "sorted = [...array].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = array.toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...array].toSorted()"}}},
		}},
		{Code: "sorted = [...array]?.sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = array?.toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...array]?.toSorted()"}}},
		}},
		{Code: "sorted = array.sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "sorted = array.toSorted()"}}},
		}},
		{Code: "sorted = array?.sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "sorted = array?.toSorted()"}}},
		}},
		{Code: "sorted = [...array].sort(compareFn)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = array.toSorted(compareFn)"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...array].toSorted(compareFn)"}}},
		}},
		{Code: "sorted = [...array]?.sort(compareFn)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = array?.toSorted(compareFn)"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...array]?.toSorted(compareFn)"}}},
		}},
		{Code: "sorted = array.sort(compareFn)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "sorted = array.toSorted(compareFn)"}}},
		}},
		{Code: "sorted = array?.sort(compareFn)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "sorted = array?.toSorted(compareFn)"}}},
		}},
		{Code: "array.sort()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array.toSorted()"}}},
		}},
		{Code: "array?.sort()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array?.toSorted()"}}},
		}},
		{Code: "[...array].sort()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "array.toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "[...array].toSorted()"}}},
		}},
		{Code: "sorted = [...(0, array)].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = (0, array).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...(0, array)].toSorted()"}}},
		}},
		{Code: "sorted = [...a + b].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "sorted = (a + b).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "sorted = [...a + b].toSorted()"}}},
		}},
		{Code: "const sorted = [...array].sort();", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "const sorted = array.toSorted();"}, {MessageId: "suggestion-not-spreading-array", Output: "const sorted = [...array].toSorted();"}}},
		}},
		{Code: "const sorted = [...iterable].sort();", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "const sorted = iterable.toSorted();"}, {MessageId: "suggestion-not-spreading-array", Output: "const sorted = [...iterable].toSorted();"}}},
		}},
		{Code: "const sorted = [...array].sort((a, b) => a - b);", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "const sorted = array.toSorted((a, b) => a - b);"}, {MessageId: "suggestion-not-spreading-array", Output: "const sorted = [...array].toSorted((a, b) => a - b);"}}},
		}},
		{Code: "array.sort();", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array.toSorted();"}}},
		}},
	})
}
