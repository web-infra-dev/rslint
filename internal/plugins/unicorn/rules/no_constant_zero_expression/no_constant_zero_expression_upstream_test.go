// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-constant-zero-expression.js
package no_constant_zero_expression_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_constant_zero_expression"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstantZeroExpressionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_constant_zero_expression.NoConstantZeroExpressionRule, []rule_tester.ValidTestCase{
		// Snapshot cases
		{Code: "x * 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x + 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x - 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x / 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x | 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x ^ 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x << 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x >> 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x / 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x % 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 % x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x ** 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0 ** x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x - x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x ^ x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "0n & x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x & 0n;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x * 0n;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "-0 * x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x * -0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x === 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x !== 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Documentation examples
		{Code: "const total = 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const flags = 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const ratio = 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Static template cases
		{Code: "Math[`PI`] * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "`abc`.length * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "const text = `abc`; text.length * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 21, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "const text = `abc`; 0;"}}},
		}},
		// Snapshot cases
		{Code: "x * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "0 * x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "0.0 * x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		{Code: "x & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "0 & x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "0 / x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "const y = value * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
		}},
		{Code: "a + b * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 5, EndLine: 1, EndColumn: 10},
		}},
		{Code: "(a + b) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
		}},
		{Code: "(x) & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		{Code: "foo.bar * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
		}},
		{Code: "foo() * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
		}},
		{Code: "const object = {value: 1};\nObject.defineProperty(object, 'value', {get() { return 1n; }});\nobject.value * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 3, Column: 1, EndLine: 3, EndColumn: 17},
		}},
		{Code: "const object = {value: 1};\nObject.defineProperty(object, 'value', {get() { return 1n; }});\nconst value = object.value;\nvalue * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 4, Column: 1, EndLine: 4, EndColumn: 10},
		}},
		{Code: "const a = 5;\na * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "const a = 5;\n0;"}}},
		}},
		{Code: "5 * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 * 5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "7 & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 & 3;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 / 5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "-5 * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "0 / -5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "0 / 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "5 /* keep */ * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
		}},
		{Code: "(x as number) * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
		}},
		{Code: "x! & 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		// Unsafe forward alias
		{Code: "const alias = condition; var condition = true; (alias ? 1 : NaN) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 48, EndLine: 1, EndColumn: 69},
		}},
		// Documentation examples
		{Code: "const total = price * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
		}},
		{Code: "const flags = mask & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
		}},
		{Code: "const ratio = 0 / count;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
		}},
	})
}
