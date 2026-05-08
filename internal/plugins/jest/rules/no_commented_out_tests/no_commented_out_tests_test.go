package no_commented_out_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCommentedOutTestsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_commented_out_tests.NoCommentedOutTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: `// foo("bar", function () {})`},
			{Code: `describe("foo", function () {})`},
			{Code: `it("foo", function () {})`},
			{Code: `describe.only("foo", function () {})`},
			{Code: `it.only("foo", function () {})`},
			{Code: `it.concurrent("foo", function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: `test.only("foo", function () {})`},
			{Code: `test.concurrent("foo", function () {})`},
			{Code: `var appliedSkip = describe.skip; appliedSkip.apply(describe)`},
			{Code: `var calledSkip = it.skip; calledSkip.call(it)`},
			{Code: `({ f: function () {} }).f()`},
			{Code: `(a || b).f()`},
			{Code: `itHappensToStartWithIt()`},
			{Code: `testSomething()`},
			{Code: `// latest(dates)`},
			{Code: `// TODO: unify with Git implementation from Shipit (?)`},
			{Code: `#!/usr/bin/env node`},
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
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `// describe("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// describe["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// describe['skip']("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it.skip("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it.only("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it.concurrent("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test.skip("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test.concurrent("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test["skip"]("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// xdescribe("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// xit("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// fit("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// xtest("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `
        // test(
        //   "foo", function () {}
        // )
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 9},
				},
			},
			{
				Code: `
        /* test
          (
            "foo", function () {}
          )
        */
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 9},
				},
			},
			{
				Code: `// it("has title but no callback")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// it()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test.someNewMethodThatMightBeAddedInTheFuture()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test["someNewMethodThatMightBeAddedInTheFuture"]()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `// test("has title but no callback")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: `
        foo()
        /*
          describe("has title but no callback", () => {})
        */
        bar()
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 3, Column: 9},
				},
			},
		},
	)
}
