package no_standalone_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_standalone_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoStandaloneExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_standalone_expect.NoStandaloneExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.any(String)`},
			{Code: `expect.extend({})`},
			{Code: `expect.not.stringContaining("value")`},
			{Code: `describe("suite", () => { test("case", () => { expect(1).toBe(1); }); });`},
			{Code: `describe("suite", () => { test("case", () => { const helper = () => { expect(1).toBe(1); }; }); });`},
			{Code: `describe("suite", () => { const helper = () => { expect(1).toBe(1); }; });`},
			{Code: `describe("suite", () => { function helper() { expect(1).toBe(1); } });`},
			{Code: `describe("suite", () => { const helper = function() { expect(1).toBe(1); }; });`},
			{Code: `test("case", () => expect(1).toBe(1));`},
			{Code: `const helper = function() { expect(1).toBe(1); };`},
			{Code: `const helper = () => expect(1).toBe(1);`},
			{Code: `function helper() { expect(1).toBe(1); }`},
			{Code: `const helpers = { assert() { expect(1).toBe(1); } };`},
			{Code: `class Helper { assert() { expect(1).toBe(1); } get value() { expect(1).toBe(1); return 1; } }`},
			{Code: `{}`},
			{Code: `test.each([1, true])("trues", value => { expect(value).toBe(true); });`},
			{Code: `test.for([1])("case", (value, context) => { context.expect(value).toBe(1); });`},
			{Code: `test.for([1])("case", value => { expect(value).toBe(1); });`},
			{Code: `const forCase = test.for([1]); forCase("case", (value, context) => { context.expect(value).toBe(1); });`},
			{Code: `test.only("only", () => { expect(1).toBe(1); });`},
			{Code: `test.concurrent("concurrent", () => { expect(1).toBe(1); });`},
			{Code: `describe.each([1])("suite", value => { test("case", () => expect(value).toBe(1)); });`},
			{Code: `test("case", ({ expect }) => { expect(1).toBe(1); });`},
			{Code: `test("case", context => { context.expect(1).toBe(1); });`},
			{Code: `test("case", () => { expect(1).toBe(1); }, 1000);`},
			{Code: `test("case", { timeout: 1 }, () => { expect(1).toBe(1); });`},
			{Code: `test("case", callback); function callback() { expect(1).toBe(1); }`},
			{Code: `import { test as t } from "@rstest/core"; t("case", () => { expect(1).toBe(1); });`},
			{Code: `import.meta.rstest.test("case", () => { import.meta.rstest.expect(1).toBe(1); });`},
			{Code: `import { test, expect } from "@rstest/playwright"; test("case", () => { expect(1).toBe(1); });`},
			{Code: `test("case", () => { expect(1).to.be.ok; });`},
			{Code: `const expect = () => ({ toBe() {} }); expect(1).toBe(1);`},
			{Code: `import { expect } from "vitest"; expect(1).toBe(1);`},
			{Code: `import { expect } from "@jest/globals"; expect(1).toBe(1);`},
			{Code: `import { test, expect } from "@playwright/test"; expect(1).toBe(1);`},
			{
				Code: `
describe("suite", () => {
  const t = Math.random() ? test.only : test;
  t("case", () => expect(true).toBe(true));
});`,
				Options: []any{map[string]any{"additionalTestBlockFunctions": []any{"t"}}},
			},
			{
				Code:    `each([[1, 2]]).test("case", (actual, expected) => { expect(actual).toBe(expected); });`,
				Options: []any{map[string]any{"additionalTestBlockFunctions": []any{"each.test"}}},
			},
			{
				Code: `
test.each` + "`" + `
  value
  ${true}
` + "`" + `("case", ({ value }) => {
  expect(value).toBe(true);
});`,
			},
			// A nested registration with no inline callback must not close the
			// scope of the test it sits in, whichever form leaves it without one.
			{Code: `test("outer", () => { test.todo("inner"); expect(1).toBe(1); });`},
			{Code: `test("outer", () => { test("inner", callback); expect(1).toBe(1); }); function callback() {}`},
			{Code: `describe("suite", () => { test("outer", () => { test.todo("inner"); expect(1).toBe(1); }); });`},
			{Code: `test.each([1])("outer", value => { test.todo("inner"); expect(value).toBe(1); });`},
			{Code: `test.for([1])("outer", value => { test.todo("inner"); expect(value).toBe(1); });`},
			{
				Code:    `test("outer", () => { t("inner"); expect(1).toBe(1); });`,
				Options: []any{map[string]any{"additionalTestBlockFunctions": []any{"t"}}},
			},
			// The fallback range covers the arguments of a registration that has
			// no callback, which is where upstream also allows an assertion.
			{Code: `test.todo(expect(1).toBe(1));`},
			// Same asymmetry in the arrow branch: an arrow passed to a call opens
			// no scope, so its exit must not close the enclosing arrow's.
			{Code: `const outer = () => foo(() => {}) || expect(1).toBe(1);`},
			// A tagged template that is not a known registration still returns the
			// call that registers the cases, so the callback of that call is inside
			// the template scope.
			{
				Code: `
myEach` + "`" + `
  value
  ${true}
` + "`" + `("case", ({ value }) => {
  expect(value).toBe(true);
});`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedExpect",
					Message:   "Expect must be inside of a test block",
					Line:      1,
					Column:    1,
					EndColumn: 18,
				}},
			},
			{
				Code:   `describe("suite", () => { expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `describe("suite", () => expect(1).toBe(1));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `describe("suite", () => { test("case", () => { expect(1).toBe(1); }); expect(2).toBe(2); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `{ expect(1).toBe(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `(() => {})("case", () => expect(true).toBe(false));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `expect.hasAssertions();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `expect().hasAssertions();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `expect.assertions(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `describe.each([1])("suite", value => { expect(value).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code: `
import { describe as d, expect as check } from "@rstest/core";
d("suite", () => { check(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `import.meta.rstest.expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `describe("suite", () => { expect(1).to.be.ok; });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code: `
import * as rstest from "@rstest/core";
rstest.describe("suite", () => { rstest.expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `import { expect } from "@rstest/core"; expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `const { expect: check } = require("@rstest/core"); check(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `const rstest = require("@rstest/core"); rstest.expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code: `
if (import.meta.rstest) {
  const { expect: check } = import.meta.rstest;
  check(1).toBe(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code: `
if (import.meta.rstest) {
  const api = import.meta.rstest;
  api.expect(1).toBe(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:   `import { expect } from "@rstest/playwright"; expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code: `
describe("suite", () => {
  const t = Math.random() ? test.only : test;
  t("case", () => expect(true).toBe(false));
});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
			{
				Code:    `each([[1, 2]]).test("case", (actual, expected) => { expect(actual).toBe(expected); });`,
				Options: []any{map[string]any{"additionalTestBlockFunctions": []any{"each"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExpect"}},
			},
		},
	)
}
