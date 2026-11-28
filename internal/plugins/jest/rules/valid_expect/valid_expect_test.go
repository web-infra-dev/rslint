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
			// Non-identifier receiver chains are not Jest `expect(...)` calls.
			{Code: "foo().expect(1);"},
			{Code: "/a/.expect(1).toBe(2);"},
			{Code: "this.expect(1);"},
			{Code: "patterns[0].expect(1).toBe(2);"},
			{Code: "new Foo().expect(1).toBe(2);"},
			// Transparent TS assertions, Promise combinators, direct const awaits, and idiomatic Promise.try.
			{Code: "test(\"valid-expect\", () => { return expect(Promise.resolve(2)).resolves.toBeDefined() as Promise<void>; });"},
			{Code: "test(\"valid-expect\", async () => { await (expect(Promise.resolve(2)).resolves.toBeDefined() as Promise<void>); });"},
			{Code: "test(\"valid-expect\", async () => { await expect(p).resolves.toBe(1).finally(() => {}); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.all([expect(p).resolves.toBe(1)]).then(() => {}); });"},
			{Code: "test(\"valid-expect\", async () => { await Promise.try(() => expect(p).resolves.toBe(1)); });"},
			{Code: "test(\"valid-expect\", async () => { const check = expect(Promise.resolve(1)).resolves.toBe(1); await check; });"},
			{Code: "const Promise = globalThis.Promise; test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "import type { Promise } from './types'; test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "interface Promise<T> { custom(): void; } test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "export {}; namespace Promise { export type Extra = number; } test('t', async () => { await Promise.all([expect(p).resolves.toBe(1)]); });"},
			{Code: "test('t', async () => { await Promise.all([Promise.all([expect(p).resolves.toBe(1)])]); });"},
			{Code: "test('t', async () => { await Promise.resolve(Promise.all([expect(p).resolves.toBe(1)])); });"},
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
				Output: []string{
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});
      `,
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]
});
      `,
				},
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
				Output: []string{
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).toResolve(),
    expect(Promise.resolve(3)).toReject(),
  ]
});
      `,
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).toResolve(),
    await expect(Promise.resolve(3)).toReject(),
  ]
});
      `,
				},
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
				Output: []string{
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).not.toResolve(),
    expect(Promise.resolve(3)).resolves.toReject(),
  ]
});
      `,
					`test("valid-expect", async () => {
  const assertions = [
    await expect(Promise.resolve(2)).not.toResolve(),
    await expect(Promise.resolve(3)).resolves.toReject(),
  ]
});
      `,
				},
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
				Output: []string{
					`test("valid-expect", async () => {
  Promise.all.x([
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `,
					`test("valid-expect", async () => {
  Promise.all.x([
    await expect(Promise.resolve(2)).resolves.not.toBeDefined(),
    await expect(Promise.resolve(3)).resolves.not.toBeDefined(),
  ]);
});
      `,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited", Line: 3, Column: 5, EndColumn: 58},
					{MessageId: "asyncMustBeAwaited", Line: 4, Column: 5, EndColumn: 58},
				},
			},
			// Promise.resolve([...]), Promise.reject(...).catch(), and sibling-deadlock autofix withholding.
			{
				Code:   `test("t", async () => { await Promise.resolve([expect(Promise.resolve(1)).resolves.toBe(1)] as const); });`,
				Output: []string{`test("t", async () => { await Promise.resolve([await expect(Promise.resolve(1)).resolves.toBe(1)] as const); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.reject(expect(p).resolves.toBe(1)).catch(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { await Promise.try((check: any) => Promise.all([check]), expect(Promise.resolve(1)).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", async () => { let started = false; const setup = { get start() { started = true; return 1; } }; await Promise.resolve([expect(Promise.resolve(true)).resolves.toBe(true), setup.start]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
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
				Code: `test("t", async () => { const assertions = [expect(Promise.resolve(1)).resolves.toBe(1), expect(Promise.resolve(2)).toResolve()]; await done; });`,
				Output: []string{
					`test("t", async () => { const assertions = [expect(Promise.resolve(1)).resolves.toBe(1), await expect(Promise.resolve(2)).toResolve()]; await done; });`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			// Red-team regressions: non-combinator Promise methods, reject propagation, unsafe return/call/nested-destructuring autofixes.
			{
				Code: `test("t", async () => { await Promise.withResolvers(expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
			},
			{
				Code: `test("t", () => { Promise.withResolvers(expect(p).resolves.toBe(1)); });`,
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
		},
	)
}

func TestJestValidExpectSourceOnlyMode(t *testing.T) {
	root := fixtures.GetRootDir()
	code := `async function run(pending: Promise<unknown>[]) { const p = Promise.resolve(1); await Promise.try(check => pending.push(check), expect(p).resolves.toBe(1)); }`
	fileName := tspath.ResolvePath(root.Dir, "jest-source-only-floating.ts")
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
