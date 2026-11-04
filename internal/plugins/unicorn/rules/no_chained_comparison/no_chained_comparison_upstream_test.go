// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-chained-comparison.js
// Parser-only repeats are deduplicated; runtime-test input is covered by the comment cases.
package no_chained_comparison_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_chained_comparison"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const comparisonMessage = "Comparison operators cannot be chained. The inner comparison evaluates to a boolean, which is then compared instead of the operands."

func TestNoChainedComparisonUpstream(t *testing.T) {
	for _, mode := range []string{"with type information", "without type information"} {
		t.Run(mode, func(t *testing.T) {
			testedRule := no_chained_comparison.NoChainedComparisonRule
			if mode == "without type information" {
				testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					ctx.TypeChecker = nil
					return no_chained_comparison.NoChainedComparisonRule.Run(ctx, options)
				}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, []rule_tester.ValidTestCase{
				// test group 3
				{Code: "/** @param {boolean} flag */ function run(flag) { return (a < b) === flag; }", FileName: "file.ts"},
				{Code: "/** @param {() => boolean} read */ function run(read) { return (a < b) !== read(); }", FileName: "file.ts"},
				{Code: "/** @param {string} text */ function run(text) { return (a < b) == text.startsWith(\"x\"); }", FileName: "file.ts"},
				{Code: "/** @param {Array<number> | Uint8Array} values */ function run(values) { return values.some(predicate) === (a < b); }", FileName: "file.ts"},
				{Code: "/** @returns {RegExp} */ function read() { return source; } (a < b) === read().global;", FileName: "file.ts"},
				// test group 6
				{Code: "/** @param {string} flag */ function run(flag: boolean) { return (a < b) === flag; }", FileName: "file.ts"},
				// snapshot cases
				{Code: "a < b", FileName: "case.js"},
				{Code: "a === b", FileName: "case.js"},
				{Code: "a < b && b < c", FileName: "case.js"},
				{Code: "a < b && b < c && c < d", FileName: "case.js"},
				{Code: "(a > 0) === (b > 0)", FileName: "case.js"},
				{Code: "(a > 0) !== (b > 0)", FileName: "case.js"},
				{Code: "(a > 0) == (b > 0)", FileName: "case.js"},
				{Code: "(a < b) === true", FileName: "case.js"},
				{Code: "(a < b) !== false", FileName: "case.js"},
				{Code: "(a < b) == true", FileName: "case.js"},
				{Code: "true === (a < b)", FileName: "case.js"},
				{Code: "a == b == true", FileName: "case.js"},
				{Code: "(a > 0) !== Boolean(b)", FileName: "case.js"},
				{Code: "(a < b) === !c", FileName: "case.js"},
				{Code: "(a < b) === (c ? true : false)", FileName: "case.js"},
				{Code: "const d = a > b; (a < b) === d;", FileName: "case.js"},
				{Code: "a < b || c < d", FileName: "case.js"},
				{Code: "!(a < b)", FileName: "case.js"},
				{Code: "foo(a < b, c)", FileName: "case.js"},
				{Code: "a < b ? c : d", FileName: "case.js"},
				{Code: "const x = a < b;", FileName: "case.js"},
				{Code: "f(a < b) === g(c < d)", FileName: "case.js"},
				{Code: "a instanceof b instanceof c", FileName: "case.js"},
				{Code: "a in b in c", FileName: "case.js"},
				{Code: "a + b + c", FileName: "case.js"},
				{Code: "(a < b) + c", FileName: "case.js"},
				{Code: "(a < b) & c", FileName: "case.js"},
				{Code: "typeof a < b", FileName: "case.js"},
				{Code: "a << b < c", FileName: "case.js"},
				{Code: "(a as number) < b", FileName: "case.ts"},
				// documentation examples
				{Code: "// ✅\nif (a < b && b < c) {}", FileName: "case.js"},
				{Code: "// ✅\nif (a === b && b === c) {}", FileName: "case.js"},
				{Code: "// ✅\n// Intentionally comparing two boolean results.\nif ((a > 0) === (b > 0)) {}", FileName: "case.js"},
			}, []rule_tester.InvalidTestCase{
				// test group 1
				{Code: "read(/* keep */) < middle < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "read(/* keep */) < middle && middle < last;"},
					}},
				}},
				{Code: "first < middle < (/* keep */ last);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "first < middle && middle < (/* keep */ last);"},
					}},
				}},
				// test group 2
				{Code: "(/* keep */ first) < middle < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "(/* keep */ first) < middle && middle < last;"},
					}},
				}},
				{Code: "read(// keep\r\n) < middle < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "read(// keep\r\n) < middle && middle < last;"},
					}},
				}},
				{Code: "first < middle < ((last /* keep */));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "first < middle && middle < ((last /* keep */));"},
					}},
				}},
				{Code: "(first as /* keep */ number) < middle < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "(first as /* keep */ number) < middle && middle < last;"},
					}},
				}},
				{Code: "first < (middle /* duplicate */) < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(/* discard */ first < middle) < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(first < middle /* discard */) < last;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				// test group 3
				{Code: "/** @param {boolean | undefined} flag */ function run(flag) { return (a < b) === flag; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 70, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "/** @param {boolean} [flag] */ function run(flag) { return (a < b) === flag; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 60, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "/** @param {() => (() => boolean)} read */ function run(read) { return (a < b) !== read(); }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 72, EndLine: 1, EndColumn: 90, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				// test group 6
				{Code: "/** @param {boolean} flag */ function run(flag: string) { return (a < b) === flag; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 66, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				// test group 8
				{Code: "a < b < c as boolean", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a < b < c) as boolean", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "(a < b && b < c) as boolean"},
					}},
				}},
				{Code: "a < b < c satisfies boolean", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a < b < c) satisfies boolean", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "(a < b && b < c) satisfies boolean"},
					}},
				}},
				// snapshot cases
				{Code: "a < b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < c"},
					}},
				}},
				{Code: "a > b > c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a > b && b > c"},
					}},
				}},
				{Code: "a <= b <= c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a <= b && b <= c"},
					}},
				}},
				{Code: "a >= b >= c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a >= b && b >= c"},
					}},
				}},
				{Code: "a < b <= c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b <= c"},
					}},
				}},
				{Code: "a < b > c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b > c"},
					}},
				}},
				{Code: "a == b == c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a == b && b == c"},
					}},
				}},
				{Code: "a === b === c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a === b && b === c"},
					}},
				}},
				{Code: "a != b != c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a != b && b != c"},
					}},
				}},
				{Code: "a !== b !== c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a !== b && b !== c"},
					}},
				}},
				{Code: "a <= b >= c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a <= b && b >= c"},
					}},
				}},
				{Code: "a === b !== c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a === b && b !== c"},
					}},
				}},
				{Code: "(a < b) < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < c"},
					}},
				}},
				{Code: "(a == b) == c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a == b && b == c"},
					}},
				}},
				{Code: "0 < x < 10", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "0 < x && x < 10"},
					}},
				}},
				{Code: "min <= value <= max", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "min <= value && value <= max"},
					}},
				}},
				{Code: "-1 < x < 1", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "-1 < x && x < 1"},
					}},
				}},
				{Code: "a < b.c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b.c && b.c < d"},
					}},
				}},
				{Code: "a < b[c] < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b[c] && b[c] < d"},
					}},
				}},
				{Code: "a < b?.c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b?.c && b?.c < d"},
					}},
				}},
				{Code: "a < b + c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b + c && b + c < d"},
					}},
				}},
				{Code: "a < b === c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a === b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a > 0) == 0", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a > 0) === isEnabled", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a == b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a === b > c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < f() < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < (b = c) < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < b++ < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < /* comment */ b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a < b) < true", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < b < false", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "(a < b) < (c < d)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < b < c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "foo === a < b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a < b < c === foo", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "x & a < b < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "a ?? b < c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				}},
				{Code: "foo === (a < b < c)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "foo === (a < b && b < c)"},
					}},
				}},
				{Code: "a && b < c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a && b < c && c < d"},
					}},
				}},
				{Code: "a || b < c < d", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a || b < c && c < d"},
					}},
				}},
				{Code: "a < b < c ? x : y", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < c ? x : y"},
					}},
				}},
				{Code: "a < (b ?? 0) < c", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < (b ?? 0) && (b ?? 0) < c"},
					}},
				}},
				{Code: "a < b < c", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < c"},
					}},
				}},
				{Code: "a < b! < c", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < b! && b! < c"},
					}},
				}},
				{Code: "a < (b as number) < c", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "a < (b as number) && (b as number) < c"},
					}},
				}},
				// test group 10
				{Code: "({value: 1}.value < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "({value: 1}.value < b && b < c);"},
					}},
				}},
				{Code: "(function () {}.length < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "(function () {}.length < b && b < c);"},
					}},
				}},
				{Code: "const result = ({value: 1}.value < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "const result = {value: 1}.value < b && b < c;"},
					}},
				}},
				{Code: "function check() { return ({value: 1}.value < b) < c; }", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "function check() { return {value: 1}.value < b && b < c; }"},
					}},
				}},
				{Code: "(({value: 1}.value < b) < c);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "({value: 1}.value < b && b < c);"},
					}},
				}},
				{Code: "(({value: 1}.value) < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "({value: 1}.value) < b && b < c;"},
					}},
				}},
				{Code: "const check = () => ({value: 1}.value < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "const check = () => ({value: 1}.value < b && b < c);"},
					}},
				}},
				{Code: "const check = () => (function () {}.length < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "const check = () => function () {}.length < b && b < c;"},
					}},
				}},
				// documentation examples
				{Code: "// ❌\nif (a < b < c) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 2, Column: 5, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "// ❌\nif (a < b && b < c) {}"},
					}},
				}},
				{Code: "// ❌\nif (a === b === c) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 2, Column: 5, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-chained-comparison/suggestion", Output: "// ❌\nif (a === b && b === c) {}"},
					}},
				}},
			})
		})
	}
}
