// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-length-check.js
package no_useless_length_check_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_length_check"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Upstream test group 1: Semicolons.
func TestNoUselessLengthCheckSemicolons(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "foo()\narray.length > 0 && (array).some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;(array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\narray.length === 0 || (array).every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;(array).every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const previous = value\narray.length !== 0 && (array.some)(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const previous = value\n;(array.some)(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo() // keep\narray.length > 0 && (array).some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo() // keep\n;(array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo();\narray.length > 0 && (array).some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo();\n(array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (ready)\narray.length > 0 && (array).some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (ready)\n(array).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\narray.length > 0 && (array as number[]).some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;(array as number[]).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\narray.length > 0 && (array).some(createPredicate())", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 7, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "foo()\n;(array).some(createPredicate())"}}},
		}},
	})
}

// Upstream test group 2: EditSafety.
func TestNoUselessLengthCheckEditSafety(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "array.length > 0 && array.some(createPredicate())", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(createPredicate())"}}},
		}},
		{Code: "array.length > 0 && array.some(predicate)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(predicate)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length === 0 || array.every(value => value > 0)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(value => value > 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(value => value > 0) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(value => value > 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 41, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Upstream autofixes this case, but a property getter can mutate the array.
		{Code: "array.every(value => value.isValid) || array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 46, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.every(value => value.isValid)"}}},
		}},
		{Code: "array.some(value => array.pop()) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 43, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(value => array.pop())"}}},
		}},
		{Code: "array.some(value => { return value > 0; }) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(value => { return value > 0; })"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.some(predicate) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 32, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(predicate)"}}},
		}},
		{Code: "array.every(predicate) || array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 33, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.every(predicate)"}}},
		}},
		{Code: "array.some(predicate) && array.length > 0 && array.some(predicate)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(predicate) && array.some(predicate)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 32, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Upstream test group 3: Snapshots.
func TestNoUselessLengthCheckSnapshots(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "if (array.length === 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length >= 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 1) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length != 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0 && condition) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for (const element of otherArray) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { before(); for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for (const element of array) {} after(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for (const element of array) {} } else { fallback(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for (const index in array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for (let index = 0; index < array.length; index++) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { while (condition) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array.length > 0) { for await (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array?.length > 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array[length] > 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (array[\"length\"] > 0) { for (const element of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = []; if (array.length > 0) { for (const array of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = []; if (array.length > 0) { for (const {value: array} of array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {array: []}; if (object.array.length > 0) { for (const object of object.array) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function iterate(array: Set<number> & {length: number}) { if (array.length > 0) { for (const element of array) {} } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f(foo: Set<number>) { return foo.length === 0 || foo.every(Boolean); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class A { #every() {} foo() { return this.length === 0 || this.#every(Boolean); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 ?? array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 && array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(array.length === 0) + (array.every(Boolean))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 1 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === \"0\" || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0. || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0x0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length == 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 === array.length || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array?.length === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.notLength === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array[length] === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || array.every?.(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || array?.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || array.every", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || array[every](Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array1.length === 0 || array2.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 ?? array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 || array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(array.length !== 0) - (array.some(Boolean))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 1 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== \"0\" && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0. && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0x0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length <= 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length != 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 !== array.length && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array?.length !== 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.notLength !== 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array[length] !== 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 && array.some?.(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 && array?.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 && array.some", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 && array.notSome(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length !== 0 && array[some](Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array1.length !== 0 && array2.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 ?? array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 || array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(array.length > 0) - (array.some(Boolean))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 1 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > \"0\" && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0. && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0x0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length >= 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 > array.length && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 < array.length && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array?.length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.notLength > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && array.some?.(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && array?.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && array.some", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && array.notSome(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && array[some](Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array1.length > 0 && array2.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (\n\tfoo &&\n\tarray.length !== 0 &&\n\tbar &&\n\tarray.some(Boolean)\n) {\n\t// ...\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(foo && array.length === 0) || array.every(Boolean) && foo", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length === 0 || (array.every(Boolean) && foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(foo || array.length > 0) && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "array.length > 0 && (array.some(Boolean) || foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "array.length === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length !== 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "((\n\t((\n\t\t(( array )).length\n\t)) === (( 0 ))\n\t||\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 3, Column: 15, EndLine: 4, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n\t||\n\t((\n\t\t(( array )).length\n\t)) === (( 0 ))\n))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 7, Column: 15, EndLine: 8, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))"}}},
		}},
		{Code: "if ((( array.length > 0 )) && array.some(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (array.some(Boolean));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (\n\tarray.length !== 0 &&\n\tarray.some(Boolean) &&\n\tfoo\n) {\n\t// ...\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (\n\tarray.some(Boolean) &&\n\tfoo\n) {\n\t// ...\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 8, EndLine: 2, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array.length === 0 || array.every(Boolean)) || foo", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array.every(Boolean)) || foo"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo || (array.length === 0 || array.every(Boolean))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo || (array.every(Boolean))"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 15, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array.length > 0 && array.some(Boolean)) && foo", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array.some(Boolean)) && foo"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo && (array.length > 0 && array.some(Boolean))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo && (array.some(Boolean))"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.every(Boolean) || array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 31, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.every(Boolean)"}}},
		}},
		{Code: "array.some(Boolean) && array.length !== 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 30, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(Boolean)"}}},
		}},
		{Code: "array.some(Boolean) && array.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 30, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.some(Boolean)"}}},
		}},
		{Code: "foo && array.length > 0 && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo && array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo || array.length === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo || array.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 14, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo || array.length === 0) || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(foo) || array.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 15, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length === 0 || (array.every(Boolean) || foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array.every(Boolean) || foo)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(foo && array.length > 0) && array.some(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(foo) && array.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length > 0 && (array.some(Boolean) && foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array.some(Boolean) && foo)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.every(Boolean) || array.length === 0 || array.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(Boolean) || array.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 31, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length === 0 || array.every(Boolean) || array.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(Boolean) || array.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 53, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array.length === 0 || array.every(Boolean)"}}},
		}},
		{Code: "array1.every(Boolean)\n|| (( array1.length === 0 || array2.length === 0 )) // Both useless\n|| array2.every(Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array1.every(Boolean)\n|| (( array1.length === 0 )) // Both useless\n|| array2.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 2, Column: 14, EndLine: 2, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "remove-length-check", Output: "array1.every(Boolean)\n|| (( array2.length === 0 )) // Both useless\n|| array2.every(Boolean)"}}},
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 2, Column: 37, EndLine: 2, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function isUselessLengthCheckNode({node, operator, siblings}) {\n\treturn (\n\t\t(\n\t\t\toperator === '||' &&\n\t\t\tzeroLengthChecks.has(node) &&\n\t\t\tsiblings.length > 0 &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarrayEveryCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t) ||\n\t\t(\n\t\t\toperator === '&&' &&\n\t\t\tnonZeroLengthChecks.has(node) &&\n\t\t\tsiblings.length > 0 &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarraySomeCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t)\n\t);\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function isUselessLengthCheckNode({node, operator, siblings}) {\n\treturn (\n\t\t(\n\t\t\toperator === '||' &&\n\t\t\tzeroLengthChecks.has(node) &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarrayEveryCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t) ||\n\t\t(\n\t\t\toperator === '&&' &&\n\t\t\tnonZeroLengthChecks.has(node) &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarraySomeCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t)\n\t);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 6, Column: 13, EndLine: 6, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 15, Column: 13, EndLine: 15, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array as any[]).length === 0 || (array as any[]).every(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array as any[]).every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(array as any[]).length > 0 && (array as any[]).some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array as any[]).some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array!.length === 0 || array!.every(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array!.every(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array!.length > 0 && array!.some(Boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array!.some(Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(array: Int8Array) { return array.length === 0 || array.every(Boolean); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function f(array: Int8Array) { return array.every(Boolean); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 45, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\n\tuse(element);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length !== 0) {\n\tfor (const element of array) {\n\t\t// Do work.\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\n\t// Do work.\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function iterate(array) {\n\tif (array.length > 0) {\n\t\tfor (const element of array) {\n\t\t\tuse(element);\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function iterate(array) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 2, Column: 12, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) for (const element of array) { use(element); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const element of array) use(element); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) use(element);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (((array.length > 0))) { for (const element of array) { use(element); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (object.array.length > 0) { for (const element of object.array) { use(element); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of object.array) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const [key, value] of array) { use(key, value); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const [key, value] of array) { use(key, value); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const element of array) { /* Keep this comment. */ use(element); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) { /* Keep this comment. */ use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n\tfor (const element of array) {\n\t\t/*\n\t\tKeep this indentation.\n\t\t*/\n\t\tuse(element);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\n\t/*\n\tKeep this indentation.\n\t*/\n\tuse(element);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0)\n\tfor (const element of array) {\n\t\tuse(element);\n\t}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\n\tuse(element);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function iterate(array) { if (array.length > 0) { for (var array of array) {} } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function iterate(array) { for (var array of array) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 37, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function iterate(array) {\n  if (array.length > 0) {\n    for (const element of array) {\n      use(element);\n    }\n  }\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function iterate(array) {\n  for (const element of array) {\n    use(element);\n  }\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 2, Column: 13, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if ((array as any[]).length > 0) { for (const element of (array as any[])) { use(element); } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of (array as any[])) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 22, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if ((<any[]>array).length > 0) { for (const element of (<any[]>array)) { use(element); } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of (<any[]>array)) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array!.length !== 0) { for (const element of array!) { use(element); } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array!) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 12, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if ((array satisfies any[]).length > 0) { for (const element of (array satisfies any[])) { use(element); } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of (array satisfies any[])) { use(element); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 29, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function iterate(array: Int8Array) { if (array.length > 0) { for (const element of array) { use(element); } } }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function iterate(array: Int8Array) { for (const element of array) { use(element); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 48, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (arrays.length > 0) {\n\tfor (const array of arrays) {\n\t\tif (array.length > 0) {\n\t\t\tfor (const element of array) {\n\t\t\t\tuse(element);\n\t\t\t}\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const array of arrays) {\n\tif (array.length > 0) {\n\t\tfor (const element of array) {\n\t\t\tuse(element);\n\t\t}\n\t}\n}", "for (const array of arrays) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 12, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 3, Column: 13, EndLine: 3, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Upstream test group 4: LoopFixSafety.
func TestNoUselessLengthCheckLoopFixSafety(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "if (array.length > 0) { /* Keep this comment. */ for (const element of array) { use(element); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length /* Keep this comment. */ > 0) { for (const element of array) { use(element); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n\tfor (const element of array) {\n\t\tconst value = `\n\t\t\tKeep this indentation.\n\t\t`;\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n\tfor (const element of array) {\n\t\t/*\nKeep this column.\n\t\t*/\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\n\t/*\nKeep this column.\n\t*/\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\n\tfor (const element of array) {\nuse(element);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\nuse(element);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) { for (const element of array) {\n\tuse(element);\n} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "before(); if (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\r\n\tfor (const element of array) {\r\n\r\n\t\tuse(element);\r\n\t}\r\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\r\n\r\n\tuse(element);\r\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length > 0) {\r\n\tfor (const element of array) {\r\n\r\n\t\tuse(element);\n\t\tuseAgain(element);\r\n\t}\r\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (const element of array) {\r\n\r\n\tuse(element);\n\tuseAgain(element);\r\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Upstream test group 5: SpaceIndentation.
func TestNoUselessLengthCheckSpaceIndentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "function run() {\n  if (array.length > 0) {\n    for (const element of array) {\n      use({\n        element,\n      });\n    }\n  }\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function run() {\n  for (const element of array) {\n    use({\n      element,\n    });\n  }\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 2, Column: 13, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Upstream test group 6: Comments.
func TestNoUselessLengthCheckComments(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "if (array.length !== 0 && /* keep */ array.some(Boolean) && foo) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.length !== 0 && array.some(Boolean) && foo) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (array.some(Boolean) && foo) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (array.some(Boolean) /* keep */ && array.length !== 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 45, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (/* keep */ array.length !== 0 && array.some(Boolean)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (/* keep */ array.some(Boolean)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 22, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Upstream test group 7: Syntax.
func TestNoUselessLengthCheckSyntax(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "(array satisfies number[]).length > 0 && (array satisfies number[]).some(value => value > 0)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(array satisfies number[]).some(value => value > 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 28, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(<number[]>array).every(value => value > 0) || (<number[]>array).length === 0", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"(<number[]>array).every(value => value > 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 66, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const element = <div>{array.length > 0 && array.some(value => value > 0)}</div>;", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const element = <div>{array.some(value => value > 0)}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 1, Column: 29, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "array.length === 0 || array.every(/* keep */ Boolean)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"array.every(/* keep */ Boolean)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Examples from the pinned rule documentation.
func TestNoUselessLengthCheckDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_length_check.NoUselessLengthCheckRule, []rule_tester.ValidTestCase{
		{Code: "// ✅\nif (array.every(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nif (array.some(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nif (array.some(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nfor (const element of array) {\n\tuse(element);\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nconst isAllTruthyOrEmpty = array.every(Boolean);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nif (array.length === 0 || anotherCheck() || array.every(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nconst isNonEmptyAllTruthyArray = array.length > 0 && array.every(Boolean);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nconst isEmptyArrayOrAllTruthy = array.length === 0 || array.some(Boolean);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "// ❌\nif (array.length === 0 || array.every(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nif (array.every(Boolean));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 2, Column: 11, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nif (array.length !== 0 && array.some(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nif (array.some(Boolean));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 11, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nif (array.length > 0 && array.some(Boolean));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nif (array.some(Boolean));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "non-zero", Message: "The non-empty check is useless as `Array#some()` returns `false` for an empty array.", Line: 2, Column: 11, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nif (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nfor (const element of array) {\n\tuse(element);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "for-of", Message: "The non-empty check is useless as `for…of` does not iterate over an empty array.", Line: 2, Column: 11, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst isAllTruthyOrEmpty = array.length === 0 || array.every(Boolean);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isAllTruthyOrEmpty = array.every(Boolean);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero", Message: "The empty check is useless as `Array#every()` returns `true` for an empty array.", Line: 2, Column: 34, EndLine: 2, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
