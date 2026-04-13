package no_delete_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeleteVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDeleteVarRule,
		// Valid cases — delete on non-Identifier expressions
		[]rule_tester.ValidTestCase{
			// Property access
			{Code: `delete x.prop;`},
			{Code: `delete foo.bar.baz;`},
			// Element access
			{Code: `delete obj["key"];`},
			{Code: `delete obj[0];`},
			// Optional chaining
			{Code: `delete a?.b;`},
			// Computed property
			{Code: "delete obj[`key`];"},
			// Parenthesized member expression (inner is not Identifier)
			{Code: `delete (x.prop);`},
			{Code: `delete ((obj["key"]));`},
			// TypeScript type assertion wrapping identifier (ESLint sees TSAsExpression, not Identifier)
			{Code: `delete (x as any);`},
			// TypeScript non-null assertion wrapping identifier
			{Code: `delete x!;`},
			// TypeScript angle-bracket assertion
			{Code: `delete (<any>x);`},
			// TypeScript satisfies expression
			{Code: `delete (x satisfies any);`},
			// Comma expression — inner result is not a plain Identifier in ESTree
			{Code: `delete (0, x);`},
		},
		// Invalid cases — delete on Identifier (including parenthesized)
		[]rule_tester.InvalidTestCase{
			// Basic
			{
				Code: `delete x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// With var declaration
			{
				Code: `var x; delete x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			// With let declaration
			{
				Code: `let y = 1; delete y;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			// Parenthesized identifier
			{
				Code: `delete (x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Double parenthesized
			{
				Code: `delete ((x));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Nested in if
			{
				Code: `var x; if (true) { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			// Nested in function
			{
				Code: `function f() { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// Nested in arrow function
			{
				Code: `const f = () => { delete x; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},
			// In for loop
			{
				Code: `for (;;) { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			// In try-catch
			{
				Code: `try { delete x; } catch(e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
				},
			},
			// In class method
			{
				Code: `class C { m() { delete x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			// Multiple violations
			{
				Code: `delete x; delete y;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			// Special identifier names
			{
				Code: `delete arguments;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `delete NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Parenthesized + nested context
			{
				Code: `if (true) { delete (x); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			// In switch case
			{
				Code: `switch(0) { case 0: delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// In while
			{
				Code: `while (false) { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			// In async function
			{
				Code: `async function f() { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			// In generator function
			{
				Code: `function* g() { delete x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			// Multi-line
			{
				Code: "delete\nx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
