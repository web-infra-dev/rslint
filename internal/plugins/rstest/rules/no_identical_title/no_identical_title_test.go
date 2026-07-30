package no_identical_title_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_identical_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIdenticalTitleRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_identical_title.NoIdenticalTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("one", () => {}); test("two", () => {});`},
			{Code: `describe("same", () => {}); test("same", () => {});`},
			{Code: `describe("outer", () => {
  test("same", () => {});
  describe("inner", () => {
    test("same", () => {});
  });
});`},
			{Code: `test("case " + name, () => {}); test("case " + name, () => {});`},
			{Code: "test(`case ${name}`, () => {}); test(`case ${name}`, () => {});"},
			{Code: `test.each([1, 2])("case %i", () => {});
test.each([3, 4])("case %i", () => {});
test.for([1, 2])("case %i", () => {});
test.for([3, 4])("case %i", () => {});`},
			{Code: "describe.each`value\n${1}`(\"case $value\", () => {});\ndescribe.for`value\n${1}`(\"case $value\", () => {});"},
			{Code: `test.for([1]).concurrent("same", () => {});
test.for([2]).concurrent("same", () => {});`},
			{Code: `describe.each([1]).only("same", () => {});
describe.each([2]).only("same", () => {});`},
			{Code: `fit("same", () => {}); fit("same", () => {});
xit("same", () => {}); xtest("same", () => {});
fdescribe("same", () => {}); xdescribe("same", () => {});`},
			{Code: `import { test } from "node:test";
test("same", () => {});
test("same", () => {});`},
			{Code: `import { test, describe } from "vitest";
test("same", () => {});
test("same", () => {});
describe("suite", () => {});
describe("suite", () => {});`},
			{Code: `const test = createRunner();
test("same", () => {});
test("same", () => {});`},
			{Code: `let customTest = test;
customTest("same", () => {});
customTest("same", () => {});`},
			{Code: `describe.fails("same", () => {});
describe.fails("same", () => {});`},
			{Code: `test.only.extend({})("same", () => {});
test.only.extend({})("same", () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("same", () => {});
test("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 6},
				},
			},
			{
				Code: `test("same", () => {});
it("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 4},
				},
			},
			{
				Code: `describe("same", () => {});
describe.only("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 2, Column: 15},
				},
			},
			{
				Code: `test.only.concurrent("same", () => {});
test.concurrent.only("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 22},
				},
			},
			{
				Code: `test.skip("same", () => {});
test.todo("same");
test.fails("same", () => {});
test.sequential("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 11},
					{MessageId: "multipleTestTitle", Line: 3, Column: 12},
					{MessageId: "multipleTestTitle", Line: 4, Column: 17},
				},
			},
			{
				Code: `test.runIf(condition)("same", () => {});
test.skipIf(condition).only("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 29},
				},
			},
			{
				Code: `describe.concurrent("same", () => {});
describe.sequential.runIf(condition)("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 2, Column: 38},
				},
			},
			{
				Code: `import { test as rstestTest, describe as rstestDescribe } from "@rstest/core";
rstestTest("same", () => {});
rstestTest.only("same", () => {});
rstestDescribe("suite", () => {});
rstestDescribe.skip("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 17},
					{MessageId: "multipleDescribeTitle", Line: 5, Column: 21},
				},
			},
			{
				Code: `const { test: rstestTest, describe: rstestDescribe } = require("@rstest/core");
rstestTest("same", () => {});
rstestTest("same", () => {});
rstestDescribe("suite", () => {});
rstestDescribe("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 12},
					{MessageId: "multipleDescribeTitle", Line: 5, Column: 16},
				},
			},
			{
				Code: `import * as rstest from "@rstest/core";
rstest.test("same", () => {});
rstest.test.only("same", () => {});
rstest.describe("suite", () => {});
rstest.describe.only("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 18},
					{MessageId: "multipleDescribeTitle", Line: 5, Column: 22},
				},
			},
			{
				Code: `const rstest = require("@rstest/core");
rstest.test("same", () => {});
rstest.test("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 13},
				},
			},
			{
				Code: `import { test } from "@rstest/core";
const alias = test;
const appTest = alias.extend({});
const adminTest = appTest.extend({});
test("same", () => {});
appTest.only("same", () => {});
adminTest.skipIf(condition)("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 6, Column: 14},
					{MessageId: "multipleTestTitle", Line: 7, Column: 29},
				},
			},
			{
				Code: `test.extend({})("same", () => {});
test.extend({}).only("same", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 22},
				},
			},
			{
				Code: `describe.each([1, 2])("case %i", () => {
  test("duplicate", () => {});
  test("duplicate", () => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 8},
				},
			},
			{
				Code: "test('same', () => {});\ntest(`same`, () => {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 6},
				},
			},
		},
	)
}
