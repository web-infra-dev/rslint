package no_setter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSetterReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSetterReturnRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Object literal setter with assignment (no return)
			{Code: `var foo = { set a(val) { val = 1; } };`},
			// Class setter with assignment (no return)
			{Code: `class A { set a(val) { val = 1; } }`},
			// Object literal setter with bare return (allowed)
			{Code: `var foo = { set a(val) { return; } };`},
			// Getter returning a value (not a setter)
			{Code: `class A { get a() { return 1; } }`},
			// Object literal getter returning a value
			{Code: `var foo = { get a() { return 1; } };`},
			// Setter with conditional bare return
			{Code: `var foo = { set a(val) { if (val) { return; } } };`},
			// Class setter with bare return
			{Code: `class A { set a(val) { return; } }`},
			// Nested function inside setter can return values
			{Code: `var foo = { set a(val) { function inner() { return 1; } } };`},
			// Arrow function inside setter can return values
			{Code: `class A { set a(val) { const fn = () => { return 1; }; } }`},
			// Function expression inside setter can return values
			{Code: `var foo = { set a(val) { var fn = function() { return 1; }; } };`},
			// Regular method returning a value
			{Code: `class A { method() { return 1; } }`},
			// Regular function returning a value
			{Code: `function fn() { return 1; }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Object literal setter returning a number
			{
				Code: `var foo = { set a(val) { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1,
						Column:    26,
					},
				},
			},
			// Class setter returning the parameter
			{
				Code: `class A { set a(val) { return val; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1,
						Column:    24,
					},
				},
			},
			// Object literal setter returning undefined explicitly
			{
				Code: `var foo = { set a(val) { return undefined; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1,
						Column:    26,
					},
				},
			},
			// Class setter with conditional return of a value
			{
				Code: `class A { set a(val) { if (val) { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "returnsValue",
						Line:      1,
						Column:    35,
					},
				},
			},
		},
	)
}
