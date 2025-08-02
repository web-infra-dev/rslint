package no_loss_of_precision

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoLossOfPrecision(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoLossOfPrecisionRule, []rule_tester.ValidTestCase{
		{Code: "const x = 12345;"},
		{Code: "const x = 123.456;"},
		{Code: "const x = -123.456;"},
		{Code: "const x = 123_456;"},
		{Code: "const x = 123_00_000_000_000_000_000_000_000;"},
		{Code: "const x = 123.000_000_000_000_000_000_000_0;"},
		{Code: "const x = 0x1234;"},
		{Code: "const x = 0b1010;"},
		{Code: "const x = 0o777;"},
		{Code: "const x = 9007199254740991;"},  // MAX_SAFE_INTEGER
		{Code: "const x = -9007199254740991;"}, // -MAX_SAFE_INTEGER
		{Code: "const x = 900719925474099.1;"},
		{Code: "const x = 9.007199254740991e15;"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "const x = 9007199254740993;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = 9_007_199_254_740_993;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = 9_007_199_254_740.993e3;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = -9007199254740993;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = 0x20000000000001;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
		{
			Code: "const x = 0o400000000000000001;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noLossOfPrecision",
				},
			},
		},
	})
}