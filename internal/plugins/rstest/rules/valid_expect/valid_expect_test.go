package valid_expect_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
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
			{Code: `const Promise = globalThis.Promise; test('starts both elements', async () => { let started = false; const p = Promise.resolve().then(() => started); await Promise.all([expect(p).resolves.toBe(true), (started = true)]); });`},
			{Code: `const { Promise: P } = globalThis; test('t', async () => { await P.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { const [check] = await Promise.resolve([expect(p).resolves.toBe(1)] as const); await check.then(() => {}); });`},
			{Code: `test("t", async () => { const [check] = await Promise.resolve([cond ? expect(p).resolves.toBe(1) : q]); await check.then(() => {}); });`},
			{Code: `test("t", async () => { const [check] = [expect(p).resolves.toBe(1)]; await check.then(() => {}); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await check; });`},
			{Code: `interface Promise<T> {} const Promise = globalThis.Promise; test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `type P = number; const P = globalThis.Promise; test("t", async () => { await P.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `import type { Promise } from "./types"; test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { const o = { [await check]() {} }; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((...args: any[]) => Promise.all(args), ...xs, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => check, ...[], expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((a: any, b: any, c: any) => c, ...[1, 2], expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => err); });`},
			{Code: `test("t", async () => { await Promise.try(function (this: unknown, check: any) { return check; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(function (a = arguments[0]) { return a; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { for await (const _ of [check]) {} }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); switch (x) { case 1: break; } await check; });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); for (const v of xs) { if (v) break; } await check; });`},
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
			// Promise.try spread-wrapped assertion, destructured default param, computed-key await, and rethrowing Promise.reject chains.
			{Code: `test("t", async () => { await Promise.try((check: any) => check, ...[expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { await Promise.try((check: any, { a = check }: any = {}) => a, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => ({ [await check]() {} }, 1), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { throw err; }).catch((err) => err); });`},
			{Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { await e; } });`},
			{Code: `test("t", async () => { try { await Promise.reject(expect(p).resolves.toBe(1)); } catch (e) { await e; } });`},
			{Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { await e; } finally { return; } });`},
			{Code: `test("t", async () => { await Promise.try((...args: any[]) => args[1], ready, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(2)).catch((e) => Promise.reject(e)).catch((e) => e); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => { await Promise.reject(e); }).catch((e) => e); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => await Promise.reject(e)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => { return await Promise.reject(e); }); });`},
			// Container aliases in `Promise.try` / `.catch`.
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch((err) => { const saved = [err]; return Promise.all(saved); }); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved: any[] = []; saved.push(check); return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]).then((arr) => Promise.all(arr)), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(function () { return arguments[0]; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved = new Set([check]); return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve().then(() => { const a = check; return a; }), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set().add(check)), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const queue = { add: (task: () => any) => Promise.resolve().then(task) }; await Promise.try((check: any) => queue.add(() => check), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]).then(([first]) => first), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const queue = { add: (q: Promise<any>) => q }; await Promise.try((check: any) => { return queue.add(check); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const prefix = [Promise.resolve(0)]; await Promise.try((check: any) => Promise.resolve([...prefix, check]).then(([...rest]) => Promise.all(rest)), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return saved[0]; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; await Promise.all(saved); saved.length = 0; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved: any[] = []; saved.length = 0; saved.push(check); await Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `export {}; interface Set { add(p: Promise<unknown>): Promise<void>; } async function run(queue: Set) { await Promise.try((check: any) => queue.add(check), expect(p).resolves.toBe(1)); }`},
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
			// A captured `Promise.resolve([...])` array whose element is NOT awaited
			// reports the diagnostic without inserting an unsafe `await` inside the array.
			{
				Code: `test("t", async () => { const [check] = await Promise.resolve([expect(p).resolves.toBe(1)] as const); log(check); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Conditional await or intervening early return still leaves the assertion potentially not awaited.
			{
				Code: `test("t", async () => { const [check] = await Promise.resolve([expect(p).resolves.toBe(1)] as const); if (cond) { await check; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const [check] = await Promise.resolve([expect(p).resolves.toBe(1)] as const); if (early) return; await check; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// `Promise.try` whose inline callback ignores its argument or never returns/awaits.
			{
				Code: `test("t", async () => { await Promise.try(() => {}, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { check.then(() => {}); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// `Promise.try` setup callback and `Promise[method]([assertion, setup])` must not autofix across setup side effects.
			{
				Code: `test("starts the operation", async () => { let started = false; const p = Promise.resolve().then(() => started); const finished = Promise.withResolvers<void>(); await Promise.try(() => { started = true; }, expect(p).resolves.toBe(true).then(finished.resolve, finished.reject)); await finished.promise; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("starts the operation", async () => { let started = false; const p = Promise.resolve().then(() => started); const method = "resolve"; await Promise[method]([expect(p).resolves.toBe(true), (started = true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// `Promise.try` returning a different parameter, and `Promise.reject(check).catch(() => {})` dropping the rejection reason.
			{
				Code: `test("t", async () => { await Promise.try((first: any, second: any) => first, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).finally(() => {}).then(undefined, () => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Statically folded `Promise[resolve]([...])` where `resolve = 'resolve'` does not await elements.
			{
				Code: `test("t", async () => { const resolve = 'resolve'; await Promise[resolve]([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { const resolve = 'resolve'; await Promise[resolve]([await expect(p).resolves.toBe(1)]); });`},
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
				Code: `test("t", async () => { let started = false; const Promise = { all: (x: any) => { started = true; return x; } }; await Promise.all([expect(p).resolves.toBe(true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const start = (_x: unknown) => { started = true; }; start(expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const run = Promise.try; await run(() => { started = true; }, expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; await Promise.try.call(Promise, () => { started = true; }, expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; await Promise.try(() => { started = true; }, ...[expect(p).resolves.toBe(true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const id = <T>(x: T) => x; await Promise.all([id(expect(p).resolves.toBe(true)), (started = true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
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
				Code: `test("t", async () => { const [check] = await Promise.resolve([cond ? expect(p).resolves.toBe(1) : q]); log(check); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const [check] = [expect(p).resolves.toBe(1)]; log(check); });`,
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
					{MessageId: "matcherNotCalled", Column: 20, EndColumn: 21},
				},
			},
			{
				Code: `expect(1).resolves[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 20, EndColumn: 21},
				},
			},
			{
				Code: `expect(1).not[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 15, EndColumn: 16},
				},
			},
			{
				Code: `expect(1).resolves['to' + 'Be'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 20, EndColumn: 31},
				},
			},
			{
				Code: `expect(1).to.be[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 17, EndColumn: 18},
				},
			},
			{
				Code: `expect(1)[0].foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 14, EndColumn: 17},
				},
			},
			{
				Code: `expect(1)[0].not;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 17},
				},
			},
			{
				Code: `expect(1)[0]['not' as const];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 19},
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
			// Resolved Promise alias preserves `return await` under `alwaysAwait: true`.
			{
				Code:    `const key = "Promise"; const P = globalThis[key]; test("t", () => { return P.all([expect(p).resolves.toBe(1), fetchData()]); });`,
				Options: map[string]any{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`const key = "Promise"; const P = globalThis[key]; test("t", async () => { return await P.all([expect(p).resolves.toBe(1), fetchData()]); });`},
			},
			// Downstream reads of assigned assertion promises withhold autofix.
			{
				Code: `test("t", () => { const check = expect(p).resolves.toBe(1); check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); if (cond) await check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { let check: any; check = expect(p).resolves.toBe(1); check.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const arr = [expect(p).resolves.toBe(1)]; arr[0].then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Conditional awaits in destructuring assignment defaults and early exits (yield, outer break) still report.
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); let x: any; [x = await check] = [1]; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); let x: any; ({ x = await check } = { x: 1 }); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1); yield 1; await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { outer: for (const v of xs) { const check = expect(p).resolves.toBe(1); for (const w of ys) { if (w) break outer; } await check; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Promise.try edge cases: preceding spread, generator callback, TS `this` parameter, array first arg, property-key `arguments`.
			{
				Code: `test("t", async () => { await Promise.try((a: any, b: any) => b, ...[1, 2], expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((a: any) => { a.then(() => {}); }, ...[], expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(function* (check: any) { return check; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(function (this: unknown) { return 1; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try([expect(p).resolves.toBe(1)] as any); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(function () { const o = { arguments: 1 }; return o; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(function () { const t: [arguments: number] = [1]; return t; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(function () { const x: typeof arguments | undefined = undefined; return x; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Exported/global/non-null-asserted assignment targets and array-literal consumers withhold autofix.
			{
				Code: `export let check: Promise<void>; test("t", () => { check = expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { globalCheck = expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const holder: { p?: Promise<void> } = {}; holder.p! = expect(p).resolves.toBe(1); return holder.p; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { let check: any = expect(p).resolves.toBe(1); check! ||= fallback; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { let arr: any; arr ??= [expect(p).resolves.toBe(1)]; return Promise.all(arr); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { [expect(p).resolves.toBe(1)][0].then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { Promise.try((arr: any) => arr[0].then(() => {}), [expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const arr = cond ? [expect(p).resolves.toBe(1)] : []; arr[0].then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Same-statement and computed-key `yield` prevent treating later `await check` as unconditional.
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1); const o = { [yield 1]() {} }; await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1), _ = yield 1; await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1); use(yield 1, await check); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = Promise.all([expect(p).resolves.toBe(1), yield 1]); await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1); await Promise.all([check, yield 1]); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1).then(yield 1); await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const check = expect(p).resolves.toBe(1); await check.then(yield 1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Non-consuming callbacks, type-only references, unused defaults, and swallowed Promise.reject variants.
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { console.error(err); return null; }); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { const x: typeof err | undefined = undefined; return x; }); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => ({ check }), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(() => {}, ...[expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const rejected = Promise.reject(expect(p).resolves.toBe(1)); await rejected.catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); try { await Promise.reject(check); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.allSettled([Promise.reject(check)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const start = () => { started = true; }; await runAll([expect(p).resolves.toBe(true), start()]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const a = expect(p).resolves.toBe(true), b = (started = true); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.allSettled([Promise.reject(expect(p).resolves.toBe(1))]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { Promise.reject(expect(p).resolves.toBe(1)); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(Promise.reject(check)).catch((err) => err).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.allSettled([Promise.resolve(Promise.reject(expect(p).resolves.toBe(1)))]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { try { Promise.reject(expect(p).resolves.toBe(1)).finally(() => {}); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { try { Promise.reject(expect(p).resolves.toBe(1)); } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let checks: Promise<void>[]; checks = [expect(p).resolves.toBe(1)]; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { try { const check = expect(p).resolves.toBe(1); await Promise.reject(Promise.reject(check)); } catch (e) { await e; } } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((...args: any[]) => args[0], ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const xs = [42]; await Promise.try((check: any, a = check) => a, expect(p).resolves.toBe(1), ...xs); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const run = (p: any) => p.then; run(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function helper(check: Promise<void>) { check = expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.reject(Promise.reject(expect(p).resolves.toBe(1))); } catch (e) { try { await e; } catch {} } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.reject(expect(p).resolves.toBe(1)); } catch (e) { try { return e; } finally { return; } } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any, { a = check }: any = { a: 1 }) => a, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { try { Promise.reject(expect(p).resolves.toBe(2)).catch((e) => Promise.reject(e)); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const promises: Promise<void>[] = [Promise.resolve()]; for (let q of promises) { q = expect(p).resolves.toBe(1); } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const record: (x: Promise<void>) => void = (x) => {}; record(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { function check(x: any) { const { then } = x; } check(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => { await Promise.reject(e); }); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.try((q) => Promise.reject(q), expect(p).resolves.toBe(1)); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `const obj = { f(x) {} } as { f(x: Promise<void>): void }; test("t", () => { obj.f(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Getter/spread sibling autofix withholding and `Promise.resolve([p])` container false negatives.
			{
				Code: `test("t", async () => { let started = false; const setup = { get start() { started = true; return 1; } }; await Promise.resolve([expect(p).resolves.toBe(true), setup.start]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(true), ...items]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(true), { ...setup }]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return Promise.resolve(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]).then(() => 1), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => new Set([check]), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((...args: any[]) => { const copy = args; return copy[0]; }, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => () => { return check; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; await Promise.resolve([expect(p).resolves.toBe(true), (globalThis as any)[(started = true, 'Promise')].resolve(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const setup = { get start() { started = true; return true; } }; await Promise.resolve([expect(p).resolves.toBe(true), expect(setup).toHaveProperty('start', true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [Promise.resolve(0), check]; return saved[0]; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return (saved as any)['length']; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved: any[] = []; saved.push(check); saved.pop(); return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([Promise.resolve(0), check]).then(([first]) => first), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((...checks: any[]) => { checks.pop(); return Promise.all(checks); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => new Set([check]).clear(), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved: any[] = []; saved.push(check); saved.length--; return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { let saved: any[] = []; return saved.push(check); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve(new Set([check])).then((items) => items.clear()), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const prefix = [Promise.resolve(0)]; await Promise.try((check: any) => Promise.resolve([...prefix, check]).then(([...rest]) => rest.push(1)), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const prefix = [Promise.resolve(0)]; await Promise.try((check: any) => Promise.resolve([...prefix, check]).then((items) => items.push(1)), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const p = Promise.resolve().then(() => started); const t = { get then() { started = true; return (accept: any) => Promise.resolve(true).then(accept); } }; await Promise.resolve([expect(p).resolves.toBe(true), expect(t).resolves.toBe(true)]); });`,
				Output: []string{
					`test("t", async () => { let started = false; const p = Promise.resolve().then(() => started); const t = { get then() { started = true; return (accept: any) => Promise.resolve(true).then(accept); } }; await Promise.resolve([expect(p).resolves.toBe(true), await expect(t).resolves.toBe(true)]); });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(true), (expect(Promise.resolve(true)) as any).toResolve()]); });`,
				Output: []string{
					`test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(true), await (expect(Promise.resolve(true)) as any).toResolve()]); });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { async function run(pending: Promise<unknown>[]) { await Promise.try((check: any) => pending.push(check), expect(p).resolves.toBe(1)); } await run([]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set().add([check])), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set().add(() => check)), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const tasks: Array<() => unknown> = []; const queue = { add(task: () => unknown) { tasks.push(task); } }; await Promise.try((check: any) => queue.add(() => check), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set([[check]])), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set([() => check])), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new WeakSet().add(check) as any), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; saved.length = 0; await Promise.all(saved); saved.push(check); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [Promise.resolve(0), check]; saved.length = 0; saved.push(check); return saved[1]; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { let saved = [check]; saved.length = 0; saved ||= [check]; await Promise.all(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async function () { const alias = arguments; delete (arguments as any)[0]; await Promise.all(alias as any); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; const alias = cond ? saved : []; saved.length = 0; await Promise.all(alias); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; const alias = saved; alias.length = 0; await Promise.all(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; const alias = saved; alias[0] = Promise.resolve(0); return saved[0]; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; saved[0] &&= Promise.resolve(1); return saved[0]; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { var saved = [check]; var saved = [Promise.resolve(1)]; return saved[0]; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; let alias = saved; saved.length = 0; alias = [check]; await Promise.all(saved); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const out = Promise.resolve().then(() => { saved.length = 0; return saved[0]; }); const saved = [check]; return out; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async function () { const alias = arguments; delete (alias as any)[0]; await Promise.all(arguments as any); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
		},
	)
}

func TestValidExpectPromiseAllLinearScaling(t *testing.T) {
	var b strings.Builder
	b.WriteString("test('many awaited Promise.all assertions', async () => {\n")
	for range 2000 {
		b.WriteString("  await Promise.all([expect(p).resolves.toBe(1)]);\n")
	}
	b.WriteString("});\n")

	var b2 strings.Builder
	b2.WriteString("test('many const-bound awaited assertions', async () => {\n")
	const n = 2000
	for i := range n {
		fmt.Fprintf(&b2, "  const c%d = expect(p).resolves.toBe(1);\n", i)
	}
	b2.WriteString("  await Promise.all([")
	for i := range n {
		if i > 0 {
			b2.WriteString(", ")
		}
		fmt.Fprintf(&b2, "c%d", i)
	}
	b2.WriteString("]);\n});\n")

	var b3 strings.Builder
	b3.WriteString("test('many un-awaited assertions in one Promise.resolve array', async () => {\n  await Promise.resolve([\n")
	for i := range n {
		fmt.Fprintf(&b3, "    expect(p).resolves.toBe(%d),\n", i)
	}
	b3.WriteString("  ]);\n});\n")

	const tryN = 1000
	var b4 strings.Builder
	b4.WriteString("test('many Promise.try assertion args with many setup statements', async () => {\n  await Promise.try((...checks: any[]) => {\n")
	for i := range tryN {
		fmt.Fprintf(&b4, "    const s%d = %d;\n", i, i)
	}
	b4.WriteString("    return Promise.all(checks);\n  }")
	for i := range tryN {
		fmt.Fprintf(&b4, ", expect(p).resolves.toBe(%d)", i)
	}
	b4.WriteString(");\n});\n")

	var b5 strings.Builder
	b5.WriteString("test('many Promise.try assertion args indexing checks[0] with many setup statements', async () => {\n  await Promise.try((...checks: any[]) => {\n")
	for i := range tryN {
		fmt.Fprintf(&b5, "    const s%d = %d;\n", i, i)
	}
	b5.WriteString("    return checks[0];\n  }")
	for i := range tryN {
		fmt.Fprintf(&b5, ", expect(p).resolves.toBe(%d)", i)
	}
	b5.WriteString(");\n});\n")

	const namedN = 400
	var b6 strings.Builder
	b6.WriteString("test('many named Promise.try args with unrelated array pops', async () => {\n  await Promise.try(async (")
	for i := range namedN {
		if i > 0 {
			b6.WriteString(", ")
		}
		fmt.Fprintf(&b6, "a%d: any", i)
	}
	b6.WriteString(") => {\n")
	for i := range namedN {
		fmt.Fprintf(&b6, "    const u%d = [1]; u%d.pop(); await a%d;\n", i, i, i)
	}
	b6.WriteString("  }")
	for i := range namedN {
		fmt.Fprintf(&b6, ", expect(p).resolves.toBe(%d)", i)
	}
	b6.WriteString(");\n});\n")

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	for _, tc := range []struct {
		name         string
		code         string
		demand       rule.EditDemand
		wantDiagSize int
	}{
		{name: "promise-all.ts", code: b.String(), demand: rule.EditDemandNone, wantDiagSize: 0},
		{name: "const-bound.ts", code: b2.String(), demand: rule.EditDemandNone, wantDiagSize: 0},
		{name: "not-awaited-array-autofix.ts", code: b3.String(), demand: rule.EditDemandAutofix, wantDiagSize: n},
		{name: "promise-try-rest-all.ts", code: b4.String(), demand: rule.EditDemandNone, wantDiagSize: 0},
		{name: "promise-try-rest-index0.ts", code: b5.String(), demand: rule.EditDemandNone, wantDiagSize: tryN - 1},
		{name: "promise-try-named-pops.ts", code: b6.String(), demand: rule.EditDemandNone, wantDiagSize: 0},
	} {
		program, sourceFile, err := helper.CreateTestProgram(tc.code, tc.name, "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     valid_expect.ValidExpectRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return valid_expect.ValidExpectRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: tc.demand,
				Report: func(d rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, d)
				},
			},
		})
		if elapsed := time.Since(start); elapsed > 10*time.Second {
			t.Fatalf("%s took %v; expected linear-time execution", tc.name, elapsed)
		}
		if len(diagnostics) != tc.wantDiagSize {
			t.Fatalf("%s reported %d diagnostics; want %d", tc.name, len(diagnostics), tc.wantDiagSize)
		}
	}
}
