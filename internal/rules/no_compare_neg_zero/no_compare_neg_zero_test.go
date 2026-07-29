package no_compare_neg_zero

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCompareNegZeroRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCompareNegZeroRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Comparisons with positive zero are allowed
			{Code: `x === 0`},
			{Code: `0 === x`},
			{Code: `x == 0`},
			{Code: `0 == x`},

			// String comparisons are allowed
			{Code: `x === '0'`},
			{Code: `'-0' === x`},
			{Code: `x == '-0'`},

			// Comparisons with other negative numbers are allowed
			{Code: `x === -1`},
			{Code: `-1 === x`},

			// Relational operators with positive zero are allowed
			{Code: `x < 0`},
			{Code: `0 <= x`},
			{Code: `x > 0`},
			{Code: `0 >= x`},

			// Inequality operators with positive zero are allowed
			{Code: `x != 0`},
			{Code: `0 !== x`},

			// Object.is() is the correct way to check for -0
			{Code: `Object.is(x, -0)`},

			// Only a direct numeric negative zero operand is compared
			{Code: `x === -0n`},
			{Code: `x === (-0 as number)`},
			{Code: `x === (-0 satisfies number)`},
			{Code: `x === -(-0)`},
			{Code: `x === +(-0)`},
			{Code: `x === (-0, y)`},
			{Code: `x ** -0`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Strict equality
			{
				Code: `x === -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Loose equality
			{
				Code: `x == -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 == x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Greater than
			{
				Code: `x > -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 > x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Greater than or equal
			{
				Code: `x >= -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 >= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Less than
			{
				Code: `x < -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 < x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Less than or equal
			{
				Code: `x <= -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `-0 <= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Inequality operators
			noCompareNegZeroInvalid(`x != -0`, "!="),
			noCompareNegZeroInvalid(`-0 != x`, "!="),
			noCompareNegZeroInvalid(`x !== -0`, "!=="),
			noCompareNegZeroInvalid(`-0 !== x`, "!=="),

			// Parentheses are transparent in ESTree
			noCompareNegZeroInvalid(`x === (-0)`, "==="),
			noCompareNegZeroInvalid(`((-0)) === x`, "==="),
			noCompareNegZeroInvalid(`x === -(0)`, "==="),
			noCompareNegZeroInvalid(`x !== (((-(((0))))))`, "!=="),
			noCompareNegZeroInvalid(`x === (/* before */ - /* after */ (0))`, "==="),

			// All numeric literal spellings whose value is zero are negative zero
			noCompareNegZeroInvalid(`x === -0.0`, "==="),
			noCompareNegZeroInvalid(`x === -0e10`, "==="),
			noCompareNegZeroInvalid(`x === -0x0`, "==="),

			// A comparison with negative zero on both sides reports once
			noCompareNegZeroInvalid(`-0 === -0`, "==="),
			noCompareNegZeroInvalid(`((-0)) !== -(((0)))`, "!=="),

			// Each nested comparison is a separate violation
			{
				Code: `x === -0 === -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// The complete comparison, not only the -0 operand, is reported
			{
				Code: `if (x !== (-0)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
		},
	)
}

func noCompareNegZeroInvalid(code string, operator string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "unexpected",
			Message:   "Do not use the '" + operator + "' operator to compare against -0.",
			Line:      1,
			Column:    1,
		}},
	}
}
