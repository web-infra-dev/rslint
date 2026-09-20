// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-unnecessary-array-splice-count.js
package no_unnecessary_array_splice_count_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_array_splice_count"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUnnecessaryArraySpliceCountUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule, []rule_tester.ValidTestCase{
		{Code: "foo.splice?.(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(foo.length, 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice()", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, foo.length - 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, foo.length, extraArgument)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(...[1], foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.not_splice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "new foo.splice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "splice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, foo.notLength)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo[splice](1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, foo[length])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.splice(1, bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.splice(1, NotInfinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.splice(1, Number.NOT_POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.splice(1, Not_Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.splice(1, Number?.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo().splice(1, foo().length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(foo as any[]).splice(1, bar.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo!.splice(1, bar!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced?.(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(foo.length, 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced()", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, foo.length - 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, foo.length, extraArgument)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(...[1], foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.not_toSpliced(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "new foo.toSpliced(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "toSpliced(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, foo.notLength)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo[toSpliced](1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, foo[length])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.toSpliced(1, bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.toSpliced(1, NotInfinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.toSpliced(1, Number.NOT_POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.toSpliced(1, Not_Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo?.toSpliced(1, Number?.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo().toSpliced(1, foo().length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(foo as any[]).toSpliced(1, bar.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo!.toSpliced(1, bar!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(foo: {splice(start: number, deleteCount: number): void; length: number}) { foo.splice(1, foo.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(foo: Set<number>) { foo.splice(1, Infinity); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ❌\nconst foo = array.toSpliced(1, string.length);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = array.toSpliced(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = array.toSpliced(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = array.toSpliced(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ❌\narray.splice(1, string.length);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\narray.splice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\narray.splice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\narray.splice(1);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "foo.splice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.splice(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.splice(1, foo.length,)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.splice(1,)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.splice(1, (( foo.length )))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.splice(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo?.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.splice(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo?.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.splice(1, Infinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.splice(1, Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.bar.splice(1, foo.bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.bar.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as any[]).splice(1, (foo as any[]).length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as any[]).splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 26, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array as string[])?.splice(1, (array as string[])?.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(array as string[])?.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `…?.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 32, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as number[]).splice(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as number[]).splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 29, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(bar as any[]).splice(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(bar as any[]).splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 26, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.splice(1, foo!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.splice(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "bar!.splice(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"bar!.splice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 16, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.toSpliced(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.toSpliced(1, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.toSpliced(1, foo.length,)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.toSpliced(1,)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.toSpliced(1, (( foo.length )))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.toSpliced(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo?.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.toSpliced(1, foo?.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo?.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.toSpliced(1, Infinity)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo?.toSpliced(1, Number.POSITIVE_INFINITY)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo?.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo.bar.toSpliced(1, foo.bar.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo.bar.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 22, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as any[]).toSpliced(1, (foo as any[]).length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as any[]).toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 29, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array as string[])?.toSpliced(1, (array as string[])?.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(array as string[])?.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `…?.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 35, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo as number[]).toSpliced(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(foo as number[]).toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 1, Column: 32, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(bar as any[]).toSpliced(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(bar as any[]).toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.", Line: 1, Column: 29, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.toSpliced(1, foo!.length)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo!.toSpliced(1, Infinity)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"foo!.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "bar!.toSpliced(1, Number.POSITIVE_INFINITY)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"bar!.toSpliced(1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(foo: number[]) { foo.splice(1, foo.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(foo: number[]) { foo.splice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 43, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(foo: readonly number[]) { foo.toSpliced(1, foo.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(foo: readonly number[]) { foo.toSpliced(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 55, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(foo: Uint8Array) { foo.splice(1, foo.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(foo: Uint8Array) { foo.splice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `foo.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 45, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(foo: Uint8Array) { foo.toSpliced(1, Infinity); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(foo: Uint8Array) { foo.toSpliced(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 1, Column: 48, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = array.toSpliced(1, Infinity);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = array.toSpliced(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 2, Column: 32, EndLine: 2, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = array.toSpliced(1, Number.POSITIVE_INFINITY);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = array.toSpliced(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `skipCount` argument is unnecessary.", Line: 2, Column: 32, EndLine: 2, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\narray.splice(1, Infinity);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\narray.splice(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `deleteCount` argument is unnecessary.", Line: 2, Column: 17, EndLine: 2, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\narray.splice(1, Number.POSITIVE_INFINITY);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\narray.splice(1);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Number.POSITIVE_INFINITY` as the `deleteCount` argument is unnecessary.", Line: 2, Column: 17, EndLine: 2, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
