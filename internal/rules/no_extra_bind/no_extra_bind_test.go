package no_extra_bind

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraBindRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraBindRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// .bind() with 2+ args is partial application, not flagged
			{Code: `var a = function(b) { return b }.bind(c, d)`},
			// Function uses this, so .bind() is necessary
			{Code: `var a = function() { this.b }.bind(c)`},
			// Function returns this, so .bind() is necessary
			{Code: `var a = function() { return this; }.bind(c)`},
			// Function uses this and returns a value, .bind() is necessary
			{Code: `var a = function() { this.b; return 1; }.bind(c)`},
			// Arrow function inside captures this from the bound function
			{Code: `var a = function() { return () => this; }.bind(b)`},
			// Not a function expression, just a variable
			{Code: `var a = f.bind(a)`},
			// Outer function uses this via inner .bind(this)
			{Code: `(function() { (function() { this.b }.bind(this)) }.bind(c))`},
			// No .bind() at all
			{Code: `var a = function() { return 1; }`},
			// .bind() with spread argument
			{Code: `var a = function() { return 1; }.bind(...args)`},

			// .call() is not .bind() - should be ignored
			{Code: `(function() { this.b; }).call(c)`},
			// .apply() is not .bind() - should be ignored
			{Code: `(function() { return 1; }).apply(c)`},
			// Method call on non-function expression
			{Code: `f.bind(a)`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Function doesn't use this
			{
				Code: `var a = function() { return 1; }.bind(b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// Arrow function with .bind() is always unnecessary
			{
				Code: `var a = (() => { return 1; }).bind(b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// this is in nested function, not in outer
			{
				Code: `var a = function() { (function(){ this.c }) }.bind(b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// .bind(this) is also unnecessary if the function doesn't use this
			{
				Code: `var a = function() { return 1; }.bind(this)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// .bind() with no arguments and no this usage
			{
				Code: `var a = function() { return 1; }.bind()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// Arrow function never needs .bind()
			{
				Code: `var a = (() => { this.b }).bind(c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// Computed property access ['bind'] is also detected
			{
				Code: `var a = function() { return 1; }['bind'](b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Arrow function with expression body and .bind()
			{
				Code: `var a = (() => 1).bind(c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			// this in nested function does not count for outer
			{
				Code: `var a = function() { function inner() { this.b; } }.bind(c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
