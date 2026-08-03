package no_disabled_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDisabledTestsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_disabled_tests.NoDisabledTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("case", () => {}); it("case", () => {}); describe("suite", () => {});`},
			{Code: `test.only("case", () => {}); test.fails("case", () => {}); test.concurrent.sequential("case", () => {});`},
			{Code: `describe.only("suite", () => {}); describe.concurrent.sequential("suite", () => {});`},
			{Code: `test.todo("case"); it.todo("case"); describe.todo("suite");`},
			{Code: `test.todo.each([1])("case"); test.todo.for([1])("case");`},
			{Code: `const todoTest = test.todo; todoTest("case");`},
			{Code: `test.skipIf(condition)("case", () => {}); test.runIf(condition)("case", () => {});`},
			{Code: `describe.skipIf(condition)("suite", () => {}); describe.runIf(condition)("suite", () => {});`},
			{Code: `test.skipIf(true)("case", () => {}); test.runIf(false)("case", () => {});`},
			{Code: `test("case", context => { context.skip(); });`},
			{Code: `test("case", ctx => { ctx.skip(); }); request.skip();`},
			{Code: `describe("empty suite"); describe.concurrent("empty suite");`},
			// an identifier is not necessarily an options object
			{Code: `test("case", options);`},
			{Code: `test("case", { timeout: 100 }, () => {});`},
			{Code: `test.todo("case", { timeout: 100 });`},
			{Code: `fit("case", () => {}); xit("case", () => {}); xtest("case", () => {}); fdescribe("suite", () => {}); xdescribe("suite", () => {});`},
			{Code: `pending();`},
			{Code: `import { pending } from "actions"; pending();`},
			{Code: `test[` + "`sk${suffix}`" + `]("case", () => {});`},
			{Code: `import { test } from "node:test"; test.skip("case", () => {});`},
			{Code: `import { test, describe } from "vitest"; test.skip("case", () => {}); describe.skip("suite", () => {});`},
			{Code: `const test = createRunner(); test.skip("case", () => {});`},
			{Code: `let skipped = test.skip; skipped("case", () => {});`},
			// a variable declaration without an initializer must not be treated as a require
			{Code: `declare const skipped: any; skipped("case", () => {});`},
			{Code: `declare const pending: any; pending();`},
			{Code: `import { customTest } from "./test-utils"; customTest.skip("case", () => {});`},
			{Code: `test.for([1]).skip("case", () => {}); describe.each([1]).skip("suite", () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `it.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `describe.skip("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `test["skip"]("case", () => {}); it[` + "`skip`" + `]("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 33},
				},
			},
			{
				Code: `test.concurrent.skip("case", () => {}); test.skip.sequential("case", () => {}); test.fails.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 41},
					{MessageId: "skippedTest", Line: 1, Column: 81},
				},
			},
			{
				Code: `test.skip.only("case", () => {}); test.only.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 35},
				},
			},
			{
				Code: `describe.concurrent.skip("suite", () => {}); describe.skip.sequential("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 46},
				},
			},
			{
				Code: `test.skip.each([1])("case", () => {}); test.skip.for([1])("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 40},
				},
			},
			{
				Code: `describe.skip.each([1])("suite", () => {}); describe.skip.for([1])("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "skippedTest", Line: 1, Column: 45},
				},
			},
			{
				Code: "test.skip.each`value\n${1}`(\"case\", () => {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `import { test as rstestTest } from "@rstest/core";
rstestTest.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 2, Column: 1},
				},
			},
			{
				Code: `const { test: rstestTest } = require("@rstest/core");
rstestTest.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 2, Column: 1},
				},
			},
			{
				Code: `import * as rstest from "@rstest/core";
rstest.test.skip("case", () => {});
rstest.describe.skip("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 2, Column: 1},
					{MessageId: "skippedTest", Line: 3, Column: 1},
				},
			},
			{
				Code: `const rstest = require("@rstest/core");
rstest.test.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 2, Column: 1},
				},
			},
			{
				Code: `const skipped = test.skip;
skipped("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 2, Column: 1},
				},
			},
			{
				Code: `const base = test;
const skipped = base.skip;
skipped.each([1])("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 3, Column: 1},
				},
			},
			{
				Code: `import * as rstest from "@rstest/core";
const skippedSuite = rstest.describe.skip;
skippedSuite("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 3, Column: 1},
				},
			},
			{
				Code: `test.extend({}).skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `const appTest = test.extend({});
const adminTest = appTest.extend({});
adminTest.concurrent.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 3, Column: 1},
				},
			},
			{
				Code: `test("case"); it("case"); test.concurrent("case"); test.fails("case");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
					{MessageId: "missingFunction", Line: 1, Column: 15},
					{MessageId: "missingFunction", Line: 1, Column: 27},
					{MessageId: "missingFunction", Line: 1, Column: 52},
				},
			},
			{
				Code: `test.each([1])("case"); test.for([1])("case");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
					{MessageId: "missingFunction", Line: 1, Column: 25},
				},
			},
			{
				Code: `test("case", { timeout: 100 }); it("case", {}); test.concurrent("case", { retry: 2 });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
					{MessageId: "missingFunction", Line: 1, Column: 33},
					{MessageId: "missingFunction", Line: 1, Column: 49},
				},
			},
			{
				Code: `test.each([1])("case", { timeout: 100 });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
				},
			},
			{
				Code: `test.skip("case", { timeout: 100 });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "missingFunction", Line: 1, Column: 1},
				},
			},
			{
				Code: `const customTest = test.extend({});
customTest("case");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 2, Column: 1},
				},
			},
			{
				Code: `test.skip("case");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
					{MessageId: "missingFunction", Line: 1, Column: 1},
				},
			},
			{
				Code: `test.todo.skip("case");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
		},
	)
}
