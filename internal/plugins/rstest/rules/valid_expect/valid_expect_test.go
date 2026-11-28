package valid_expect_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
			{Code: `foo().expect(1);`},
			{Code: `/a/.expect(1).toBe(2);`},
			{Code: `this.expect(1);`},
			{Code: `new Foo().expect(1).toBe(2);`},

			// Transparent TS assertions, Promise combinators, direct const awaits, and idiomatic Promise.try.
			{Code: `test("t", async () => { await (expect(p).resolves.toBe(1) as Promise<void>); });`},
			{Code: `test("t", () => { return expect(p).resolves.toBe(1) as Promise<void>; });`},
			{Code: `test("t", async () => { await expect(p).resolves.toBe(1).finally(() => {}); });`},
			{Code: `test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`},
			{Code: `test("t", async () => { await Promise.try(() => expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await check; });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); switch (x) { case 1: break; } await check; });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); for (const v of xs) { if (v) break; } await check; });`},
			{Code: `test("t", async () => { await (log(), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await (cond ? expect(p).resolves.toBe(1) : expect(q).resolves.toBe(2)).then(() => {}); });`},
			{Code: `const Promise = globalThis.Promise; test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `const { Promise: P } = globalThis; test("t", async () => { await P.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `import type { Promise } from "./types"; test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `interface Promise<T> {} const Promise = globalThis.Promise; test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `type P = number; const P = globalThis.Promise; test("t", async () => { await P.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `interface Promise<T> { custom(): void; } test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `export {}; namespace Promise { export type Extra = number; } test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { await Promise.all([Promise.all([expect(p).resolves.toBe(1)])]); });`},
			{Code: `test("t", async () => { await Promise.resolve(Promise.all([expect(p).resolves.toBe(1)])); });`},
			{Code: `test("t", async () => { const assertion = expect(p).resolves.toBe(1); await Promise.all([assertion.finally(cleanup)]); });`},
			{Code: `test("t", async () => { const assertion = expect(p).resolves.toBe(1); await Promise.all([Promise.resolve(assertion)]); });`},
			{Code: `test("t", async () => { let result: any; const assertion = expect(p).resolves.toBe(1); result = await assertion; });`},
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
				Code: `import.meta.rstest!.expect(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},

			// Trailing index access / non-Promise method calls / cast keys.
			{
				Code: `expect(1)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `(expect(1) as any)[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 20, EndColumn: 21},
				},
			},
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1)(); });`,
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

			// Promise.resolve([...]), Promise.reject(...).catch(), and Promise.try callback args do not await the inner assertion.
			{
				Code: `test("t", async () => { await Promise.resolve([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.resolve([await expect(p).resolves.toBe(1)]); });`},
			},
			{
				Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { return Promise.all([check]); }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); if (cond) return; await check; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},

			// Conservative autofix safety: non-fixable functions, early return, try/catch, and sibling deadlocks.
			{
				Code: `function check(p: Promise<number>): any { expect(p).resolves.toBe(1); }`,
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
				Code: `if (cond) function check() { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code:    `test("t", () => { if (cond) { return expect(p).resolves.toBe(1); } next(); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { if (cond) { return await expect(p).resolves.toBe(1); } next(); });`},
			},
			{
				Code: `test("t", async () => { try { expect(p).resolves.toBe(1); } catch {} });`,
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
				Code: `test("t", async () => { let started = false; const Promise = { all: (x: any) => { started = true; return x; } }; await Promise.all([expect(p).resolves.toBe(true)]); });`,
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
				Code: `test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.extend({ toStart() { gate.resolve(1); return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), (expect(Promise.resolve(1)).resolves as any).toStart()]); await done.promise; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.extend({ toStart() { gate.resolve(1); return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), await (expect(Promise.resolve(1)).resolves as any).toStart()]); await done.promise; });`,
				},
			},
			{
				Code: `test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), expect(Promise.resolve(() => { gate.resolve(1); throw new Error('boom'); })).resolves.toThrow('boom')]); await done.promise; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), await expect(Promise.resolve(() => { gate.resolve(1); throw new Error('boom'); })).resolves.toThrow('boom')]); await done.promise; });`,
				},
			},
			{
				Code: `test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.addEqualityTesters([() => { gate.resolve(1); return true; }]); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), expect(Promise.resolve(1)).resolves.toEqual(1)]); await done.promise; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.addEqualityTesters([() => { gate.resolve(1); return true; }]); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), await expect(Promise.resolve(1)).resolves.toEqual(1)]); await done.promise; });`,
				},
			},
			// Red-team regressions: return/call/shadowed-expect/computed-extend/catch-var/using/class-extends/overloads/arrow-await/reject-const/withResolvers/nested-globalThis.
			{
				Code: `function makeChecks(p: any) { return [expect(p).resolves.toBe(1)]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const handle = (p: any) => p.then(() => {}); handle(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const Promise = { all: (arr: any) => arr[0].then(() => {}) }; Promise.all([expect(globalThis.Promise.resolve(1)).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `import { expect as check } from "@rstest/core"; test("t", async () => { const expect = () => ({ resolves: { toBe: () => Promise.resolve() } }); await Promise.resolve([check(p).resolves.toBe(1), expect(Promise.resolve(1)).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { expect['ext' + 'end']({ toBe() { return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(p).resolves.toBe(1), expect(Promise.resolve(1)).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`test("t", async () => { expect['ext' + 'end']({ toBe() { return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(p).resolves.toBe(1), await expect(Promise.resolve(1)).resolves.toBe(1)]); });`,
				},
			},
			{
				Code: `test("t", () => { try { work(); } catch (e) { var e = expect(p).resolves.toBe(1); e.then(() => {}); } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { using x = expect(p).resolves.toBe(1) as any; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `class C extends (function() { expect(p).resolves.toBe(1); } as any) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function f(x: number): void; function f(x: number) { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code:    `run(() => { return expect(p).resolves.toBe(1) as unknown as Promise<string>; });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{
					`run(async () => { return await (expect(p).resolves.toBe(1) as unknown as Promise<string>); });`,
				},
			},
			{
				Code: `const await = 1; test("t", () => { const f = (x = await) => x; expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = Promise.reject(expect(p).resolves.toBe(1)); await check.catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve(cond ? Promise.reject(expect(p).resolves.toBe(1)) : q).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.withResolvers(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.withResolvers([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `const { fake: { Promise: P } } = globalThis as any; test("t", async () => { await P.all([expect(p).resolves.toBe(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Round 2 red-team regressions: multi-arg/lazy-thenable Promise.all, declare function wrap, Promise.try.call, Promise.allSettled/any(Promise.reject(...)).
			{
				Code: `test("t", () => { const starter = { then(ok: (v: number) => void) { ok(1); } }; Promise.all([starter, cond ? expect(p).resolves.toBe(1) : undefined]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const iter = { *[Symbol.iterator]() { yield 1; } }; await Promise.all(iter as any, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `declare function wrap(p: Promise<void>): Promise<void>; test("t", async () => { await Promise.all([wrap(expect(Promise.resolve(1)).resolves.toBe(1))]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { Promise.try.call(Promise, (p: Promise<void>) => p.then(() => {}), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
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
				Code: `test("t", async () => { await Promise.allSettled([cond ? Promise.reject(expect(p).resolves.toBe(1)) : q]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.any([Promise.reject(expect(p).resolves.toBe(1)), Promise.resolve(1)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", done => { expect(p).resolves.toBe(1).then(() => done(), done); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function Worker() { expect(p).resolves.toBe(1); } new Worker();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `const cb: () => number = () => { expect(p).resolves.toBe(1); return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `function makeGate(cond: boolean) { expect(p).resolves.toBe(1); if (cond) return; return Promise.withResolvers<number>(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code:    `import * as r from "@rstest/core"; test("t", () => { return r.expect(p).resolves.toBe(1); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`import * as r from "@rstest/core"; test("t", async () => { await r.expect(p).resolves.toBe(1); });`},
			},
			{
				Code:    `import { expect as check } from "@rstest/core"; test("t", () => { return check(p).resolves.toBe(1); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`import { expect as check } from "@rstest/core"; test("t", async () => { await check(p).resolves.toBe(1); });`},
			},
			{
				Code:    `const fn = (cond: boolean) => { return cond ? 1 : expect(p).resolves.toBe(1); };`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
		},
	)
}

func TestValidExpectPromiseAllLinearScaling(t *testing.T) {
	var b1 strings.Builder
	b1.WriteString("test('many awaited Promise.all assertions', async () => {\n")
	for range 2000 {
		b1.WriteString("  await Promise.all([expect(p).resolves.toBe(1)]);\n")
	}
	b1.WriteString("});\n")

	var b2 strings.Builder
	b2.WriteString("test('many const-bound awaited assertions', async () => {\n")
	for i := range 2000 {
		fmt.Fprintf(&b2, "  const c%d = expect(p).resolves.toBe(1);\n", i)
	}
	b2.WriteString("  await Promise.all([")
	for i := range 2000 {
		if i > 0 {
			b2.WriteString(", ")
		}
		fmt.Fprintf(&b2, "c%d", i)
	}
	b2.WriteString("]);\n});\n")

	var b3 strings.Builder
	b3.WriteString("test('many un-awaited assertions in one Promise.resolve array', async () => {\n  await Promise.resolve([\n")
	for i := range 2000 {
		fmt.Fprintf(&b3, "    expect(p).resolves.toBe(%d),\n", i)
	}
	b3.WriteString("  ]);\n});\n")

	var b4 strings.Builder
	b4.WriteString("test('many pure literal assertions in one Promise.resolve array', async () => {\n  await Promise.resolve([\n")
	for i := range 500 {
		fmt.Fprintf(&b4, "    expect(Promise.resolve(%d)).resolves.toBe(%d),\n", i, i)
	}
	b4.WriteString("  ]);\n});\n")

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	for _, tc := range []struct {
		name         string
		code         string
		wantDiagSize int
	}{
		{name: "many-promise-all", code: b1.String(), wantDiagSize: 0},
		{name: "many-const-bound", code: b2.String(), wantDiagSize: 0},
		{name: "many-floating-array", code: b3.String(), wantDiagSize: 2000},
		{name: "many-pure-floating-array", code: b4.String(), wantDiagSize: 500},
	} {
		program, sourceFile, err := helper.CreateTestProgram(tc.code, tc.name+".ts", "tsconfig.json")
		if err != nil {
			t.Fatalf("%s: CreateTestProgram: %v", tc.name, err)
		}
		var diagnostics []rule.RuleDiagnostic
		start := time.Now()
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
				Demand: rule.EditDemandAutofix,
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

func TestValidExpectSourceOnlyMode(t *testing.T) {
	root := fixtures.GetRootDir()
	for _, tc := range []struct {
		name         string
		code         string
		wantDiagSize int
		wantHasFix   bool
	}{
		{
			name:         "floating-promise-try-arg.ts",
			code:         `async function run(pending: Promise<unknown>[]) { const p = Promise.resolve(1); await Promise.try(check => pending.push(check), expect(p).resolves.toBe(1)); }`,
			wantDiagSize: 1,
		},
		{
			name:         "awaited-const-binding.ts",
			code:         `async function run() { const p = Promise.resolve(1); const check = expect(p).resolves.toBe(1); await check; }`,
			wantDiagSize: 0,
		},
		{
			name:         "source-only-typeof-query.ts",
			code:         `test("t", () => { const p = expect(Promise.resolve(1)).resolves.toBe(1); type P = typeof p; });`,
			wantDiagSize: 1,
			wantHasFix:   false,
		},
		{
			name:         "source-only-unread-const-with-prop-name.ts",
			code:         `test("t", () => { const res = expect(p).resolves.toBe(1); const obj = { res: 1 }; });`,
			wantDiagSize: 1,
			wantHasFix:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fileName := tspath.ResolvePath(root.Dir, tc.name)
			fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: tc.code})
			host := utils.CreateCompilerHost(root.Dir, fs)
			sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames:   []string{fileName},
				Host:            host,
				CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
				SingleThreaded:  true,
			})
			if err != nil {
				t.Fatalf("NewFromRoots: %v", err)
			}
			if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
				t.Fatal("expected a source-only Program with no TypeChecker")
			}
			lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
				Programs:         []*lintprogram.Program{sourceProgram},
				TargetsByProgram: [][]string{{fileName}},
				SingleThreaded:   true,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:     valid_expect.ValidExpectRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return valid_expect.ValidExpectRule.Run(ctx, nil)
						},
					}}
				},
			})
			if err != nil {
				t.Fatalf("PrepareLintPlan: %v", err)
			}
			var diagnostics []rule.RuleDiagnostic
			if _, err := linter.RunLinter(linter.RunLinterOptions{
				SingleThreaded: true,
				LintPlan:       lintPlan,
				Consumer: rule.DiagnosticConsumer{
					Demand: rule.EditDemandAutofix,
					Report: func(d rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, d)
					},
				},
			}); err != nil {
				t.Fatalf("RunLinter: %v", err)
			}
			if len(diagnostics) != tc.wantDiagSize {
				t.Fatalf("%s reported %d diagnostics; want %d", tc.name, len(diagnostics), tc.wantDiagSize)
			}
			if tc.wantDiagSize == 1 {
				gotHasFix := len(diagnostics[0].Fixes()) > 0
				if gotHasFix != tc.wantHasFix {
					t.Fatalf("%s hasFix = %v; want %v", tc.name, gotHasFix, tc.wantHasFix)
				}
			}
		})
	}
}

func TestValidExpectMultiRuleAtomicSyncAutofix(t *testing.T) {
	// A sibling rule editing the arrow head (`x =>` -> `(x) =>`) conflicts with
	// valid-expect's `async ` insertion. Because every fixable diagnostic in a
	// sync function carries its own `async ` fix atomically, neither assertion
	// emits a bare `await` into a synchronous function on pass 1.
	code := "test('t', x => {\n  expect(p1).resolves.toBe(1);\n  expect(p2).resolves.toBe(2);\n});\n"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "multi-rule-atomic.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(program),
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{
				{
					Name:     "synthetic/arrow-parens",
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return rule.RuleListeners{
							ast.KindArrowFunction: func(node *ast.Node) {
								params := node.Parameters()
								if len(params) != 1 {
									return
								}
								r := utils.TrimNodeTextRange(ctx.SourceFile, params[0])
								ctx.ReportNodeWithFixes(params[0], rule.RuleMessage{Id: "parens", Description: "wrap"}, rule.RuleFixReplaceRange(r, "("+ctx.SourceFile.Text()[r.Pos():r.End()]+")"))
							},
						}
					},
				},
				{
					Name:     valid_expect.ValidExpectRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return valid_expect.ValidExpectRule.Run(ctx, nil)
					},
				},
			}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			},
		},
	})
	fixed, _, didFix := linter.ApplyRuleFixes(code, diagnostics)
	if !didFix {
		t.Fatal("expected pass 1 to apply arrow-parens fix")
	}
	wantPass1 := "test('t', (x) => {\n  expect(p1).resolves.toBe(1);\n  expect(p2).resolves.toBe(2);\n});\n"
	if fixed != wantPass1 {
		t.Fatalf("pass 1 output mismatch:\ngot:\n%s\nwant:\n%s", fixed, wantPass1)
	}
}
