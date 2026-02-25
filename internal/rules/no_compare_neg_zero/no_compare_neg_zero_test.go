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
		},
	)
}
