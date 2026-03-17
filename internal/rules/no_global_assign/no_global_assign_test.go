package no_global_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoGlobalAssignRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Lowercase identifier is not a builtin
			{Code: `string = 'hello world';`},

			// Shadowed by var declaration
			{Code: `var String; String = 'test';`},

			// Shadowed by let declaration
			{Code: `let Array; Array = 1;`},

			// Shadowed by function parameter
			{Code: `function foo(String) { String = 'test'; }`},

			// Shadowed by function declaration
			{Code: `function Object() {} Object = 'test';`},

			// Exception option
			{Code: `Object = 0;`, Options: map[string]interface{}{"exceptions": []interface{}{"Object"}}},

			// Read-only usage (not a write reference)
			{Code: `var x = String(123);`},

			// Property access (not an identifier assignment)
			{Code: `var x = Math.PI;`},

			// Not a builtin name
			{Code: `foo = 'bar';`},

			// Shadowed by class declaration
			{Code: `class Array {} Array = 1;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Direct assignment to builtin
			{
				Code: `String = 'hello world';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Postfix increment on builtin
			{
				Code: `String++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Array
			{
				Code: `Array = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Number
			{
				Code: `Number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Compound assignment
			{
				Code: `Math += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Prefix decrement
			{
				Code: `--Object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 3},
				},
			},

			// Assignment to undefined
			{
				Code: `undefined = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to NaN
			{
				Code: `NaN = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Assignment to Infinity
			{
				Code: `Infinity = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
				},
			},

			// Multiple globals assigned
			{
				Code: `String = 1; Array = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 1},
					{MessageId: "globalShouldNotBeModified", Line: 1, Column: 13},
				},
			},
		},
	)
}
