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

			// Reassigning var inside function expression body (var, not function decl)
			{Code: `var foo = function() { foo = 1; };`},

			// Local var shadows function name
			{Code: `function foo() { var foo = 1; }`},

			// Import with shadowed variable
			{Code: `import bar from 'bar'; function foo() { var foo = bar; }`},

			// let shadows function name
			{Code: `function foo() { let foo = 1; foo = 2; }`},

			// const shadows function name
			{Code: `function foo() { const foo = 1; }`},

			// Inner function declaration shadows, no reassignment of outer
			{Code: `function foo() { function foo() {} }`},

			// Catch clause parameter shadows function name
			{Code: `function foo() {} try {} catch(foo) { foo = 1; }`},
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

			// Reassign before declaration with numeric value (hoisting)
			{
				Code: `foo = 1; function foo() { };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 1},
				},
			},

			// Increment operator (no semicolon after declaration)
			{
				Code: `function foo() {} foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 19},
				},
			},

			// Array destructuring with array literal
			{
				Code: `function foo() {} [foo] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
				},
			},

			// Array destructuring before function declaration (hoisting)
			{
				Code: `[foo] = bar; function foo() { };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 2},
				},
			},

			// Object destructuring with default value
			{
				Code: `({x: foo = 0} = bar); function foo() { };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 6},
				},
			},

			// Array destructuring inside function body
			{
				Code: `function foo() { [foo] = bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 19},
				},
			},

			// Object destructuring with default in nested IIFE
			{
				Code: `(function() { ({x: foo = 0} = bar); function foo() { }; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 20},
				},
			},

			// Shorthand object destructuring
			{
				Code: `function foo() {} ({foo} = bar);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 21},
				},
			},

			// for-in assignment
			{
				Code: `function foo() {} for (foo in bar) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 24},
				},
			},

			// for-of assignment
			{
				Code: `function foo() {} for (foo of bar) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 24},
				},
			},

			// Spread in object destructuring
			{
				Code: `function foo() {} ({...foo} = bar);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 24},
				},
			},

			// Spread in array destructuring
			{
				Code: `function foo() {} [...foo] = bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 23},
				},
			},

			// Prefix decrement
			{
				Code: `function foo() {} --foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 21},
				},
			},

			// Postfix decrement
			{
				Code: `function foo() {} foo--;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 19},
				},
			},

			// Inner function declaration shadows, but reassignment targets inner
			{
				Code: `function foo() { function foo() {} foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 36},
				},
			},

			// try block reassignment (not shadowed by catch parameter)
			{
				Code: `function foo() {} try { foo = 1; } catch(foo) { foo = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "isAFunction", Line: 1, Column: 25},
				},
			},
		},
	)
}
