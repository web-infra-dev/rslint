package no_loss_of_precision

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLossOfPrecisionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLossOfPrecisionRule,
		// Valid cases - numbers that don't lose precision
		[]rule_tester.ValidTestCase{
			// Basic integers and decimals
			{Code: `var x = 12345`},
			{Code: `var x = 123.456`},
			{Code: `var x = -123.456`},
			{Code: `var x = 0`},
			{Code: `var x = 0.0`},
			{Code: `var x = -0`},

			// Scientific notation
			{Code: `var x = 123e34`},
			{Code: `var x = 123e-34`},
			{Code: `var x = 12.3e34`},
			{Code: `var x = -12.3e34`},

			// Edge cases within safe precision
			{Code: `var x = 9007199254740991`},  // MAX_SAFE_INTEGER
			{Code: `var x = -9007199254740991`}, // MIN_SAFE_INTEGER
			{Code: `var x = 12300000000000000000000000`},
			{Code: `var x = 0.00000000000000000000000123`},

			// With numeric separators (ES2021)
			{Code: `var x = 12_34_56`},
			{Code: `var x = 12_3.4_56`},
			{Code: `var x = 1_230000000_00000000_00000_000`},
			{Code: `var x = 9007_1992547409_91`},

			// Binary format (safe precision)
			{Code: `var x = 0b11111111111111111111111111111111111111111111111111111`},

			// Octal format (safe precision)
			{Code: `var x = 0o377777777777777777`},

			// Hex format (safe precision)
			{Code: `var x = 0x1FFFFFFFFFFFFF`},
			{Code: `var x = 0X1FFFFFFFFFFFFF`},
		},
		// Invalid cases - numbers that lose precision
		[]rule_tester.InvalidTestCase{
			// Integers exceeding safe integer limit
			{
				Code: `var x = 9007199254740993`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = -9007199254740993`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 10},
				},
			},

			// Very large integers
			{
				Code: `var x = 5123000000000000000000000000001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Scientific notation causing precision loss
			{
				Code: `var x = 9007199254740.993e3`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = 9.007199254740993e15`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Decimal precision beyond JavaScript limits
			{
				Code: `var x = 900719.9254740994`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = 1.0000000000000000000000123`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = .1230000000000000000000000`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Exponent causes Infinity
			{
				Code: `var x = 2e999`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Binary exceeds safe integer
			{
				Code: `var x = 0b100000000000000000000000000000000000000000000000000001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Octal exceeds safe integer
			{
				Code: `var x = 0o400000000000000001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// Hex exceeds safe integer
			{
				Code: `var x = 0x20000000000001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = 0X20000000000001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},

			// With numeric separators
			{
				Code: `var x = 900719925474099_3`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = 9.0_0719925_474099_3e15`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = 0X2_000000000_0001`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-loss-of-precision", Line: 1, Column: 9},
				},
			},
		},
	)
}
