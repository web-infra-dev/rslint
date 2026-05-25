package no_with

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWithRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoWithRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Normal code
			{Code: `foo.bar()`},
			// "with" as property name
			{Code: `obj.with(1)`},
			// "with" as method name in object literal
			{Code: `var obj = { with: function() {} }; obj.with();`},
			// "with" in string literal
			{Code: `var s = "with";`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Basic with statement
			{
				Code: `with(foo) { bar() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// Single-statement body (no block)
			{
				Code: `with(foo) bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// Nested with inside with — two errors
			{
				Code: `with(a) { with(b) { c() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
					{MessageId: "unexpectedWith", Line: 1, Column: 11},
				},
			},
			// with inside function
			{
				Code: `function f() { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 16},
				},
			},
			// with inside arrow function
			{
				Code: `var f = () => { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 17},
				},
			},
			// with inside if block
			{
				Code: `if (true) { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 13},
				},
			},
			// with inside for loop
			{
				Code: `for (;;) { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 12},
				},
			},
			// with inside while loop
			{
				Code: `while (true) { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 16},
				},
			},
			// with inside try/catch
			{
				Code: `try { with(obj) { x; } } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 7},
				},
			},
			// with inside switch case
			{
				Code: `switch(a) { case 1: with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 21},
				},
			},
			// with with member expression
			{
				Code: `with(a.b) { c(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// with with call expression
			{
				Code: `with(a()) { b(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// Multiple sequential with statements — two errors
			{
				Code: "with(a) { x; }\nwith(b) { y; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
					{MessageId: "unexpectedWith", Line: 2, Column: 1},
				},
			},
			// Multi-line with statement
			{
				Code: "with (obj) {\n  foo();\n  bar();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// with with empty body
			{
				Code: `with(obj) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
				},
			},
			// with inside class method
			{
				Code: `class C { method() { with(obj) { x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 22},
				},
			},
			// with inside class constructor
			{
				Code: `class C { constructor() { with(obj) { x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 27},
				},
			},
			// with inside static block
			{
				Code: `class C { static { with(obj) { x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 20},
				},
			},
			// with inside do...while
			{
				Code: `do { with(obj) { x; } } while(true)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 6},
				},
			},
			// with inside for...in
			{
				Code: `for (var k in obj) { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 22},
				},
			},
			// with inside for...of
			{
				Code: `for (var v of arr) { with(v) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 22},
				},
			},
			// with inside else branch
			{
				Code: `if (false) {} else { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 22},
				},
			},
			// with inside finally block
			{
				Code: `try {} finally { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 18},
				},
			},
			// with inside catch block
			{
				Code: `try {} catch(e) { with(obj) { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 19},
				},
			},
			// with inside labeled statement
			{
				Code: `label: with(obj) { x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 8},
				},
			},
			// Multi-byte characters (emoji surrogate pair) to verify UTF-16 code unit counting
			{
				Code: "/* 🚀 */ with(obj) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 10},
				},
			},
			// deeply nested: with inside if inside function inside with
			{
				Code: `with(a) { function f() { if (true) { with(b) { x; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedWith", Line: 1, Column: 1},
					{MessageId: "unexpectedWith", Line: 1, Column: 38},
				},
			},
		},
	)
}
