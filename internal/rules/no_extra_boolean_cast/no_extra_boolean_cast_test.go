package no_extra_boolean_cast

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraBooleanCastRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraBooleanCastRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// !! not in boolean context
			{Code: `var foo = !!bar;`},
			{Code: `function foo() { return !!bar; }`},
			// Boolean() not in boolean context
			{Code: `var foo = Boolean(bar);`},
			// No extra cast
			{Code: `if (foo) {}`},
			// !! in for initializer (not condition)
			{Code: `for(!!foo;;) {}`},
			// new Boolean() outside boolean context is fine
			{Code: `var x = new Boolean(foo);`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// !! in if test
			{
				Code: `if (!!foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 5},
				},
			},
			// !! in while test
			{
				Code: `while (!!foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 8},
				},
			},
			// !! in do-while test
			{
				Code: `do {} while (!!foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 14},
				},
			},
			// !! in for condition
			{
				Code: `for (;!!foo;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 7},
				},
			},
			// Boolean() in if test
			{
				Code: `if (Boolean(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// Boolean() in while test
			{
				Code: `while (Boolean(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 8},
				},
			},
			// Boolean() as operand of !
			{
				Code: `!Boolean(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 2},
				},
			},
			// !! in ternary condition
			{
				Code: `!!foo ? bar : baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 1},
				},
			},
			// !!! - inner !! is operand of outer !
			{
				Code: `!!!foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 2},
				},
			},
			// new Boolean() in boolean context
			{
				Code: `if (new Boolean(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// !! nested inside Boolean() call
			{
				Code: `Boolean(!!foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 9},
				},
			},
			// Boolean() nested inside new Boolean()
			{
				Code: `new Boolean(Boolean(foo))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 13},
				},
			},
		},
	)
}
