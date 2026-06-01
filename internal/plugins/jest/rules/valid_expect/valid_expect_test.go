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
		},
	)
}
