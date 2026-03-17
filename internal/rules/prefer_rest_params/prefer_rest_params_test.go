package prefer_rest_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferRestParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRestParamsRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `arguments;`},
			{Code: `function foo(arguments: any) { arguments; }`},
			{Code: `function foo() { var arguments: any; arguments; }`},
			{Code: `var foo = () => arguments;`},
			{Code: `function foo(...args: any[]) { args; }`},
			{Code: `function foo() { arguments.length; }`},
			{Code: `function foo() { arguments.callee; }`},
			// var arguments in nested block — hoisted to function scope, shadows built-in
			{Code: `function foo() { if (true) { var arguments: any = []; } arguments; }`},
			// let arguments in block — reference INSIDE the block refers to block-scoped variable
			{Code: `function foo() { { let arguments: any = 1; arguments; } }`},
			// const arguments in block — same as above
			{Code: `function foo() { { const arguments: any = 1; arguments; } }`},
			// for-of with let arguments — loop variable shadows inside loop body
			{Code: `function foo() { for (let arguments of []) { arguments; } }`},
			// for-in with let arguments — same as above
			{Code: `function foo() { for (let arguments in {}) { arguments; } }`},
			// catch clause parameter named arguments — shadows inside catch body
			{Code: `function foo() { try {} catch(arguments) { arguments; } }`},
			// function expression named "arguments" — name does NOT shadow implicit arguments
			// (function name is in FunctionNameScope, not function body scope)
			{Code: `var foo = function arguments() { };`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo() { arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			{
				Code: `function foo() { arguments[0]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			{
				Code: `function foo() { arguments[1]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			// Computed member access with Symbol
			{
				Code: `function foo() { arguments[Symbol.iterator]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 18},
				},
			},
			// Storing arguments in a variable
			{
				Code: `function foo() { var x = arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 26},
				},
			},
			// let arguments in a block does NOT shadow the implicit arguments outside the block
			{
				Code: `function foo() { { let arguments: any = 1; } arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 46},
				},
			},
			// Arrow function — arguments refers to enclosing non-arrow function
			{
				Code: `function foo() { var f = () => arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 32},
				},
			},
			// Function expression
			{
				Code: `var foo = function() { arguments; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 24},
				},
			},
			// Class constructor
			{
				Code: `class C { constructor() { arguments; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 27},
				},
			},
			// Destructuring parameter — does NOT shadow implicit arguments
			{
				Code: `function foo({ arguments: a }: any) { arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 39},
				},
			},
			// Function expression named "arguments" — name does NOT shadow body arguments
			{
				Code: `var foo = function arguments() { arguments; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 34},
				},
			},
			// Catch parameter is block-scoped — does NOT shadow arguments outside catch
			{
				Code: `function foo() { try {} catch(arguments) {} arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 45},
				},
			},
			// Shorthand property in object literal — IS a reference to arguments
			{
				Code: `function foo() { var x = { arguments }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 28},
				},
			},
			// Method
			{
				Code: `var obj = { method() { arguments; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 24},
				},
			},
			// Getter
			{
				Code: `var obj = { get x() { arguments; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 23},
				},
			},
			// Setter
			{
				Code: `var obj = { set x(v: any) { arguments; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 29},
				},
			},
			// Nested arrows — arguments passes through multiple arrow boundaries
			{
				Code: `function foo() { var f = () => () => arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 38},
				},
			},
			// typeof arguments
			{
				Code: `function foo() { typeof arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 25},
				},
			},
			// Spread arguments
			{
				Code: `function foo() { bar(...arguments); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 25},
				},
			},
		},
	)
}
