package valid_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_expect.ValidExpectRule,
		[]rule_tester.ValidTestCase{
			// Basic well-formed assertions.
			{Code: `expect(value).toBe(1);`},

			// Reading a property of a completed assertion is not part of the
			// chain, so the assertion stays well-formed.
			{Code: `expect(value).toBe(1).message;`},
			{Code: `const r = expect(value).toHaveBeenCalled().message;`},
			{Code: `expect(value).toBe(1)[key];`},
			{Code: `expect(value).toHaveReturned()[key]();`},
			{Code: `test("t", async () => { (await expect(value).resolves.toBe(1)).message; });`},
			{Code: `expect(value).not.toBe(1);`},
			{Code: `expect("something").toEqual("else");`},

			// Static members carry no assertion factory: no arg/await checks.
			{Code: `expect.hasAssertions();`},
			{Code: `expect.assertions(1);`},
			{Code: `expect.any(Number);`},
			{Code: `expect.stringContaining("x");`},
			{Code: `expect.resolves.toBe(1);`},
			{Code: `test("t", () => { expect.toResolve(); });`},

			// vitest maxArgs allowance: message second argument and poll/element
			// options are not excess arguments.
			{Code: `expect(value, "message").toBe(1);`},
			{Code: "expect(value, `msg ${x}`).toBe(1);"},
			{Code: `test("t", async () => { await expect.poll(() => value, { timeout: 1 }).toBe(1); });`},
			{Code: `test("t", async () => { await expect.element(locator, { timeout: 1 }).toBeVisible(); });`},

			// Chai property matchers are valid without a call.
			{Code: `expect(value).to.be.ok;`},
			{Code: `expect(value).to.be.a("string");`},
			{Code: `expect(spy).to.have.been.called;`},

			// Async assertions awaited or returned.
			{Code: `test("t", async () => { await expect(p).resolves.toBe(1); });`},
			{Code: `test("t", () => { return expect(p).rejects.toThrow(); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.true; });`},
			{Code: `test("t", () => { return expect(p).resolves.to.be.true; });`},
			{Code: `test("t", () => expect(p).resolves.to.be.true);`},
			// A Chai chain may carry several matchers; the await/return check
			// looks at the outermost expression, not the first matcher.
			// Parentheses are nodes in the TypeScript AST but not in ESTree, so
			// the await/return check must look past them to match upstream.
			{Code: `test("t", async () => { await (expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", () => { return (expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", () => { return (condition ? expect(p).resolves.toBe(1) : expect(q).toBe(1)); });`},
			{Code: `test("t", async () => { await ((expect(p).resolves.to.be.true)); });`},
			{Code: `test("t", async () => { await Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`},
			{Code: `test("t", async () => { await (expect(p).resolves.toBe(1)).then(() => {}); });`},
			{Code: `test("t", async () => { await import.meta.rstest.expect(p).resolves.toBe(1).then(() => {}); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.an("object").that.is.ok; });`},
			{Code: `test("t", () => { return expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			{Code: `expect(value).not.toBe(1);`},
			{Code: `test("t", async () => { await expect.poll(() => value).toBe(1); });`},
			// poll/element are not treated as async by valid-expect (matching
			// vitest: only resolves/rejects modifiers and asyncMatchers require
			// await), so a poll chain without await is not reported here.
			{Code: `test("t", () => { expect.poll(() => value).toBe(1); });`},
			{Code: `expect(1, 2).toBe(1);`, Options: []any{map[string]any{"maxArgs": 2}}},
			{
				Code:    `test("t", async () => { await expect(p).toResolveEventually(); });`,
				Options: []any{map[string]any{"asyncMatchers": []any{"toResolveEventually"}}},
			},

			// Provenance: non-Rstest expect and local shadow are ignored.
			{Code: `import { expect } from "vitest"; expect(1);`},
			{Code: `const expect = () => {}; expect(1);`},
			{Code: `import r = require('@rstest/core'); const expect = 'test'; r[expect]('works', () => {});`},
			{Code: `import * as r from '@rstest/core'; const expect = 'test'; r[expect]('works', () => {});`},
			// TypeScript assertion wrappers before the matcher are erased at
			// runtime and must not trigger a false `matcherNotFound` warning.
			{Code: `import.meta.rstest!.expect(1)!.toBe(1);`},
			{Code: `expect(1)!.toBe(1);`},
			{Code: `(expect(1) as any).toBe(1);`},
			{Code: `(expect(1) satisfies unknown).toBe(1);`},
			// Object properties mutated through a helper call are not static constants.
			{Code: `import * as r from '@rstest/core'; const config = { name: 'expect' }; function selectSuite(x: any) { x.name = 'test'; } selectSuite(config); const key = config.name; r[key]('registers a test', () => {});`},
			// Outer TypeScript assertion wrappers preserve the returned/awaited Promise.
			{Code: `import * as r from '@rstest/core'; const key = 'expect'; r.test('value matches', () => { return r[key](Promise.resolve(1)).resolves.toBe(1) as Promise<void>; });`},
			{Code: `import * as r from '@rstest/core'; const key = 'expect'; r.test('value matches', async () => { await (r[key](Promise.resolve(1)).resolves.toBe(1) as Promise<void>); });`},
			{Code: `test('t', () => { return expect(p).resolves.toBe(1) satisfies Promise<void>; });`},
			{Code: `test('t', () => { return <Promise<void>>expect(p).resolves.toBe(1); });`},
			{Code: `test('t', async () => { await expect(p).resolves.toBe(1)!; });`},
			{Code: `test('t', () => expect(p).resolves.toBe(1) as Promise<void>);`},
			{Code: `test('t', () => { return Promise.all([expect(p).resolves.toBe(1) as Promise<void>]); });`},
			{Code: `test('t', async () => { const resolve = 'all'; await Promise[resolve]([expect(p).resolves.toBe(true), startWork()]); });`},
			{Code: `test('t', async () => { await Promise.try((check: any) => check.then(() => {}), expect(p).resolves.toBe(1)); });`},
			{Code: `test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`},
			{Code: `test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]).finally(() => {}); });`},
			{Code: `test('t', async () => { await expect(p).resolves.toBe(1).finally(() => {}); });`},
			{Code: `test('t', async () => { await Promise.all([Promise.resolve(expect(p).resolves.toBe(1))]); });`},
			{Code: `test('t', async () => { await (Promise as typeof Promise).all([expect(p).resolves.toBe(1)]); });`},
			{Code: `test('t', async () => { await globalThis.Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `test('t', async () => { await Promise.all([cond ? expect(p).resolves.toBe(1) : Promise.resolve()]); });`},
			{Code: `test('t', async () => { await (log(), expect(p).resolves.toBe(1)); });`},
			{Code: `test('t', async () => { await expect(p)['resolves' as const].toBe(1); });`},
			{Code: `test('t', async () => { await (cond ? expect(p).resolves.toBe(1) : expect(q).resolves.toBe(2)).then(() => {}); });`},
			{Code: `test('t', async () => { await (log(), expect(p).resolves.toBe(1)).then(() => {}); });`},
			{Code: `test('t', async () => { await Promise.all([(log(), expect(p).resolves.toBe(1))]); });`},
			{Code: `test('t', async () => { await Promise.resolve((log(), expect(p).resolves.toBe(1))); });`},
			{Code: `expect(1)['to' + 'Be'](1);`},
			{Code: `expect(1)[getMatcher()](1);`},
			{Code: `expect(1).not['to' + 'Be'](1);`},
			{Code: `test('t', async () => { await expect(p).resolves['to' + 'Be'](1); });`},
			{Code: `expect(1)[0]();`},
			{Code: `import * as r from "@rstest/core"; const expect = "test"; r[expect as any]("works", () => {});`},
			{Code: `const [expect] = require("@rstest/core"); expect(1);`},
		},
		[]rule_tester.InvalidTestCase{
			// --- arg count ---
			{
				Code: `expect().toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs", Column: 7, EndColumn: 8},
				},
			},
			{
				Code: `expect(1, 2, 3).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs", Column: 11, EndColumn: 14},
				},
			},
			{
				Code: `expect(value, 123).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs"},
				},
			},
			// jest would report tooManyArgs here; rstest allows the message arg,
			// so this appears ONLY in the valid list above. A three-arg call still
			// reports because the allowance is a single trailing message.

			// --- broken chains (reason from the expect parser) ---
			{
				Code: `expect(value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect(value).toBe;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `expect(value).resolves;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect(value).foo.toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},

			// --- async must be awaited ---
			{
				Code: `test("t", async () => { expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.toBe(1); });`},
			},
			{
				Code: `test("t", async () => { expect(p).rejects.toThrow(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).rejects.toThrow(); });`},
			},
			{
				Code: `test("t", async () => { expect(p).resolves.to.be.true; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.to.be.true; });`},
			},
			{
				// A parenthesized assertion that nothing awaits still reports,
				// and the fix goes inside the parentheses.
				Code: `test("t", async () => { (expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { (await expect(p).resolves.toBe(1)); });`},
			},
			{
				Code:    `test("t", () => { return (expect(p).resolves.toBe(1)); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe(1)); });`},
			},
			{
				// A multi-matcher chain still reports when it is not awaited,
				// and the fix covers the whole chain.
				Code: `test("t", async () => { expect(p).resolves.to.be.a("string").that.contains("x"); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			},
			{
				Code: `test("t", async () => { const all = Promise.all([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { const all = await Promise.all([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]); });`},
			},
			{
				Code: `test("t", async () => { Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`},
			},
			{
				Code: `test("t", async () => { Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)])); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)])); });`},
			},
			{
				Code:    `test("t", () => { return expect(p).resolves.toBe(1); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.toBe(1); });`},
			},
			{
				Code:    `expect(value).toBe(1);`,
				Options: []any{map[string]any{"minArgs": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs"},
				},
			},
			{
				Code:    `test("t", () => { expect(p).toResolveEventually(); });`,
				Options: []any{map[string]any{"asyncMatchers": []any{"toResolveEventually"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).toResolveEventually(); });`},
			},

			// --- provenance (resolved by the expect parser) ---
			{
				Code: `import { expect as check } from "@rstest/core"; check(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `test("t", ({ expect }) => { expect(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `test("t", () => { import.meta.rstest.expect(p).resolves.toBe(1).then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 1, Column: 19, EndColumn: 79},
				},
				Output: []string{`test("t", async () => { await import.meta.rstest.expect(p).resolves.toBe(1).then(() => {}); });`},
			},
			{
				Code: `(expect(1) as any);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `(expect(1) as any).toBe;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `test("t", () => { (expect(p) as any).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p) as any).resolves.toBe(1); });`},
			},
			{
				Code: `test("t", () => { expect(p).resolves.toBe(1) as Promise<void>; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe(1) as Promise<void>); });`},
			},
			{
				Code:    `test("t", () => { return expect(p).resolves.toBe(1) as Promise<void>; });`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe(1) as Promise<void>); });`},
			},
			{
				Code:    `test("t", () => { return (expect(p).resolves.toBe(1) as Promise<void>); });`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe(1) as Promise<void>); });`},
			},
			// Early return with alwaysAwait must keep `return await` so control
			// flow does not fall through to subsequent statements.
			{
				Code: `test.each([true])('early return', (ready) => {
  const p = Promise.resolve(1);
  if (ready) return expect(p).resolves.toBe(1) as Promise<void>;
  throw new Error('fell through');
});`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test.each([true])('early return', async (ready) => {
  const p = Promise.resolve(1);
  if (ready) return await (expect(p).resolves.toBe(1) as Promise<void>);
  throw new Error('fell through');
});`},
			},
			// `Promise.resolve([...])` resolves to the array without awaiting its elements.
			{
				Code: `test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(1)] as const); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.resolve([await expect(p).resolves.toBe(1)] as const); });`},
			},
			// A captured `Promise.resolve([...])` array may be awaited downstream;
			// report the diagnostic without inserting an unsafe `await` inside the array.
			{
				Code: `test("t", async () => { const [check] = await Promise.resolve([expect(p).resolves.toBe(1)] as const); await check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// A conditional test position discards the assertion promise rather than
			// forwarding it; report without autofixing since awaiting inverts truthiness.
			{
				Code: `test("t", () => { return (expect(p).resolves.toBe(1) as unknown) ? Promise.resolve() : Promise.reject(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { return (expect(p).resolves.toBe(1).then(() => false) as unknown) ? onReady() : onError(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Node's `createRequire` and ambient `declare const require` still load `@rstest/core`.
			{
				Code: `import { createRequire } from 'node:module'; const require = createRequire(import.meta.url); const { expect: check } = require('@rstest/core'); check(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `declare const require: any; const { expect: check } = require('@rstest/core'); check(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			// Non-fixable function-like scopes report without producing broken syntax.
			{
				Code: `class C { constructor() { expect(p).resolves.toBe(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `class C { get x() { expect(p).resolves.toBe(1); return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `class C { set x(v: any) { expect(p).resolves.toBe(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function* gen() { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function check(p: Promise<number>): void { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function f(x = expect(p).resolves.toBe(1)) { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `class C { static { expect(p).resolves.toBe(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `class C { x = expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Safe `async` insertion positions for exported functions, static methods,
			// async generators, and identifier-adjacent arrow functions.
			{
				Code: `export function check(p: Promise<number>) { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`export async function check(p: Promise<number>) { await expect(p).resolves.toBe(1); }`},
			},
			{
				Code: `class C { static check(p: Promise<number>) { expect(p).resolves.toBe(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`class C { static async check(p: Promise<number>) { await expect(p).resolves.toBe(1); } }`},
			},
			{
				Code: `async function* gen() { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`async function* gen() { await expect(p).resolves.toBe(1); }`},
			},
			{
				Code: `function make() { return() => { expect(p).resolves.toBe(1); }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`function make() { return async () => { await expect(p).resolves.toBe(1); }; }`},
			},
			{
				Code:    `test("t", (ok) => { if (ok) return(expect(p).resolves.toBe(1)); throw new Error(); });`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async (ok) => { if (ok) return await (expect(p).resolves.toBe(1)); throw new Error(); });`},
			},
			{
				Code:    `function checkAndTag(p: Promise<number>): Promise<void> { return expect(p).resolves.toBe(1); }`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`async function checkAndTag(p: Promise<number>): Promise<void> { return await expect(p).resolves.toBe(1); }`},
			},
			// Unsafe await positions report without autofixing.
			{
				Code: `test("t", () => { if (true && expect(p).resolves.toBe(1)) {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { switch (expect(p).resolves.toBe(1)) {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { 'then' in expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.reject([expect(p).resolves.toBe(1)]); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Promise wrapper chaining, shadowing, and non-all methods.
			{
				Code: `test("t", () => { Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`},
			},
			{
				Code: `test("t", () => { Promise.all([expect(p).resolves.toBe(1)]).finally(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).finally(() => {}); });`},
			},
			{
				Code: `test("t", async () => { await new Promise((r) => r(1)).then(expect(p).resolves.toBe(1) as any); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await new Promise((r) => r(1)).then(await (expect(p).resolves.toBe(1) as any)); });`},
			},
			{
				Code: `test("t", async () => { const Promise = { all: (x: any) => x }; await Promise.all([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { const Promise = { all: (x: any) => x }; await Promise.all([await expect(p).resolves.toBe(1)]); });`},
			},
			{
				Code: `test("t", async () => { await (Promise as any).withResolvers([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (Promise as any).withResolvers([await expect(p).resolves.toBe(1)]); });`},
			},
			{
				Code: `test("t", async () => { await (Promise as any).resolve(1, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (Promise as any).resolve(1, await expect(p).resolves.toBe(1)); });`},
			},
			// Trailing index access, non-Promise property reads, and cast element keys.
			{
				Code: `expect(1)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `test("t", async () => { expect(p).resolves.toBe(1)[0](); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await (expect(p).resolves.toBe(1) as any).message; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1).toString(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { expect(p)['resolves' as const].toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p)['resolves' as const].toBe(1); });`},
			},
			// Non-fixable return types (`: any` / `: unknown`), `await` identifier, Annex B, `new (function(){})()`.
			{
				Code: `function check(p: Promise<number>): any { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `const check = (p: Promise<number>): unknown => { expect(p).resolves.toBe(1); return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", (await: number) => { expect(Promise.resolve(await)).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const await = 1; expect(Promise.resolve(await)).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `if (cond) function check() { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `label: function check() { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `new (function () { expect(p).resolves.toBe(1); })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Unsafe await positions: extends, decorators, wrapped member/call/in/instanceof/typeof, logical assignment, destructuring, equality, wrapped Promise.resolve array.
			{
				Code: `test("t", () => { class C extends expect(p).resolves.toBe(1) {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { @expect(p).resolves.toBe(1) class C {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { (cond ? expect(p).resolves.toBe(1) : q).then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (cond ? expect(p).resolves.toBe(1) : q).then(() => {}); });`},
			},
			{
				Code: `test("t", () => { (ready && expect(p).resolves.toBe(1)).then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { "then" in (cond ? expect(p).resolves.toBe(1) : q); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { let x; if (x ??= expect(p).resolves.toBe(1)) { run(); } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const { then } = expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { ({ then } = expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const [check] = await Promise.resolve([cond ? expect(p).resolves.toBe(1) : q]); await check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const [check] = [expect(p).resolves.toBe(1)]; await check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { if (expect(p).resolves.toBe(1) === undefined) throw new Error('bad'); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Value-returning helpers & custom matchers must keep `return await` under `alwaysAwait: true`.
			{
				Code:    `expect.extend({ toMatchAsync: (r: any) => { return expect(r).resolves.toBe(1).then(() => ({ pass: true, message: () => "" })); } });`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`expect.extend({ toMatchAsync: async (r: any) => { return await expect(r).resolves.toBe(1).then(() => ({ pass: true, message: () => "" })); } });`},
			},
			{
				Code:    `const verifyAndFetch = (p: any) => { return Promise.all([expect(p).resolves.toBe(1), fetchUser()]); };`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`const verifyAndFetch = async (p: any) => { return await Promise.all([expect(p).resolves.toBe(1), fetchUser()]); };`},
			},
			{
				Code:    `const check = (p: any) => { return expect(p).resolves.toBe(1).then(() => "ok"); };`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`const check = async (p: any) => { return await expect(p).resolves.toBe(1).then(() => "ok"); };`},
			},
			// Computed method key containing async assertion must defer inner sync method cleanly across 2 passes (never emit bare `await` in sync method).
			{
				Code: `test("t", () => { ({ [expect(p).resolves.toBe(1)]() { expect(q).resolves.toBe(2); expect(r).resolves.toBe(3); } }); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`test("t", async () => { ({ [await expect(p).resolves.toBe(1)]() { expect(q).resolves.toBe(2); expect(r).resolves.toBe(3); } }); });`,
					`test("t", async () => { ({ async [await expect(p).resolves.toBe(1)]() { await expect(q).resolves.toBe(2); await expect(r).resolves.toBe(3); } }); });`,
				},
			},
			// Trailing call / non-Chai method chained directly onto matcher promise.
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1)(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1).hasOwnProperty("then"); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1).foo(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { expect(p).resolves['to' + 'Be'](1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves['to' + 'Be'](1); });`},
			},
			{
				Code: `(expect(1) as any)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `expect()(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs"},
				},
			},
			{
				Code: `import type { require } from "./types"; const { expect: check } = require("@rstest/core"); check(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			// Cast computed identifier keys must NOT be mistaken for `.resolves` / `.toResolve` async assertions.
			{
				Code: `const resolves = "toBe"; test("t", () => { expect(1)[resolves as any].toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},
			{
				Code: `const toResolve = "toBe"; test("t", () => { expect(1)[toResolve as any](1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect(1)[resolves as any];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			// `await` as function name, inner function declaration name, or computed method key must suppress `async` autofix.
			{
				Code: `const f = function await() { expect(p).resolves.toBe(1); };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function run() { function await() {} expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function run() { const o = { [await]() {} }; expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Unknown modifier before a dynamic computed matcher call must still report `modifierUnknown`.
			{
				Code: `expect(1).foo['to' + 'Be'](1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},
			{
				Code: `test("t", () => { expect(p).resolves.foo['to' + 'Be'](1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},
			{
				Code: `expect(1)[resolves]['to' + 'Be'](1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},
		},
	)
}
