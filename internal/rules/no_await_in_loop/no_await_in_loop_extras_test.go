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
					{MessageId: "unexpectedAwait", Line: 1, Column: 38, EndLine: 1, EndColumn: 58},
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

// The controls and regressions below were independently checked against ESLint
// v10.8.1, including complete diagnostic ranges. JS keeps the upstream suite.
func TestNoAwaitInLoopBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAwaitInLoopRule,
		[]rule_tester.ValidTestCase{
			// Loop setup inside outer loops
			{Code: "async function outer() { for (var x = await f() in xs) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for (var x = await f() in xs) {} } }"},
			{Code: "async function outer() { for (var x in await xs) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for (var x in await xs) {} } }"},
			{Code: "async function outer() { for (var x of await xs) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for (var x of await xs) {} } }"},
			{Code: "async function outer() { for (var x = await f();;) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for (var x = await f();;) {} } }"},
			{Code: "async function outer() { for (await using x = f();;) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for (await using x = f();;) {} } }"},
			{Code: "async function outer() { for await (const x of xs) {} }"},
			{Code: "async function outer() { for await (const y of ys) { for await (const x of xs) {} } }"},
			// Resource disposal in each loop family
			{Code: "async function outer() { while (true) { async function inner() { await using x = f(); } } }"},
			{Code: "async function outer() { for (;;) { async function inner() { await using x = f(); } } }"},
			{Code: "async function outer() { for (const y of ys) { async function inner() { await using x = f(); } } }"},
			{Code: "async function outer() { do { async function inner() { await using x = f(); } } while (true); }"},
			// Async iteration boundaries and computed names
			{Code: "async function outer() { for await (await using x of xs) {} }"},
			{Code: "async function outer() { while (true) { class C { async method() { await f(); } } } }"},
			{Code: "async function outer() { while (true) { const obj = { async method() { await f(); } }; } }"},
		},
		[]rule_tester.InvalidTestCase{
			// Loop setup inside outer loops
			{
				Code: "async function outer() { while (true) { for (var x = await f() in xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 54, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: "async function outer() { while (true) { for (var x in await xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 55, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: "async function outer() { while (true) { for (var x of await xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 55, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: "async function outer() { while (true) { for (var x = await f();;) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 54, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: "async function outer() { while (true) { for (await using x = f();;) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 46, EndLine: 1, EndColumn: 65},
				},
			},
			{
				Code: "async function outer() { while (true) { for await (const x of xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 41, EndLine: 1, EndColumn: 69},
				},
			},
			// Resource disposal in each loop family
			{
				Code: "async function outer() { while (true) { await using x = f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 41, EndLine: 1, EndColumn: 61},
				},
			},
			{
				Code: "async function outer() { while (true) { await using x = await f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 41, EndLine: 1, EndColumn: 67},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 57, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code: "async function outer() { while (true) { await using x = f(), y = g(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 41, EndLine: 1, EndColumn: 70},
				},
			},
			{
				Code: "async function outer() { for (;;) { await using x = f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 37, EndLine: 1, EndColumn: 57},
				},
			},
			{
				Code: "async function outer() { for (;;) { await using x = await f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 37, EndLine: 1, EndColumn: 63},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 53, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code: "async function outer() { for (;;) { await using x = f(), y = g(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 37, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code: "async function outer() { for (const y of ys) { await using x = f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 48, EndLine: 1, EndColumn: 68},
				},
			},
			{
				Code: "async function outer() { for (const y of ys) { await using x = await f(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 48, EndLine: 1, EndColumn: 74},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 64, EndLine: 1, EndColumn: 73},
				},
			},
			{
				Code: "async function outer() { for (const y of ys) { await using x = f(), y = g(); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 48, EndLine: 1, EndColumn: 77},
				},
			},
			{
				Code: "async function outer() { do { await using x = f(); } while (true); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 31, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code: "async function outer() { do { await using x = await f(); } while (true); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 31, EndLine: 1, EndColumn: 57},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 47, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code: "async function outer() { do { await using x = f(), y = g(); } while (true); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 31, EndLine: 1, EndColumn: 60},
				},
			},
			// Async iteration boundaries and computed names
			{
				Code: "async function outer() { for (await using x of xs) {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 31, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code: "async function outer() { while (true) { for await (await using x of xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 41, EndLine: 1, EndColumn: 75},
				},
			},
			{
				Code: "async function outer() { for await (const x of xs) { for (await using y of ys) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 59, EndLine: 1, EndColumn: 72},
				},
			},
			{
				Code: "async function outer() { while (true) { class C { [await f()]() {} } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 52, EndLine: 1, EndColumn: 61},
				},
			},
			{
				Code: "async function outer() { while (true) { const obj = { [await f()]() {} }; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 56, EndLine: 1, EndColumn: 65},
				},
			},
			{
				Code: "async function outer() { while (true) { class C { get [await f()]() { return 1; } } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 56, EndLine: 1, EndColumn: 65},
				},
			},
			{
				Code: "async function outer() { while (true) { class C { set [await f()](x) {} } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 56, EndLine: 1, EndColumn: 65},
				},
			},
			{
				Code: "async function outer() { while (true) { /* 😀 */\r\n await using x = f();\r\n await (f()); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 2, Column: 2, EndLine: 2, EndColumn: 22},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 3, Column: 2, EndLine: 3, EndColumn: 13},
				},
			},
			{
				Code: "async function outer() { while (true) { for (/* 😀 */ await using x of xs) { await f(); } } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 55, EndLine: 1, EndColumn: 68},
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 78, EndLine: 1, EndColumn: 87},
				},
			},
			{
				Code: "async function outer() { while (true) { for (obj[await f()] of xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 50, EndLine: 1, EndColumn: 59},
				},
			},
			{
				Code: "async function outer() { while (true) { for (obj[await f()] in xs) {} } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAwait", Message: "Unexpected `await` inside a loop.", Line: 1, Column: 50, EndLine: 1, EndColumn: 59},
				},
			},
		},
	)
}
