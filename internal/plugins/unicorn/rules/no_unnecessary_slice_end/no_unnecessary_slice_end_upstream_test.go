// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-unnecessary-slice-end.js
package no_unnecessary_slice_end_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_slice_end"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUnnecessarySliceEndUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_slice_end.NoUnnecessarySliceEndRule, []rule_tester.ValidTestCase{
		{Code: "foo.slice?.(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(foo.length, 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice()", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, foo.length - 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, foo.length, extraArgument)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(...[1], foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.not_slice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "new foo.slice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "slice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, foo.notLength)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo[slice](1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, foo[length])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.slice(1, bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.slice(1, NotInfinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.slice(1, Number.NOT_POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.slice(1, Not_Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.slice(1, Number?.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo().slice(1, foo().length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(foo as any[]).slice(1, bar.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo!.slice(1, bar!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = string.slice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = string.slice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = string.slice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "foo.slice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo.length` as the `end` argument is unnecessary.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.slice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo.length` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.slice(1, foo.length,)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.slice(1,)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo.length` as the `end` argument is unnecessary.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.slice(1, (( foo.length )))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo.length` as the `end` argument is unnecessary.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.slice(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo?.length` as the `end` argument is unnecessary.", Line: 1, Column: 14, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.slice(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `foo?.length` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.slice(1, Infinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Infinity` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.slice(1, Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.bar.slice(1, foo.bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.bar.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as any[]).slice(1, (foo as any[]).length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as any[]).slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 25, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array as string[])?.slice(1, (array as string[])?.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(array as string[])?.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `…?.length` as the `end` argument is unnecessary.", Line: 1, Column: 31, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as number[]).slice(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as number[]).slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Infinity` as the `end` argument is unnecessary.", Line: 1, Column: 28, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(bar as any[]).slice(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(bar as any[]).slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.", Line: 1, Column: 25, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.slice(1, foo!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.slice(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Infinity` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "bar!.slice(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"bar!.slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(string: string) { return string.slice(1, string.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(string: string) { return string.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `string.length` as the `end` argument is unnecessary.", Line: 1, Column: 53, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(bytes: Uint8Array) { return bytes.slice(1, bytes.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(bytes: Uint8Array) { return bytes.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `bytes.length` as the `end` argument is unnecessary.", Line: 1, Column: 55, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = string.slice(1, string.length);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = string.slice(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `string.length` as the `end` argument is unnecessary.", Line: 2, Column: 29, EndLine: 2, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = string.slice(1, Infinity);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = string.slice(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Infinity` as the `end` argument is unnecessary.", Line: 2, Column: 29, EndLine: 2, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = string.slice(1, Number.POSITIVE_INFINITY);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = string.slice(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Number.POSITIVE_INFINITY` as the `end` argument is unnecessary.", Line: 2, Column: 29, EndLine: 2, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
