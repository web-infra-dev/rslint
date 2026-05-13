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
				Code: "// test(\n//   \"foo\", function () {}\n// )\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: "/* test\n  (\n    \"foo\", function () {}\n  )\n*/\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
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
				Code: "foo()\n/*\n  describe(\"has title but no callback\", () => {})\n*/\nbar()\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			// Multi-byte (CJK) and surrogate-pair (emoji) text before the comment: AST
			// positions are UTF-8 byte offsets; LSP/ESLint columns stay UTF-16 elsewhere.
			{
				Code: "const 中文 = 1;\n// describe(\"x\", () => {})\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			{
				Code: "const e = \"🚀\";\n// test(\"x\", () => {})\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			{
				Code: "const 中文 = 1;\n/* test(\"x\", () => {}) */\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			{
				Code: "const e = \"🚀\";\n/* describe() */\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			// Same-line comments: assert 1-based UTF-16 columns (IDE/ESLint), which
			// differ from naive UTF-8 byte columns when CJK or surrogate pairs precede.
			{
				Code: `const 中文 = 1; // describe("a", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 15},
				},
			},
			{
				Code: `const e = "🚀"; // test("x", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 17},
				},
			},
			{
				Code: `const 中文 = 1; /* test("x") */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 15},
				},
			},
			{
				Code: `const e = "🚀"; /* describe() */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 17},
				},
			},
		},
	)
}
