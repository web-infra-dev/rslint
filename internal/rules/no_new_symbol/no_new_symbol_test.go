package no_new_symbol

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewSymbolRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewSymbolRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `var foo = Symbol('foo');`},
			{Code: `new foo(Symbol);`},
			{Code: `new foo(bar, Symbol);`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = new Symbol('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewSymbol",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: `new Symbol()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewSymbol",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
