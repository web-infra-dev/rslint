package no_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `const JOE = 'schmoe';`},
			{Code: `let moo = 'car';`},
			{Code: `const JOE = 'schmoe'; let moo = 'car';`},
			{Code: `for (let i = 0; i < 10; i++) {}`},
			{Code: `for (const x of [1,2]) {}`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code: `var foo = bar, toast = most;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
				},
			},
			{
				Code: `var foo = bar; var baz = quux;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 1},
					{MessageId: "unexpectedVar", Line: 1, Column: 16},
				},
			},
			{
				Code: `if (true) { var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 13},
				},
			},

			// for-loop initializer
			{
				Code: `for (var i = 0; i < 10; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 6},
				},
			},

			// for-in
			{
				Code: `for (var x in obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 6},
				},
			},

			// for-of
			{
				Code: `for (var x of [1,2]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 6},
				},
			},
		},
	)
}
