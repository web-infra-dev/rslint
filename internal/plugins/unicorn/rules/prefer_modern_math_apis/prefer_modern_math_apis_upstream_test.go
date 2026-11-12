package prefer_modern_math_apis_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_modern_math_apis"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-modern-math-apis.js
// Every upstream group and documentation example is retained. Expected messages,
// ranges and edits were checked with eslint-plugin-unicorn 79.0.0 / ESLint 10.14.0.
func TestPreferModernMathApisUpstream(t *testing.T) {
	t.Run("upstream line 7", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(Math.E * Math.E);", FileName: "input.js",
					Output: []string{"Math.abs(Math.E);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.LN10 * Math.LN10);", FileName: "input.js",
					Output: []string{"Math.abs(Math.LN10);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.LN2 * Math.LN2);", FileName: "input.js",
					Output: []string{"Math.abs(Math.LN2);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.LOG10E * Math.LOG10E);", FileName: "input.js",
					Output: []string{"Math.abs(Math.LOG10E);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.LOG2E * Math.LOG2E);", FileName: "input.js",
					Output: []string{"Math.abs(Math.LOG2E);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.PI * Math.PI);", FileName: "input.js",
					Output: []string{"Math.abs(Math.PI);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.SQRT1_2 * Math.SQRT1_2);", FileName: "input.js",
					Output: []string{"Math.abs(Math.SQRT1_2);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Math.SQRT2 * Math.SQRT2);", FileName: "input.js",
					Output: []string{"Math.abs(Math.SQRT2);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(-Math.PI) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(-Math.PI);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
					}},
				{Code: "Math.LOG10E * Math.log(-Math.E);", FileName: "input.js",
					Output: []string{"Math.log10(-Math.E);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					}},
				{Code: "const value = Math.PI; Math.sqrt(value * value);", FileName: "input.js",
					Output: []string{"const value = Math.PI; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
					}},
				{Code: "const value = -Math.SQRT2; Math.log(value) / Math.LN10;", FileName: "input.js",
					Output: []string{"const value = -Math.SQRT2; Math.log10(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 55},
					}},
				{Code: "Math.sqrt(Math.PI * Math.PI + Infinity ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(Math.PI, Infinity);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(NaN ** 2 + Math.SQRT2 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(NaN, Math.SQRT2);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "void Math.sqrt(Math.E ** 2 + Math.PI ** 2);", FileName: "input.js",
					Output: []string{"void Math.hypot(Math.E, Math.PI);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
				{Code: "Math.sqrt((Math./* keep */PI) * Math.PI);", FileName: "input.js",
					Output: []string{"Math.abs((Math./* keep */PI));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "delete (Math.log(getValue()) / Math.LN2);", FileName: "input.js",
					Output: []string{"delete (Math.log2(getValue()));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 40},
					}},
				{Code: "delete (Math.LOG10E * Math.log(value));", FileName: "input.js",
					Output: []string{"delete (Math.log10(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 38},
					}},
				{Code: "delete Math.sqrt(1e200 * 1e200);", FileName: "input.js",
					Output: []string{"delete Math.abs(1e200);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
					}},
				{Code: "delete Math.sqrt(3 ** 2 + 4 ** 2);", FileName: "input.js",
					Output: []string{"delete Math.hypot(3, 4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
					}},
				{Code: "delete (condition && Math.log(value) / Math.LN2);", FileName: "input.js",
					Output: []string{"delete (condition && Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 48},
					}},
			})
	})
	t.Run("upstream line 34", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt((Math.PI as number) * Math.PI);", FileName: "input.ts",
					Output: []string{"Math.abs((Math.PI as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = Math.LN2 as number; Math.sqrt(value * value);", FileName: "input.ts",
					Output: []string{"const value = Math.LN2 as number; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44},
					}},
				{Code: "Math.log(-(Math.SQRT2 as number)) / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(-(Math.SQRT2 as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
					}},
			})
	})
	t.Run("upstream line 48", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(Math.E) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Math.E);"}}},
					}},
				{Code: "Math.log(Math.LN2) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(Math.LN2);"}}},
					}},
				{Code: "Math.sqrt(Math.PI ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Math.PI);"}}},
					}},
				{Code: "Math.sqrt(Math.E * Math.E + Math.PI * Math.PI);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Math.E, Math.PI);"}}},
					}},
				{Code: "Math.log(Math.custom) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Math.custom);"}}},
					}},
				{Code: "Math.log(Math.random) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Math.random);"}}},
					}},
				{Code: "Math.sqrt(Math[\"PI\"] * Math[\"PI\"]);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Math[\"PI\"]);"}}},
					}},
				{Code: "Math.log(Math?.PI) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Math?.PI);"}}},
					}},
				{Code: "Math.sqrt(Math.PI * /* keep */ Math.PI);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "delete Math.sqrt(value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "delete Math.abs(value);"}}},
					}},
				{Code: "delete (Math.log(/* keep */ value) / Math.LN2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 46},
					}},
				{Code: "delete (Math.log(value) / Math.LN2).property;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "delete (Math.log2(value)).property;"}}},
					}},
			})
	})
	t.Run("upstream line 67", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(Number.NaN) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(Number.NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					}},
				{Code: "Math.log(Number./* keep */NaN) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(Number./* keep */NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
					}},
				{Code: "Math.log(Number.POSITIVE_INFINITY) * Math.LOG10E;", FileName: "input.js",
					Output: []string{"Math.log10(Number.POSITIVE_INFINITY);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
					}},
				{Code: "Math.LOG2E * Math.log(Number.NEGATIVE_INFINITY);", FileName: "input.js",
					Output: []string{"Math.log2(Number.NEGATIVE_INFINITY);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
					}},
				{Code: "Math.log(Number.MIN_SAFE_INTEGER) / Math.LN10;", FileName: "input.js",
					Output: []string{"Math.log10(Number.MIN_SAFE_INTEGER);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
					}},
				{Code: "Math.log(-Number.MIN_VALUE) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(-Number.MIN_VALUE);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
					}},
				{Code: "Math.log(-Number.NEGATIVE_INFINITY) / Math.LN10;", FileName: "input.js",
					Output: []string{"Math.log10(-Number.NEGATIVE_INFINITY);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
					}},
				{Code: "Math.sqrt(Number.NaN ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(Number.NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.POSITIVE_INFINITY ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(Number.POSITIVE_INFINITY);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.NEGATIVE_INFINITY ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(Number.NEGATIVE_INFINITY);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((-Number.POSITIVE_INFINITY) ** 2);", FileName: "input.js",
					Output: []string{"Math.abs((-Number.POSITIVE_INFINITY));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((+Number.POSITIVE_INFINITY) ** 2);", FileName: "input.js",
					Output: []string{"Math.abs((+Number.POSITIVE_INFINITY));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.EPSILON * Number.EPSILON);", FileName: "input.js",
					Output: []string{"Math.abs(Number.EPSILON);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.MAX_SAFE_INTEGER * Number.MAX_SAFE_INTEGER);", FileName: "input.js",
					Output: []string{"Math.abs(Number.MAX_SAFE_INTEGER);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.MIN_SAFE_INTEGER * Number.MIN_SAFE_INTEGER);", FileName: "input.js",
					Output: []string{"Math.abs(Number.MIN_SAFE_INTEGER);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.MAX_VALUE * Number.MAX_VALUE + Number.NaN ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(Number.MAX_VALUE, Number.NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Number.NEGATIVE_INFINITY ** 2 + 3 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(Number.NEGATIVE_INFINITY, 3);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = Number.NaN; Math.log(value) / Math.LN2;", FileName: "input.js",
					Output: []string{"const value = Number.NaN; Math.log2(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 53},
					}},
				{Code: "const value = -Number.POSITIVE_INFINITY; Math.sqrt(value ** 2);", FileName: "input.js",
					Output: []string{"const value = -Number.POSITIVE_INFINITY; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 51},
					}},
				{Code: "void Math.sqrt(Number.MAX_VALUE ** 2 + Number.MIN_VALUE ** 2);", FileName: "input.js",
					Output: []string{"void Math.hypot(Number.MAX_VALUE, Number.MIN_VALUE);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
			})
	})
	t.Run("upstream line 97", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(Number.NaN as number) / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(Number.NaN as number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
					}},
				{Code: "const value = Number.POSITIVE_INFINITY as number; Math.sqrt(value ** 2);", FileName: "input.ts",
					Output: []string{"const value = Number.POSITIVE_INFINITY as number; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 60},
					}},
				{Code: "Math.log(-(Number.MIN_VALUE as number)) * Math.LOG2E;", FileName: "input.ts",
					Output: []string{"Math.log2(-(Number.MIN_VALUE as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
					}},
			})
	})
	t.Run("upstream line 111", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(Number.MAX_VALUE * Number.MAX_VALUE);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Number.MAX_VALUE);"}}},
					}},
				{Code: "Math.sqrt(Number.MIN_VALUE * Number.MIN_VALUE);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Number.MIN_VALUE);"}}},
					}},
				{Code: "Math.sqrt(Number.EPSILON ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Number.EPSILON);"}}},
					}},
				{Code: "Math.sqrt(Number.NaN ** 2 + Number.POSITIVE_INFINITY ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Number.NaN, Number.POSITIVE_INFINITY);"}}},
					}},
				{Code: "Math.log(Number.EPSILON) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.EPSILON);"}}},
					}},
				{Code: "Math.log(Number.MIN_VALUE) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.MIN_VALUE);"}}},
					}},
				{Code: "Math.log(Number.length) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.length);"}}},
					}},
				{Code: "Math.log(Number.custom) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.custom);"}}},
					}},
				{Code: "Math.log(Number[\"NaN\"]) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number[\"NaN\"]);"}}},
					}},
				{Code: "Math.log(Number?.NaN) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number?.NaN);"}}},
					}},
				{Code: "Math.log(Number.prototype) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.prototype);"}}},
					}},
				{Code: "Math.log(other.NaN) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(other.NaN);"}}},
					}},
				{Code: "Math.log(Number.NaN /* keep */ + value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.NaN /* keep */ + value);"}}},
					}},
				{Code: "Math.sqrt(Number.NaN ** /* keep */ 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
			})
	})
	t.Run("upstream line 132", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(NaN ** 2 + 3 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(NaN, 3);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1e200 * 1e200 + NaN ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(1e200, NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(\"not a number\" ** 2 + 0 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(\"not a number\", 0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = NaN; Math.sqrt((0 ** 2 + value ** 2) + 4 * 4);", FileName: "input.js",
					Output: []string{"const value = NaN; Math.hypot(0, value, 4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
					}},
				{Code: "Math.log(undefined) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(undefined);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					}},
				{Code: "const value = undefined; Math.log(value) * Math.LOG10E;", FileName: "input.js",
					Output: []string{"const value = undefined; Math.log10(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 55},
					}},
				{Code: "Math.sqrt(undefined ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(undefined);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(undefined ** 2 + 4 * 4);", FileName: "input.js",
					Output: []string{"Math.hypot(undefined, 4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "for (Math.log(value) / Math.LN2; condition; update()) {}", FileName: "input.js",
					Output: []string{"for (Math.log2(value); condition; update()) {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 32},
					}},
				{Code: "for (; condition; Math.LOG10E * Math.log(getValue())) {}", FileName: "input.js",
					Output: []string{"for (; condition; Math.log10(getValue())) {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 53},
					}},
				{Code: "for (start(); condition; next(), Math.log(value) / Math.LN10) {}", FileName: "input.js",
					Output: []string{"for (start(); condition; next(), Math.log10(value)) {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 61},
					}},
				{Code: "typeof (Math.log(value) * Math.LOG2E);", FileName: "input.js",
					Output: []string{"typeof (Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 37},
					}},
				{Code: "const type = typeof (Math.LOG10E * Math.log(getValue()));", FileName: "input.js",
					Output: []string{"const type = typeof (Math.log10(getValue()));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 56},
					}},
				{Code: "void (condition ? Math.log(value) / Math.LN2 : fallback());", FileName: "input.js",
					Output: []string{"void (condition ? Math.log2(value) : fallback());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 45},
					}},
				{Code: "void (condition ? fallback() : Math.log(value) / Math.LN10);", FileName: "input.js",
					Output: []string{"void (condition ? fallback() : Math.log10(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 59},
					}},
				{Code: "void (condition && Math.log(value) / Math.LN2);", FileName: "input.js",
					Output: []string{"void (condition && Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46},
					}},
				{Code: "const type = typeof (condition || Math.log(value) * Math.LOG10E);", FileName: "input.js",
					Output: []string{"const type = typeof (condition || Math.log10(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 64},
					}},
				{Code: "const result = (condition ?? Math.log(value) / Math.LN2, next());", FileName: "input.js",
					Output: []string{"const result = (condition ?? Math.log2(value), next());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 56},
					}},
				{Code: "void (first ? second && (next(), Math.log(getValue()) / Math.LN2) : fallback());", FileName: "input.js",
					Output: []string{"void (first ? second && (next(), Math.log2(getValue())) : fallback());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 65},
					}},
				{Code: "for (; condition; flag ? Math.log(value) / Math.LN2 : fallback()) {}", FileName: "input.js",
					Output: []string{"for (; condition; flag ? Math.log2(value) : fallback()) {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 52},
					}},
				{Code: "void Math.sqrt(4 ** 2);", FileName: "input.js",
					Output: []string{"void Math.abs(4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
				{Code: "void Math.sqrt(1e200 * 1e200);", FileName: "input.js",
					Output: []string{"void Math.abs(1e200);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
				{Code: "typeof Math.sqrt(3 * 3 + 4 * 4);", FileName: "input.js",
					Output: []string{"typeof Math.hypot(3, 4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
					}},
				{Code: "void Math.sqrt(NaN ** 2 + Infinity ** 2);", FileName: "input.js",
					Output: []string{"void Math.hypot(NaN, Infinity);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
				{Code: "const result = (Math.sqrt(3 ** 2 + 4 ** 2), next());", FileName: "input.js",
					Output: []string{"const result = (Math.hypot(3, 4), next());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
					}},
			})
	})
	t.Run("upstream line 167", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt((NaN as number) ** 2 + 3 ** 2);", FileName: "input.ts",
					Output: []string{"Math.hypot((NaN as number), 3);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(undefined as undefined) / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(undefined as undefined);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
					}},
				{Code: "for ((Math.log(value) / Math.LN2) as number; condition; update()) {}", FileName: "input.ts",
					Output: []string{"for ((Math.log2(value)) as number; condition; update()) {}"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33},
					}},
				{Code: "typeof ((Math.log(value) / Math.LN2) satisfies number);", FileName: "input.ts",
					Output: []string{"typeof ((Math.log2(value)) satisfies number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 36},
					}},
				{Code: "void ((condition ? (Math.log(value) / Math.LN2) as number : fallback()) satisfies number);", FileName: "input.ts",
					Output: []string{"void ((condition ? (Math.log2(value)) as number : fallback()) satisfies number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 47},
					}},
			})
	})
	t.Run("upstream line 183", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(NaN ** 2 + value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(NaN, value);"}}},
					}},
				{Code: "Math.sqrt(undefined ** 2 + Infinity ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(undefined, Infinity);"}}},
					}},
				{Code: "const result = condition ? Math.log(value) / Math.LN2 : fallback();", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = condition ? Math.log2(value) : fallback();"}}},
					}},
				{Code: "const result = condition && Math.log(value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = condition && Math.log2(value);"}}},
					}},
				{Code: "void ((Math.log(value) / Math.LN2) ? first() : second());", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void ((Math.log2(value)) ? first() : second());"}}},
					}},
				{Code: "void ((Math.log(value) / Math.LN2) && next());", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void ((Math.log2(value)) && next());"}}},
					}},
				{Code: "for (; Math.log(value) / Math.LN2; update()) {}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "for (; Math.log2(value); update()) {}"}}},
					}},
				{Code: "for (const value = Math.log(input) / Math.LN2; condition; update()) {}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "for (const value = Math.log2(input); condition; update()) {}"}}},
					}},
				{Code: "for (; condition; update()) { Math.log(value) / Math.LN2; }", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "for (; condition; update()) { Math.log2(value); }"}}},
					}},
				{Code: "void Math.sqrt(getValue() ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void Math.abs(getValue());"}}},
					}},
				{Code: "void Math.sqrt(1n ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void Math.abs(1n);"}}},
					}},
				{Code: "typeof Math.sqrt(value * value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "typeof Math.abs(value);"}}},
					}},
				{Code: "void (Math.sqrt(1e-200 * 1e-200) && next());", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void (Math.abs(1e-200) && next());"}}},
					}},
				{Code: "void (Math.sqrt(1e-200 * 1e-200) ? first() : second());", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void (Math.abs(1e-200) ? first() : second());"}}},
					}},
				{Code: "void Math.sqrt(getValue() ** 2 + 3 ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void Math.hypot(getValue(), 3);"}}},
					}},
				{Code: "void Math.sqrt(4 ** /* keep */ 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
					}},
				{Code: "for (Math.log(/* keep */ value) / Math.LN2; condition; update()) {}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 43},
					}},
			})
	})
	t.Run("upstream line 207", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(0 ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((-0) ** 2);", FileName: "input.js",
					Output: []string{"Math.abs((-0));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Infinity ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(Infinity);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((-Infinity) ** 2);", FileName: "input.js",
					Output: []string{"Math.abs((-Infinity));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(NaN ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(NaN);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(\"4\" * \"4\");", FileName: "input.js",
					Output: []string{"Math.abs(\"4\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(true * true);", FileName: "input.js",
					Output: []string{"Math.abs(true);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = -0; Math.sqrt(value ** 2);", FileName: "input.js",
					Output: []string{"const value = -0; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28},
					}},
				{Code: "const value = Infinity; Math.sqrt(value ** 2);", FileName: "input.js",
					Output: []string{"const value = Infinity; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34},
					}},
				{Code: "Math.sqrt(0 ** 2 + 0 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(0, 0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((0 * 0 + (-0) ** 2) + 0 ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(0, (-0), 0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(Infinity ** 2 + 3 * 3);", FileName: "input.js",
					Output: []string{"Math.hypot(Infinity, 3);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(3 ** 2 + (-Infinity) ** 2);", FileName: "input.js",
					Output: []string{"Math.hypot(3, (-Infinity));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = -Infinity; Math.sqrt(4 * 4 + value ** 2);", FileName: "input.js",
					Output: []string{"const value = -Infinity; Math.hypot(4, value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
					}},
				{Code: "Math.sqrt(1e999 * 1e999 + 1e200 * 1e200);", FileName: "input.js",
					Output: []string{"Math.hypot(1e999, 1e200);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(true) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(true);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					}},
				{Code: "Math.log(false) * Math.LOG10E;", FileName: "input.js",
					Output: []string{"Math.log10(false);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
					}},
				{Code: "Math.LOG2E * Math.log(null);", FileName: "input.js",
					Output: []string{"Math.log2(null);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					}},
				{Code: "Math.log(\"1\") / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(\"1\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					}},
				{Code: "Math.log(`1`) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(`1`);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					}},
				{Code: "Math.sqrt(`0` ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(`0`);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(`not a number` ** 2);", FileName: "input.js",
					Output: []string{"Math.abs(`not a number`);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(\"-1\") * Math.LOG2E;", FileName: "input.js",
					Output: []string{"Math.log2(\"-1\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					}},
				{Code: "Math.log(\"not a number\") / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(\"not a number\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
					}},
				{Code: "Math.log(\"Infinity\") / Math.LN10;", FileName: "input.js",
					Output: []string{"Math.log10(\"Infinity\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					}},
				{Code: "const value = true; Math.log(value) / Math.LN2;", FileName: "input.js",
					Output: []string{"const value = true; Math.log2(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 47},
					}},
				{Code: "void (Math.log(value) / Math.LN2);", FileName: "input.js",
					Output: []string{"void (Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33},
					}},
				{Code: "void (Math.log(getValue()) * Math.LOG10E);", FileName: "input.js",
					Output: []string{"void (Math.log10(getValue()));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 41},
					}},
				{Code: "void (Math.LOG2E * Math.log(value));", FileName: "input.js",
					Output: []string{"void (Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 35},
					}},
				{Code: "const result = (Math.log(value) / Math.LN10, next());", FileName: "input.js",
					Output: []string{"const result = (Math.log10(value), next());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 44},
					}},
				{Code: "const result = (Math.LOG10E * Math.log(value), next());", FileName: "input.js",
					Output: []string{"const result = (Math.log10(value), next());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 46},
					}},
				{Code: "const result = ((next(), Math.log(value) / Math.LN2), final());", FileName: "input.js",
					Output: []string{"const result = ((next(), Math.log2(value)), final());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 52},
					}},
				{Code: "void (next(), Math.log(value) / Math.LN2);", FileName: "input.js",
					Output: []string{"void (next(), Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 41},
					}},
				{Code: "void (Math.log((/* keep */ value)) / Math.LN2);", FileName: "input.js",
					Output: []string{"void (Math.log2((/* keep */ value)));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 46},
					}},
				{Code: "void (Math.log(1n) / Math.LN2);", FileName: "input.js",
					Output: []string{"void (Math.log2(1n));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30},
					}},
				{Code: "async function calculate() { return void (Math.log(await getValue()) / Math.LN2); }", FileName: "input.js",
					Output: []string{"async function calculate() { return void (Math.log2(await getValue())); }"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 80},
					}},
			})
	})
	t.Run("upstream line 253", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt((0 as number) ** 2);", FileName: "input.ts",
					Output: []string{"Math.abs((0 as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = Infinity as number; Math.sqrt(value ** 2 + 4 ** 2);", FileName: "input.ts",
					Output: []string{"const value = Infinity as number; Math.hypot(value, 4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44},
					}},
				{Code: "Math.log(true as boolean) / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(true as boolean);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
					}},
				{Code: "void ((Math.log(value) / Math.LN2) as number);", FileName: "input.ts",
					Output: []string{"void ((Math.log2(value)) as number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 34},
					}},
				{Code: "const result = ((Math.log(value) / Math.LN2) satisfies number, next());", FileName: "input.ts",
					Output: []string{"const result = ((Math.log2(value)) satisfies number, next());"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 44},
					}},
			})
	})
	t.Run("upstream line 269", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(Infinity ** 2 + \"not a number\" ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Infinity, \"not a number\");"}}},
					}},
				{Code: "Math.sqrt(NaN ** 2 + (-Infinity) ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(NaN, (-Infinity));"}}},
					}},
				{Code: "Math.sqrt(Infinity ** 2 + value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Infinity, value);"}}},
					}},
				{Code: "Math.sqrt(0 ** 2 + getValue() ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(0, getValue());"}}},
					}},
				{Code: "Math.sqrt(1 ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1);"}}},
					}},
				{Code: "Math.sqrt((-1) ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((-1));"}}},
					}},
				{Code: "Math.sqrt(0 ** /* keep */ 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(2n) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(2n);"}}},
					}},
				{Code: "Math.log(/1/) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(/1/);"}}},
					}},
				{Code: "Math.log({valueOf() { return 1; }}) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2({valueOf() { return 1; }});"}}},
					}},
				{Code: "const result = (next(), Math.log(value) / Math.LN2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = (next(), Math.log2(value));"}}},
					}},
				{Code: "void (Math.log(value) / Math.LN2 + next());", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void (Math.log2(value) + next());"}}},
					}},
				{Code: "Math.log(value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value);"}}},
					}},
				{Code: "void (Math.log(/* keep */ value) / Math.LN2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 44},
					}},
				{Code: "void Math.sqrt(value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "void Math.abs(value);"}}},
					}},
			})
	})
	t.Run("upstream line 291", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(4 * 4);", FileName: "input.js",
					Output: []string{"Math.abs(4);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(0 * 0);", FileName: "input.js",
					Output: []string{"Math.abs(0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(0.5 * 0.5);", FileName: "input.js",
					Output: []string{"Math.abs(0.5);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1.5 * 1.5);", FileName: "input.js",
					Output: []string{"Math.abs(1.5);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1e150 * 1e150);", FileName: "input.js",
					Output: []string{"Math.abs(1e150);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1e-150 * 1e-150);", FileName: "input.js",
					Output: []string{"Math.abs(1e-150);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1e999 * 1e999);", FileName: "input.js",
					Output: []string{"Math.abs(1e999);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value = 4; Math.sqrt(value * value);", FileName: "input.js",
					Output: []string{"const value = 4; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
					}},
				{Code: "const value = -4; Math.sqrt(value * value);", FileName: "input.js",
					Output: []string{"const value = -4; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28},
					}},
				{Code: "const value = -0; Math.sqrt(value * value);", FileName: "input.js",
					Output: []string{"const value = -0; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28},
					}},
				{Code: "Math.sqrt(((4)) * ((4)),);", FileName: "input.js",
					Output: []string{"Math.abs(((4)),);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((/* keep */ 4) * 4);", FileName: "input.js",
					Output: []string{"Math.abs((/* keep */ 4));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(1) * Math.LOG10E;", FileName: "input.js",
					Output: []string{"Math.log10(1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					}},
				{Code: "Math.LOG10E * Math.log(0);", FileName: "input.js",
					Output: []string{"Math.log10(0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					}},
				{Code: "Math.log(-1) / Math.LN10;", FileName: "input.js",
					Output: []string{"Math.log10(-1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					}},
				{Code: "Math.log(-0) * Math.LOG2E;", FileName: "input.js",
					Output: []string{"Math.log2(-0);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					}},
				{Code: "Math.LOG2E * Math.log(1);", FileName: "input.js",
					Output: []string{"Math.log2(1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					}},
				{Code: "Math.log(1) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
					}},
				{Code: "Math.log(+1) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2(+1);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
					}},
				{Code: "Math.log(1e999) / Math.LN10;", FileName: "input.js",
					Output: []string{"Math.log10(1e999);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					}},
				{Code: "Math.log(-1e999) * Math.LOG2E;", FileName: "input.js",
					Output: []string{"Math.log2(-1e999);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
					}},
				{Code: "const value = 1; Math.log(value) / Math.LN2;", FileName: "input.js",
					Output: []string{"const value = 1; Math.log2(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 44},
					}},
				{Code: "const value = -1; Math.LOG10E * Math.log(value);", FileName: "input.js",
					Output: []string{"const value = -1; Math.log10(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 48},
					}},
				{Code: "Math.log((/* keep */ 1)) / Math.LN2;", FileName: "input.js",
					Output: []string{"Math.log2((/* keep */ 1));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
					}},
			})
	})
	t.Run("upstream line 326", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(1e200 * 1e200);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1e200);"}}},
					}},
				{Code: "Math.sqrt(1e-200 * 1e-200);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1e-200);"}}},
					}},
				{Code: "Math.sqrt(1e-160 * 1e-160);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1e-160);"}}},
					}},
				{Code: "Math.sqrt(4 ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(4);"}}},
					}},
				{Code: "Math.sqrt(3 * 3 + 4 * 4);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(3, 4);"}}},
					}},
				{Code: "Math.log(2) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(2);"}}},
					}},
				{Code: "Math.log(\"2\") / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(\"2\");"}}},
					}},
				{Code: "Math.log(1n) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(1n);"}}},
					}},
				{Code: "Math.log({valueOf() { return 1; }}) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10({valueOf() { return 1; }});"}}},
					}},
				{Code: "Math.log(getValue()) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(getValue());"}}},
					}},
				{Code: "let value = 1; value = 2; Math.log(value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "let value = 1; value = 2; Math.log2(value);"}}},
					}},
				{Code: "const value = getValue(); Math.log(value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = getValue(); Math.log2(value);"}}},
					}},
				{Code: "const value = 4; const alias = value; Math.sqrt(alias * alias);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 4; const alias = value; Math.abs(alias);"}}},
					}},
				{Code: "const value = 2 + 2; Math.sqrt(value * value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 2 + 2; Math.abs(value);"}}},
					}},
				{Code: "const value = {number: 4}; Math.sqrt(value.number * value.number);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = {number: 4}; Math.abs(value.number);"}}},
					}},
				{Code: "const value = 4; function calculate(value) { return Math.sqrt(value * value); }", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 58, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 4; function calculate(value) { return Math.abs(value); }"}}},
					}},
				{Code: "const value = 1; { let value = 2; Math.log(value) / Math.LN2; }", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; { let value = 2; Math.log2(value); }"}}},
					}},
				{Code: "Math.sqrt(4 * /* keep */ 4);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(/* keep */ 1) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
					}},
			})
	})
	t.Run("upstream line 351", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(1 as number) / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(1 as number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					}},
				{Code: "Math.sqrt((4 as number) * (4 as number));", FileName: "input.ts",
					Output: []string{"Math.abs((4 as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const value: number = 4; Math.sqrt(value * value);", FileName: "input.ts",
					Output: []string{"const value: number = 4; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
					}},
				{Code: "const value = 4; Math.sqrt((value as number) * value);", FileName: "input.ts",
					Output: []string{"const value = 4; Math.abs((value as number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
					}},
				{Code: "const value = 1; Math.log(value as number) / Math.LN2;", FileName: "input.ts",
					Output: []string{"const value = 1; Math.log2(value as number);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 54},
					}},
				{Code: "const value = -1 as number; Math.log(value) / Math.LN10;", FileName: "input.ts",
					Output: []string{"const value = -1 as number; Math.log10(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 56},
					}},
				{Code: "Math.log((1 satisfies number)) * Math.LOG2E;", FileName: "input.ts",
					Output: []string{"Math.log2((1 satisfies number));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
					}},
			})
	})
	t.Run("upstream line 369", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "const x = 1e200, y = 1; Math.sqrt(x * x + y * y);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const x = 1e200, y = 1; Math.hypot(x, y);"}}},
					}},
				{Code: "Math.log(3) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(3);"}}},
					}},
				{Code: "Math.log(1000) / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(1000);"}}},
					}},
				{Code: "Math.log(9) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(9);"}}},
					}},
				{Code: "Math.sqrt(1e200 ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1e200);"}}},
					}},
				{Code: "Math.sqrt(1e-200 ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1e-200);"}}},
					}},
			})
	})
	t.Run("upstream line 388", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(`2`) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(`2`);"}}},
					}},
				{Code: "Math.log(`${1}`) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(`${1}`);"}}},
					}},
				{Code: "Math.sqrt(`${0}` ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(`${0}`);"}}},
					}},
				{Code: "const {length} = \"Infinity\"; Math.log(length) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const {length} = \"Infinity\"; Math.log2(length);"}}},
					}},
				{Code: "const [value] = \"4\"; Math.sqrt(value * value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const [value] = \"4\"; Math.abs(value);"}}},
					}},
			})
	})
	t.Run("upstream line 406", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt((a ** 2 + b ** 2) + (c ** 2 + d ** 2));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c, d);"}}},
					}},
				{Code: "Math.sqrt(a ** 2 + ((b ** 2 + c ** 2) + (d * d + e ** 2)));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c, d, e);"}}},
					}},
				{Code: "Math.sqrt(((a ** 2 + b ** 2) + (c ** 2 + d ** 2)) + ((e ** 2 + f ** 2) + (g ** 2 + h ** 2)));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c, d, e, f, g, h);"}}},
					}},
				{Code: "Math.sqrt((((a ** 2 + b ** 2))) + (((c ** 2 + d ** 2))),);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c, d,);"}}},
					}},
				{Code: "Math.sqrt(((a as number) ** 2 + b ** 2) + (c ** 2 + d ** 2));", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((a as number), b, c, d);"}}},
					}},
				{Code: "Math.sqrt((a ** 2 + b ** 2) /* keep */ + (c ** 2 + d ** 2));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b /* keep */, c, d);"}}},
					}},
				{Code: "Math.sqrt((a ** /* keep */ 2 + b ** 2) + (c ** 2 + d ** 2));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
			})
	})
	t.Run("upstream line 428", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{
				{Code: "Math.log(x) * Math.log(x)", FileName: "input.js"},
				{Code: "Math.log(x) * Math.log(x)", FileName: "input.js"},
				{Code: "Math.LOG10E * Math.LOG10E", FileName: "input.js"},
				{Code: "Math.LOG2E * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(x) * Math[LOG10E]", FileName: "input.js"},
				{Code: "Math.log(x) * Math[LOG10E]", FileName: "input.js"},
				{Code: "Math.log(x) * LOG10E", FileName: "input.js"},
				{Code: "Math.log(x) * LOG10E", FileName: "input.js"},
				{Code: "Math[log](x) * Math.LOG10E", FileName: "input.js"},
				{Code: "Math[log](x) * Math.LOG2E", FileName: "input.js"},
				{Code: "foo.Math.log(x) * Math.LOG10E", FileName: "input.js"},
				{Code: "foo.Math.log(x) * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(x) * foo.Math.LOG10E", FileName: "input.js"},
				{Code: "Math.log(x) * foo.Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(x) * Math.NOT_LOG10E", FileName: "input.js"},
				{Code: "Math.log(x) * Math.NOT_LOG10E", FileName: "input.js"},
				{Code: "Math.log(x) * Math?.LOG10E", FileName: "input.js"},
				{Code: "Math.log(x) * Math?.LOG10E", FileName: "input.js"},
				{Code: "Math?.log(x) * Math.LOG10E", FileName: "input.js"},
				{Code: "Math?.log(x) * Math.LOG2E", FileName: "input.js"},
				{Code: "log(x) * Math.LOG10E", FileName: "input.js"},
				{Code: "log(x) * Math.LOG2E", FileName: "input.js"},
				{Code: "new Math.log(x) * Math.LOG10E", FileName: "input.js"},
				{Code: "new Math.log(x) * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.not_log(x) + Math.LOG10E", FileName: "input.js"},
				{Code: "Math.not_log(x) + Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(x)[Math.LOG10E]", FileName: "input.js"},
				{Code: "Math.log(x)[Math.LOG2E]", FileName: "input.js"},
				{Code: "Math.log() * Math.LOG10E", FileName: "input.js"},
				{Code: "Math.log() * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(x, extraArgument) * Math.LOG10E", FileName: "input.js"},
				{Code: "Math.log(x, extraArgument) * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.log(...x) * Math.LOG10E", FileName: "input.js"},
				{Code: "Math.log(...x) * Math.LOG2E", FileName: "input.js"},
				{Code: "Math.LN10 / Math.LN10", FileName: "input.js"},
				{Code: "Math.LN2 / Math.LN2", FileName: "input.js"},
				{Code: "Math.log(x) /Math[LN10]", FileName: "input.js"},
				{Code: "Math.log(x) /Math[LN10]", FileName: "input.js"},
				{Code: "Math.log(x) / LN10", FileName: "input.js"},
				{Code: "Math.log(x) / LN10", FileName: "input.js"},
				{Code: "Math[log](x) / Math.LN10", FileName: "input.js"},
				{Code: "Math[log](x) / Math.LN2", FileName: "input.js"},
				{Code: "foo.Math.log(x) / Math.LN10", FileName: "input.js"},
				{Code: "foo.Math.log(x) / Math.LN2", FileName: "input.js"},
				{Code: "Math.log(x) / foo.Math.LN10", FileName: "input.js"},
				{Code: "Math.log(x) / foo.Math.LN2", FileName: "input.js"},
				{Code: "Math.log(x) / Math.log(x)", FileName: "input.js"},
				{Code: "Math.log(x) / Math.log(x)", FileName: "input.js"},
				{Code: "Math.log(x) / Math.NOT_LN10", FileName: "input.js"},
				{Code: "Math.log(x) / Math.NOT_LN10", FileName: "input.js"},
				{Code: "Math.log(x) / Math?.LN10", FileName: "input.js"},
				{Code: "Math.log(x) / Math?.LN10", FileName: "input.js"},
				{Code: "Math?.log(x) / Math.LN10", FileName: "input.js"},
				{Code: "Math?.log(x) / Math.LN2", FileName: "input.js"},
				{Code: "log(x) / Math.LN10", FileName: "input.js"},
				{Code: "log(x) / Math.LN2", FileName: "input.js"},
				{Code: "new Math.log(x) / Math.LN10", FileName: "input.js"},
				{Code: "new Math.log(x) / Math.LN2", FileName: "input.js"},
				{Code: "Math.not_log(x) + Math.LN10", FileName: "input.js"},
				{Code: "Math.not_log(x) + Math.LN2", FileName: "input.js"},
				{Code: "Math.log(x)[Math.LN10]", FileName: "input.js"},
				{Code: "Math.log(x)[Math.LN2]", FileName: "input.js"},
				{Code: "Math.log() / Math.LN10", FileName: "input.js"},
				{Code: "Math.log() / Math.LN2", FileName: "input.js"},
				{Code: "Math.log(x, extraArgument) / Math.LN10", FileName: "input.js"},
				{Code: "Math.log(x, extraArgument) / Math.LN2", FileName: "input.js"},
				{Code: "Math.log(...x) / Math.LN10", FileName: "input.js"},
				{Code: "Math.log(...x) / Math.LN2", FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(x) * Math.LOG10E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.log(x) * Math.LOG2E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.LOG10E * Math.log(x)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.LOG2E * Math.log(x)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.log(x) / Math.LN10", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.log(x) / Math.LN2", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.log((( 0,x ))) * Math.LOG10E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10((( 0,x )))"}}},
					}},
				{Code: "Math.log((( 0,x ))) * Math.LOG2E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2((( 0,x )))"}}},
					}},
				{Code: "Math.LOG10E * Math.log((( 0,x )))", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10((( 0,x )))"}}},
					}},
				{Code: "Math.LOG2E * Math.log((( 0,x )))", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2((( 0,x )))"}}},
					}},
				{Code: "Math.log((( 0,x ))) / Math.LN10", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10((( 0,x )))"}}},
					}},
				{Code: "Math.log((( 0,x ))) / Math.LN2", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2((( 0,x )))"}}},
					}},
				{Code: "function foo(x) {\n\treturn (\n\t\tMath.log(x)\n\t\t\t/ Math.LN10\n\t);\n}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 3, Column: 3, EndLine: 4, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(x) {\n\treturn (\n\t\tMath.log10(x)\n\t);\n}"}}},
					}},
				{Code: "function foo(x) {\n\treturn (\n\t\tMath.log(x)\n\t\t\t/ Math.LN2\n\t);\n}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 3, Column: 3, EndLine: 4, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function foo(x) {\n\treturn (\n\t\tMath.log2(x)\n\t);\n}"}}},
					}},
			})
	})
	t.Run("upstream line 486", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{
				{Code: "Math.notSqrt(a ** 2 + b ** 2)", FileName: "input.js"},
				{Code: "NotMath.sqrt(a ** 2 + b ** 2)", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2 - b ** 2)", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2 + 2 ** b)", FileName: "input.js"},
				{Code: "Math.sqrt(a * c + b * c)", FileName: "input.js"},
				{Code: "Math.sqrt((++a) * (++a))", FileName: "input.js"},
				{Code: "Math.sqrt(Math.pow(a, 2) + Math.pow(b, 2))", FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(a * a + b * b)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b)"}}},
					}},
				{Code: "Math.sqrt(a ** 2 + b ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b)"}}},
					}},
				{Code: "Math.sqrt(a * a + b ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b)"}}},
					}},
				{Code: "Math.sqrt(a * a + b * b + c * c)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c)"}}},
					}},
				{Code: "Math.sqrt(a ** 2 + b ** 2 + c ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c)"}}},
					}},
				{Code: "Math.sqrt(a * a)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(a)"}}},
					}},
				{Code: "Math.sqrt(a ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(a)"}}},
					}},
				{Code: "Math.sqrt(a * a,)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(a,)"}}},
					}},
				{Code: "Math.sqrt(a ** 2,)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(a,)"}}},
					}},
				{Code: "Math.sqrt((a, b) ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((a, b))"}}},
					}},
				{Code: "Math.sqrt((++a) ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((++a))"}}},
					}},
				{Code: "Math.sqrt(a * a + b * b,)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b,)"}}},
					}},
				{Code: "Math.sqrt(a ** 2 + b ** 2,)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b,)"}}},
					}},
				{Code: "Math.sqrt((( a ** 2 )) + (( b ** 2 + c ** 2 )) + (( d )) * (( d )) + (( e )) ** (( 2 )))", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((( a )),  b, c , (( d )), (( e )))"}}},
					}},
			})
	})
	t.Run("upstream line 516", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(/* keep */ x) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
					}},
				{Code: "Math.log(x) // keep\n\t* Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 2, EndColumn: 15},
					}},
				{Code: "Math.log(x) / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x);"}}},
					}},
				{Code: "Math.log(x) / /* keep */ Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
					}},
				{Code: "Math.sqrt(a ** /* keep */ 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(a ** // keep\n\t2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log(x /* keep */ + 1) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x /* keep */ + 1);"}}},
					}},
				{Code: "Math.log((/* keep */ x)) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2((/* keep */ x));"}}},
					}},
			})
	})
	t.Run("upstream line 532", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(value) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value);"}}},
					}},
				{Code: "Math.LOG10E * Math.log(value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value);"}}},
					}},
				{Code: "Math.log(value) * Math.LOG2E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value);"}}},
					}},
				{Code: "Math.LOG2E * Math.log(value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value);"}}},
					}},
				{Code: "Math.log(value) / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value);"}}},
					}},
				{Code: "Math.log(value) / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value);"}}},
					}},
				{Code: "Math.sqrt(value * value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(value);"}}},
					}},
				{Code: "Math.sqrt(value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(value);"}}},
					}},
				{Code: "Math.sqrt(first * first + second * second);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, second);"}}},
					}},
				{Code: "Math.sqrt(first ** 2 + second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, second);"}}},
					}},
			})
	})
	t.Run("upstream line 555", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "const x = 1e-200, y = 1e-200; Math.sqrt(x * x + y * y);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const x = 1e-200, y = 1e-200; Math.hypot(x, y);"}}},
					}},
				{Code: "Math.sqrt(Infinity ** 2 + NaN ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Infinity, NaN);"}}},
					}},
				{Code: "Math.sqrt(Number.MAX_VALUE ** 2 + Number.MIN_VALUE ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(Number.MAX_VALUE, Number.MIN_VALUE);"}}},
					}},
				{Code: "const x = 1e200; Math.sqrt(x * x);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const x = 1e200; Math.abs(x);"}}},
					}},
				{Code: "const x = 1e-200; Math.sqrt(x * x);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const x = 1e-200; Math.abs(x);"}}},
					}},
				{Code: "Math.LOG2E * Math.log(9);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(9);"}}},
					}},
				{Code: "Math.LOG10E * Math.log(3);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(3);"}}},
					}},
			})
	})
	t.Run("upstream line 569", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{
				{Code: "Math.log?.(x) * Math.LOG10E;", FileName: "input.js"},
				{Code: "Math.LOG2E * Math.log?.(x);", FileName: "input.js"},
				{Code: "Math.log?.(x) / Math.LN2;", FileName: "input.js"},
				{Code: "Math[\"log\"](x) * Math.LOG10E;", FileName: "input.js"},
				{Code: "Math.log(x) / Math[\"LN10\"];", FileName: "input.js"},
				{Code: "Math.sqrt();", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2, b ** 2);", FileName: "input.js"},
				{Code: "Math.sqrt(...squares);", FileName: "input.js"},
				{Code: "Math.sqrt?.(a ** 2 + b ** 2);", FileName: "input.js"},
				{Code: "Math?.sqrt(a ** 2 + b ** 2);", FileName: "input.js"},
				{Code: "Math[\"sqrt\"](a ** 2 + b ** 2);", FileName: "input.js"},
				{Code: "new Math.sqrt(a ** 2 + b ** 2);", FileName: "input.js"},
				{Code: "math.sqrt(a ** 2 + b ** 2);", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2 + b);", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2 + b ** 3);", FileName: "input.js"},
				{Code: "Math.sqrt(a ** 2 * b ** 2);", FileName: "input.js"},
				{Code: "Math.sqrt(getValue() * getValue());", FileName: "input.js"},
				{Code: "Math.sqrt(a * b + c * c);", FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{})
	})
	t.Run("upstream line 594", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "const result = 2 * (Math.log(value) / Math.LN10);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = 2 * (Math.log10(value));"}}},
					}},
				{Code: "const result = (Math.log(value) * Math.LOG2E) ** 2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = (Math.log2(value)) ** 2;"}}},
					}},
				{Code: "Math.log(getValue()) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(getValue());"}}},
					}},
				{Code: "async function calculate() { return Math.log(await getValue()) / Math.LN2; }", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 74, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "async function calculate() { return Math.log2(await getValue()); }"}}},
					}},
				{Code: "function * calculate() { return Math.log(yield value) * Math.LOG2E; }", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function * calculate() { return Math.log2(yield value); }"}}},
					}},
				{Code: "Math.log(value = nextValue) / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value = nextValue);"}}},
					}},
				{Code: "Math.LOG10E * Math.log(condition ? first : second);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(condition ? first : second);"}}},
					}},
				{Code: "(Math).log(value,) / (Math.LN2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value);"}}},
					}},
				{Code: "const distance = Math.sqrt(point.x * point.x + point.y * point.y);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const distance = Math.hypot(point.x, point.y);"}}},
					}},
				{Code: "Math.sqrt(values[index] * values[index] + other ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(values[index], other);"}}},
					}},
				{Code: "Math.sqrt((first + second) ** 2 + (third ? fourth : fifth) ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((first + second), (third ? fourth : fifth));"}}},
					}},
				{Code: "Math.sqrt(getValue() ** 2 + other ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(getValue(), other);"}}},
					}},
				{Code: "Math.sqrt((first, second) ** 2 + third ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((first, second), third);"}}},
					}},
				{Code: "Math.sqrt(value ** 2).toFixed(2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(value).toFixed(2);"}}},
					}},
				{Code: "const result = <span>{Math.log(value) / Math.LN10}</span>;", FileName: "input.jsx", Tsx: true,
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = <span>{Math.log10(value)}</span>;"}}},
					}},
				{Code: "const result = <>{Math.sqrt(first ** 2 + second ** 2)}</>;", FileName: "input.jsx", Tsx: true,
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = <>{Math.hypot(first, second)}</>;"}}},
					}},
				{Code: "Math.log(Math.sqrt(value ** 2)) / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(Math.sqrt(value ** 2));"}}},
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log(Math.abs(value)) / Math.LN10;"}}},
					}},
				{Code: "Math.sqrt((Math.log(value) * Math.LOG2E) ** 2 + other ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((Math.log(value) * Math.LOG2E), other);"}}},
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.sqrt((Math.log2(value)) ** 2 + other ** 2);"}}},
					}},
			})
	})
	t.Run("upstream line 625", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(value as number) * Math.LOG10E;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value as number);"}}},
					}},
				{Code: "Math.LOG2E * Math.log(<number>value);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(<number>value);"}}},
					}},
				{Code: "Math.log(value!) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(value!);"}}},
					}},
				{Code: "Math.log(value satisfies number) / Math.LN10;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value satisfies number);"}}},
					}},
				{Code: "Math.log(getValue<number>()) * Math.LOG2E;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(getValue<number>());"}}},
					}},
				{Code: "Math.sqrt((value as number) ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((value as number));"}}},
					}},
				{Code: "Math.sqrt((<number>value) ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((<number>value));"}}},
					}},
				{Code: "Math.sqrt(value! * value!);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(value!);"}}},
					}},
				{Code: "Math.sqrt((value satisfies number) ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((value satisfies number));"}}},
					}},
				{Code: "Math.sqrt(first! ** 2 + (second as number) ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first!, (second as number));"}}},
					}},
				{Code: "Math.sqrt(getValue<number>() ** 2 + second ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(getValue<number>(), second);"}}},
					}},
				{Code: "Math.sqrt(value ** 2) satisfies number;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(value) satisfies number;"}}},
					}},
			})
	})
	t.Run("upstream line 644", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.LOG10E /* keep */ * Math.log(value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
					}},
				{Code: "Math.LOG2E * Math.log(/* keep */ value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
					}},
				{Code: "Math.log(value) /* keep */ / Math.LN2;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
					}},
				{Code: "Math /* keep */ .log(value) * Math.LOG10E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
					}},
				{Code: "Math.log(value) * Math./* keep */LOG2E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
					}},
				{Code: "Math.sqrt(value * /* keep */ value);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(first ** 2 + second ** /* keep */ 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt((first ** 2 + second ** 2) + (third * /* keep */ third + fourth ** 2));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.log((/* keep */ value)) * Math.LOG2E;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2((/* keep */ value));"}}},
					}},
				{Code: "Math.LOG10E * Math.log(value /* keep */ + offset);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(value /* keep */ + offset);"}}},
					}},
				{Code: "Math.sqrt(/* keep */ first ** 2 + second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/* keep */ first, second);"}}},
					}},
				{Code: "Math.sqrt((/* keep */ first) ** 2 + second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((/* keep */ first), second);"}}},
					}},
				{Code: "Math.sqrt(first ** 2 + /* keep */ second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, /* keep */ second);"}}},
					}},
				{Code: "Math /* keep */ .sqrt(value ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math /* keep */ .abs(value);"}}},
					}},
				{Code: "Math.sqrt(first ** 2 // keep\n  + second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first // keep\n, second);"}}},
					}},
				{Code: "Math.sqrt(first ** 2 + // keep\n  second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, // keep\n  second);"}}},
					}},
				{Code: "Math.log(Math.sqrt(value ** 2)) /* keep */ / Math.LN10;", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 55},
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log(Math.abs(value)) /* keep */ / Math.LN10;"}}},
					}},
				{Code: "Math.sqrt((Math.log(/* keep */ value) * Math.LOG2E) ** 2 + other ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((Math.log(/* keep */ value) * Math.LOG2E), other);"}}},
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 51},
					}},
				{Code: "function distance() {\r\n  return Math.sqrt(\r\n    first ** 2 +\r\n    second ** 2\r\n  );\r\n}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 2, Column: 15, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function distance() {\r\n  return Math.hypot(\r\n    first,\r\n    second\r\n  );\r\n}"}}},
					}},
				{Code: "function logarithm() {\r\n  return (\r\n    Math.log(value /* keep */ + offset)\r\n      / Math.LN10\r\n  );\r\n}", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 3, Column: 5, EndLine: 4, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function logarithm() {\r\n  return (\r\n    Math.log10(value /* keep */ + offset)\r\n  );\r\n}"}}},
					}},
			})
	})
	t.Run("documentation examples", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{
				{Code: "Math.log10(x)", FileName: "input.js"},
				{Code: "Math.log2(x)", FileName: "input.js"},
				{Code: "Math.hypot(a, b)", FileName: "input.js"},
				{Code: "Math.abs(x)", FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(x) * Math.LOG10E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) * Math.LOG10E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.LOG10E * Math.log(x)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.LOG10E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.log(x) / Math.LN10", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log10(…)` over `Math.log(…) / Math.LN10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log10(x)"}}},
					}},
				{Code: "Math.log(x) * Math.LOG2E", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) * Math.LOG2E`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.LOG2E * Math.log(x)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.LOG2E * Math.log(…)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.log(x) / Math.LN2", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(x)"}}},
					}},
				{Code: "Math.sqrt(a * a + b * b)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b)"}}},
					}},
				{Code: "Math.sqrt(a ** 2 + b ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b)"}}},
					}},
				{Code: "Math.sqrt(x ** 2)", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(x)"}}},
					}},
			})
	})
}
