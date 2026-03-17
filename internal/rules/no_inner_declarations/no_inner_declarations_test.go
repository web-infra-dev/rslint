package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInnerDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// Default mode ("functions")
			{Code: `function doSomething() { }`},
			{Code: `function doSomething() { function somethingElse() { } }`},
			{Code: `(function() { function doSomething() { } }());`},
			{Code: `if (test) { var fn = function() { }; }`},
			{Code: `if (test) var x = 42;`},
			{Code: `var x = 1;`},
			{Code: `var fn = function() { };`},
			{Code: `function foo() { if (test) { var x = 1; } }`},

			// "both" mode
			{Code: `function doSomething() { }`, Options: []interface{}{"both"}},
			{Code: `function doSomething() { var x = 1; }`, Options: []interface{}{"both"}},
			{Code: `var x = 1;`, Options: []interface{}{"both"}},
			{Code: `var fn = function() { };`, Options: []interface{}{"both"}},
			{Code: `function foo() { var x = 1; }`, Options: []interface{}{"both"}},

			// let/const should never be flagged even in "both" mode
			{Code: `if (test) { let x = 1; }`, Options: []interface{}{"both"}},
			{Code: `if (test) { const x = 1; }`, Options: []interface{}{"both"}},
		},
		[]rule_tester.InvalidTestCase{
			// Default mode ("functions") - function declarations in blocks
			{
				Code: `if (foo) function f(){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `function bar() { if (foo) function f(){}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `while (test) { function doSomething() { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// "both" mode - var declarations in blocks
			{
				Code:    `if (foo) { var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `function bar() { if (foo) var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `if (foo) var a;`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
		},
	)
}
