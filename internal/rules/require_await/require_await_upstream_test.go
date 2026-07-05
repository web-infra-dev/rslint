package require_await

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func missingAwaitError(message string, line int, column int, output string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "missingAwait",
		Message:   message,
		Line:      line,
		Column:    column,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: "removeAsync",
				Output:    output,
			},
		},
	}
}

// TestRequireAwaitUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/require-await.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// require_await_extras_test.go file.
func TestRequireAwaitUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		[]rule_tester.ValidTestCase{
			{Code: `async function foo() { await doSomething() }`},
			{Code: `(async function() { await doSomething() })`},
			{Code: `async () => { await doSomething() }`},
			{Code: `async () => await doSomething()`},
			{Code: `({ async foo() { await doSomething() } })`},
			{Code: `class A { async foo() { await doSomething() } }`},
			{Code: `(class { async foo() { await doSomething() } })`},
			{Code: `async function foo() { await (async () => { await doSomething() }) }`},

			// ---- empty functions are ok. ----
			{Code: `async function foo() {}`},
			{Code: `async () => {}`},

			// ---- normal functions are ok. ----
			{Code: `function foo() { doSomething() }`},

			// ---- for-await-of ----
			{Code: `async function foo() { for await (x of xs); }`},

			// ---- global await ----
			{Code: `await foo()`},
			{Code: `
for await (let num of asyncIterable) {
    console.log(num);
}
`},

			{Code: `async function* run() { yield * anotherAsyncGenerator() }`},
			{Code: `
async function* run() {
    await new Promise(resolve => setTimeout(resolve, 100));
    yield 'Hello';
    console.log('World');
}
`},
			{Code: `async function* run() { }`},
			{Code: `const foo = async function *(){}`},
			{Code: `const foo = async function *(){ console.log("bar") }`},
			{Code: `async function* run() { console.log("bar") }`},
			{Code: `await using resource = getResource();`},
			{Code: `async function run() { await using resource = getResource(); }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `async function foo() { doSomething() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'foo' has no 'await' expression.", 1, 1, `function foo() { doSomething() }`),
				},
			},
			{
				Code: `(async function() { doSomething() })`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function has no 'await' expression.", 1, 2, `(function() { doSomething() })`),
				},
			},
			{
				Code: `async () => { doSomething() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 10, `() => { doSomething() }`),
				},
			},
			{
				Code: `async () => doSomething()`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 10, `() => doSomething()`),
				},
			},
			{
				Code: `({ async foo() { doSomething() } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'foo' has no 'await' expression.", 1, 4, `({ foo() { doSomething() } })`),
				},
			},
			{
				Code: `class A { async foo() { doSomething() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'foo' has no 'await' expression.", 1, 11, `class A { foo() { doSomething() } }`),
				},
			},
			{
				Code: `(class { async foo() { doSomething() } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'foo' has no 'await' expression.", 1, 10, `(class { foo() { doSomething() } })`),
				},
			},
			{
				Code: `(class { async ''() { doSomething() } })`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method '' has no 'await' expression.", 1, 10, `(class { ''() { doSomething() } })`),
				},
			},
			{
				Code: `async function foo() { async () => { await doSomething() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'foo' has no 'await' expression.", 1, 1, `function foo() { async () => { await doSomething() } }`),
				},
			},
			{
				Code: `async function foo() { await (async () => { doSomething() }) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 40, `async function foo() { await (() => { doSomething() }) }`),
				},
			},
			{
				Code: `const obj = { async: async function foo() { bar(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'async' has no 'await' expression.", 1, 15, `const obj = { async: function foo() { bar(); } }`),
				},
			},
			{
				Code: `async    /* test */ function foo() { doSomething() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'foo' has no 'await' expression.", 1, 1, `/* test */ function foo() { doSomething() }`),
				},
			},
			{
				Code: `class A {
    a = 0
    async [b](){ return 0; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    a = 0
    ;[b](){ return 0; }
}`),
				},
			},
			{
				Code: `class A {
    a
    async [b](){ return 0; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    a
    [b](){ return 0; }
}`),
				},
			},
			{
				Code: `class A {
    a = 0
    async in(){ return 0; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'in' has no 'await' expression.", 3, 5, `class A {
    a = 0
    ;in(){ return 0; }
}`),
				},
			},
			{
				Code: `const obj = {
    foo,
    async in(){ return 0; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'in' has no 'await' expression.", 3, 5, `const obj = {
    foo,
    in(){ return 0; }
}`),
				},
			},
			{
				Code: `foo
    async () => { return 0; }
`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 2, 14, `foo
    ;() => { return 0; }
`),
				},
			},
			{
				Code: `class A {
    foo() {}
    async [bar] () { baz; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    foo() {}
    [bar] () { baz; }
}`),
				},
			},
			{
				Code: `async function run() { using resource = getResource(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'run' has no 'await' expression.", 1, 1, `function run() { using resource = getResource(); }`),
				},
			},
		},
	)
}
