// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-global-number-constants.js
package prefer_global_number_constants_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_global_number_constants"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferGlobalNumberConstantsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_number_constants.PreferGlobalNumberConstantsRule, []rule_tester.ValidTestCase{
		{Code: "const foo = NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = Infinity;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = -Infinity;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = Number.MAX_SAFE_INTEGER;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const foo = object.Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "Number.NaN = 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "Number.POSITIVE_INFINITY ||= 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "[Number.NEGATIVE_INFINITY] = [];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const Number = {\n\tNaN: 1,\n\tPOSITIVE_INFINITY: 2,\n\tNEGATIVE_INFINITY: -2,\n};\nconst foo = Number.NaN + Number.POSITIVE_INFINITY + Number.NEGATIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function foo() {\n\tconst NaN = 1;\n\tconst value = Number.NaN;\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function foo() {\n\tconst Infinity = 1;\n\tconst positive = Number.POSITIVE_INFINITY;\n\tconst negative = Number.NEGATIVE_INFINITY;\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const NaN = 1; const value = Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const Infinity = 1; const positive = Number.POSITIVE_INFINITY; const negative = Number.NEGATIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = NaN;\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = Infinity;\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nconst foo = -Infinity;\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const foo = Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = window.Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = Number[\"NaN\"];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = Number.POSITIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = Infinity;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = Number.NEGATIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = Number.NEGATIVE_INFINITY.toString();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = {value: Number.NaN};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = {value: NaN};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = {[Number.POSITIVE_INFINITY]: Number.NEGATIVE_INFINITY};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const foo = {[Infinity]: Number.NEGATIVE_INFINITY};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 1, Column: 42, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = Number /* comment */ .NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = Number.NaN;\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = NaN;\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 2, Column: 13, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = Number.POSITIVE_INFINITY;\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nconst foo = Infinity;\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 2, Column: 13, EndLine: 2, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nconst foo = Number.NEGATIVE_INFINITY;\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 2, Column: 13, EndLine: 2, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
