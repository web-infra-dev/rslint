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
			// let arguments in a block does NOT shadow the implicit arguments
			{
				Code: `function foo() { { let arguments: any = 1; } arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams", Line: 1, Column: 46},
				},
			},
		},
	)
}
