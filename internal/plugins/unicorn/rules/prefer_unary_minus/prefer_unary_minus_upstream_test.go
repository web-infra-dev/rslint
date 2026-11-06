// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-unary-minus.js
package prefer_unary_minus_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_unary_minus"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferUnaryMinusUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_unary_minus.PreferUnaryMinusRule, []rule_tester.ValidTestCase{
		// Snapshot cases.
		{Code: "x * 2;", FileName: "case.js"},
		{Code: "x / 2;", FileName: "case.js"},
		{Code: "x * -2;", FileName: "case.js"},
		{Code: "x / -2;", FileName: "case.js"},
		{Code: "-1 * -1;", FileName: "case.js"},
		{Code: "-1 / -1;", FileName: "case.js"},
		{Code: "-1 / x;", FileName: "case.js"},
		{Code: "-1 / (a + b);", FileName: "case.js"},
		{Code: "x * -1n;", FileName: "case.js"},
		{Code: "x / -1n;", FileName: "case.js"},
		{Code: "-1n * x;", FileName: "case.js"},
		{Code: "-1n / x;", FileName: "case.js"},
		{Code: "x ** -1;", FileName: "case.js"},
		{Code: "x - 1;", FileName: "case.js"},
		{Code: "x + -1;", FileName: "case.js"},
		{Code: "x % -1;", FileName: "case.js"},
		{Code: "x & -1;", FileName: "case.js"},
		// Documentation examples.
		{Code: "const inverted = -value;", FileName: "case.js"},
		{Code: "const flipped = -someNumber;", FileName: "case.js"},
		{Code: "const divided = -value;", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		// Trailing operator keywords.
		{Code: "getObject(/* keep */).value * (-1)in object;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value in object;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		{Code: "getObject(/* keep */).value / (-1)instanceof Number;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value instanceof Number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		{Code: "-1 * (getObject(/* keep */).value)in object;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value in object;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		{Code: "getObject(/* keep */).value * (-1)instanceof Number;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value instanceof Number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		// Preserved operand comments.
		{Code: "getValue(/* keep */) * -1;", FileName: "case.js", Output: []string{"-getValue(/* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
		}},
		{Code: "-1 * values[/* keep */ index];", FileName: "case.js", Output: []string{"-values[/* keep */ index];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
		}},
		{Code: "(left + /* keep */ right) / -1;", FileName: "case.js", Output: []string{"-(left + /* keep */ right);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
		}},
		{Code: "(condition ? /* keep */ left : right) * -1;", FileName: "case.js", Output: []string{"-(condition ? /* keep */ left : right);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
		}},
		{Code: "getValue(\r\n  // keep\r\n) / -1;", FileName: "case.js", Output: []string{"-getValue(\r\n  // keep\r\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 3, EndColumn: 7},
		}},
		// Removed comments and TypeScript.
		{Code: "(/* keep */ value) * -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		{Code: "(value /* keep */) / -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		{Code: "value * -/* keep */1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "-1 /* keep */ * value;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
		}},
		{Code: "(value as /* keep */ number) * -1;", FileName: "case.ts", Output: []string{"-(value as /* keep */ number);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
		}},
		// Snapshot cases.
		{Code: "x * -1;", FileName: "case.js", Output: []string{"-x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "-1 * x;", FileName: "case.js", Output: []string{"-x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "x / -1;", FileName: "case.js", Output: []string{"-x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "a.b * -1;", FileName: "case.js", Output: []string{"-a.b;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
		}},
		{Code: "foo() * -1;", FileName: "case.js", Output: []string{"-foo();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		{Code: "5 * -1;", FileName: "case.js", Output: []string{"-5;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "-1 * 5;", FileName: "case.js", Output: []string{"-5;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "foo?.bar * -1;", FileName: "case.js", Output: []string{"-foo?.bar;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "tag`str` * -1;", FileName: "case.js", Output: []string{"-tag`str`;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "--x * -1;", FileName: "case.js", Output: []string{"-(--x);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
		}},
		{Code: "x++ * -1;", FileName: "case.js", Output: []string{"-(x++);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
		}},
		{Code: "-y * -1;", FileName: "case.js", Output: []string{"-(-y);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		{Code: "+x * -1;", FileName: "case.js", Output: []string{"-+x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		{Code: "(a + b) * -1;", FileName: "case.js", Output: []string{"-(a + b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "-1 * (a + b);", FileName: "case.js", Output: []string{"-(a + b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(a ? b : c) * -1;", FileName: "case.js", Output: []string{"-(a ? b : c);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
		}},
		{Code: "(a, b) * -1;", FileName: "case.js", Output: []string{"-(a, b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
		}},
		{Code: "(a = b) * -1;", FileName: "case.js", Output: []string{"-(a = b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(a + b) / -1;", FileName: "case.js", Output: []string{"-(a + b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(x) * -1;", FileName: "case.js", Output: []string{"-x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
		}},
		{Code: "(x * -1) ** 2;", FileName: "case.js", Output: []string{"(-x) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
		}},
		{Code: "a-x*-1;", FileName: "case.js", Output: []string{"a- -x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 7},
		}},
		{Code: "a - x * -1;", FileName: "case.js", Output: []string{"a - -x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "foo\nx * -1;", FileName: "case.js", Output: []string{"foo\n;-x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
		}},
		{Code: "function f() { return x * -1; }", FileName: "case.js", Output: []string{"function f() { return -x; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
		}},
		{Code: "x * -1 * -1;", FileName: "case.js", Output: []string{"-x * -1;", "-(-x);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "a + b * -1;", FileName: "case.js", Output: []string{"a + -b;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "x /* c */ * -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "x * -1 /* c */;", FileName: "case.js", Output: []string{"-x /* c */;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "(x as number) * -1;", FileName: "case.ts", Output: []string{"-(x as number);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
		}},
		{Code: "-1 * (x as number);", FileName: "case.ts", Output: []string{"-(x as number);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
		}},
		{Code: "x! * -1;", FileName: "case.ts", Output: []string{"-(x!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		// Erased assertions before exponentiation.
		// tsgo rejects unparenthesized assertions before ** (TS1005); retain the upstream expectation.
		{Code: "value * -1 as any ** 2;", FileName: "case.ts", Skip: true, Output: []string{"(-value) as any ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		{Code: "(value * -1 as any) ** 2;", FileName: "case.ts", Output: []string{"((-value) as any) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		// tsgo rejects unparenthesized assertions before ** (TS1005); retain the upstream expectation.
		{Code: "-1 * value satisfies number as any ** 2;", FileName: "case.ts", Skip: true, Output: []string{"(-value) satisfies number as any ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		// tsgo rejects unparenthesized assertions before ** (TS1005); retain the upstream expectation.
		{Code: "value / -1 as any ** 2;", FileName: "case.ts", Skip: true, Output: []string{"(-value) as any ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		{Code: "value * -1 as number;", FileName: "case.ts", Output: []string{"-value as number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		{Code: "value * -1 satisfies number;", FileName: "case.ts", Output: []string{"-value satisfies number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
		}},
		{Code: "(value * -1) as number;", FileName: "case.ts", Output: []string{"(-value) as number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		// tsgo rejects unparenthesized assertions before ** (TS1005); retain the upstream expectation.
		{Code: "value /* keep */ * -1 as any ** 2;", FileName: "case.ts", Skip: true, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
		}},
		// Runtime evaluation cases.
		{Code: "getObject(/* keep */).value / (-1)in object;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value in object;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		{Code: "-1 * (getObject(/* keep */).value)instanceof Number;", FileName: "case.js", Output: []string{"-getObject(/* keep */).value instanceof Number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
		}},
		{Code: "-1 * getValue(/* keep */);", FileName: "case.js", Output: []string{"-getValue(/* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
		}},
		{Code: "getValue(/* keep */) / -1;", FileName: "case.js", Output: []string{"-getValue(/* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
		}},
		// Documentation examples.
		{Code: "const inverted = value * -1;", FileName: "case.js", Output: []string{"const inverted = -value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
		}},
		{Code: "const flipped = -1 * someNumber;", FileName: "case.js", Output: []string{"const flipped = -someNumber;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 32},
		}},
		{Code: "const divided = value / -1;", FileName: "case.js", Output: []string{"const divided = -value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27},
		}},
	})
}
