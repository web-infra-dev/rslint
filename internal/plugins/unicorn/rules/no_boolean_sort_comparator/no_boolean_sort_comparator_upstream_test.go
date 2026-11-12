// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-boolean-sort-comparator.js
// The same cases run through multiple upstream parsers. Each input is listed once.
package no_boolean_sort_comparator_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_boolean_sort_comparator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-boolean-sort-comparator"
const message = "Do not return a boolean from a sort comparator."
const suggestionID = "no-boolean-sort-comparator/suggestion"

func TestNoBooleanSortComparatorUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// JSDoc and Reflect cases
		{Code: "function run(Reflect) { values.sort(Reflect.set); }", FileName: "file.ts"},
		{Code: "values.sort(Reflect.get);", FileName: "file.ts"},
		{Code: "values.toSorted(Object.preventExtensions);", FileName: "file.ts"},
		{Code: "/** @param {Set<object>} values */ function run(values) { values.sort(Reflect.isExtensible); }", FileName: "file.ts"},
		{Code: "/** @returns {number} */ function compare(a, b) { return source; } values.sort(compare);", FileName: "file.ts"},
		{Code: "/** @returns {?boolean} */ function compare(a, b) { return source; } values.sort(compare);", FileName: "file.ts"},
		{Code: "/** @returns {boolean} */ async function compare(a, b) { return source; } values.sort(compare);", FileName: "file.ts"},
		{Code: "/** @returns {boolean} */ let compare = (a, b) => source; compare = replacement; values.sort(compare);", FileName: "file.ts"},
		{Code: "/** @param {Set<number>} values */ function run(values) { values.toSorted((a, b) => a > b); }", FileName: "file.ts"},
		// Native return types take precedence over JSDoc
		{Code: "/** @type {() => boolean} */ const compare = ((): number => source) satisfies Function; values.sort(compare);", FileName: "file.ts"},
		// Upstream snapshot cases
		{Code: "function f(set: Set<number>) { set.sort((a, b) => a > b); }", FileName: "case.ts"},
		{Code: "array.sort()", FileName: "case.js"},
		{Code: "array.toSorted()", FileName: "case.js"},
		{Code: "array.sort(compareFunction)", FileName: "case.js"},
		{Code: "array.sort((a, b) => a - b)", FileName: "case.js"},
		{Code: "array.sort((a, b) => b - a)", FileName: "case.js"},
		{Code: "array.sort((a, b) => a.score - b.score)", FileName: "case.js"},
		{Code: "array.sort((a, b) => a.localeCompare(b))", FileName: "case.js"},
		{Code: "array.toSorted((a, b) => a.localeCompare(b))", FileName: "case.js"},
		{Code: "array.sort((a, b) => a > b ? 1 : -1)", FileName: "case.js"},
		{Code: "array.sort((a, b) => a > b ? 1 : a < b ? -1 : 0)", FileName: "case.js"},
		{Code: "array.sort((a, b) => a.foo - b.foo || a.bar - b.bar)", FileName: "case.js"},
		{Code: "array.sort((a, b) => Math.random() - 0.5)", FileName: "case.js"},
		{Code: "array.sort(...comparators)", FileName: "case.js"},
		{Code: "array.sort(...comparators, extraArgument)", FileName: "case.js"},
		{Code: "array[sort]((a, b) => a > b)", FileName: "case.js"},
		{Code: "new Set().sort((a, b) => a > b)", FileName: "case.js"},
		{Code: "({sort() {}}).sort((a, b) => a > b)", FileName: "case.js"},
		{Code: "const collection = new Set(); collection.sort((a, b) => a > b)", FileName: "case.js"},
		{Code: "const object = {}; object.sort((a, b) => a > b)", FileName: "case.js"},
		{Code: "const function_ = () => {}; function_.sort((a, b) => a > b)", FileName: "case.js"},
		{Code: "array.sort(async (a, b) => a > b)", FileName: "case.js"},
		{Code: "array.sort(function * (a, b) { return a > b; })", FileName: "case.js"},
		{Code: "array.sort((a, b) => unknownBoolean)", FileName: "case.js"},
		{Code: "const Boolean = () => 0;\narray.sort(Boolean)", FileName: "case.js"},
		{Code: "const isFinite = () => 0;\narray.sort(isFinite)", FileName: "case.js"},
		{Code: "const Number = {isFinite() { return 0; }};\narray.sort(Number.isFinite)", FileName: "case.js"},
		{Code: "const Number = {isNaN() { return 0; }};\narray.sort(Number[\"isNaN\"])", FileName: "case.js"},
		{Code: "const Array = {isArray() { return 0; }};\narray.sort(Array.isArray)", FileName: "case.js"},
		{Code: "const Object = {is() { return 0; }};\narray.sort(Object.is)", FileName: "case.js"},
		{Code: "array.sort((a, b) => {\n\tif (a > b) {\n\t\treturn true;\n\t}\n\n\treturn false;\n})", FileName: "case.js"},
		{Code: "function compare(a, b) { return a - b; }\narray.sort(compare)", FileName: "case.js"},
		// Documentation examples
		{Code: "// ✅\narray.sort((a, b) => a - b);", FileName: "case.js"},
		{Code: "// ✅\narray.sort((a, b) => a.score - b.score);", FileName: "case.js"},
		{Code: "// ✅\nstrings.toSorted((a, b) => a.localeCompare(b));", FileName: "case.js"},
		{Code: "// ✅ Sorting strings\narray.sort((a, b) => a.localeCompare(b));", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// JSDoc and Reflect cases
		invalidComparator("values.sort(Reflect.defineProperty);", "file.ts", 1, 13, 1, 35),
		invalidComparator("values.sort(Reflect.isExtensible);", "file.ts", 1, 13, 1, 33),
		invalidComparator("values.sort(Reflect.preventExtensions);", "file.ts", 1, 13, 1, 38),
		invalidComparator("values.sort(Reflect.set);", "file.ts", 1, 13, 1, 24),
		invalidComparator("values.sort(Reflect.setPrototypeOf);", "file.ts", 1, 13, 1, 35),
		invalidComparator("values.toSorted(Reflect[\"isExtensible\"]);", "file.ts", 1, 17, 1, 40),
		invalidComparator("/** @returns {boolean} */ function compare(a, b) { return source; } values.sort(compare);", "file.ts", 1, 81, 1, 88),
		invalidComparator("/** @param {(a: number, b: number) => boolean} compare */ function run(compare) { values.toSorted(compare); }", "file.ts", 1, 99, 1, 106),
		invalidComparator("/** @type {{(a: number, b: number): boolean; (): boolean}} */ const compare = source; values.sort(compare);", "file.ts", 1, 99, 1, 106),
		invalidComparator("values.sort(/** @returns {boolean} */ (a, b) => source);", "file.ts", 1, 39, 1, 55),
		invalidComparator("/** @param {Float32Array} values */ function run(values) { values.toSorted((a, b) => a < b); }", "file.ts", 1, 76, 1, 91, "/** @param {Float32Array} values */ function run(values) { values.toSorted((a, b) => b - a); }"),
		// Native return types take precedence over JSDoc
		invalidComparator("/** @type {() => number} */ const compare = ((): boolean => source) satisfies Function; values.sort(compare);", "file.ts", 1, 101, 1, 108),
		// Upstream snapshot cases
		invalidComparator("array.sort((a, b) => a > b)", "case.js", 1, 12, 1, 27, "array.sort((a, b) => a - b)"),
		invalidComparator("[].sort((a, b) => a > b)", "case.js", 1, 9, 1, 24, "[].sort((a, b) => a - b)"),
		invalidComparator("Array.from(iterable).sort((a, b) => a > b)", "case.js", 1, 27, 1, 42, "Array.from(iterable).sort((a, b) => a - b)"),
		invalidComparator("new Array().toSorted((a, b) => a > b)", "case.js", 1, 22, 1, 37, "new Array().toSorted((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => a >= b)", "case.js", 1, 12, 1, 28, "array.sort((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => a < b)", "case.js", 1, 12, 1, 27, "array.sort((a, b) => b - a)"),
		invalidComparator("array.sort((a, b) => a <= b)", "case.js", 1, 12, 1, 28, "array.sort((a, b) => b - a)"),
		invalidComparator("array.sort((a, b) => b > a)", "case.js", 1, 12, 1, 27, "array.sort((a, b) => b - a)"),
		invalidComparator("array.sort((a, b) => b >= a)", "case.js", 1, 12, 1, 28, "array.sort((a, b) => b - a)"),
		invalidComparator("array.sort((a, b) => b < a)", "case.js", 1, 12, 1, 27, "array.sort((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => b <= a)", "case.js", 1, 12, 1, 28, "array.sort((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => a.score > b.score)", "case.js", 1, 12, 1, 39, "array.sort((a, b) => a.score - b.score)"),
		invalidComparator("array.sort((a, b) => a.score >= b.score)", "case.js", 1, 12, 1, 40, "array.sort((a, b) => a.score - b.score)"),
		invalidComparator("array.sort((a, b) => a[0] < b[0])", "case.js", 1, 12, 1, 33, "array.sort((a, b) => b[0] - a[0])"),
		invalidComparator("array.sort((a, b) => a[i] <= b[i])", "case.js", 1, 12, 1, 34, "array.sort((a, b) => b[i] - a[i])"),
		invalidComparator("array.sort((a, b) => a.foo.bar > b.foo.bar)", "case.js", 1, 12, 1, 43, "array.sort((a, b) => a.foo.bar - b.foo.bar)"),
		invalidComparator("array.sort((a, b) => a > c)", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort((a, b) => a.foo > b.bar)", "case.js", 1, 12, 1, 35),
		invalidComparator("array.sort((a, b) => a[i] > b[j])", "case.js", 1, 12, 1, 33),
		invalidComparator("array.sort((a, b) => a > b, extraArgument)", "case.js", 1, 12, 1, 27, "array.sort((a, b) => a - b, extraArgument)"),
		invalidComparator("array[\"sort\"]((a, b) => a > b)", "case.js", 1, 15, 1, 30, "array[\"sort\"]((a, b) => a - b)"),
		invalidComparator("array[\"toSorted\"]((a, b) => a > b)", "case.js", 1, 19, 1, 34, "array[\"toSorted\"]((a, b) => a - b)"),
		invalidComparator("const method = \"sort\";\narray[method]((a, b) => a > b)", "case.js", 2, 15, 2, 30, "const method = \"sort\";\narray[method]((a, b) => a - b)"),
		invalidComparator("array.toSorted((a, b) => a > b)", "case.js", 1, 16, 1, 31, "array.toSorted((a, b) => a - b)"),
		invalidComparator("array.sort?.((a, b) => a > b)", "case.js", 1, 14, 1, 29, "array.sort?.((a, b) => a - b)"),
		invalidComparator("array?.sort((a, b) => a > b)", "case.js", 1, 13, 1, 28, "array?.sort((a, b) => a - b)"),
		invalidComparator("array.sort(function (a, b) { return a > b; })", "case.js", 1, 12, 1, 45, "array.sort((a, b) => a - b)"),
		asScript(invalidComparator("array.sort(function (a, a) { return a > a; })", "case.js", 1, 12, 1, 45)),
		invalidComparator("array.sort((a, b) => {\n\treturn a > b;\n})", "case.js", 1, 12, 3, 2, "array.sort((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => (a) > (b))", "case.js", 1, 12, 1, 31, "array.sort((a, b) => a - b)"),
		invalidComparator("array.sort((a, b) => {\n\t// Compare\n\treturn a > b;\n})", "case.js", 1, 12, 4, 2),
		invalidComparator("array.sort((a, b) => a === b)", "case.js", 1, 12, 1, 29),
		invalidComparator("strings.toSorted((a, b) => a === b)", "case.js", 1, 18, 1, 35),
		invalidComparator("array.sort((a, b) => a !== b)", "case.js", 1, 12, 1, 29),
		invalidComparator("array.sort((a, b) => true)", "case.js", 1, 12, 1, 26),
		invalidComparator("array.sort((a, b) => false)", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort((a, b) => Boolean(a - b))", "case.js", 1, 12, 1, 36),
		invalidComparator("array.sort((a, b) => !(a - b))", "case.js", 1, 12, 1, 30),
		invalidComparator("array.sort((a, b) => a > b && c > d)", "case.js", 1, 12, 1, 36),
		invalidComparator("array.sort((a, b) => a > b ? c > d : e > f)", "case.js", 1, 12, 1, 43),
		invalidComparator("array.sort((a, b) => a in b)", "case.js", 1, 12, 1, 28),
		invalidComparator("array.sort((a, b) => a instanceof b)", "case.js", 1, 12, 1, 36),
		invalidComparator("array.sort(Boolean)", "case.js", 1, 12, 1, 19),
		invalidComparator("array.toSorted(Boolean)", "case.js", 1, 16, 1, 23),
		invalidComparator("array.sort(isFinite)", "case.js", 1, 12, 1, 20),
		invalidComparator("array.sort(isNaN)", "case.js", 1, 12, 1, 17),
		invalidComparator("array.sort(Array.isArray)", "case.js", 1, 12, 1, 25),
		invalidComparator("array.sort(ArrayBuffer.isView)", "case.js", 1, 12, 1, 30),
		invalidComparator("array.sort(Atomics.isLockFree)", "case.js", 1, 12, 1, 30),
		invalidComparator("array.sort(Number.isFinite)", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort(Number.isInteger)", "case.js", 1, 12, 1, 28),
		invalidComparator("array.sort(Number.isNaN)", "case.js", 1, 12, 1, 24),
		invalidComparator("array.sort(Number[\"isNaN\"])", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort(Number.isSafeInteger)", "case.js", 1, 12, 1, 32),
		invalidComparator("array.sort(Object.hasOwn)", "case.js", 1, 12, 1, 25),
		invalidComparator("array.sort(Object.is)", "case.js", 1, 12, 1, 21),
		invalidComparator("array.sort(Object.isExtensible)", "case.js", 1, 12, 1, 31),
		invalidComparator("array.sort(Object.isFrozen)", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort(Object.isSealed)", "case.js", 1, 12, 1, 27),
		invalidComparator("array.sort(Reflect.deleteProperty)", "case.js", 1, 12, 1, 34),
		invalidComparator("array.sort(Reflect.has)", "case.js", 1, 12, 1, 23),
		withURL(invalidComparator("array.sort(URL.canParse)", "case.js", 1, 12, 1, 24)),
		invalidComparator("const compare = (a, b) => a > b;\narray.sort(compare)", "case.js", 2, 12, 2, 19),
		invalidComparator("const compare = function (a, b) { return a > b; };\narray.sort(compare)", "case.js", 2, 12, 2, 19),
		invalidComparator("function compare(a, b) { return a > b; }\narray.sort(compare)", "case.js", 2, 12, 2, 19),
		invalidComparator("array.sort((a: number, b: number): boolean => a > b)", "case.ts", 1, 12, 1, 52),
		invalidComparator("array.sort((Boolean as (value: unknown) => boolean))", "case.ts", 1, 13, 1, 51),
		invalidComparator("const compare: (a: number, b: number) => boolean = (a, b) => compareUnknown(a, b);\narray.sort(compare)", "case.ts", 2, 12, 2, 19),
		invalidComparator("type Comparator = (a: number, b: number) => boolean;\narray.sort(((a, b) => a > b) as Comparator)", "case.ts", 2, 12, 2, 43),
		invalidComparator("type Comparator = (a: number, b: number) => boolean;\narray.sort(compare as Comparator)", "case.ts", 2, 12, 2, 33),
		invalidComparator("array.sort(((a: number, b: number) => a > b) as (a: number, b: number) => boolean)", "case.ts", 1, 12, 1, 82),
		invalidComparator("array.sort(compare as (a: number, b: number) => boolean)", "case.ts", 1, 12, 1, 56),
		invalidComparator("array.sort((a: number, b: number): boolean => compare(a, b))", "case.ts", 1, 12, 1, 60),
		invalidComparator("array.sort(function (a: number, b: number): boolean {\n\treturn compare(a, b);\n})", "case.ts", 1, 12, 3, 2),
		invalidComparator("array.sort((a: number, b: number) => (a > b) as boolean)", "case.ts", 1, 12, 1, 56),
		invalidComparator("array.sort((a: number, b: number) => <boolean>(a > b))", "case.ts", 1, 12, 1, 54),
		invalidComparator("array.sort((a: number, b: number) => (a > b) satisfies boolean)", "case.ts", 1, 12, 1, 63),
		invalidComparator("array.sort((a: number, b: number) => (a > b)!)", "case.ts", 1, 12, 1, 46),
		invalidComparator("array.sort(<T extends {score: number}>(a, b) => a.score > b.score)", "case.ts", 1, 12, 1, 66),
		invalidComparator("array.sort((a?, b?) => a > b)", "case.ts", 1, 12, 1, 29),
		invalidComparator("array.sort((a: number, b: number): boolean => {\n\tif (a > b) {\n\t\treturn true;\n\t}\n\n\treturn false;\n})", "case.ts", 1, 12, 7, 2),
		invalidComparator("array.sort((a: unknown, b: unknown) => (a as number) > (b as number))", "case.ts", 1, 12, 1, 69),
		invalidComparator("array.sort(Boolean!)", "case.ts", 1, 12, 1, 20),
		invalidComparator("array.sort(([a], [b]) => a > b)", "case.js", 1, 12, 1, 31),
		invalidComparator("function f(array: Int8Array) { array.sort((a, b) => a > b); }", "case.ts", 1, 43, 1, 58, "function f(array: Int8Array) { array.sort((a, b) => a - b); }"),
		invalidComparator("array.sort((a, b) => a.score > b)", "case.js", 1, 12, 1, 33),
		invalidComparator("array.sort((a, b) => (a as number) > b)", "case.ts", 1, 12, 1, 39),
		invalidComparator("array.sort((compare as (a: number, b: number) => boolean)!)", "case.ts", 1, 12, 1, 59),
		// Documentation examples
		invalidComparator("// ❌\narray.sort((a, b) => a > b);", "case.js", 2, 12, 2, 27, "// ❌\narray.sort((a, b) => a - b);"),
		invalidComparator("// ❌\narray.sort((a, b) => a.score >= b.score);", "case.js", 2, 12, 2, 40, "// ❌\narray.sort((a, b) => a.score - b.score);"),
		invalidComparator("// ❌\nstrings.toSorted((a, b) => a === b);", "case.js", 2, 18, 2, 35),
	}
	run := func(t *testing.T, testedRule rule.Rule) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, valid, invalid)
	}
	t.Run("with type information", func(t *testing.T) {
		run(t, no_boolean_sort_comparator.NoBooleanSortComparatorRule)
	})
	t.Run("without type information", func(t *testing.T) {
		testedRule := no_boolean_sort_comparator.NoBooleanSortComparatorRule
		testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			ctx.TypeChecker = nil
			return no_boolean_sort_comparator.NoBooleanSortComparatorRule.Run(ctx, options)
		}
		run(t, testedRule)
	})
}

func invalidComparator(code, file string, line, column, endLine, endColumn int, output ...string) rule_tester.InvalidTestCase {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if len(output) > 0 {
		expected.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: suggestionID, Output: output[0]}}
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: file, Errors: []rule_tester.InvalidTestCaseError{expected}}
}

func asScript(test rule_tester.InvalidTestCase) rule_tester.InvalidTestCase {
	test.LanguageOptions.SourceType = "script"
	return test
}

func withURL(test rule_tester.InvalidTestCase) rule_tester.InvalidTestCase {
	test.Globals = map[string]any{"URL": "readonly"}
	return test
}
