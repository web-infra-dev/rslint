package no_obj_calls

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoObjCallsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjCallsRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `var x = Math.random();`},
			{Code: `var x = JSON.parse(foo);`},
			{Code: `Reflect.get(foo, 'x');`},
			{Code: `new Intl.Segmenter();`},
			{Code: `var x = Math;`},
			// Shadowed variable — should not be flagged
			{Code: `function f() { var Math = 1; Math(); }`},
			{Code: `function f(JSON: any) { JSON(); }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `var x = JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
		},
	)
}
