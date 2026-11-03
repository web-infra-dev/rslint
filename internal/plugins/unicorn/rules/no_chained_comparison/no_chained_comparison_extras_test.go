package no_chained_comparison_test

import (
	"reflect"
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
