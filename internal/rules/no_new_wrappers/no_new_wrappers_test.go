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
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `var a = new Object();`},
			{Code: `var a = String('test'), b = String.fromCharCode(32);`},
			// Locally shadowed — should not report
			{Code: `function test(Number: any) { return new Number; }`},
			{Code: `function test() { var Boolean: any = function(){}; return new Boolean(true); }`},
			{Code: `function test() { let String: any = class {}; return new String('x'); }`},
			// Non-wrapper constructors
			{Code: `var a = new Map();`},
			{Code: `var a = new Date();`},
		},
		// Invalid cases
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
			// Multiple in one statement
			{
				Code: `var a = new String('a'); var b = new Number(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noConstructor", Line: 1, Column: 9},
					{MessageId: "noConstructor", Line: 1, Column: 34},
				},
			},
		},
	)
}
