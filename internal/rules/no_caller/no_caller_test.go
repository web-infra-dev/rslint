package no_caller

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCallerRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCallerRule,
		[]rule_tester.ValidTestCase{
			// Basic valid cases
			{Code: `var x = arguments.length`},
			{Code: `var x = arguments`},
			{Code: `var x = arguments[0]`},
			{Code: `var x = arguments[caller]`},
			// Computed access with string literals (not PropertyAccessExpression)
			{Code: `var x = arguments["callee"]`},
			{Code: `var x = arguments["caller"]`},
			// callee/caller on non-arguments objects
			{Code: `var obj = { callee: 1 }; var x = obj.callee`},
			{Code: `var obj = { caller: 1 }; var x = obj.caller`},
			// Similar but non-matching property names
			{Code: `var x = arguments.called`},
			{Code: `var x = arguments.call`},
			{Code: `var x = arguments.callees`}, // cspell:ignore callees
			{Code: `var x = arguments.callers`},
			// Comma operator result - not an arguments identifier
			{Code: `var x = (0, arguments).callee`},
			// Ternary result - not an arguments identifier
			{Code: `var x = (true ? arguments : null).callee`},
			// TypeScript non-null assertion wraps identifier (not flagged, matching ESLint+TS parser)
			{Code: `function nonNull() { arguments!.callee; }`},
			// TypeScript as assertion wraps identifier (not flagged)
			{Code: `function typeAssert() { (arguments as any).callee; }`},
			// TypeScript angle bracket assertion wraps identifier (not flagged)
			{Code: `function angleBracket() { (<any>arguments).callee; }`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic cases
			{
				Code: `var x = arguments.callee`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = arguments.caller`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			// Chained property access
			{
				Code: `var x = arguments.callee.bind(this)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = arguments.caller.toString()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			// In function call
			{
				Code: `function foo() { arguments.callee(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			// In function expression
			{
				Code: `var bar = function() { return arguments.callee; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			// In IIFE
			{
				Code: `(function() { arguments.callee; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// Nested function - inner scope
			{
				Code: `function outer() {
  function inner() {
    arguments.callee;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 5},
				},
			},
			// In conditional
			{
				Code: `function cond() { if (true) { arguments.callee; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			// In loop
			{
				Code: `function loop() { for (var i = 0; i < 10; i++) { arguments.caller; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 50},
				},
			},
			// In ternary
			{
				Code: `function tern() { var x = true ? arguments.callee : null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},
			// In logical expression
			{
				Code: `function logic() { var x = arguments.callee || null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			// In assignment
			{
				Code: `function assign() { var x; x = arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 32},
				},
			},
			// In return statement
			{
				Code: `function ret() { return arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			// In throw statement
			{
				Code: `function thrw() { throw arguments.callee; }`, // cspell:ignore thrw
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			// As function argument
			{
				Code: `function arg() { console.log(arguments.callee); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			// In arrow function (arguments from outer scope)
			{
				Code: `function arrowOuter() { var fn = () => arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},
			// In object literal value
			{
				Code: `function objLit() { var o = { fn: arguments.callee }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			// In array literal
			{
				Code: `function arrLit() { var a = [arguments.callee]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			// In template literal expression
			{
				Code: "function tmpl() { var s = `${arguments.callee}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			// In typeof
			{
				Code: `function typ() { var t = typeof arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// In void
			{
				Code: `function vd() { void arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			// In delete
			{
				Code: `function del() { delete arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			// In comma expression
			{
				Code: `function comma() { (0, arguments.callee); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// In switch case
			{
				Code: `function sw() { switch(arguments.callee) { default: break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// In while condition
			{
				Code: `function wh() { while(arguments.callee) { break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Multiple occurrences - callee calling caller
			{
				Code: `function nested() { arguments.callee(arguments.caller); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},
			// In class method
			{
				Code: `class MyClass { method() { arguments.callee; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			// In try-catch
			{
				Code: `function tryCatch() { try { arguments.callee; } catch(e) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
			// Multiple on same line
			{
				Code: `function multi() { arguments.callee; arguments.caller; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},
			// Parenthesized arguments - ESLint sees through parens
			{
				Code: `var x = (arguments).callee`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = ((arguments)).callee`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			// arguments as parameter name - still flagged (syntactic check)
			{
				Code: `function paramName(arguments) { arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// In generator function
			{
				Code: `function* gen() { arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// In async function
			{
				Code: `async function asyncFn() { arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			// In getter
			{
				Code: `var obj = { get x() { return arguments.callee; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			// In setter
			{
				Code: `var obj = { set x(v) { arguments.caller; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// In method shorthand
			{
				Code: `var obj = { m() { arguments.callee; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// Deeply nested control flow
			{
				Code: `function deep() {
  if (true) {
    while (true) {
      for (var i = 0; i < 1; i++) {
        try {
          arguments.callee;
        } catch(e) {}
      }
    }
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 6, Column: 11},
				},
			},
			// Nested object value
			{
				Code: `function nestedObj() { var o = { a: { b: { c: arguments.callee } } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 47},
				},
			},
			// In tagged template
			{
				Code: "function tagged() { String.raw`${arguments.callee}`; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},
			// With new expression
			{
				Code: `function newExpr() { new (arguments.callee)(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},
			// In spread
			{
				Code: `function spread() { var a = [...arguments.callee]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// As computed property key
			{
				Code: `function compKey() { var o = { [arguments.callee]: 1 }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 33},
				},
			},
			// In sequence with assignment
			{
				Code: `function seq() { var x; x = (arguments.callee, 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			// Optional chaining - still PropertyAccessExpression in TS AST
			{
				Code: `function optChain() { arguments?.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			{
				Code: `function optChain2() { arguments?.caller; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			// var shadowing arguments - still flagged (syntactic check)
			{
				Code: `function varShadow() { var arguments = {}; arguments.callee; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 44},
				},
			},
			// arguments in catch clause - still flagged
			{
				Code: `function catchArgs() { try {} catch(arguments) { arguments.callee; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 50},
				},
			},
			// Top-level arguments.callee
			{
				Code: `arguments.callee`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `arguments.caller`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
