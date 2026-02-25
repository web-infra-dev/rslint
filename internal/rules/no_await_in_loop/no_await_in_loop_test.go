package no_await_in_loop

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAwaitInLoopRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAwaitInLoopRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Basic async function with await (not in loop)
			{Code: `async function foo() { await bar; }`},

			// For-in/for-of with await in iterable (init position is OK)
			{Code: `async function foo() { for (var bar in await baz) { } }`},
			{Code: `async function foo() { for (var bar of await baz) { } }`},

			// For-await-of loops (await is expected here)
			{Code: `async function foo() { for await (var bar of await baz) { } }`},
			{Code: `async function foo() { for await (var x of xs) { await f(x) } }`},

			// For loop with await in init
			{Code: `async function foo() { for (var i = await bar; i < n; i++) { } }`},

			// Nested function blocks (blocking scope) - await in nested function is OK
			{Code: `async function foo() { while (true) { async function bar() { await baz; } } }`},
			{Code: `async function foo() { while (true) { var bar = async () => { await baz; } } }`},
			{Code: `async function foo() { while (true) { var bar = async function() { await baz; } } }`},
			{Code: `async function foo() { while (true) { class bar { async baz() { await qux; } } } }`},

			// Regular for loops without await in body/condition/update
			{Code: `async function foo() { for (var i = 0; i < 10; i++) { } }`},
			{Code: `async function foo() { for (var i = 0; i < 10; i++) { bar(); } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// While loops with await
			{
				Code: `async function foo() { while (baz) { await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 38},
				},
			},
			{
				Code: `async function foo() { while (await foo()) { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 31},
				},
			},
			{
				Code: `async function foo() { while (baz) { for await (x of xs); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 42},
				},
			},

			// Do-while loops
			{
				Code: `async function foo() { do { await bar; } while (baz); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 29},
				},
			},
			{
				Code: `async function foo() { do { } while (await bar); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 38},
				},
			},

			// For-of/for-in loops with await in body
			{
				Code: `async function foo() { for (var bar of baz) { await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 47},
				},
			},
			{
				Code: `async function foo() { for (var bar in baz) { await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 47},
				},
			},

			// For loops with await in condition/update
			{
				Code: `async function foo() { for (var i = 0; await foo(i); i++) { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 40},
				},
			},
			{
				Code: `async function foo() { for (var i = 0; i < n; i = await bar) { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 51},
				},
			},
			{
				Code: `async function foo() { for (var i = 0; i < n; i++) { await bar; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 54},
				},
			},

			// Deep nesting
			{
				Code: `async function foo() { while (true) { if (bar) { await baz; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 50},
				},
			},
			{
				Code: `async function foo() { while (true) { while (true) { await bar; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Line: 1, Column: 54},
				},
			},
		},
	)
}
