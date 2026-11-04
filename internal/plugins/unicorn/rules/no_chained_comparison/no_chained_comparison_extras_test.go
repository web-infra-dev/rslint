package no_chained_comparison_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_chained_comparison"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics and suggestions were checked against Unicorn v79.0.0
// with TypeScript type information. The upstream suite also covers syntax-only mode.
func TestNoChainedComparisonExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_chained_comparison.NoChainedComparisonRule, []rule_tester.ValidTestCase{
		// Use instantiated call results and flow-narrowed property values.
		{Code: "declare const a: number, b: number; declare function id<T>(x: T): T; (a === b) === id(true);", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare function id<T>(x: T): T; (a === b) === id<boolean>(true);", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; const id = <T>(x: T): T => x; (a === b) === id(true);", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare function id<T>(x: T): T; const alias = id; (a === b) === alias(true);", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare const object: { id<T>(x: T): T }; (a === b) === object.id(true);", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare const object: { flag?: boolean }; if (object.flag !== undefined) { (a === b) === object.flag; }", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare const object: { flag?: boolean }; if (object[\"flag\"] !== undefined) { (a === b) === object[\"flag\"]; }", FileName: "case.ts"},
		{Code: "declare const a: number, b: number; declare const object: { flag: boolean | null }; if (object.flag !== null) { (a === b) === object.flag; }", FileName: "case.ts"},
		// JSDoc callable initializers and aliases, checked against Unicorn v79.0.0.
		{Code: "/** @param {() => number} read */ function run(read = (): boolean => true) { return (a < b) === read(); }", FileName: "case.ts"},
		// Computed keys use their values, rather than the identifier spelling.
		{Code: "/** @param {RegExp} value */ function run(value) { const key = \"global\"; return (a < b) === value[key]; }", FileName: "case.ts"},
		// Static results precede syntax; JSDoc optional values can be narrowed.
		{Code: "(a < b) === (0 || false);", FileName: "case.ts"},
		{Code: "(a < b) === (null ?? true);", FileName: "case.ts"},
		{Code: "(a < b) === (true as number);", FileName: "case.ts"},
		{Code: "/** @param {boolean} [flag] */ function run(flag) { if (flag !== undefined) return (a < b) === flag; }", FileName: "case.js"},
		// JSDoc casts and authored TypeScript wrappers retain their boolean semantics.
		{Code: "(a < b) === /** @type {number} */ (!c);", FileName: "case.js"},
		{Code: "(a < b) === /** @type {unknown} */ (!c);", FileName: "case.js"},
		{Code: "function run(flag?: boolean) { if (flag !== undefined) return (a < b) === flag; }", FileName: "case.ts"},
		{Code: "let flag = 1; (a < b) === ((flag) satisfies boolean);", FileName: "case.ts"},
		// Grouping and syntax boundaries
		{Code: "((a < b) as number) < c;", FileName: "case.ts"},
		{Code: "((a < b) satisfies boolean) === c;", FileName: "case.ts"},
		{Code: "((a < b)!) < c;", FileName: "case.ts"},
		// Intentional boolean values and unknown operands
		{Code: "declare const flag: boolean; (a < b) === flag;", FileName: "case.ts"},
		{Code: "type Flag = true | false; declare const flag: Flag; (a < b) === flag;", FileName: "case.ts"},
		{Code: "function run<T extends boolean>(flag: T) { return (a < b) === flag; }", FileName: "case.ts"},
		{Code: "(a < b) === (!c && !d);", FileName: "case.ts"},
		{Code: "(a < b) === (!c || !d);", FileName: "case.ts"},
		{Code: "(a < b) === (!c ?? !d);", FileName: "case.ts"},
		{Code: "(a < b) === (flag = !c);", FileName: "case.ts"},
		{Code: "(a < b) === (read(), !c);", FileName: "case.ts"},
		{Code: "(a < b) === delete object.key;", FileName: "case.ts"},
		{Code: "(a < b) === ((!c) as unknown);", FileName: "case.ts"},
		{Code: "(a < b) === ((!c) satisfies unknown);", FileName: "case.ts"},
		{Code: "const flag = (a < b); const alias = flag; (c < d) !== alias;", FileName: "case.ts"},
		{Code: "const object = {enabled: true}; (a < b) === object.enabled;", FileName: "case.ts"},
		{Code: "(a < b) === /test/.global;", FileName: "case.ts"},
		{Code: "(a < b) === [1, 2].includes(c);", FileName: "case.ts"},
		{Code: "(a < b) === \"text\"[\"startsWith\"](c);", FileName: "case.ts"},
		{Code: "(a < b) === new Set().has(c);", FileName: "case.ts"},
		{Code: "let flag = Boolean(c); (a < b) === flag;", FileName: "case.ts"},
	}, []rule_tester.InvalidTestCase{
		// Actual numeric results must still report and retain their suggestions.
		{Code: "declare const a: number, b: number; declare function id<T>(x: T): T; (a === b) === id(1);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 70, EndLine: 1, EndColumn: 89, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "declare const a: number, b: number; declare function id<T>(x: T): T; a === b && b === id(1);"},
			}},
		}},
		{Code: "declare const a: number, b: number; declare const object: { value?: number }; if (object.value !== undefined) { (a === b) === object.value; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 113, EndLine: 1, EndColumn: 139, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "declare const a: number, b: number; declare const object: { value?: number }; if (object.value !== undefined) { a === b && b === object.value; }"},
			}},
		}},
		// JSDoc callable initializers and aliases, checked against Unicorn v79.0.0.
		{Code: "/** @type {() => boolean} */ const read = async () => true; (a < b) === read();", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 61, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @type {() => boolean} */ const read = function* () { return true; }; (a < b) === read();", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 74, EndLine: 1, EndColumn: 92, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @param {() => boolean} read */ function run(read = (): number => 0) { return (a < b) === read(); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 82, EndLine: 1, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @param {() => boolean} read */ function run(read = (() => 0) as () => number) { return (a < b) === read(); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 92, EndLine: 1, EndColumn: 110, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @param {() => boolean} read */ function run(read = async () => true) { return (a < b) === read(); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 83, EndLine: 1, EndColumn: 101, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @param {() => boolean} read */ function run(read = function* () { return true; }) { return (a < b) === read(); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 96, EndLine: 1, EndColumn: 114, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @type {number} */ const flag = true; const alias = flag; (a < b) === alias;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 62, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// In scripts, let can be an identifier. Keep let[0] from becoming a
		// lexical declaration when the inner comparison's parentheses go away.
		{Code: "for ((let[0] < middle) < last; false;) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "for ((let[0] < middle && middle < last); false;) {}"},
			}},
		}},
		// Computed keys use their values, rather than the identifier spelling.
		{Code: "const key = \"value\"; const object = {key: true, value: 0}; (a < b) === object[key];", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 60, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// JSDoc casts and authored TypeScript wrappers retain their boolean semantics.
		{Code: "(a < b) === /** @type {boolean} */ (c);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const result = /** @type {number} */ (a < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 38, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "const result = /** @type {number} */ a < b && b < c;"},
			}},
		}},
		{Code: "const result = /** @satisfies {boolean} */ (a < b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 44, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "const result = /** @satisfies {boolean} */ a < b && b < c;"},
			}},
		}},
		{Code: "a < b < /** @type {number} */ (c);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a < /** @type {number} */ (b) < c;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "declare const flag: boolean | undefined; (a < b) === flag!;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 42, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Grouping and syntax boundaries
		{Code: "min < ((value)) <= max;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "min < ((value)) && ((value)) <= max;"},
			}},
		}},
		{Code: "a < b < (true as boolean);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < (true as boolean);"},
			}},
		}},
		{Code: "a < b < c!;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < b && b < c!;"},
			}},
		}},
		{Code: "a < ((b, c)) < d;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < ((b, c)) && ((b, c)) < d;"},
			}},
		}},
		{Code: "a < (b ? c : d) < e;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < (b ? c : d) && (b ? c : d) < e;"},
			}},
		}},
		{Code: "a < (b satisfies number) < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < (b satisfies number) && (b satisfies number) < c;"},
			}},
		}},
		{Code: "(<number>a) < b < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "(<number>a) < b && b < c;"},
			}},
		}},
		{Code: "a ?? (b < c < d);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 7, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a ?? (b < c && c < d);"},
			}},
		}},
		{Code: "result = a < b < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "result = a < b && b < c;"},
			}},
		}},
		{Code: "result ||= a < b < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "result ||= a < b && b < c;"},
			}},
		}},
		{Code: "(before(), a < b < c);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "(before(), a < b && b < c);"},
			}},
		}},
		{Code: "consume(a < b < c);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "consume(a < b && b < c);"},
			}},
		}},
		{Code: "const element = <Widget ok={min < value < max} />;", FileName: "case.tsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 29, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "const element = <Widget ok={min < value && value < max} />;"},
			}},
		}},
		{Code: "class C { #value = 1; run() { return min < this.#value < max; } }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 38, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "class C { #value = 1; run() { return min < this.#value && this.#value < max; } }"},
			}},
		}},
		{Code: "a < b?.[key] < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "a < b?.[key] && b?.[key] < c;"},
			}},
		}},
		{Code: "const 注 = \"😀\"; 最小 < 值 < 最大;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "const 注 = \"😀\"; 最小 < 值 && 值 < 最大;"},
			}},
		}},
		{Code: "const prefix = \"😀\";\nif (\n  low <\n  value < high\n) {}", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 3, Column: 3, EndLine: 4, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "const prefix = \"😀\";\nif (\n  low <\n  value && value < high\n) {}"},
			}},
		}},
		// Side effects and comments
		{Code: "a < b[read()] < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a < b?.() < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a < ++b < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a < new Number(b) < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Unlike upstream, the existing side-effect checker recognizes tag
		// calls. Duplicating the middle operand would execute the tag twice.
		{Code: "a < tag`value` < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "async function run() { return a < await read() < c; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 31, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function* run() { return a < (yield b) < c; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a < (b += 1) < c;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "first < middle /* discard */ < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "first < middle < /* discard */ last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "first < middle < (/* preserve */ last);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "first < middle && middle < (/* preserve */ last);"},
			}},
		}},
		{Code: "(/* preserve */ first) < (middle) < ((last));", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "(/* preserve */ first) < (middle) && (middle) < ((last));"},
			}},
		}},
		{Code: "first < /* eslint-disable-line no-alert */ middle < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "first < String.raw`middle` < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "first < String.raw`middle` && String.raw`middle` < last;"},
			}},
		}},
		// Replacement at expression starts
		{Code: "(class {}.length < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "(class {}.length < middle && middle < last);"},
			}},
		}},
		{Code: "(async function () {}.length < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "(async function () {}.length < middle && middle < last);"},
			}},
		}},
		{Code: "export default (function () {}.length < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "export default (function () {}.length < middle && middle < last);"},
			}},
		}},
		{Code: "export default (class {}.length < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "export default (class {}.length < middle && middle < last);"},
			}},
		}},
		{Code: "export default ({value: 1}.value < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "export default {value: 1}.value < middle && middle < last;"},
			}},
		}},
		{Code: "for (({value: 1}.value < middle) < last;;) {}", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "for ({value: 1}.value < middle && middle < last;;) {}"},
			}},
		}},
		{Code: "(/x/.lastIndex < middle) < last;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "no-chained-comparison/suggestion", Output: "/x/.lastIndex < middle && middle < last;"},
			}},
		}},
		// Intentional boolean values and unknown operands
		{Code: "declare const flag: boolean | undefined; (a < b) === flag;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 42, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "declare const flag: any; (a < b) === flag;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "declare const flag: unknown; (a < b) === flag;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 30, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function run(Boolean: (value: unknown) => number) { return (a < b) === Boolean(c); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 60, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a < b) === (c && !d);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a < b) === ((!c) as number);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "async function run() { return (a < b) === await Promise.resolve(true); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 31, EndLine: 1, EndColumn: 70, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function* run() { return (a < b) === (yield true); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @returns {boolean} */ function* read() { yield true; } (a < b) === read();", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 60, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @returns {boolean} */ async function read() { return true; } (a < b) === read();", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 66, EndLine: 1, EndColumn: 84, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @param {(() => boolean) | undefined} read */ function run(read) { return (a < b) === read(); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 78, EndLine: 1, EndColumn: 96, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},

		{Code: "/** @type {number} */ const flag = true; (a < b) === flag;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 42, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const flag = flag; (a < b) === flag;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a < b) === new Number(1);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a < b) === unknownObject?.enabled;", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoChainedComparisonArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		code       string
		suggestion bool
	}{
		{"a < b < c;", true},
		{"a < read() < c;", false},
		{"a < /* keep */ b < c;", false},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_chained_comparison.NoChainedComparisonRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_chained_comparison.NoChainedComparisonRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if diagnostic.FixesPtr != nil {
					t.Errorf("demand %d produced an automatic fix", demand)
				}
				wantSuggestion := testCase.suggestion && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if (diagnostic.Suggestions != nil) != wantSuggestion || (wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions)) {
					t.Errorf("demand %d: unexpected suggestions", demand)
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.code, []rule.RuleDiagnostic{diagnostic})
				if fixed || output != testCase.code {
					t.Errorf("demand %d applied a suggestion as an automatic fix", demand)
				}
			}
			if testCase.suggestion {
				if len(*all.Suggestions) != 1 {
					t.Fatalf("got %d suggestions, want 1", len(*all.Suggestions))
				}
				suggestion := (*all.Suggestions)[0]
				if !reflect.DeepEqual(suggestion.Message, rule.RuleMessage{
					Id: "no-chained-comparison/suggestion", Description: "Replace with `a < b && b < c`.",
					Data: map[string]string{"replacement": "a < b && b < c"},
				}) {
					t.Errorf("unexpected suggestion message: %#v", suggestion.Message)
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.code, []rule.RuleDiagnostic{{FixesPtr: &suggestion.FixesArr}})
				if !fixed || output != "a < b && b < c;" {
					t.Errorf("unexpected suggestion output: %q", output)
				}
			}
		})
	}
}

