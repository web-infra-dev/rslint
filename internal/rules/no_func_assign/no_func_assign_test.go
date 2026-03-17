package no_func_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFuncAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoFuncAssignRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Variable shadows function name inside function body
			{Code: `function foo() { var foo = bar; }`},

			// Parameter shadows function name
			{Code: `function foo(foo) { foo = bar; }`},

			// Variable declaration shadows function name inside function body
			{Code: `function foo() { var foo; foo = bar; }`},

			// Variable assignment (not a function declaration)
			{Code: `var foo = function() {}; foo = bar;`},

			// Function used without reassignment
			{Code: `function foo() {} foo();`},

			// Function used as argument
			{Code: `function foo() {} bar(foo);`},

			// Arrow function assigned to variable
			{Code: `var foo = () => {}; foo = bar;`},

			// Let assignment (not a function declaration)
			{Code: `let foo = function() {}; foo = bar;`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Direct reassignment after function declaration
			{
				Code: `function foo() {}; foo = bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
				},
			},

			// Reassignment inside the function body
			{
				Code: `function foo() { foo = bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 18},
				},
			},

			// Reassignment before function declaration (hoisting)
			{
				Code: `foo = bar; function foo() { };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 1},
				},
			},

			// Array destructuring assignment
			{
				Code: `function foo() {}; [foo] = bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 21},
				},
			},

			// Increment operator
			{
				Code: `function foo() {}; foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
				},
			},

			// Compound assignment
			{
				Code: `function foo() {}; foo += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
				},
			},

			// Multiple reassignments
			{
				Code: `function foo() {}; foo = bar; foo = baz;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
					{MessageId: "isAFunction", Line: 1, Column: 31},
				},
			},

			// Named function expression reassignment inside body
			{
				Code: `var a = function foo() { foo = 123; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 26},
				},
			},
		},
	)
}
