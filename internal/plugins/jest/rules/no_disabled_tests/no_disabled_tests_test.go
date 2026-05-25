package no_disabled_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_disabled_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDisabledTestsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_disabled_tests.NoDisabledTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("foo", function () {})`},
			{Code: `it("foo", function () {})`},
			{Code: `describe.only("foo", function () {})`},
			{Code: `it.only("foo", function () {})`},
			{Code: `it.each("foo", () => {})`},
			{Code: `it.concurrent("foo", function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: `test.only("foo", function () {})`},
			{Code: `test.concurrent("foo", function () {})`},
			{Code: "describe[`${\"skip\"}`](\"foo\", function () {})"},
			{Code: `it.todo("fill this later")`},
			{Code: `var appliedSkip = describe.skip; appliedSkip.apply(describe)`},
			{Code: `var calledSkip = it.skip; calledSkip.call(it)`},
			{Code: `({ f: function () {} }).f()`},
			{Code: `(a || b).f()`},
			{Code: `itHappensToStartWithIt()`},
			{Code: `testSomething()`},
			{Code: `xitSomethingElse()`},
			{Code: `xitiViewMap()`},
			{Code: `
				import { pending } from "actions"

				test("foo", () => {
				  expect(pending()).toEqual({})
				})
			`},
			{Code: `
				const { pending } = require("actions")

				test("foo", () => {
				  expect(pending()).toEqual({})
				})
			`},
			{Code: `
				test("foo", () => {
				  const pending = getPending()
				  expect(pending()).toEqual({})
				})
			`},
			{Code: `
				test("foo", () => {
				  expect(pending()).toEqual({})
				})

				function pending() {
				  return {}
				}
			`},
			{
				Code: `
					import { test } from './test-utils';

					test('something');
				`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe.skip("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `describe.skip.each([1, 2, 3])("%s", (a, b) => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xdescribe.each([1, 2, 3])("%s", (a, b) => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: "describe[`skip`](\"foo\", function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `describe["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `it.skip("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `it["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `test.skip("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: "it.skip.each``(\"foo\", function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: "test.skip.each``(\"foo\", function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `it.skip.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `test.skip.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `test["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xdescribe("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xit("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xtest("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: "xit.each``(\"foo\", function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: "xtest.each``(\"foo\", function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xit.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `xtest.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `it("has title but no callback")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
				},
			},
			{
				Code: `test("has title but no callback")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 1, Column: 1},
				},
			},
			{
				Code: `it("contains a call to pending", function () { pending() })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 48},
				},
			},
			{
				Code: `pending();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `describe("contains a call to pending", function () { pending() })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "skippedTest", Line: 1, Column: 54},
				},
			},
			{
				Code: "import { test } from '@jest/globals';\n\ntest('something');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingFunction", Line: 3, Column: 1},
				},
			},
		},
	)
}
