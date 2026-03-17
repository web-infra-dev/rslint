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
		},
	)
}
