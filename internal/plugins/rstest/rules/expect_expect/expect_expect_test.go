package expect_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/expect_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExpectExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&expect_expect.ExpectExpectRule,
		[]rule_tester.ValidTestCase{
			// Basic assertions.
			{Code: `test("case", () => { expect(value).toBe(1); });`},
			{Code: `it("case", () => { expect(value).toBe(1); });`},
			{Code: `test("case", () => expect(value).toBe(1));`},
			// Rstest global `assert` counts by default (assertFunctionNames includes assert).
			{Code: `test("case", () => { assert.equal(value, 1); });`},
			// Assertion in a promise callback.
			{Code: `it("case", () => somePromise().then(() => expect(true).toBeDefined()));`},

			// `.todo` has no callback and is exempt.
			{Code: `test.todo("later");`},
			{Code: `it.todo("later");`},
			// Alias `.todo` regression: the final call-site Members are empty, so the
			// exemption must come from the semantic Todo field, not Members.
			{Code: `const todoTest = test.todo;
todoTest("later");`},

			// Named callback by reference.
			{Code: `it("case", run); function run() { expect(true).toBeDefined(); }`},
			{Code: `function run() { expect(true).toBeDefined(); } test("case", run);`},
			{Code: `const run = () => { expect(true).toBeDefined(); }; test("case", run);`},
			{Code: `test("case", run, TIMEOUT); function run() { expect(true).toBeDefined(); }`},
			{Code: `test("case", { timeout: 100 }, run); function run() { expect(true).toBeDefined(); }`},
			{Code: `it("first", shared);
test("second", shared);
function shared() { expect(true).toBeDefined(); }`},

			// Provenance: named import, alias, require, namespace, import.meta.rstest.
			{Code: `import { test, expect } from "@rstest/core";
test("case", () => { expect(value).toBe(1); });`},
			{Code: `import { test as check, expect } from "@rstest/core";
check("case", () => { expect(value).toBe(1); });`},
			{Code: `const { test, expect } = require("@rstest/core");
test("case", () => { expect(value).toBe(1); });`},
			{Code: `if (import.meta.rstest) {
  const { test, expect } = import.meta.rstest;
  test("case", () => { expect(value).toBe(1); });
}`},

			// Expect forms the callee-text patterns cannot match: the analysis
			// resolves them, exactly as rstest/no-conditional-expect does.
			{Code: `test("case", context => { context.expect(value).toBe(1); });`},
			{Code: `test.for([{ enabled: true }])("case", (row, context) => { context.expect(row).toBeDefined(); });`},
			{Code: `import * as rstest from "@rstest/core";
rstest.test("case", () => { rstest.expect(value).toBe(1); });`},
			{Code: `import { test, expect as check } from "@rstest/core";
test("case", () => { check(value).toBe(1); });`},
			{Code: `if (import.meta.rstest) {
  import.meta.rstest.test("case", () => {
    import.meta.rstest.expect(value).toBe(1);
  });
}`},

			// Custom test / parameterized.
			{Code: `const appTest = test.extend({ user: async ({}, use) => use({}) });
appTest("case", ({ expect }) => { expect(value).toBe(1); });`},
			{Code: `test.each([[1]])("case", value => { expect(value).toBe(1); });`},
			{Code: `test.for([[1]])("case", (row, context) => { expect(row).toBe(1); });`},

			// describe / hooks are not test registrations.
			{Code: `describe("suite", () => {});`},
			{Code: `beforeEach(() => {});`},

			// Non-Rstest test is not required to assert.
			{Code: `import { test } from "vitest";
test("case", () => {});`},
			// Local shadow of test is not a Rstest test.
			{Code: `const test = () => {};
test("case", () => {});`},

			// Options.
			{
				Code:    `it("case", () => expectSaga(mySaga).returns());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expectSaga"}}},
			},
			{
				Code:    `test("case", () => request.get().foo().expect(200));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.**.expect"}}},
			},
			{
				Code: `theoretically("case", theories, theory => { expect(f(theory.input)).toBe(theory.expected); });`,
				Options: []interface{}{map[string]interface{}{
					"additionalTestBlockFunctions": []interface{}{"theoretically"},
				}},
			},
			{
				Code:    `foo.todo("eventual test");`,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"foo.todo"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `it("case", () => { doStuff(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			// Named callback with no assertion.
			{
				Code: `it("case", run); function run() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			// Aliased test with empty body.
			{
				Code: `const custom = test;
custom("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
				},
			},
			// Parameterized empty body. The reported node is the registration
			// callee `test.each([[1]])`, spanning columns 1-17.
			{
				Code: `test.each([[1]])("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// Custom test empty body.
			{
				Code: `const appTest = test.extend({});
appTest("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				},
			},
			// A member call named `expect` that the analysis does not resolve to a
			// Rstest expect stays unrecognized: the resolving hook only widens to
			// what the analysis can prove.
			{
				Code: `test("case", () => { agent.expect(200); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			// Assertion name does not match.
			{
				Code:    `test("case", () => { notExpect(value); });`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			// `expect` alone is not in the configured assertFunctionNames.
			{
				Code:    `it("case", () => expectSaga(mySaga).returns());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			// additionalTestBlockFunctions block with no assertion.
			{
				Code:    `theoretically("case", theories, theory => { const output = f(theory.input); });`,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"theoretically"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
		},
	)
}
