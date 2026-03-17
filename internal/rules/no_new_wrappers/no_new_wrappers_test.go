package no_new_wrappers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewWrappersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewWrappersRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `var a = new Object();`},
			{Code: `var a = String('test'), b = String.fromCharCode(32);`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var a = new String('hello');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = new Number(10);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = new Boolean(false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 9},
				},
			},
		},
	)
}
