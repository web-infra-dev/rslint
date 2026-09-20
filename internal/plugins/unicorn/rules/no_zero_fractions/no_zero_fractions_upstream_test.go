// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-zero-fractions.js
package no_zero_fractions_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_zero_fractions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoZeroFractionsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_zero_fractions.NoZeroFractionsRule, []rule_tester.ValidTestCase{
		{Code: "const foo = \"123.1000\"", FileName: "case.js", Options: []any{}},
		{Code: "foo(\"123.1000\")", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1 + 2", FileName: "case.js", Options: []any{}},
		{Code: "const foo = -1", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123123123", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1.1", FileName: "case.js", Options: []any{}},
		{Code: "const foo = -1.1", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123123123.4", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1e3", FileName: "case.js", Options: []any{}},
		{Code: "1 .toString()", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = -1;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123_456;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123.111;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123e20;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1.1;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = -1.1;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 123.456;", FileName: "case.js", Options: []any{}},
		{Code: "const foo = 1e3;", FileName: "case.js", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const foo = 1.0", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.0 + 1", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1 + 1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo(1.0 + 1)", FileName: "case.js", Options: []any{}, Output: []string{"foo(1 + 1)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 6, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.00", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.00000", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.10", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1.1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 16, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = -1.0", FileName: "case.js", Options: []any{}, Output: []string{"const foo = -1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 123123123.0", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 123123123"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 22, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 123.11100000000", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 123.111"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 20, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a[.0]", FileName: "case.js", Options: []any{}, Output: []string{"a[0]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 4, EndLine: 1, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = +1.", FileName: "case.js", Options: []any{}, Output: []string{"const foo = +1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = -1.", FileName: "case.js", Options: []any{}, Output: []string{"const foo = -1"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.e10", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1e10"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = +1.e-10", FileName: "case.js", Options: []any{}, Output: []string{"const foo = +1e-10"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = -1.e+10", FileName: "case.js", Options: []any{}, Output: []string{"const foo = -1e+10"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = (1.).toString()", FileName: "case.js", Options: []any{}, Output: []string{"const foo = (1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.0E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.0E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.123_000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.123E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.123_000E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.123E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e+1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e+1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e+1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e-1;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e-1;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e-1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e+0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e+0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e+0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e-0;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e-0;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e-0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e+10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e+10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e+10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400e-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400e-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4e-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400E-10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400E-10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4E-10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+123_000.000_400E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"+123_000.000_4E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "-123_000.000_400E-10_10;", FileName: "case.js", Options: []any{}, Output: []string{"-123_000.000_4E-10_10;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "1.00.toFixed(2)", FileName: "case.js", Options: []any{}, Output: []string{"(1).toFixed(2)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "1.00 .toFixed(2)", FileName: "case.js", Options: []any{}, Output: []string{"(1) .toFixed(2)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(1.00).toFixed(2)", FileName: "case.js", Options: []any{}, Output: []string{"(1).toFixed(2)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 3, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "1.00?.toFixed(2)", FileName: "case.js", Options: []any{}, Output: []string{"(1)?.toFixed(2)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log()\n1..toString()", FileName: "case.js", Options: []any{}, Output: []string{"console.log()\n;(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log()\na[1.].toString()", FileName: "case.js", Options: []any{}, Output: []string{"console.log()\na[1].toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 2, Column: 4, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log()\n1.00e10.toString()", FileName: "case.js", Options: []any{}, Output: []string{"console.log()\n1e10.toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log()\na[1.00e10].toString()", FileName: "case.js", Options: []any{}, Output: []string{"console.log()\na[1e10].toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 4, EndLine: 2, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a = .0;", FileName: "case.js", Options: []any{}, Output: []string{"a = 0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 6, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a = .0.toString()", FileName: "case.js", Options: []any{}, Output: []string{"a = (0).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 6, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo(){return.0}", FileName: "case.js", Options: []any{}, Output: []string{"function foo(){return 0}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo(){return.0.toString()}", FileName: "case.js", Options: []any{}, Output: []string{"function foo(){return (0).toString()}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo(){return.0+.1}", FileName: "case.js", Options: []any{}, Output: []string{"function foo(){return 0+.1}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log()\n.0.toString()", FileName: "case.js", Options: []any{}, Output: []string{"console.log()\n;(0).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.0;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 1.;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "dangling-dot", Message: "Don't use a dangling dot in the number.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = -1.0;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = -1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 123_456.000_000;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 123_456;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 20, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 123.111000000;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 123.111;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = 123.00e20;", FileName: "case.js", Options: []any{}, Output: []string{"const foo = 123e20;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
