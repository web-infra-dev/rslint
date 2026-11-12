// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-abs.js
package prefer_math_abs_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_abs"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMathAbsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Parentheses and comments that cannot be kept.
		{Code: "const value = 1n; (/* keep */ value) < 0 ? -value : value;", FileName: "file.js"},

		// BigInt initializers and default values from annotated calls.
		{Code: "/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "/** @returns {bigint | number} */ function read() { return source; } function run(value = read()) { return value > limit || value < -limit; }", FileName: "file.js"},
		{Code: "/** @param {() => bigint} read */ function run(read, value = read()) { return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "/** @returns {bigint} */ function read() { return source; } let value = read(); value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "/** @returns {bigint | number} */ function read() { return source; } var value = read(); value > limit || value < -limit;", FileName: "file.js"},
		{Code: "/** @returns {bigint} */ function read() { return source; } let value = -read(); value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "/** @returns {bigint} */ function read() { return source; } function run(value, limit = read()) { return value > limit || value < -limit; }", FileName: "file.js"},

		// JSDoc declarations, unions, and operations.
		{Code: "/** @param {bigint} input */ function run(input) { const value = input + 1n; return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "/** @type {bigint} */ let value; value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "/** @param {number | bigint} value */ function run(value) { return 0 < value ? value : -value; }", FileName: "file.js"},
		{Code: "/** @param {?bigint} value */ function run(value) { return value >= 0 ? value : -value; }", FileName: "file.js"},
		{Code: "/** @param {bigint} [value] */ function run(value) { return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "/** @param {bigint} value */ function run(value, limit) { return value > limit || value < -limit; }", FileName: "file.js"},
		{Code: "/** @param {bigint} limit */ function run(value, limit) { return -limit >= value || limit <= value; }", FileName: "file.js"},
		{Code: "/** @returns {bigint} */ function read() { return source; } const value = read(); value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "/** @param {bigint} input */ function run(input, limit) { const value = ~input; return value > limit || value < -limit; }", FileName: "file.js"},

		// TypeScript declarations and JSDoc shared by both upstream project modes.
		{Code: "function read(): bigint { return source; } let value = read(); value < 0 ? -value : value; function run(value = read()) { return value > limit || value < -limit; }", FileName: "file.ts"},
		{Code: "/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }", FileName: "file.ts"},
		{Code: "/** @param {bigint} value */ function run(value) { return value < 0 ? -value : value; }", FileName: "file.ts"},
		{Code: "/** @param {number | bigint} value */ function run(value) { return 0 < value ? value : -value; }", FileName: "file.ts"},
		{Code: "/** @param {bigint} limit */ function run(value, limit) { return -limit >= value || limit <= value; }", FileName: "file.ts"},

		// Runtime BigInt examples.
		{Code: "/** @returns {bigint} */ function read() { return -5n; } function run(value = read()) { return value < 0 ? -value : value; } run();", FileName: "file.js"},
		{Code: "/** @returns {bigint} */ function read() { return -5n; } let value = read(); value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "/** @param {bigint} value */ function run(value) { return value < 0 ? -value : value; } run(-5n);", FileName: "file.js"},
		{Code: "/** @param {bigint} value */ function run(value, limit) { return value > limit || value < -limit; } run(-5n, 2);", FileName: "file.js"},

		// JavaScript snapshot cases.
		{Code: "const absolute = Math.abs(value);", FileName: "file.js"},
		{Code: "value < 0 ? value : -value", FileName: "file.js"},
		{Code: "value > 0 ? -value : value", FileName: "file.js"},
		{Code: "value < 1 ? -value : value", FileName: "file.js"},
		{Code: "value < 0 ? -otherValue : value", FileName: "file.js"},
		{Code: "value < 0 ? -value : otherValue", FileName: "file.js"},
		{Code: "value < 0n ? -value : value", FileName: "file.js"},
		{Code: "value < BigInt(0) ? -value : value", FileName: "file.js"},
		{Code: "const value = 1n; value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "const value = BigInt(1); value < 0 ? -value : value;", FileName: "file.js"},
		{Code: "function foo(value = 1n) { return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "function foo(value = BigInt(1)) { return value < 0 ? -value : value; }", FileName: "file.js"},
		{Code: "value > limit || value < limit", FileName: "file.js"},
		{Code: "value > limit || otherValue < -limit", FileName: "file.js"},
		{Code: "value > limit || value <= -limit", FileName: "file.js"},
		{Code: "value > limit || value < -otherLimit", FileName: "file.js"},
		{Code: "value > 0n || value < -0n", FileName: "file.js"},
		{Code: "value > BigInt(0) || value < -BigInt(0)", FileName: "file.js"},
		{Code: "const value = 1n; const limit = 2n; value > limit || value < -limit;", FileName: "file.js"},
		{Code: "(value + 1) < 0 ? -(value + 1) : value + 1", FileName: "file.js"},
		{Code: "Math.abs(value) > limit", FileName: "file.js"},
		{Code: "Math.abs(value) >= limit", FileName: "file.js"},
		{Code: "value > limit && value < -limit", FileName: "file.js"},

		// TypeScript snapshot cases.
		{Code: "function foo(value: bigint) {\n\treturn value < 0 ? -value : value;\n}", FileName: "file.ts"},
		{Code: "function foo(value) {\n\treturn (value as bigint) < 0 ? -(value as bigint) : (value as bigint);\n}", FileName: "file.ts"},
		{Code: "function foo(value) {\n\treturn (<bigint>value) < 0 ? -(<bigint>value) : (<bigint>value);\n}", FileName: "file.ts"},
		{Code: "function foo(value: bigint, limit: bigint) {\n\treturn value > limit || value < -limit;\n}", FileName: "file.ts"},
		{Code: "const value: bigint = 1n;\nvalue < 0 ? -value : value;", FileName: "file.ts"},
		{Code: "function foo(value: number | bigint) {\n\treturn value < 0 ? -value : value;\n}", FileName: "file.ts"},
		{Code: "function foo(value) {\n\treturn (value as number | bigint) < 0 ? -(value as number | bigint) : (value as number | bigint);\n}", FileName: "file.ts"},
		{Code: "function foo(value: bigint & {}) {\n\treturn value < 0 ? -value : value;\n}", FileName: "file.ts"},
		{Code: "const value: 1n = 1n;\nvalue < 0 ? -value : value;", FileName: "file.ts"},
		{Code: "function foo(value) {\n\treturn (value satisfies bigint) < 0 ? -(value satisfies bigint) : (value satisfies bigint);\n}", FileName: "file.ts"},
		{Code: "function foo(value) {\n\treturn (value as bigint)! < 0 ? -(value as bigint)! : (value as bigint)!;\n}", FileName: "file.ts"},

		// Upstream documentation examples.
		{Code: "const absolute = Math.abs(value);", FileName: "file.js"},
		{Code: "if (Math.abs(number) > MAXIMUM) {\n}", FileName: "file.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Comments kept in the selected operands.
		absCase("(/* keep */ value) < 0 ? -value : value;", "Math.abs((/* keep */ value));", "file.js", 1, 1, 1, 40),
		absCase(
			"value > (/* keep */ limit) || value < -limit;",
			"Math.abs(value) > (/* keep */ limit);",
			"file.js", 1, 1, 1, 45,
		),
		absCase(
			"value < -limit || (/* keep */ value) > limit;",
			"Math.abs((/* keep */ value)) > limit;",
			"file.js", 1, 1, 1, 45,
		),
		absCase(
			"(value as /* keep */ number) < 0 ? -value : value;",
			"Math.abs((value as /* keep */ number));",
			"file.ts", 1, 1, 1, 50,
		),

		// Parentheses and comments that cannot be kept.
		absCase("((value /* keep */)) >= 0 ? value : -value;", "Math.abs(((value /* keep */)));", "file.js", 1, 1, 1, 43),
		absCase(
			"value > (limit // keep\r\n) || value < -limit;",
			"Math.abs(value) > (limit // keep\r\n);",
			"file.js", 1, 1, 2, 20,
		),
		absCase(
			"limit as /* keep */ number < value || value < -(limit as number);",
			"Math.abs(value) > (limit as /* keep */ number);",
			"file.ts", 1, 1, 1, 65,
		),
		absCase("value /* keep */ < 0 ? -value : value;", "", "file.js", 1, 1, 1, 38),
		absCase("value < 0 ? -(/* keep */ value) : value;", "", "file.js", 1, 1, 1, 40),
		absCase("value < 0 ? -value : (/* keep */ value);", "", "file.js", 1, 1, 1, 40),
		absCase("value > limit || (/* keep */ value) < -limit;", "", "file.js", 1, 1, 1, 45),
		absCase("value > limit || value < -(/* keep */ limit);", "", "file.js", 1, 1, 1, 45),
		absCase("(/* keep */ object.value) < 0 ? -object.value : object.value;", "", "file.js", 1, 1, 1, 61),
		absCase("value > (/* keep */ object.limit) || value < -object.limit;", "", "file.js", 1, 1, 1, 59),

		// Suggestion-only rewrite for coercion and signed zero.
		absCase("(/* keep */ value) < 0 ? -value : value;", "Math.abs((/* keep */ value));", "file.js", 1, 1, 1, 40),

		// BigInt initializers and default values from annotated calls.
		absCase(
			"/** @returns {number} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }",
			"/** @returns {number} */ function read() { return source; } function run(value = read()) { return Math.abs(value); }",
			"file.js", 1, 99, 1, 125,
		),
		absCase(
			"/** @returns {number} */ function read() { return source; } let value = read(); value > limit || value < -limit;",
			"/** @returns {number} */ function read() { return source; } let value = read(); Math.abs(value) > limit;",
			"file.js", 1, 81, 1, 112,
		),

		// JSDoc declarations, unions, and operations.
		absCase(
			"/** @param {number} value */ function run(value) { return value < 0 ? -value : value; }",
			"/** @param {number} value */ function run(value) { return Math.abs(value); }",
			"file.js", 1, 59, 1, 85,
		),
		absCase(
			"/** @param {number} value */ function run(value, limit) { return value > limit || value < -limit; }",
			"/** @param {number} value */ function run(value, limit) { return Math.abs(value) > limit; }",
			"file.js", 1, 66, 1, 97,
		),

		// TypeScript declarations and JSDoc, with and without a project.
		absCase(
			"/** @param {bigint} value */ function run(value: number) { return value < 0 ? -value : value; }",
			"/** @param {bigint} value */ function run(value: number) { return Math.abs(value); }",
			"file.ts", 1, 67, 1, 93,
		),

		// TypeScript assertions used as the range limit.
		absCase(
			"limit as number < value || value < -(limit as number)",
			"Math.abs(value) > (limit as number)",
			"file.ts", 1, 1, 1, 54,
		),
		absCase(
			"value <= -(limit as number) || limit as number <= value",
			"Math.abs(value) >= (limit as number)",
			"file.ts", 1, 1, 1, 56,
		),
		absCase(
			"(limit as number) < value || value < -(limit as number)",
			"Math.abs(value) > (limit as number)",
			"file.ts", 1, 1, 1, 56,
		),
		absCase(
			"limit satisfies number < value || value < -(limit satisfies number)",
			"Math.abs(value) > (limit satisfies number)",
			"file.ts", 1, 1, 1, 68,
		),
		absCase(
			"value <= -(limit satisfies number) || limit satisfies number <= value",
			"Math.abs(value) >= (limit satisfies number)",
			"file.ts", 1, 1, 1, 70,
		),
		absCase(
			"(limit satisfies number) < value || value < -(limit satisfies number)",
			"Math.abs(value) > (limit satisfies number)",
			"file.ts", 1, 1, 1, 70,
		),

		// JavaScript snapshot cases.
		absCase("value < 0 ? -value : value", "Math.abs(value)", "file.js", 1, 1, 1, 27),
		absCase("value <= 0 ? -value : value", "Math.abs(value)", "file.js", 1, 1, 1, 28),
		absCase("value > 0 ? value : -value", "Math.abs(value)", "file.js", 1, 1, 1, 27),
		absCase("value >= 0 ? value : -value", "Math.abs(value)", "file.js", 1, 1, 1, 28),
		absCase("value < 0 ? 0 - value : value", "Math.abs(value)", "file.js", 1, 1, 1, 30),
		absCase("value <= 0 ? 0 - value : value", "Math.abs(value)", "file.js", 1, 1, 1, 31),
		absCase("value < 0 ? value * -1 : value", "Math.abs(value)", "file.js", 1, 1, 1, 31),
		absCase("value < 0 ? -1 * value : value", "Math.abs(value)", "file.js", 1, 1, 1, 31),
		absCase("object.value < 0 ? -object.value : object.value", "", "file.js", 1, 1, 1, 48),
		absCase("0 > value ? -value : value", "Math.abs(value)", "file.js", 1, 1, 1, 27),
		absCase("0 >= value ? -value : value", "Math.abs(value)", "file.js", 1, 1, 1, 28),
		absCase("0 < value ? value : -value", "Math.abs(value)", "file.js", 1, 1, 1, 27),
		absCase("0 <= value ? value : -value", "Math.abs(value)", "file.js", 1, 1, 1, 28),
		absCase(
			"function foo() {return value < 0 ? -value : value;}",
			"function foo() {return Math.abs(value);}",
			"file.js", 1, 24, 1, 50,
		),
		absCase(
			"function foo() {return(value < 0 ? -value : value);}",
			"function foo() {return (Math.abs(value));}",
			"file.js", 1, 24, 1, 50,
		),
		absCase(
			"const absolute = value > limit || value < -limit;",
			"const absolute = Math.abs(value) > limit;",
			"file.js", 1, 18, 1, 49,
		),
		absCase(
			"const absolute = value >= limit || value <= -limit;",
			"const absolute = Math.abs(value) >= limit;",
			"file.js", 1, 18, 1, 51,
		),
		absCase(
			"const absolute = value < -limit || value > limit;",
			"const absolute = Math.abs(value) > limit;",
			"file.js", 1, 18, 1, 49,
		),
		absCase(
			"const absolute = value <= -limit || value >= limit;",
			"const absolute = Math.abs(value) >= limit;",
			"file.js", 1, 18, 1, 51,
		),
		absCase(
			"const absolute = limit < value || -limit > value;",
			"const absolute = Math.abs(value) > limit;",
			"file.js", 1, 18, 1, 49,
		),
		absCase(
			"const absolute = limit <= value || -limit >= value;",
			"const absolute = Math.abs(value) >= limit;",
			"file.js", 1, 18, 1, 51,
		),
		absCase(
			"function foo() {return value > limit || value < -limit;}",
			"function foo() {return Math.abs(value) > limit;}",
			"file.js", 1, 24, 1, 55,
		),
		absCase(
			"function foo() {return(value > limit || value < -limit);}",
			"function foo() {return (Math.abs(value) > limit);}",
			"file.js", 1, 24, 1, 55,
		),
		absCase("const absolute = object.value > limit || object.value < -limit;", "", "file.js", 1, 18, 1, 63),
		absCase("const absolute = value > object.limit || value < -object.limit;", "", "file.js", 1, 18, 1, 63),
		absCase("value < 0 ? -/* comment */value : value", "", "file.js", 1, 1, 1, 40),
		absCase("value > limit || value < -/* comment */limit", "", "file.js", 1, 1, 1, 45),
		absCase(
			"if (number > POSITIVE_CONSTANT || number < -POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}",
			"if (Math.abs(number) > POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}",
			"file.js", 1, 5, 1, 62,
		),
		absCase(
			"if (number >= POSITIVE_CONSTANT || number <= -POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}",
			"if (Math.abs(number) >= POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}",
			"file.js", 1, 5, 1, 64,
		),
		absCase(
			"import {value} from \"./value.js\"; value < 0 ? -value : value;",
			"import {value} from \"./value.js\"; Math.abs(value);",
			"file.js", 1, 35, 1, 61,
		),

		// TypeScript snapshot cases.
		absCase(
			"function foo(value: number) {\n\treturn value < 0 ? -value : value;\n}",
			"function foo(value: number) {\n\treturn Math.abs(value);\n}",
			"file.ts", 2, 9, 2, 35,
		),
		absCase(
			"function foo(value) {\n\treturn (value satisfies number) < 0 ? -(value satisfies number) : (value satisfies number);\n}",
			"function foo(value) {\n\treturn Math.abs((value satisfies number));\n}",
			"file.ts", 2, 9, 2, 92,
		),

		// Suggestion-only ternaries.
		absCase("const a = b < 0 ? -b : b;", "const a = Math.abs(b);", "file.js", 1, 11, 1, 25),
		absCase("const a = b <= 0 ? -b : b;", "const a = Math.abs(b);", "file.js", 1, 11, 1, 26),
		absCase("const a = b > 0 ? b : -b;", "const a = Math.abs(b);", "file.js", 1, 11, 1, 25),
		absCase("const a = b >= 0 ? b : -b;", "const a = Math.abs(b);", "file.js", 1, 11, 1, 26),

		// Upstream documentation examples.
		absCase("const absolute = value < 0 ? -value : value;", "const absolute = Math.abs(value);", "file.js", 1, 18, 1, 44),
		absCase(
			"const absolute = value <= 0 ? 0 - value : value;",
			"const absolute = Math.abs(value);",
			"file.js", 1, 18, 1, 48,
		),
		absCase("const absolute = value > 0 ? value : -value;", "const absolute = Math.abs(value);", "file.js", 1, 18, 1, 44),
		absCase(
			"if (number > MAXIMUM || number < -MAXIMUM) {\n}",
			"if (Math.abs(number) > MAXIMUM) {\n}",
			"file.js", 1, 5, 1, 42,
		),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_abs.PreferMathAbsRule, valid, invalid)
}

func absCase(code, suggestion, filename string, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "prefer-math-abs",
		Message:   "Prefer `Math.abs()` to simplify this expression.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
	if suggestion != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "prefer-math-abs/suggestion",
			Output:    suggestion,
		}}
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: filename, Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
}
