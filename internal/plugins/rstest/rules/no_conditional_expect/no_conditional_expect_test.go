package no_conditional_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_conditional_expect.NoConditionalExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("case", () => { expect(value).toBe(1); });`},
			{Code: `test("case", () => {
  if (condition) setup();
  expect(value).toBe(1);
});`},
			{Code: `test("case", () => {
  const expected = condition ? 1 : 2;
  expect(value).toBe(expected);
});`},
			{Code: `test("case", () => {
  try { run(); } catch { recover(); } finally {
    expect(value).toBe(1);
  }
});`},
			{Code: `describe("suite", () => {
  if (condition) expect(value).toBe(1);
});`},
			{Code: `beforeEach(() => {
  if (condition) expect(value).toBe(1);
});`},
			{Code: `import { expect } from "vitest";
test("case", () => { if (condition) expect(value).toBe(1); });`},
			{Code: `import { expect } from "@jest/globals";
test("case", () => { if (condition) expect(value).toBe(1); });`},
			{Code: `import { test, expect } from "@playwright/test";
test("case", () => { if (condition) expect(value).toBe(1); });`},
			{Code: `const expect = createAssertionLibrary();
test("case", () => { if (condition) expect(value); });`},
			{Code: `const custom = { expect };
test("case", () => { if (condition) custom.expect(value).toBe(1); });`},
			{Code: `test.each([[1]])("case", value => {
  if (value) custom.expect(value);
});`},
			{Code: `test.each([{ expect: customExpect }])("case", ({ expect }) => {
  if (condition) expect(value);
});`},
			{Code: `import { expect, test } from "@rstest/playwright";
test.describe("suite", () => {
  if (condition) expect(value).toBe(1);
});`},
			{Code: `import { expect, test } from "@rstest/playwright";
test.beforeEach(() => {
  if (condition) expect(value).toBe(1);
});`},
			// A declaration without an initializer must not reach SkipParentheses
			// while the expect root is being resolved.
			{Code: `let logger;
export function run() { logger(); }`},
			// Same, for a callback argument that resolves to an uninitialized
			// declaration.
			{Code: `let cb; cb = () => {}; test("case", cb);`},
			// An unresolvable callback name starts the pending walk, which visits
			// every declaration in the file including uninitialized ones.
			{Code: `test("case", callback);
let state;`},
		},
		[]rule_tester.InvalidTestCase{
			// `import.meta.rstest` is typed optional, so `?.` is the idiomatic
			// access and must not disable recognition of either the test call or
			// the expect call.
			{
				Code: `import.meta.rstest.test("case", () => {
  if (condition) import.meta.rstest?.expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import.meta.rstest?.test("case", () => {
  if (condition) import.meta.rstest?.expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", () => {
  if (condition) expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect", Line: 2, Column: 18},
				},
			},
			{
				Code: `it("case", () => {
  condition && expect(value).toBe(1);
  condition || expect(other).toBe(2);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect", Line: 2, Column: 16},
					{MessageId: "conditionalExpect", Line: 3, Column: 16},
				},
			},
			{
				Code: `test("case", () => {
  condition ? expect(value).toBe(1) : expect(value).toBe(2);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", async () => {
  try { await request(); } catch (error) {
    expect(error).toBeInstanceOf(Error);
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", async () => {
  await request().catch(error => expect(error).toBeInstanceOf(Error));
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				// An unresolvable timeout in the third position must not stop the
				// second argument from being treated as the callback.
				Code: `test("case", ({ expect }) => { if (condition) expect(value).toBe(1); }, TIMEOUT);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const TIMEOUT = 5000;
test("case", ({ expect }) => { if (condition) expect(value).toBe(1); }, TIMEOUT);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				// The `(name, options, fn)` shape still resolves the third argument
				// through the pending walk.
				Code: `test("case", { timeout: 1 }, handler);
function handler({ expect }) { if (condition) expect(value).toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import { expect as check, test } from "@rstest/core";
test("case", () => { if (condition) check(value).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const { expect: check, test } = require("@rstest/core");
test("case", () => { if (condition) check(value).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import * as rstest from "@rstest/core";
rstest.test("case", () => {
  if (condition) rstest.expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const rstest = require("@rstest/core");
rstest.test("case", () => {
  if (condition) rstest["expect"](value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `if (import.meta.rstest) {
  import.meta.rstest.test("case", () => {
    if (condition) import.meta.rstest.expect(value).toBe(1);
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `if (import.meta.rstest) {
  const { test, expect: check } = import.meta.rstest;
  test("case", () => { if (condition) check(value).toBe(1); });
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `if (import.meta.rstest) {
  const api = import.meta.rstest;
  api.test("case", () => { if (condition) api.expect(value).toBe(1); });
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", context => {
  if (condition) context.expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", ({ expect: check }) => {
  if (condition) check(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test.for([{ enabled: true }])("case", (row, context) => {
  if (row.enabled) context.expect(row).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", { timeout: 100 }, ({ expect }) => {
  if (condition) expect(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const appTest = test.extend({});
appTest.concurrent("case", ({ expect }) => {
  if (condition) expect.soft(value).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", async () => {
  if (condition) await expect.poll(readValue).toBe(1);
  if (other) await expect.element(locator).toBeVisible();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", () => {
  if (condition) {
    expect.assertions(1);
    expect.hasAssertions();
    expect.unreachable();
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import { expect, test } from "@rstest/playwright";
test.fail("case", async ({ page }) => {
  if (condition) await expect.soft(page).toHaveTitle("Dashboard");
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import { expect as check, test as playwrightTest } from "@rstest/playwright";
playwrightTest.extend({}).for([{ enabled: true }])("case", (row, context) => {
  if (row.enabled) check(context).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `import * as playwright from "@rstest/playwright";
playwright.test("case", async ({ page }) => {
  if (condition) await playwright.expect(page).toHaveTitle("Dashboard");
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const forCase = test.for([{ enabled: true }]);
forCase("case", (row, context) => {
  if (row.enabled) context.expect(row).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				// Leaving the inner catch must not clear the outer one.
				Code: `test("case", async () => {
  await a().catch(() => {
    b().catch(() => { expect(1).toBe(1); });
    expect(2).toBe(2);
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", () => {
  if (condition) log(expect(value).toBe(1));
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", () => {
  if (condition) expect(value).toBe(expect.stringContaining("x"));
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `test("case", () => {
  if (condition) results[expect(value).toBe(1)] = 1;
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `function callback() {
  if (condition) expect(value).toBe(1);
}
test("case", callback);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
			{
				Code: `const callback = () => {
  if (condition) expect(value).toBe(1);
};
test("case", { retry: 1 }, callback);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "conditionalExpect"},
				},
			},
		},
	)
}
