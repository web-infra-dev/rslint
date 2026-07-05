package require_await

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRequireAwaitExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestRequireAwaitExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: async generator container forms ----
			// ---- Real-user: eslint/eslint#12459 async generators can delegate without await ----
			{Code: `class Stream { async *read() { yield* anotherAsyncGenerator(); } }`},
			{Code: `const makeStream = async function* () { yield* anotherAsyncGenerator(); };`},

			// ---- Dimension 4: empty function body forms ----
			// Locks in upstream exitFunction() arm 4: empty async functions are ignored.
			{Code: `class A { async method() {} field = async () => {}; }`},
			{Code: `async function onlyComment() { /* no executable statements */ }`},

			// ---- Dimension 4: await in computed keys belongs to the enclosing function ----
			// Locks in upstream AwaitExpression listener: computed class keys are outside the method body.
			{Code: `async function outer() { class C { [await key()]() {} } }`},
			{Code: `async function outer() { class C { async [await key()]() { await work(); } } }`},
			{Code: `async function outer() { const obj = { [await key()]() {} }; }`},
			{Code: `async function outer() { class C { [await key()] = value; } }`},

			// ---- Dimension 4: for-await and await-using branch lock-ins ----
			// Locks in upstream ForOfStatement arm: only awaited for-of marks the function.
			{Code: `const run = async () => { for await (const item of items) consume(item); };`},
			// Locks in upstream VariableDeclaration arm: await using marks the function.
			{Code: `const run = async () => { await using resource = getResource(); resource.use(); };`},

			// ---- Dimension 4: declaration forms without bodies ----
			// N/A: ESLint core does not parse TS overloads, but rslint should not report body-absent declarations.
			{Code: `declare function load(): Promise<void>;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized async arrow expression ----
			{
				Code: `const f = (async () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 21, `const f = (() => { return 1; });`),
				},
			},

			// ---- Dimension 4: TS type-parameter wrapper on arrow function ----
			{
				Code: `const f = async <T>(value: T) => value;`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 31, `const f = <T>(value: T) => value;`),
				},
			},
			{
				Code: `const f = async <T extends PromiseLike<number>>(value: T) => value;`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 59, `const f = <T extends PromiseLike<number>>(value: T) => value;`),
				},
			},

			// ---- Dimension 4: class field arrow/function expression forms ----
			{
				Code: `class A { field = async () => value; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'field' has no 'await' expression.", 1, 11, `class A { field = () => value; }`),
				},
			},
			{
				Code: `class A { field = async function work() { return value; }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'field' has no 'await' expression.", 1, 11, `class A { field = function work() { return value; }; }`),
				},
			},
			{
				Code: `class A { [key] = async () => value; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 1, 11, `class A { [key] = () => value; }`),
				},
			},

			// ---- Dimension 4: same-kind nesting boundary ----
			// Locks in upstream enterFunction()/exitFunction() stack behavior: inner await does not satisfy outer.
			{
				Code: `async function outer() { async function inner() { await work(); } return inner; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'outer' has no 'await' expression.", 1, 1, `function outer() { async function inner() { await work(); } return inner; }`),
				},
			},
			// Locks in upstream enterFunction()/exitFunction() stack behavior: outer await does not satisfy inner.
			{
				Code: `async function outer() { await setup(); return async () => result; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 57, `async function outer() { await setup(); return () => result; }`),
				},
			},
			{
				Code: `async function outer() { await setup(); function middle() { return async () => value; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 77, `async function outer() { await setup(); function middle() { return () => value; } }`),
				},
			},
			// Locks in upstream enterFunction()/exitFunction() stack behavior: await in an async method does not satisfy the outer function.
			{
				Code: `async function outer() { class C { async method() { await work(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'outer' has no 'await' expression.", 1, 1, `function outer() { class C { async method() { await work(); } } }`),
				},
			},
			// Locks in upstream ForOfStatement arm: an awaited for-of in an inner async function does not satisfy the outer function.
			{
				Code: `async function outer() { async function inner() { for await (const item of items) consume(item); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'outer' has no 'await' expression.", 1, 1, `function outer() { async function inner() { for await (const item of items) consume(item); } }`),
				},
			},
			// Locks in upstream VariableDeclaration arm: await-using in an inner async function does not satisfy the outer function.
			{
				Code: `async function outer() { async function inner() { await using resource = getResource(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'outer' has no 'await' expression.", 1, 1, `function outer() { async function inner() { await using resource = getResource(); } }`),
				},
			},

			// ---- Dimension 4: for-of without await and using without await do not satisfy the rule ----
			// Locks in upstream ForOfStatement arm: non-await for-of is ignored.
			{
				Code: `async function run() { for (const item of items) consume(item); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'run' has no 'await' expression.", 1, 1, `function run() { for (const item of items) consume(item); }`),
				},
			},
			// Locks in upstream VariableDeclaration arm: plain using is ignored.
			{
				Code: `async function run() { using resource = getResource(); resource.use(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'run' has no 'await' expression.", 1, 1, `function run() { using resource = getResource(); resource.use(); }`),
				},
			},

			// ---- Real-user: eslint/eslint#10829 returning another promise is still reported ----
			{
				Code: `async function fetchJSON() { return fetch(url).then(r => r.json()); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'fetchJSON' has no 'await' expression.", 1, 1, `function fetchJSON() { return fetch(url).then(r => r.json()); }`),
				},
			},

			// ---- Real-user: eslint/eslint#10000 throw-only async functions are still reported ----
			{
				Code: `async function fail() { throw new Error("Failure!"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'fail' has no 'await' expression.", 1, 1, `function fail() { throw new Error("Failure!"); }`),
				},
			},
			// ---- Real-user: async APIs often return promises directly; core require-await still reports them ----
			{
				Code: `export default async function loadUser(id) { return api.users.get(id); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async function 'loadUser' has no 'await' expression.", 1, 16, `export default function loadUser(id) { return api.users.get(id); }`),
				},
			},
			{
				Code: `const handler = async /* preserve */ (event) => ({ statusCode: 200, body: event.body });`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 46, `const handler = /* preserve */ (event) => ({ statusCode: 200, body: event.body });`),
				},
			},
			{
				Code: `export default async () => ({ ok: true });`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async arrow function has no 'await' expression.", 1, 25, `export default () => ({ ok: true });`),
				},
			},

			// ---- Dimension 4: computed keys are outside the async method body ----
			// The await in the computed key satisfies the outer function, not the method.
			{
				Code: `async function outer() { class C { async [await key()]() { return 1; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 1, 36, `async function outer() { class C { [await key()]() { return 1; } } }`),
				},
			},
			{
				Code: `const obj = { async [name]() { return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 1, 15, `const obj = { [name]() { return 1; } };`),
				},
			},
			{
				Code: `async function outer() { class C { [await key()] = async () => value; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 1, 36, `async function outer() { class C { [await key()] = () => value; } }`),
				},
			},

			// ---- Branch lock-in: class-body semicolon after object-literal field initializer ----
			// Locks in upstream exitFunction() suggestion arm: class member continuation after `async`.
			{
				Code: `class A {
    field = {}
    async [name]() { return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    field = {}
    ;[name]() { return value; }
}`),
				},
			},
			{
				Code: `class A {
    field = 0
    async /* keep */ [name]() { return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    field = 0
    ;/* keep */ [name]() { return value; }
}`),
				},
			},
			{
				Code: `class A {
    field
    async in(){ return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'in' has no 'await' expression.", 3, 5, `class A {
    field
    in(){ return value; }
}`),
				},
			},
			{
				Code: `class A {
    field: string
    async [name]() { return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    field: string
    ;[name]() { return value; }
}`),
				},
			},
			{
				Code: `class A {
    field = 0
    async instanceof(){ return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method 'instanceof' has no 'await' expression.", 3, 5, `class A {
    field = 0
    ;instanceof(){ return value; }
}`),
				},
			},
			{
				Code: `const C = class {
    field = class {}
    async [name]() { return value; }
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `const C = class {
    field = class {}
    ;[name]() { return value; }
};`),
				},
			},
			{
				Code: `class A {
    field = () => {}
    async [name]() { return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    field = () => {}
    [name]() { return value; }
}`),
				},
			},
			{
				Code: `class A {
    field = value++
    async [name]() { return value; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Async method has no 'await' expression.", 3, 5, `class A {
    field = value++
    [name]() { return value; }
}`),
				},
			},
			{
				Code: `class A { static async #secret() { return value; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					missingAwaitError("Static private async method '#secret' has no 'await' expression.", 1, 11, `class A { static #secret() { return value; } }`),
				},
			},
		},
	)
}