func TestNoChainedComparisonScriptFunctionsWithoutTypeInfo(t *testing.T) {
	testedRule := no_chained_comparison.NoChainedComparisonRule
	testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return no_chained_comparison.NoChainedComparisonRule.Run(ctx, options)
	}
	script := rule.LanguageOptions{SourceType: "script"}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, []rule_tester.ValidTestCase{
		{Code: `function run() { "use strict"; function read() { return true; } return (a < b) === read(); }`, FileName: "case.js", LanguageOptions: script},
	}, []rule_tester.InvalidTestCase{
		{Code: `function read() { return true; } (a < b) === read();`, FileName: "case.js", LanguageOptions: script, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: `"use strict"; function read() { return true; } (a < b) === read();`, FileName: "case.js", LanguageOptions: script, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 48, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: `function run() { function read() { return true; } return (a < b) === read(); }`, FileName: "case.js", LanguageOptions: script, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 58, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoChainedComparisonBuiltinsWithoutTypeInfo(t *testing.T) {
	testedRule := no_chained_comparison.NoChainedComparisonRule
	testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return no_chained_comparison.NoChainedComparisonRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, []rule_tester.ValidTestCase{
		{Code: "const values = new Set(); (a < b) === values.has(c);", FileName: "case.ts"},
		{Code: "const values = new Map(); (a < b) === values.has(c);", FileName: "case.ts"},
		{Code: "const values = new WeakSet(); (a < b) === values.has(c);", FileName: "case.ts"},
		{Code: "const values = new Set(); (a < b) === values.delete(c);", FileName: "case.ts"},
		{Code: "function read(v: string): boolean; function read(v: number): boolean; function read(v: unknown) { return true; } (a < b) === read(c);", FileName: "case.ts"},
		{Code: "const read = function named(): boolean { return (a < b) === named(); };", FileName: "case.ts"},
	}, []rule_tester.InvalidTestCase{
		{Code: "class Set { has() { return 0; } } const values = new Set(); (a < b) === values.has(c);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Line: 1, Column: 61, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

// Checked against Unicorn v79.0.0 in each mode. Shadowed built-ins retain their
// local declarations' behavior (see the documented upstream difference).
func TestNoChainedComparisonReviewRegressions(t *testing.T) {
	cases := []struct {
		name, code              string
		sourceValid, typedValid bool
	}{
		{"parameter default", "function f(read = () => true) { return (a < b) === read(); } f(() => 0);", false, true},
		{"parameter callback annotation", "function f(read: () => boolean = () => true) { return (a < b) === read(); }", true, true},
		{"parameter JSDoc callback", "/** @param {() => boolean} read */ function f(read = () => true) { return (a < b) === read(); }", true, true},
		{"stable variable callback", "const read = () => true; (a < b) === read();", true, true},
		{"JSDoc native number assertion", "/** @type {boolean} */ const flag = 1 as number; (a < b) === flag;", false, false},
		{"JSDoc native boolean assertion", "/** @type {number} */ const flag = true as boolean; (a < b) === flag;", true, true},
		{"JSDoc native number alias", "/** @type {boolean} */ const flag = 1 as number; const alias = flag; (a < b) === alias;", false, false},
		{"JSDoc native boolean alias", "/** @type {number} */ const flag = true as boolean; const alias = flag; (a < b) === alias;", true, true},
		{"JSDoc native angle number", "/** @type {boolean} */ const flag = <number>1; (a < b) === flag;", false, false},
		{"JSDoc native angle boolean", "/** @type {number} */ const flag = <boolean>true; (a < b) === flag;", true, true},
		{"JSDoc non-nullable boolean", "/** @param {!boolean} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"JSDoc non-nullable number", "/** @param {!number} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"JSDoc boolean intersection", "/** @param {boolean & {brand: string}} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"native boolean intersection", "function f(flag: boolean & {brand: string}) { return (a < b) === flag; }", false, false},
		{"JSDoc number intersection", "/** @param {number & {brand: string}} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"JSDoc nullable boolean", "/** @param {?boolean} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"Closure boolean callback", "/** @param {function():boolean} read */ function f(read) { return (a < b) === read(); }", true, true},
		{"Closure number callback", "/** @param {function():number} read */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boolean variable", "/** @type {function():boolean} */ const read = value; (a < b) === read();", true, true},
		{"Closure number variable", "/** @type {function():number} */ const read = value; (a < b) === read();", false, false},
		{"Closure unknown return", "/** @param {function():*} read */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure function return", "/** @param {function():function():boolean} read */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boolean argument", "/** @param {function(string):boolean} read */ function f(read) { return (a < b) === read(\"x\"); }", true, true},
		{"Closure union return", "/** @param {function():(boolean|number)} read */ function f(read) { return (a < b) === read(); }", false, false},
		{"regexp arrow factory", "const f = (): RegExp => /x/; (a < b) === f().global;", true, true},
		{"regexp function factory", "const f = function(): RegExp { return /x/; }; (a < b) === f().global;", true, true},
		{"set arrow factory", "const f = (): Set<number> => new Set(); (a < b) === f().has(1);", true, true},
		{"map arrow factory", "const f = (): Map<number, number> => new Map(); (a < b) === f().has(1);", true, true},
		{"regexp async factory", "const f = async (): Promise<RegExp> => /x/; (a < b) === f().global;", false, false},
		{"regexp generator factory", "const f = function*(): Generator<RegExp> { yield /x/; }; (a < b) === f().global;", false, false},
		{"regexp factory write", "let f = (): RegExp => /x/; f = () => value; (a < b) === f().global;", false, true},
		{"regexp parameter default", "function f(read = (): RegExp => /x/) { return (a < b) === read().global; }", false, true},
		{"regexp function declaration", "function f(): RegExp { return /x/; } (a < b) === f().global;", true, true},
		{"regexp callback parameter", "function f(read: () => RegExp) { return (a < b) === read().global; }", false, true},
		{"URLSearchParams direct has", "(a < b) === new URLSearchParams().has(\"x\");", true, true},
		{"URLSearchParams alias has", "const value = new URLSearchParams(); const alias = value; (a < b) === alias.has(\"x\");", true, true},
		{"URLSearchParams parameter has", "function f(value: URLSearchParams) { return (a < b) === value.has(\"x\"); }", true, true},
		{"URLSearchParams JSDoc parameter has", "/** @param {URLSearchParams} value */ function f(value) { return (a < b) === value.has(\"x\"); }", true, true},
		{"URLSearchParams shadow has", "function f() { class URLSearchParams { has() { return 0; } } (a < b) === new URLSearchParams().has(\"x\"); }", false, false},
		{"URLSearchParams delete control", "(a < b) === new URLSearchParams().delete(\"x\");", false, false},
		{"Headers direct has", "(a < b) === new Headers().has(\"x\");", true, true},
		{"Headers alias has", "const value = new Headers(); const alias = value; (a < b) === alias.has(\"x\");", true, true},
		{"Headers parameter has", "function f(value: Headers) { return (a < b) === value.has(\"x\"); }", true, true},
		{"Headers JSDoc parameter has", "/** @param {Headers} value */ function f(value) { return (a < b) === value.has(\"x\"); }", true, true},
		{"Headers shadow has", "function f() { class Headers { has() { return 0; } } (a < b) === new Headers().has(\"x\"); }", false, false},
		{"Headers delete control", "(a < b) === new Headers().delete(\"x\");", false, false},
		{"FormData direct has", "(a < b) === new FormData().has(\"x\");", true, true},
		{"FormData alias has", "const value = new FormData(); const alias = value; (a < b) === alias.has(\"x\");", true, true},
		{"FormData parameter has", "function f(value: FormData) { return (a < b) === value.has(\"x\"); }", true, true},
		{"FormData JSDoc parameter has", "/** @param {FormData} value */ function f(value) { return (a < b) === value.has(\"x\"); }", true, true},
		{"FormData shadow has", "function f() { class FormData { has() { return 0; } } (a < b) === new FormData().has(\"x\"); }", false, false},
		{"FormData delete control", "(a < b) === new FormData().delete(\"x\");", false, false},
		{"Closure spaced callback", "/** @param {function (string): boolean} read */ function f(read) { return (a < b) === read(\"x\"); }", true, true},
		{"Closure optional callback", "/** @param {function():boolean} [read] */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure multiple parameters", "/** @param {function():number} wrong\n * @param {function():boolean} read */ function f(wrong, read) { return (a < b) === read(); }", true, true},
		{"Closure alias result", "type Flag = boolean; /** @param {function():Flag} read */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure non-nullable result", "/** @param {function():!boolean} read */ function f(read) { return (a < b) === read(); }", true, true},
		{"Closure native precedence", "/** @param {function():boolean} read */ function f(read: () => number) { return (a < b) === read(); }", false, false},
		{"Closure mutable callback", "/** @type {function():boolean} */ let read = value; read = () => 0; (a < b) === read();", false, false},
		{"JSDoc number static assertion", "/** @type {boolean} */ const flag = true as number; (a < b) === flag;", true, true},
		{"JSDoc boolean satisfies assertion", "/** @type {number} */ const flag = (true as boolean) satisfies boolean; (a < b) === flag;", true, true},
		{"regexp factory assertion", "const f = (((): RegExp => /x/)) as () => number; (a < b) === f().global;", false, false},
		{"boolean callback assertion", "const f = (() => 0) as () => boolean; (a < b) === f();", false, true},
		{"numeric callback assertion", "const f = (((): boolean => true)) as () => number; (a < b) === f();", false, false},
		{"shadowed annotated receiver", "function run() { class Headers { has() { return 0; } } /** @param {Headers} value */ function f(value) { return (a < b) === value.has(\"x\"); } }", false, false},
		{"shadowed exported receiver", "export class Headers { has() { return 0; } } (a < b) === new Headers().has(\"x\");", false, false},
		{"stable let callback", "let read = () => true; (a < b) === read();", true, true},
		{"stable let function expression", "let read = function() { return true; }; (a < b) === read();", true, true},
		{"stable let factory", "let read = (): RegExp => /x/; (a < b) === read().global;", true, true},
		{"written let callback", "let read = () => true; read = () => 0; (a < b) === read();", false, true},
		{"written let factory", "let read = (): RegExp => /x/; read = () => value; (a < b) === read().global;", false, true},
		{"stable var callback", "var read = () => true; (a < b) === read();", true, true},
		{"stable var function expression", "var read = function() { return true; }; (a < b) === read();", true, true},
		{"stable var factory", "var read = (): RegExp => /x/; (a < b) === read().global;", true, true},
		{"written var callback", "var read = () => true; read = () => 0; (a < b) === read();", false, true},
		{"written var factory", "var read = (): RegExp => /x/; read = () => value; (a < b) === read().global;", false, true},
		{"intersection any", "/** @param {boolean & any} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection reversed any", "/** @param {any & boolean} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection unknown", "/** @param {boolean & unknown} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection never", "/** @param {boolean & never} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection number", "/** @param {boolean & number} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection branded any", "/** @param {boolean & {brand: string} & any} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection nested brands", "/** @param {boolean & ({brand: string} & {label: string})} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"intersection nested any", "/** @param {boolean & ({brand: string} & any)} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection boolean literal", "/** @param {true & {brand: string}} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"intersection only brands", "/** @param {{brand: string} & {label: string}} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection callable member", "/** @param {boolean & {(): boolean}} flag */ function f(flag) { return (a < b) === flag; }", false, false},
		{"intersection method brand", "/** @param {boolean & {brand(): string}} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"Closure boundary dotted property", "/** @param {function():boolean} read.callback */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary indexed property", "/** @param {function():boolean} read[\"callback\"] */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary optional dotted property", "/** @param {function():boolean} [read.callback] */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary optional indexed property", "/** @param {function():boolean} [read[\"callback\"]] */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary hyphenated name", "/** @param {function():boolean} read-callback */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary plain description", "/** @param {function():boolean} read callback description */ function f(read) { return (a < b) === read(); }", true, true},
		{"Closure boundary spaced dotted description", "/** @param {function():boolean} read .callback description */ function f(read) { return (a < b) === read(); }", true, true},
		{"Closure boundary Unicode whitespace", "/** @param {function():boolean} read\u00a0callback description */ function f(read) { return (a < b) === read(); }", true, true},
		{"Closure boundary non-whitespace", "/** @param {function():boolean} read\u0085callback */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure boundary optional description", "/** @param {function():boolean} [read] callback description */ function f(read) { return (a < b) === read(); }", false, false},
		{"Closure property then parameter", "/** @param {function():number} read.callback\n * @param {function():boolean} read */ function f(read) { return (a < b) === read(); }", true, true},
		{"intersection non-nullable brand", "/** @param {boolean & (!{brand: string})} flag */ function f(flag) { return (a < b) === flag; }", true, true},
		{"intersection generic object brand", "/** @param {boolean & Object<string, unknown>} flag */ function f(flag) { return (a < b) === flag; }", true, true},
	}
	for _, typed := range []bool{false, true} {
		t.Run(fmt.Sprintf("type-information-%t", typed), func(t *testing.T) {
			testedRule := no_chained_comparison.NoChainedComparisonRule
			testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
				if !typed {
					ctx.TypeChecker = nil
				}
				return no_chained_comparison.NoChainedComparisonRule.Run(ctx, options)
			}
			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					valid := testCase.sourceValid
					if typed {
						valid = testCase.typedValid
					}
					if valid {
						rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule,
							[]rule_tester.ValidTestCase{{Code: testCase.code, FileName: "case.ts"}}, nil)
					} else {
						rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, nil,
							[]rule_tester.InvalidTestCase{{Code: testCase.code, FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "no-chained-comparison/error", Message: comparisonMessage, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
							}}})
					}
				})
			}
		})
	}
}

func TestNoChainedComparisonRepeatedAliases(t *testing.T) {
	for _, callable := range []bool{false, true} {
		var source strings.Builder
		source.WriteString("const b0 = x > 0;\n")
		if callable {
			source.Reset()
			source.WriteString("const b0 = () => x > 0;\n")
		}
		call := ""
		if callable {
			call = "()"
		}
		for index := 1; index <= 64; index++ {
			body := fmt.Sprintf("b%d%s && b%d%s", index-1, call, index-1, call)
			if callable {
				body = "() => " + body
			}
			fmt.Fprintf(&source, "const b%d = %s;\n", index, body)
		}
		fmt.Fprintf(&source, "(a < b) === b64%s; (c < d) === b64%s;", call, call)
		for _, typed := range []bool{false, true} {
			t.Run(fmt.Sprintf("callable-%t-type-information-%t", callable, typed), func(t *testing.T) {
				testedRule := no_chained_comparison.NoChainedComparisonRule
				testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					if !typed {
						ctx.TypeChecker = nil
					}
					return no_chained_comparison.NoChainedComparisonRule.Run(ctx, options)
				}
				rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule,
					[]rule_tester.ValidTestCase{{Code: source.String(), FileName: "case.ts"}}, nil)
			})
		}
	}
}
