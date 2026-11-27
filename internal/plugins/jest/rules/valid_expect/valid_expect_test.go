package valid_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
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
				Output: []string{"test(\"valid-expect\", async () => { await Promise.try(() => {}, await expect(Promise.resolve(1)).resolves.toBe(1)); });"},
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
				Output: []string{`test("t", async () => { await Promise.try(function (this: unknown) { return 1; }, await expect(p).resolves.toBe(1)); });`},
			},
			{
				Code: `test("t", async () => { await Promise.try([expect(p).resolves.toBe(1)] as any); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.try([await expect(p).resolves.toBe(1)] as any); });`},
			},
			{
				Code: `test("t", async () => { await Promise.try(function () { const o = { arguments: 1 }; return o; }, expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.try(function () { const o = { arguments: 1 }; return o; }, await expect(p).resolves.toBe(1)); });`},
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
		},
	)
}
