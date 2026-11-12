// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-constants.js
package prefer_math_constants_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_constants"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMathConstantsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Already using the constant.
		{Code: "Math.PI", FileName: "file.js"},
		{Code: "const tau = Math.PI * 2;", FileName: "file.js"},
		{Code: "Math.E", FileName: "file.js"},
		// Too few fraction digits.
		{Code: "3.1", FileName: "file.js"},
		{Code: "2.71", FileName: "file.js"},
		{Code: "1.41", FileName: "file.js"},
		{Code: "0.69", FileName: "file.js"},
		{Code: "2.30", FileName: "file.js"},
		{Code: "1.44", FileName: "file.js"},
		{Code: "0.43", FileName: "file.js"},
		{Code: "0.70", FileName: "file.js"},
		// Different integer part, unrelated digits, extra precision or different rounding.
		{Code: "4.14", FileName: "file.js"},
		{Code: "5.718", FileName: "file.js"},
		{Code: "1.23456", FileName: "file.js"},
		{Code: "0.12345", FileName: "file.js"},
		{Code: "3.1415926535897932", FileName: "file.js"},
		{Code: "1.4143", FileName: "file.js"},
		// Literal types, including negative ones.
		{Code: "type Tau = 3.14;", FileName: "file.ts"},
		{Code: "type Tau = -3.14;", FileName: "file.ts"},
		{Code: "let value: -3.14;", FileName: "file.ts"},
		{Code: "const value = foo as -3.14;", FileName: "file.ts"},
		{Code: "function f(value: -3.14) {}", FileName: "file.ts"},
		{Code: "const x = y as 3.14;", FileName: "file.ts"},
		// Other numeric spellings.
		{Code: "3", FileName: "file.js"},
		{Code: "0x1F", FileName: "file.js"},
		{Code: "0b101", FileName: "file.js"},
		{Code: "3n", FileName: "file.js"},
		{Code: "3.14e2", FileName: "file.js"},
		{Code: "1_000", FileName: "file.js"},
		// Non-computed property keys.
		{Code: "const object = {3.14: 1};", FileName: "file.js"},
		{Code: "class Foo {3.14 = 1}", FileName: "file.js"},
		{Code: "class Foo {3.14() {}}", FileName: "file.js"},
		{Code: "interface Foo {3.14: number}", FileName: "file.ts"},
		// Documentation examples.
		{Code: "const area = Math.PI * radius ** 2;", FileName: "file.js"},
		{Code: "const growth = amount * Math.E;", FileName: "file.js"},
		{Code: "const diagonal = side * Math.SQRT2;", FileName: "file.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// PI.
		approximationCase("3.14", "3.14", "Math.PI", "file.js"),
		approximationCase("const area = 3.14 * radius ** 2;", "3.14", "Math.PI", "file.js"),
		approximationCase("3.141", "3.141", "Math.PI", "file.js"),
		approximationCase("3.1416", "3.1416", "Math.PI", "file.js"),
		approximationCase("3.141592653589793", "3.141592653589793", "Math.PI", "file.js"),
		// E, including rounding.
		approximationCase("2.718", "2.718", "Math.E", "file.js"),
		approximationCase("const growth = amount * 2.718;", "2.718", "Math.E", "file.js"),
		approximationCase("2.718281828459045", "2.718281828459045", "Math.E", "file.js"),
		approximationCase("2.7183", "2.7183", "Math.E", "file.js"),
		// LN2, LN10, LOG2E, LOG10E, SQRT2 and SQRT1_2.
		approximationCase("0.6931", "0.6931", "Math.LN2", "file.js"),
		approximationCase("0.6931471805599453", "0.6931471805599453", "Math.LN2", "file.js"),
		approximationCase("2.3025", "2.3025", "Math.LN10", "file.js"),
		approximationCase("2.3026", "2.3026", "Math.LN10", "file.js"),
		approximationCase("1.4426", "1.4426", "Math.LOG2E", "file.js"),
		approximationCase("1.4427", "1.4427", "Math.LOG2E", "file.js"),
		approximationCase("0.4342", "0.4342", "Math.LOG10E", "file.js"),
		approximationCase("1.4142", "1.4142", "Math.SQRT2", "file.js"),
		approximationCase("0.7071", "0.7071", "Math.SQRT1_2", "file.js"),
		// Negative, leading-dot, separators, and nested expressions.
		approximationCase("-3.14", "3.14", "Math.PI", "file.js"),
		approximationCase(".6931", ".6931", "Math.LN2", "file.js"),
		approximationCase("3.141_592", "3.141_592", "Math.PI", "file.js"),
		approximationCase("1 / 3.14", "3.14", "Math.PI", "file.js"),
		approximationCase("foo(3.14)", "3.14", "Math.PI", "file.js"),
		approximationCase("[2.718]", "2.718", "Math.E", "file.js"),
		// Computed keys, JSX and comments.
		approximationCase("const object = {[3.14]: 1};", "3.14", "Math.PI", "file.js"),
		approximationCase("class Foo {[3.14] = 1}", "3.14", "Math.PI", "file.js"),
		approximationCase("object[3.14]", "3.14", "Math.PI", "file.js"),
		approximationCase("<div>{3.14}</div>", "3.14", "Math.PI", "file.jsx"),
		approximationCase("foo(/* before */ 3.14 /* after */)", "3.14", "Math.PI", "file.js"),
		// TypeScript runtime expressions.
		approximationCase("enum Foo {a = 3.14}", "3.14", "Math.PI", "file.ts"),
		approximationCase("const x = 3.14 satisfies number;", "3.14", "Math.PI", "file.ts"),
		approximationCase("const x = 3.14 as number;", "3.14", "Math.PI", "file.ts"),
		approximationCase("const x = <number>3.14;", "3.14", "Math.PI", "file.ts"),
		approximationCase("const value = -3.14 as number;", "3.14", "Math.PI", "file.ts"),
		// Documentation example.
		approximationCase("const diagonal = side * 1.4142;", "1.4142", "Math.SQRT2", "file.js"),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_constants.PreferMathConstantsRule, valid, invalid)
}

// These cases have a single matching literal on one ASCII line.
func approximationCase(code, literal, constant, filename string) rule_tester.InvalidTestCase {
	before, after, _ := strings.Cut(code, literal)
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: filename,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "prefer-math-constants/error",
			Message:   "Prefer `" + constant + "` over the approximate value `" + literal + "`.",
			Line:      1,
			Column:    len(before) + 1,
			EndLine:   1,
			EndColumn: len(before) + len(literal) + 1,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "prefer-math-constants/suggestion",
				Output:    before + constant + after,
			}},
		}},
	}
}
