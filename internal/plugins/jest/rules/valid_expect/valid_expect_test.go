package valid_expect_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_expect"
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
			{Code: "expect.hasAssertions"},
			{Code: "expect.hasAssertions()"},
			{Code: "expect.extend;"},
			{Code: "expect.resolves;"},
			{Code: "expect.resolves.toBe(2);"},
			{Code: "expect(\"something\").toEqual(\"else\");"},
			{Code: "expect(true).toBeDefined();"},
			{Code: "expect([1, 2, 3]).toEqual([1, 2, 3]);"},
			{Code: "expect(undefined).not.toBeDefined();"},
			{Code: "import { expect } from 'chai'; expect(foo).to.equal(bar);"},
			{Code: "test(\"valid-expect\", () => { return expect(Promise.resolve(2)).resolves.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", () => { return expect(Promise.reject(2)).rejects.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", () => { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });"},
			// Parentheses are nodes in the TypeScript AST but not in ESTree, so
			// the await/return check must look past them to match upstream.
			{Code: "test(\"valid-expect\", async () => { await (expect(Promise.resolve(2)).resolves.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", () => { return (expect(Promise.resolve(2)).resolves.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", () => { return (condition ? expect(Promise.resolve(2)).resolves.toBeDefined() : expect(2).toBeDefined()); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.all([(expect(Promise.resolve(2)).resolves.toBeDefined()), expect(Promise.resolve(3)).resolves.toBeDefined()]); });"},
			{Code: "test(\"valid-expect\", async () => { await (expect(Promise.resolve(2)).resolves.toBeDefined()).then(() => {}); });"},
			{Code: "test(\"valid-expect\", () => { return expect(Promise.resolve(2)).resolves.toBeDefined() as Promise<void>; });"},
			{Code: "test(\"valid-expect\", async () => { await (expect(Promise.resolve(2)).resolves.toBeDefined() as Promise<void>); });"},
			{Code: "test(\"valid-expect\", async () => { const resolve = 'all'; await Promise[resolve]([expect(Promise.resolve(true)).resolves.toBe(true), startWork()]); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.try((check: any) => check.then(() => {}), expect(Promise.resolve(1)).resolves.toBe(1)); });"},
			{Code: "const Promise = globalThis.Promise; test('starts both elements', async () => { let started = false; const p = Promise.resolve().then(() => started); await Promise.all([expect(p).resolves.toBe(true), (started = true)]); });"},
			{Code: "test(\"valid-expect\", async () => { const [check] = await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1)] as const); await check.then(() => {}); });"},
			{Code: "test(\"valid-expect\", async () => { const [check] = [expect(Promise.resolve(1)).resolves.toBe(1)]; await check.then(() => {}); });"},
			{Code: "test(\"valid-expect\", async () => { const check = expect(Promise.resolve(1)).resolves.toBe(1); await check; });"},
			{Code: "interface Promise<T> {} const Promise = globalThis.Promise; test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "type P = number; const P = globalThis.Promise; test('t', async () => { await P.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "import type { Promise } from './types'; test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "test('t', async () => { await Promise.try(async (check: any) => { const o = { [await check]() {} }; }, expect(p).resolves.toBe(1)); });"},
			{Code: "test('t', async () => { await Promise.try((...args: any[]) => Promise.all(args), ...xs, expect(p).resolves.toBe(1)); });"},
			{Code: "test('t', async () => { await Promise.try((check: any) => check, ...[], expect(p).resolves.toBe(1)); });"},
			{Code: "test('t', async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => err); });"},
			{Code: "test('t', async () => { await Promise.try(function (this: unknown, check: any) { return check; }, expect(p).resolves.toBe(1)); });"},
			{Code: "test('t', async () => { await Promise.try(async (check: any) => { for await (const _ of [check]) {} }, expect(p).resolves.toBe(1)); });"},
			{Code: "test('t', async () => { const check = expect(p).resolves.toBe(1); switch (x) { case 1: break; } await check; });"},
			{Code: "test('t', async () => { const check = expect(p).resolves.toBe(1); for (const v of xs) { if (v) break; } await check; });"},
			{Code: "test(\"valid-expect\", () => { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", function () { return expect(Promise.resolve(2)).resolves.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", function () { return expect(Promise.resolve(2)).rejects.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", function () { return Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", function () { return Promise.resolve(expect(Promise.resolve(2)).rejects.not.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", () => expect(Promise.resolve(2)).resolves.toBeDefined());", Options: map[string]interface{}{"alwaysAwait": true}},
			{Code: "test(\"valid-expect\", () => expect(Promise.resolve(2)).resolves.toBeDefined());"},
			{Code: "test(\"valid-expect\", () => expect(Promise.reject(2)).rejects.toBeDefined());"},
			{Code: "test(\"valid-expect\", () => expect(Promise.reject(2)).resolves.not.toBeDefined());"},
			{Code: "test(\"valid-expect\", () => expect(Promise.reject(2)).rejects.not.toBeDefined());"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", async function () { await expect(Promise.reject(2)).resolves.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", async function () { await expect(Promise.reject(2)).rejects.not.toBeDefined(); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.resolve(expect(Promise.reject(2)).rejects.not.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.reject(expect(Promise.reject(2)).rejects.not.toBeDefined()); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.all([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.race([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.allSettled([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.any([expect(Promise.reject(2)).rejects.not.toBeDefined(), expect(Promise.reject(2)).rejects.not.toBeDefined()]); });"},
			{Code: "test(\"valid-expect\", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")); });"},
			{Code: "test(\"valid-expect\", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")).then(() => console.log(\"another valid case\")); });"},
			{Code: "test(\"valid-expect\", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log(\"valid-case\")); });"},
			{Code: "test(\"valid-expect\", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")).catch(() => console.log(\"another valid case\")); });"},
			{Code: "test(\"valid-expect\", async () => { return expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")).then(() => console.log(\"another valid case\")); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().catch(() => console.log(\"valid-case\")); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => console.log(\"valid-case\")).catch(() => console.log(\"another valid case\")); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.reject(2)).resolves.not.toBeDefined().then(() => { expect(someMock).toHaveBeenCalledTimes(1); }); });"},
			{Code: `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
    return expect(Promise.resolve(2)).resolves.toBe(1);
  });
});
    `},
			{Code: `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
  });
});
    `},
			{Code: `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => expect(Promise.resolve(2)).resolves.toBe(1));
});
    `},
			{Code: `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).toBe(true)
      : expect(obj).resolves.not.toThrow();
  }
});
    `},
			{Code: `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});
    `},
			{Code: `expect.extend({
  toResolve(obj) {
    return this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});
    `},
			{Code: "expect(1).toBe(2);", Options: map[string]interface{}{"maxArgs": 2}},
			{Code: "expect(1, \"1 !== 2\").toBe(2);", Options: map[string]interface{}{"maxArgs": 2}},
			{Code: "expect(1, \"1 !== 2\").toBe(2);", Options: map[string]interface{}{"maxArgs": 2, "minArgs": 2}},
			{Code: "test(\"valid-expect\", () => { expect(2).not.toBe(2); });", Options: map[string]interface{}{"asyncMatchers": []interface{}{"toRejectWith"}}},
			{Code: "test(\"valid-expect\", () => { expect(Promise.reject(2)).toRejectWith(2); });", Options: map[string]interface{}{"asyncMatchers": []interface{}{"toResolveWith"}}},
			{Code: "test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).toResolve(); });", Options: map[string]interface{}{"asyncMatchers": []interface{}{"toResolveWith"}}},
			{Code: "test(\"valid-expect\", async () => { expect(Promise.resolve(2)).toResolve(); });", Options: map[string]interface{}{"asyncMatchers": []interface{}{"toResolveWith"}}},
			{Code: "expect().pass();", Options: map[string]interface{}{"minArgs": 0}},
			{Code: "this.expect(1).toBe(2);"},
			{Code: "/a/.expect(1).toBe(2);"},
			{Code: "this.expect(1);"},
			{Code: "patterns[0].expect(1).toBe(2);"},
			{Code: "new Foo().expect(1).toBe(2);"},
			{Code: `test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`},
			{Code: `test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).finally(() => {}); });`},
			{Code: `test("t", async () => { await expect(p).resolves.toBe(1).finally(() => {}); });`},
			{Code: `test("t", async () => { await expect(p).resolves[('toBe')](1); });`},
			{Code: `test("t", async () => { await (expect(p).resolves.toBe)(1); });`},
			{Code: `test("t", async () => { await (expect)(p).resolves.toBe(1); });`},
			// Promise.try spread-wrapped assertion, destructured default param, computed-key await, and rethrowing Promise.reject chains.
			{Code: `test("t", async () => { await Promise.try((check: any) => check, ...[expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { await Promise.try((check: any, { a = check }: any = {}) => a, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => ({ [await check]() {} }, 1), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { throw err; }).catch((err) => err); });`},
			{Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { await e; } });`},
			{Code: `test("t", async () => { try { await Promise.reject(expect(p).resolves.toBe(1)); } catch (e) { await e; } });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => (check ||= fallback), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.resolve(...[expect(p).resolves.toBe(1)]); });`},
			{Code: `test("t", async () => { await Promise.resolve(...[], expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((...[first, second]: any[]) => second, ready, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { await e; } finally { return; } });`},
			{Code: `test("t", async () => { try { try { const check = expect(p).resolves.toBe(1); await Promise.reject(Promise.reject(check)); } catch (e) { await e; } } catch (e2) { await e2; } });`},
			{Code: `test("t", async () => { await Promise.try((...args: any[]) => args[1], ready, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch((...args: any[]) => args[0]); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(2)).catch((e) => Promise.reject(e)).catch((e) => e); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => { await Promise.reject(e); }).catch((e) => e); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => await Promise.reject(e)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(async (e) => { return await Promise.reject(e); }); });`},
			// Container aliases in `Promise.try` / `.catch`.
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch((err) => { const saved = [err]; return Promise.all(saved); }); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set().add(check)), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const queue = { add: (task: () => any) => Promise.resolve().then(task) }; await Promise.try((check: any) => queue.add(() => check), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]).then(([first]) => first), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [check]; return saved[0]; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; await Promise.all(saved); saved.length = 0; }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved: any[] = []; saved.length = 0; saved.push(check); await Promise.all(saved); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const jobs = new Map([['key', check]]); return Promise.all(jobs.values()); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(Array.from({ 0: check, length: 1 })), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => Promise.all({ *[Symbol.iterator]() { yield check; } }), expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await Promise.try((check: any) => { const a = [check]; let b = a; b = []; b.length = 0; return Promise.all(a); }, expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { const dispatch: (p: Promise<unknown>) => void = p => p; await Promise.try((check: any) => dispatch(check), expect(p).resolves.toBe(1)); });`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "expect().toBe(2);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs"},
				},
			},
			{
				Code: "expect().toBe(true);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs", Column: 7, EndColumn: 8},
				},
			},
			{
				Code: "expect().toEqual(\"something\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs", Column: 7, EndColumn: 8},
				},
			},
			{
				Code: "expect(\"something\", \"else\").toEqual(\"something\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs", Column: 21, EndColumn: 26},
				},
			},
			{
				Code:    "expect(\"something\", \"else\", \"entirely\").toEqual(\"something\");",
				Options: map[string]interface{}{"maxArgs": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs", Column: 29, EndColumn: 38},
				},
			},
			{
				Code:    "expect(\"something\", \"else\", \"entirely\").toEqual(\"something\");",
				Options: map[string]interface{}{"maxArgs": 2, "minArgs": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs", Column: 29, EndColumn: 38},
				},
			},
			{
				Code:    "expect(\"something\", \"else\", \"entirely\").toEqual(\"something\");",
				Options: map[string]interface{}{"maxArgs": 2, "minArgs": 1},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs", Column: 29, EndColumn: 38},
				},
			},
			{
				Code:    "expect(\"something\").toEqual(\"something\");",
				Options: map[string]interface{}{"minArgs": 2},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs", Column: 7, EndColumn: 8},
				},
			},
			{
				Code:    "expect(\"something\", \"else\").toEqual(\"something\");",
				Options: map[string]interface{}{"maxArgs": 1, "minArgs": 3},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs", Column: 7, EndColumn: 8},
					{MessageId: "tooManyArgs", Column: 21, EndColumn: 26},
				},
			},
			{
				Code: "expect(\"something\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 1, EndColumn: 20},
				},
			},
			{
				Code: "expect();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 1, EndColumn: 9},
				},
			},
			{
				Code: "expect(true).toBeDefined;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 14, EndColumn: 25},
				},
			},
			{
				Code: "expect(true).not.toBeDefined;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 18, EndColumn: 29},
				},
			},
			{
				Code: "expect(true).nope.toBeDefined;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 19, EndColumn: 30},
				},
			},
			{
				Code: "expect(true).nope.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown", Column: 1, EndColumn: 32},
				},
			},
			{
				Code: "expect(true).not.resolves.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown", Column: 1, EndColumn: 40},
				},
			},
			{
				Code: "expect(true).not.not.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown", Column: 1, EndColumn: 35},
				},
			},
			{
				Code: "expect(1).not.not.each();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown", Column: 1, EndColumn: 25},
				},
			},
			{
				Code: "expect`value`();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 1, EndColumn: 16},
				},
			},
			{
				Code: "expect(true).resolves.not.exactly.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown", Column: 1, EndColumn: 48},
				},
			},
			{
				Code: "expect(true).resolves;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 22},
				},
			},
			{
				Code: "expect(true).rejects;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 21},
				},
			},
			{
				Code: "expect(true).not;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 17},
				},
			},
			{
				Code: "expect(Promise.resolve(2)).resolves.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 1, EndColumn: 50},
				},
			},
			{
				Code: "expect(Promise.resolve(2)).rejects.toBeDefined();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 1, EndColumn: 49},
				},
			},
			{
				Code:    "expect(Promise.resolve(2)).resolves.toBeDefined();",
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 1, EndColumn: 50},
				},
			},
			{
				Code:    "test(\"valid-expect\", () => { return (expect(Promise.resolve(2)).resolves.toBeDefined()); });",
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{"test(\"valid-expect\", async () => { await (expect(Promise.resolve(2)).resolves.toBeDefined()); });"},
			},
			{
				Code:    "test(\"early return\", (ready) => { if (ready) return expect(Promise.resolve(1)).resolves.toBe(1) as Promise<void>; throw new Error('fell through'); });",
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{"test(\"early return\", async (ready) => { if (ready) return await (expect(Promise.resolve(1)).resolves.toBe(1) as Promise<void>); throw new Error('fell through'); });"},
			},
			{
				Code: "test(\"valid-expect\", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1)] as const); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{"test(\"valid-expect\", async () => { await Promise.resolve([await expect(Promise.resolve(1)).resolves.toBe(1)] as const); });"},
			},
			{
				Code: "test(\"valid-expect\", async () => { const [check] = await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1)] as const); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test(\"valid-expect\", async () => { await Promise.try(() => {}, expect(Promise.resolve(1)).resolves.toBe(1)); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test('starts the operation', async () => { let started = false; const p = Promise.resolve().then(() => started); const finished = Promise.withResolvers<void>(); await Promise.try(() => { started = true; }, expect(p).resolves.toBe(true).then(finished.resolve, finished.reject)); await finished.promise; });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test('starts the operation', async () => { let started = false; const p = Promise.resolve().then(() => started); const method = 'resolve'; await Promise[method]([expect(p).resolves.toBe(true), (started = true)]); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test('t', async () => { await Promise.try((first: any, second: any) => first, ready, expect(p).resolves.toBe(1)); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test('t', async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch(() => {}); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test(\"valid-expect\", async () => { const resolve = 'resolve'; await Promise[resolve]([expect(Promise.resolve(1)).resolves.toBe(1)]); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{"test(\"valid-expect\", async () => { const resolve = 'resolve'; await Promise[resolve]([await expect(Promise.resolve(1)).resolves.toBe(1)]); });"},
			},
			{
				Code: "test(\"valid-expect\", () => { return (expect(Promise.resolve(1)).resolves.toBe(1) as unknown) ? Promise.resolve() : Promise.reject(); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "test(\"valid-expect\", () => { return (expect(Promise.resolve(1)).resolves.toBe(1).then(() => false) as unknown) ? onReady() : onError(); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "import { createRequire } from 'node:module'; const require = createRequire(import.meta.url); const { expect: check } = require('@jest/globals'); check(1);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : expect(obj).resolves.not.toThrow();
  }
});
      `,
				Output: []string{`expect.extend({
  async toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : await expect(obj).resolves.not.toThrow();
  }
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 9, EndColumn: 43},
				},
			},
			{
				Code: `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});
      `,
				Output: []string{`expect.extend({
  async toResolve(obj) {
    this.isNot
      ? await expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(true);
  }
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 9, EndColumn: 43},
				},
			},
			{
				Code: `expect.extend({
  toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});
      `,
				Output: []string{`expect.extend({
  async toResolve(obj) {
    this.isNot
      ? expect(obj).toBe(true)
      : anotherCondition
      ? await expect(obj).resolves.not.toThrow()
      : expect(obj).toBe(false)
  }
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 9, EndColumn: 43},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30, EndColumn: 79},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).toResolve(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).toResolve(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 1, Column: 30},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).toResolve(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).toResolve(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 1, Column: 30},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).toReject(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).toReject(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 1, Column: 30},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).not.toReject(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).not.toReject(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 1, Column: 30},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30, EndColumn: 83},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).rejects.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).rejects.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30, EndColumn: 78},
				},
			},
			{
				Code:   "test(\"valid-expect\", () => { expect(Promise.resolve(2)).rejects.not.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).rejects.not.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30, EndColumn: 82},
				},
			},
			{
				Code:   "test(\"valid-expect\", async () => { expect(Promise.resolve(2)).resolves.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).resolves.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 36, EndColumn: 85},
				},
			},
			{
				Code:   "test(\"valid-expect\", async () => { expect(Promise.resolve(2)).resolves.not.toBeDefined(); });",
				Output: []string{"test(\"valid-expect\", async () => { await expect(Promise.resolve(2)).resolves.not.toBeDefined(); });"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 36, EndColumn: 89},
				},
			},
			{
				Code:    "test(\"valid-expect\", () => { expect(Promise.reject(2)).toRejectWith(2); });",
				Output:  []string{"test(\"valid-expect\", async () => { await expect(Promise.reject(2)).toRejectWith(2); });"},
				Options: map[string]interface{}{"asyncMatchers": []interface{}{"toRejectWith"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30},
				},
			},
			{
				Code:    "test(\"valid-expect\", () => { expect(Promise.reject(2)).rejects.toBe(2); });",
				Output:  []string{"test(\"valid-expect\", async () => { await expect(Promise.reject(2)).rejects.toBe(2); });"},
				Options: map[string]interface{}{"asyncMatchers": []interface{}{"toRejectWith"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 30},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 2, Column: 3, EndColumn: 56},
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 3, EndColumn: 51},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 3, EndColumn: 51},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `},
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 2, Column: 3, EndColumn: 56},
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 10, EndColumn: 58},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 2, Column: 3, EndColumn: 56},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  return expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).resolves.not.toBeDefined();
  await expect(Promise.resolve(1)).rejects.toBeDefined();
});
      `},
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 10, EndColumn: 58},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  await expect(Promise.resolve(2)).toResolve();
  return expect(Promise.resolve(1)).toReject();
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await expect(Promise.resolve(2)).toResolve();
  await expect(Promise.resolve(1)).toReject();
});
      `},
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 10},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndColumn: 73},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndColumn: 72},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.reject(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndColumn: 72},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.x(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndColumn: 67},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.resolve(expect(Promise.resolve(2)).resolves.not.toBeDefined());
});
      `},
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndColumn: 73},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.all([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.all([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndLine: 5, EndColumn: 5},
				},
			},
			{
				Code:   `test("valid-expect", () => { Promise.all([(expect(Promise.resolve(2)).resolves.toBeDefined()), expect(Promise.resolve(3)).resolves.toBeDefined()]); });`,
				Output: []string{`test("valid-expect", async () => { await Promise.all([(expect(Promise.resolve(2)).resolves.toBeDefined()), expect(Promise.resolve(3)).resolves.toBeDefined()]); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code:   `test("valid-expect", () => { Promise.all(([expect(Promise.resolve(2)).resolves.toBeDefined(), expect(Promise.resolve(3)).resolves.toBeDefined()])); });`,
				Output: []string{`test("valid-expect", async () => { await Promise.all(([expect(Promise.resolve(2)).resolves.toBeDefined(), expect(Promise.resolve(3)).resolves.toBeDefined()])); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.x([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.x([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited", Line: 2, Column: 3, EndLine: 5, EndColumn: 5},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.all([foo(expect(Promise.resolve(2)).resolves.not.toBeDefined())]);
});
      `,
				Output: []string{`test("valid-expect", async () => {
  Promise.all([foo(await expect(Promise.resolve(2)).resolves.not.toBeDefined())]);
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 2, Column: 20, EndColumn: 73},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  await Promise.all([foo(expect(Promise.resolve(2)).resolves.not.toBeDefined())]);
});
      `,
				Output: []string{`test("valid-expect", async () => {
  await Promise.all([foo(await expect(Promise.resolve(2)).resolves.not.toBeDefined())]);
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 2, Column: 26, EndColumn: 79},
				},
			},
			{
				Code: `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});
      `,
				Output: []string{`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5, EndLine: 3, EndColumn: 58},
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5, EndLine: 4, EndColumn: 58},
				},
			},
			{
				Code: `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).toResolve(),
    expect(Promise.resolve(3)).toReject(),
  ]
});
      `,
				Output: []string{`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).toResolve(),
    await expect(Promise.resolve(3)).toReject(),
  ]
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5},
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5},
				},
			},
			{
				Code: `test("valid-expect", () => {
  const assertions = [
    expect(Promise.resolve(2)).not.toResolve(),
    expect(Promise.resolve(3)).resolves.toReject(),
  ]
});
      `,
				Output: []string{`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).not.toResolve(),
    await expect(Promise.resolve(3)).resolves.toReject(),
  ]
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5},
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5},
				},
			},
			{
				Code: "expect(Promise.resolve(2)).resolves.toBe;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 37, EndColumn: 41},
				},
			},
			{
				Code: `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(() => {
    expect(Promise.resolve(2)).resolves.toBe(1);
  });
});
      `,
				Output: []string{`test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
  });
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5, EndLine: 3, EndColumn: 48},
				},
			},
			{
				Code: `test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
    expect(Promise.resolve(4)).resolves.toBe(4);
  });
});
      `,
				Output: []string{`test("valid-expect", () => {
  return expect(functionReturningAPromise()).resolves.toEqual(1).then(async () => {
    await expect(Promise.resolve(2)).resolves.toBe(1);
    await expect(Promise.resolve(4)).resolves.toBe(4);
  });
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5, EndLine: 4, EndColumn: 48},
				},
			},
			{
				Code: `test("valid-expect", async () => {
  await expect(Promise.resolve(1));
});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 9, EndColumn: 35},
				},
			},
			{
				Code: "expect(true).assertions;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 14, EndColumn: 24},
				},
			},
			{
				Code: "expect(1)[0];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 11, EndColumn: 12},
				},
			},
			{
				Code: "expect(1)[0].foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled", Column: 14, EndColumn: 17},
				},
			},
			{
				Code: "expect(1)[0].not;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound", Column: 14, EndColumn: 17},
				},
			},
			{
				Code: "import { expect as pleaseExpect } from '@jest/globals'; pleaseExpect(Promise.resolve(2)).resolves.toBe(2);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Column: 57, EndColumn: 106},
				},
			},
			{
				Code: `test("valid-expect", () => {
  Promise.all.x([
    expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `,
				Output: []string{`test("valid-expect", async () => {
  Promise.all.x([
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5, EndColumn: 58},
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5, EndColumn: 58},
				},
			},
			{
				Code: `test("t", () => { expect(p).resolves[('toBe')](1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves[('toBe')](1); });`},
			},
			{
				Code: `test("t", () => { (expect(p).resolves.toBe)(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe)(1); });`},
			},
			{
				Code: `test("t", () => { (expect)(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect)(p).resolves.toBe(1); });`},
			},
			{
				Code: `expect(1)[resolves];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `test("t", () => { Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });`},
			},
			{
				Code: `class C { constructor() { expect(p).resolves.toBe(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `export function check(p: Promise<number>) { expect(p).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`export async function check(p: Promise<number>) { await expect(p).resolves.toBe(1); }`},
			},
			{
				Code: `test("t", () => { if (true && expect(p).resolves.toBe(1)) {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { expect(p).resolves.toBe(1).then(() => {}).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.toBe(1).then(() => {}).catch(() => {}); });`},
			},
			{
				Code: `test("t", () => { expect(p).resolves.toBe(1)(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await expect(p).resolves.toBe(1)(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
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
			// Resolved Promise alias preserves `return await` under `alwaysAwait: true`.
			{
				Code:    `const key = "Promise"; const P = globalThis[key]; test("t", () => { return P.all([expect(p).resolves.toBe(1), fetchData()]); });`,
				Options: map[string]interface{}{"alwaysAwait": true},
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
				Code: `test("t", () => { [expect(p).resolves.toBe(1)][0].then(() => {}); });`,
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
			// Non-consuming callback references, type-only references, unused defaults, and spread/leading-spread Promise.try cases.
			{
				Code: `test("t", async () => { await Promise.try((first: any, second: any) => { console.log(second); return first; }, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (first: any, second: any) => (second, await first), ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => void check, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((first: any, second: any, unused = second) => first, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((first: any, second: any) => (first as typeof second) || 1, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const xs: unknown[] = []; await Promise.try((_ignored: any, ...rest: any[]) => Promise.all(rest), ...xs, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(...[], expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(...[], () => {}, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Swallowed Promise.reject variants (aliased, comma/ternary, try/catch, allSettled, any, non-callable pass-through, throwing finally).
			{
				Code: `test("t", async () => { const rejected = Promise.reject(expect(p).resolves.toBe(1)); await rejected.catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await (0, Promise.reject(check)).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
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
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.any([Promise.reject(check), Promise.resolve('ok')]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).then(undefined, void 0).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).then(...[], undefined, () => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).finally(() => { throw new Error('override'); }).catch((err) => err); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Unsafe autofix withheld for sibling calls, constructors, spreads, comma LHS, object literals, and multi-declarators.
			{
				Code: `test("t", async () => { let started = false; const start = () => { started = true; }; await run(expect(p).resolves.toBe(true), start()); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; await Promise.resolve([expect(p).resolves.toBe(true)], (started = true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; (expect(p).resolves.toBe(true), (started = true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const obj = { a: expect(p).resolves.toBe(true), b: (started = true) }; await obj.a; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Post-fix adversarial regressions: Promise.reject inside allSettled/try, indirect/let/shorthand/duplicate callees, returned containers, property access on tracked Promise, statementAlwaysThrows, cbUnknown, alwaysAwait/finally in try, rest/destructuring/default/spread tracking, nested rejectDepth, destructuring yield, type annotations.
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
				Code: `test("t", async () => { let started = false; (() => { started = true; }).call(null, expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const { try: pTry } = Promise; await pTry.call(Promise, () => { started = true; }, expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function getChecks(p: Promise<number>) { return [expect(p).resolves.toBe(1)]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err: any) => err.message); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((...args: any[]) => args.length, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { if (cond) return; throw err; }).catch((e) => e); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch((err) => { throw new Error(String(err)); }).catch((e) => e); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const xs: any[] = []; await Promise.try(...xs, () => {}, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code:    `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { return e; } });`,
				Options: map[string]interface{}{"alwaysAwait": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).catch(({ message }: any) => message); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((...[first, second]: any[]) => first, ready, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any, a = check) => a, expect(p).resolves.toBe(1), 42); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => ({ ...check }), expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; await Promise.resolve([expect(p).resolves.toBe(true), expect(() => { started = true; }).not.toThrow()]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { const check: Promise<void> = expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(Promise.reject(check)).catch((err) => err).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const check = expect(p).resolves.toBe(1); await Promise.reject(check).then(undefined, {} as any).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function* gen() { const [check, other = yield 1] = [expect(p).resolves.toBe(1)]; await check; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Final sanity regressions: wrapped/chained Promise.reject & try-finally autofix withholding, typed assignment/mutated callees/destructured params/thenable/multi-decl, try-catch-finally depth, rest index mismatch, and omitted/non-callable/spread Promise.try.
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
				Code: `test("t", async () => { let started = false; const api = { all(x: any) { return x; } }; api.all = () => { started = true; }; await api.all([expect(p).resolves.toBe(true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; function run(x: any) { return x; } run = () => { started = true; }; await run(expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const run = ({ a = (started = true) }: any) => a; await run(expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const helper = { then(x: any) { started = true; } }; helper.then(expect(p).resolves.toBe(true)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const a = expect(p).resolves.toBe(true), { x = (started = true) } = {}; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { throw new Error(); } catch { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } finally { return; } });`,
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
				Code: `test("t", async () => { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { return e; } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { try { const check = expect(p).resolves.toBe(1); await Promise.reject(check); } catch (e) { return e; } } finally { return; } });`,
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
				Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch((...args: any[]) => args[1]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(...[, () => {}], expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(null as any, expect(p).resolves.toBe(1)); });`,
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
			// Sign-off audit regressions: parameter dereference/throw/typed-param/typed-let autofix withholding, nested try inside catch, non-empty default object in Promise.try, `.catch((e) => Promise.reject(e))`, and dynamic `Promise[key]`.
			{
				Code: `test("t", async () => { const run = (p: any) => p.then; run(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const run = (p: any) => { if (!p) throw 'missing'; }; run(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const run = (p: Promise<void>) => p; run(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `async function helper(check: Promise<void>) { check = expect(Promise.resolve(1)).resolves.toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let check = Promise.resolve(); check = expect(Promise.resolve(1)).resolves.toBe(1); });`,
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
				Code: `test("t", async () => { try { await Promise.reject(expect(p).resolves.toBe(2)).catch((e) => Promise.reject(e)); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `declare const key: string; test("t", async () => { await (Promise as any)[key](expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Verification audit regressions: for-of loop var, contextually typed / cast callees, body destructuring, async/aliased/Promise.try re-rejections, and dynamic Promise[key].call.
			{
				Code: `test("t", async () => { const promises: Promise<void>[] = [Promise.resolve()]; for (let q of promises) { q = expect(Promise.resolve(1)).resolves.toBe(1); } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const record: (x: Promise<void>) => void = (x) => {}; record(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const obj: { record(x: Promise<void>): void } = { record(x) {} }; obj.record(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { (((x: any) => {}) as (x: Promise<void>) => void)(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { function check(x: any) { const { then } = x; } check(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { Promise.reject(expect(Promise.resolve(1)).resolves.toBe(1)).catch(async (e) => { await Promise.reject(e); }); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.reject(expect(Promise.resolve(1)).resolves.toBe(1)).catch(async (e) => { await Promise.reject(e); }); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.reject(expect(Promise.resolve(1)).resolves.toBe(1)).catch((e) => { const q = Promise.reject(e); return q; }); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { try { await Promise.try((q) => Promise.reject(q), expect(Promise.resolve(1)).resolves.toBe(1)); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const key = getKey(); (Promise as any)[key].call(Promise, () => { started = true; }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `const obj = { f(x) {} } as { f(x: Promise<void>): void }; test("t", () => { obj.f(expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Getter/spread sibling autofix withholding and `Promise.resolve([p])` container false negatives.
			{
				Code: `test("t", async () => { let started = false; const setup = { get start() { started = true; return 1; } }; await Promise.resolve([expect(Promise.resolve(true)).resolves.toBe(true), setup.start]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve([expect(Promise.resolve(true)).resolves.toBe(true), ...items]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve([check]), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const setup = { get start() { started = true; return true; } }; await Promise.resolve([expect(Promise.resolve(true)).resolves.toBe(true), expect(setup).toHaveProperty('start', true)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved = [Promise.resolve(0), check]; return saved[0]; }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const saved: any[] = []; saved.push(check); saved.pop(); return Promise.all(saved); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { let saved: any[] = []; return saved.push(check); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.resolve(new Set([check])).then((items) => items.clear()), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
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
				Code: `expect.extend({ toResolve() { return Promise.resolve({ pass: true, message: () => '' }); } }); test("t", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), expect(Promise.resolve(1)).toResolve()]); });`,
				Output: []string{
					`expect.extend({ toResolve() { return Promise.resolve({ pass: true, message: () => '' }); } }); test("t", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), await expect(Promise.resolve(1)).toResolve()]); });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { async function run(pending: Promise<unknown>[]) { await Promise.try((check: any) => pending.push(check), expect(Promise.resolve(1)).resolves.toBe(1)); } await run([]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set().add([check])), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const tasks: Array<() => unknown> = []; const queue = { add(task: () => unknown) { tasks.push(task); } }; await Promise.try((check: any) => queue.add(() => check), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: "expect[`extend`]({ toResolve() { return Promise.resolve({ pass: true, message: () => '' }); } }); test(\"t\", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), expect(Promise.resolve(1)).toResolve()]); });",
				Output: []string{
					"expect[`extend`]({ toResolve() { return Promise.resolve({ pass: true, message: () => '' }); } }); test(\"t\", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), await expect(Promise.resolve(1)).toResolve()]); });",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(new Set([[check]])), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; saved.length = 0; await Promise.all(saved); saved.push(check); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try(async (check: any) => { const saved = [check]; const alias = saved; alias.length = 0; await Promise.all(saved); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; (globalThis as any).startWork = () => { started = true; }; const p = Promise.resolve().then(() => started); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(p).resolves.toBe(true).then(done.resolve, done.reject), expect(1).toResolve()]); await done.promise; });`,
				Output: []string{
					`test("t", async () => { let started = false; (globalThis as any).startWork = () => { started = true; }; const p = Promise.resolve().then(() => started); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(p).resolves.toBe(true).then(done.resolve, done.reject), await expect(1).toResolve()]); await done.promise; });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), expect(Promise.resolve(1)).toResolve()]); });`,
				Output: []string{
					`test("t", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1), await expect(Promise.resolve(1)).toResolve()]); });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const a: any[] = []; const done = Promise.all(a); a.push(check); a.pop(); return done; }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const a = [check, Promise.resolve(0)]; return a.pop(); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const jobs = new Map([[check, 1]]); return Promise.all(jobs.values()); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const assertions = [expect(Promise.resolve(1)).resolves.toBe(1), expect(Promise.resolve(2)).toResolve()]; await done; });`,
				Output: []string{
					`test("t", async () => { const assertions = [expect(Promise.resolve(1)).resolves.toBe(1), await expect(Promise.resolve(2)).toResolve()]; await done; });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; { const assertions = [expect(Promise.resolve(2)).resolves.toBe(2), expect(Promise.resolve(3)).toReject()]; } expect(started).toBe(true); });`,
				Output: []string{
					`test("t", async () => { let started = false; { const assertions = [expect(Promise.resolve(2)).resolves.toBe(2), await expect(Promise.resolve(3)).toReject()]; } expect(started).toBe(true); });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.extend({ toStart() { gate.resolve(1); return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), expect(Promise.resolve(1)).resolves.toStart()]); await done.promise; });`,
				Output: []string{
					`test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); expect.extend({ toStart() { gate.resolve(1); return { pass: true, message: () => '' }; } }); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), await expect(Promise.resolve(1)).resolves.toStart()]); await done.promise; });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), expect(Promise.resolve(() => { gate.resolve(1); throw new Error('boom'); })).resolves.toThrow('boom')]); await done.promise; });`,
				Output: []string{
					`test("t", async () => { const gate = Promise.withResolvers<number>(); const done = Promise.withResolvers<void>(); await Promise.resolve([expect(gate.promise).resolves.toBe(1).then(done.resolve, done.reject), await expect(Promise.resolve(() => { gate.resolve(1); throw new Error('boom'); })).resolves.toThrow('boom')]); await done.promise; });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { const queue: Promise<unknown>[] = []; const enqueue: (p: Promise<unknown>) => void = p => queue.push(p); await Promise.try((check: any) => enqueue(check), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => { const jobs = new Map([['job', check]]); jobs.set('job', Promise.resolve(0)); return Promise.all(jobs.values()); }, expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all({ *[Symbol.iterator]() { return check; } }), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all(Array.from({ 0: check, length: 1, *[Symbol.iterator]() { yield 0; } })), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
		},
	)
}

func TestJestValidExpectSourceOnlyMode(t *testing.T) {
	root := fixtures.GetRootDir()
	code := `async function run(pending: Promise<unknown>[]) { const p = Promise.resolve(1); await Promise.try(check => pending.push(check), expect(p).resolves.toBe(1)); }`
	fileName := tspath.ResolvePath(root.Dir, "jest-source-only-typed-array.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
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
			Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			},
		},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("reported %d diagnostics; want 1", len(diagnostics))
	}
}
