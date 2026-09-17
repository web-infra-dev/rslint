// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-negation-in-equality-check.js
package no_negation_in_equality_check_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_negation_in_equality_check"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNegationInEqualityCheckUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negation_in_equality_check.NoNegationInEqualityCheckRule, []rule_tester.ValidTestCase{
		{Code: "!foo instanceof bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "+foo === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!(foo === bar)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!!foo === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!!!foo === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "foo === !bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nif (foo !== bar) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "// ✅\nif (!(foo === bar)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "!foo === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo !== bar"}}},
		}},
		{Code: "!foo !== bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo === bar"}}},
		}},
		{Code: "!foo == bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo != bar"}}},
		}},
		{Code: "!foo != bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo == bar"}}},
		}},
		{Code: "function x() {\n\treturn!foo === bar;\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 8, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function x() {\n\treturn foo !== bar;\n}"}}},
		}},
		{Code: "function x() {\n\treturn!\n\t\tfoo === bar;\n\tthrow!\n\t\tfoo === bar;\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 8, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function x() {\n\treturn  (\n\t\tfoo !== bar);\n\tthrow!\n\t\tfoo === bar;\n}"}}},
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 4, Column: 7, EndLine: 4, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function x() {\n\treturn!\n\t\tfoo === bar;\n\tthrow  (\n\t\tfoo !== bar);\n}"}}},
		}},
		{Code: "foo\n!(a) === b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;(a) !== b"}}},
		}},
		{Code: "foo\n![a, b].join('') === c", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;[a, b].join('') !== c"}}},
		}},
		{Code: "foo\n! [a, b].join('') === c", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n; [a, b].join('') !== c"}}},
		}},
		{Code: "foo\n!/* comment */[a, b].join('') === c", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;/* comment */[a, b].join('') !== c"}}},
		}},
		{Code: "// ❌\nif (!foo === bar) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 5, EndLine: 2, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "// ❌\nif (foo !== bar) {}"}}},
		}},
		{Code: "// ❌\nif (!foo !== bar) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 5, EndLine: 2, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "// ❌\nif (foo === bar) {}"}}},
		}},
	})
}
