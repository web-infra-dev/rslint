package no_empty_character_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyCharacterClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyCharacterClassRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `var foo = /^abc[a-zA-Z]/;`},
			{Code: `var regExp = new RegExp("^abc[]");`},
			{Code: `var foo = /^abc/;`},
			{Code: `var foo = /[\[]/;`},
			{Code: `var foo = /[\]]/;`},
			{Code: `var foo = /[^]/;`},
			{Code: `var foo = /\[]/`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = /^abc[]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /foo[]bar/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[]]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /\[[]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
		},
	)
}
