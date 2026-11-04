// Additional literal, scope, AST and edit-boundary coverage for Unicorn v79.0.0.
package no_constant_zero_expression_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_constant_zero_expression"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstantZeroExpressionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_constant_zero_expression.NoConstantZeroExpressionRule, []rule_tester.ValidTestCase{
		{Code: "const zero = 0; x * zero;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x * +0; +0 / x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x * (1 - 1);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x *= 0; x &= 0; x /= 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x && 0; x >>> 0; 0 | x; 0 ^ x; 0 + x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "\"0\" * x; x & false; null / x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x * (0 as number);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(0 as number) / x;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x & 0!;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(0 satisfies number) * x;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x & <number>0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(0 as const) * 5;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const value = <number>0 * 5;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Numeric literal forms, including underflow to zero.
		{Code: "x * ((0));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
		}},
		{Code: "0x0 & 7;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "5 * 0b0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0o0 / 5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0e3 * 5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0.0_0 * 5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "1e-999 * x;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		// Signed zero, NaN and JavaScript coercion.
		{Code: "0 * -5;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "-0 & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 / -0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "Infinity * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "0 / Infinity;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 / -Infinity;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "NaN & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "false * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "null * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "undefined & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "'' * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "'abc' * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
		}},
		{Code: "[] * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "0 / 2n;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		// Member access and scope-sensitive suggestion safety.
		{Code: "obj?.value * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "obj['value'] & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
		}},
		{Code: "class C { #value = 3; f() { return this.#value * 0; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 36, EndLine: 1, EndColumn: 51},
		}},
		{Code: "Math.PI * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "Math['PI'] * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "const Math = {PI: 3}; Math.PI * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 23, EndLine: 1, EndColumn: 34},
		}},
		{Code: "function f(Math) { return Math.PI * 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 27, EndLine: 1, EndColumn: 38},
		}},
		{Code: "Math.random() * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
		}},
		{Code: "let a = 3; a * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
		}},
		{Code: "const a = 3; const b = a; b * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 27, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "const a = 3; const b = a; 0;"}}},
		}},
		{Code: "const a = 5; { const a = -5; a * 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35},
		}},
		{Code: "a * 0; const a = 3;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
		}},
		{Code: "const a = b; var b = 3; a * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 25, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const a = [1]; a[0] * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
		}},
		// Side effects are preserved even on unselected paths.
		{Code: "(foo(), 5) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "(value = 5) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
		}},
		{Code: "(false && foo()) & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "(true ? 5 : foo()) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
		}},
		// Comment boundaries and UTF-16 diagnostic ranges.
		{Code: "/* before */ (5 * 0) /* after */;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 15, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "/* before */ (0) /* after */;"}}},
		}},
		{Code: "(5 /* inside */) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "5 * (/* keep */ 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
		}},
		{Code: "5 * 0 // trailing\n;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0 // trailing\n;"}}},
		}},
		{Code: "// line\n5 *\n0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 2, Column: 1, EndLine: 3, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "// line\n0;"}}},
		}},
		{Code: "'😀'; 5 * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 7, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "'😀'; 0;"}}},
		}},
		// Authored TypeScript wrappers and JavaScript JSDoc wrappers.
		{Code: "(5 as number) * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "(5 satisfies number) * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "5! & 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "(<number>5) * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "/** @type {number} */ (5) * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "/** @type {number} */ 0;"}}},
		}},
		{Code: "5 * /** @type {number} */ (0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const view = <div>{5 * 0}</div>;", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 20, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "const view = <div>{0}</div>;"}}},
		}},
		// Nested expressions each report their own diagnostic.
		{Code: "0 * 5 * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0 * 0;"}}},
		}},
		{Code: "function f(Infinity, NaN) { 0 / Infinity; NaN & 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41},
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
		}},
		{Code: "const Infinity = 5; 0 / Infinity;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "const Infinity = 5; 0;"}}},
		}},
		// Authored TS receivers, optional chains and execution boundaries.
		{Code: "(Math as any).PI * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "(Math!).PI * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "(Math satisfies object).PI * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
		}},
		{Code: "(5 * 0) as number;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 2, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "(0) as number;"}}},
		}},
		{Code: "({value: 5} as const).value * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
		}},
		{Code: "const text = `abc` as const; text.length * 0;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 30, EndLine: 1, EndColumn: 45},
		}},
		{Code: "null?.value & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
		}},
		{Code: "const object = null; object?.value & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 22, EndLine: 1, EndColumn: 39},
		}},
		{Code: "({value: 5}).value * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "0;"}}},
		}},
		{Code: "({get value() { return 5; }}).value & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
		}},
		{Code: "let value = 5; const alias = value; alias * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 37, EndLine: 1, EndColumn: 46},
		}},
		{Code: "0 * Math[\"P\" + \"I\"];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
		}},
		{Code: "const value = Symbol(); value & 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
		}},
		{Code: "class C { static value = 5 * 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "class C { static value = 0; }"}}},
		}},
		{Code: "async function f() { return (await value) * 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 29, EndLine: 1, EndColumn: 46},
		}},
		{Code: "function* f() { return (yield 5) & 0; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
		}},
		{Code: "(5 * 0)();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 2, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "(0)();"}}},
		}},
		{Code: "(5 * 0).toString();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 2, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "(0).toString();"}}},
		}},
		{Code: "/** @type {Math} */ (Math).PI * 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 21, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-constant-zero-expression/suggestion", Output: "/** @type {Math} */ 0;"}}},
		}},
		{Code: "0 * /** @type {Math} */ (Math).PI;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-constant-zero-expression/error", Message: "This operation always evaluates to `0` (or `-0`/`NaN`), which is likely a mistake.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
		}},
	})
}

func TestNoConstantZeroExpressionSuggestionsFollowDemand(t *testing.T) {
	for _, test := range []struct{ code, output string }{
		{"5 * 0", "0"},
		{"-0 & 0", "0"},
		{"-5 * 0", ""},
		{"5 /* keep */ * 0", ""},
		{"foo() * 0", ""},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_constant_zero_expression.NoConstantZeroExpressionRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_constant_zero_expression.NoConstantZeroExpressionRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if diagnostic.FixesPtr != nil {
					t.Errorf("demand %d produced an autofix", demand)
				}
				wantSuggestion := test.output != "" && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestion {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced unexpected suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				suggestion := (*diagnostic.Suggestions)[0]
				if suggestion.Message.Id != "no-constant-zero-expression/suggestion" || suggestion.Message.Description != "Replace this expression with `0`." {
					t.Errorf("demand %d: wrong suggestion message", demand)
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, (*diagnostic.Suggestions)[:1])
				if !fixed || output != test.output {
					t.Errorf("demand %d: unexpected suggestion %q", demand, output)
				}
			}
		})
	}
}
