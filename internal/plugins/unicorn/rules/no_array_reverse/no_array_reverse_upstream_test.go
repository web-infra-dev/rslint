// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-array-reverse.js
package no_array_reverse_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_reverse"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoArrayReverseUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_reverse.NoArrayReverseRule, []rule_tester.ValidTestCase{
		{Code: "function f(foo: Set<number>) { foo.reverse(); }", FileName: "case.ts", Options: []any{}},
		{Code: "reversed =[...array].toReversed()", FileName: "case.js", Options: []any{}},
		{Code: "reversed =array.toReversed()", FileName: "case.js", Options: []any{}},
		{Code: "reversed =[...array].reverse", FileName: "case.js", Options: []any{}},
		{Code: "reversed =[...array].reverse?.()", FileName: "case.js", Options: []any{}},
		{Code: "array.reverse()", FileName: "case.js", Options: []any{}},
		{Code: "array.reverse?.()", FileName: "case.js", Options: []any{}},
		{Code: "array?.reverse()", FileName: "case.js", Options: []any{}},
		{Code: "if (true) array.reverse()", FileName: "case.js", Options: []any{}},
		{Code: "reversed = array.reverse(extraArgument)", FileName: "case.js", Options: []any{}},
		{Code: "const reversed = [...array].toReversed();", FileName: "case.js", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "reversed = [...array].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = array.toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...array].toReversed()"}}},
		}},
		{Code: "reversed = [...array]?.reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = array?.toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...array]?.toReversed()"}}},
		}},
		{Code: "reversed = array.reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "reversed = array.toReversed()"}}},
		}},
		{Code: "reversed = array?.reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "reversed = array?.toReversed()"}}},
		}},
		{Code: "array.reverse()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array.toReversed()"}}},
		}},
		{Code: "array?.reverse()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array?.toReversed()"}}},
		}},
		{Code: "[...array].reverse()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "array.toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "[...array].toReversed()"}}},
		}},
		{Code: "reversed = [...(0, array)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = (0, array).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...(0, array)].toReversed()"}}},
		}},
		{Code: "reversed = [...a + b].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = (a + b).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...a + b].toReversed()"}}},
		}},
		{Code: "reversed = [...a ? b : c].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = (a ? b : c).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...a ? b : c].toReversed()"}}},
		}},
		{Code: "reversed = [...(a + b)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = (a + b).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...(a + b)].toReversed()"}}},
		}},
		{Code: "reversed = [...new Set(array)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "reversed = new Set(array).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "reversed = [...new Set(array)].toReversed()"}}},
		}},
		{Code: "const reversed = [...array].reverse();", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "const reversed = array.toReversed();"}, {MessageId: "suggestion-not-spreading-array", Output: "const reversed = [...array].toReversed();"}}},
		}},
		{Code: "array.reverse();", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "array.toReversed();"}}},
		}},
	})
}
