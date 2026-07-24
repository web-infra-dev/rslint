package no_commented_out_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_commented_out_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCommentedOutTests(t *testing.T) {
	invalid := func(code string) rule_tester.InvalidTestCase {
		return rule_tester.InvalidTestCase{
			Code: code,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentedTests", Line: 1, Column: 1},
			},
		}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_commented_out_tests.NoCommentedOutTestsRule,
		[]rule_tester.ValidTestCase{
			// Active Rstest calls are not comments.
			{Code: `test("foo", () => {})`},
			{Code: `it.skip("foo", () => {})`},
			{Code: `describe.only("foo", () => {})`},
			{Code: `it.only.fails("foo", () => {})`},
			{Code: `describe.only.concurrent("foo", () => {})`},
			{Code: "test.each`\nvalue\n${1}\n`(\"$value\", ({ value }) => {})"},
			{Code: `describe.each<Row>(rows)("foo", ({ value }) => {})`},

			// Jest aliases are not part of the Rstest API.
			{Code: `// fit("foo", () => {})`},
			{Code: `// xit("foo", () => {})`},
			{Code: `// xtest("foo", () => {})`},
			{Code: `// fdescribe("foo", () => {})`},
			{Code: `// xdescribe("foo", () => {})`},

			// Similar names and APIs outside the supported Rstest roots are ignored.
			{Code: `// suite("foo", () => {})`},
			{Code: `// myTest("foo", () => {})`},
			{Code: `// rstest.test("foo", () => {})`},
			{Code: `// beforeEach(() => {})`},
			{Code: `// testSomething()`},
			{Code: `// latest(items)`},
			{Code: "// test`not a parameterized Rstest API`"},
			{Code: `// test.only`},
			{Code: `// describe.concurrent`},
		},
		[]rule_tester.InvalidTestCase{
			// Direct test and suite calls.
			invalid(`// test("foo", () => {})`),
			invalid(`// it("foo", () => {})`),
			invalid(`// describe("foo", () => {})`),

			// Every Rstest test modifier.
			invalid(`// test.only("foo", () => {})`),
			invalid(`// test.skip("foo", () => {})`),
			invalid(`// test.todo("foo")`),
			invalid(`// test.fails("foo", () => {})`),
			invalid(`// test.concurrent("foo", () => {})`),
			invalid(`// test.sequential("foo", () => {})`),
			invalid(`// test.runIf(condition)("foo", () => {})`),
			invalid(`// test.skipIf(condition)("foo", () => {})`),

			// Every Rstest suite modifier.
			invalid(`// describe.only("foo", () => {})`),
			invalid(`// describe.skip("foo", () => {})`),
			invalid(`// describe.todo("foo")`),
			invalid(`// describe.concurrent("foo", () => {})`),
			invalid(`// describe.sequential("foo", () => {})`),
			invalid(`// describe.runIf(condition)("foo", () => {})`),
			invalid(`// describe.skipIf(condition)("foo", () => {})`),

			// Getter and conditional modifiers can be chained.
			invalid(`// it.only.fails("foo", () => {})`),
			invalid(`// test.skip.concurrent("foo", () => {})`),
			invalid(`// test.concurrent.only("foo", () => {})`),
			invalid(`// test.concurrent.runIf(condition)("foo", () => {})`),
			invalid(`// describe.only.concurrent("foo", () => {})`),
			invalid(`// describe.concurrent.only("foo", () => {})`),
			invalid(`// describe.concurrent.skipIf(condition)("foo", () => {})`),
			invalid(`// describe.skipIf(condition).concurrent("foo", () => {})`),

			// Array-based parameterized tests and suites.
			invalid(`// test.each(rows)("foo", ({ value }) => {})`),
			invalid(`// test.for(rows)("foo", ({ value }) => {})`),
			invalid(`// it.each(rows)("foo", ({ value }) => {})`),
			invalid(`// it.for(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.each(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.for(rows)("foo", ({ value }) => {})`),
			invalid(`// it.only.each(rows)("foo", ({ value }) => {})`),
			invalid(`// it.concurrent.for(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.only.each(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.skip.for(rows)("foo", ({ value }) => {})`),

			// Explicit type arguments are supported by each and for.
			invalid(`// test.each<Row>(rows)("foo", ({ value }) => {})`),
			invalid(`// test.for<Row>(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.each<Row>(rows)("foo", ({ value }) => {})`),
			invalid(`// describe.for<Row>(rows)("foo", ({ value }) => {})`),
			invalid(`// test.for<{ value: Map<string, number> }>(rows)("foo", ({ value }) => {})`),

			// Tagged-template parameterized tests and suites.
			invalid("// test.each`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// test.for`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// it.each`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// it.for`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// describe.each`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// describe.for`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// test.for<Row>`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// describe.each<Row>`value | expected`(\"foo\", ({ value }) => {})"),
			invalid("// it.concurrent.for<Row>`value | expected`(\"foo\", ({ value }) => {})"),

			// Property and bracket access can be chained.
			invalid(`// describe['skip']("foo", () => {})`),
			invalid(`// test["only"]["concurrent"]("foo", () => {})`),
			invalid(`// describe['only'].each(rows)("foo", ({ value }) => {})`),
			invalid("// test[\"for\"]`value | expected`(\"foo\", ({ value }) => {})"),

			// Block comments may contain multiline calls and chains.
			invalid("/*\n  describe(\"foo\", () => {})\n*/"),
			invalid("/*\n  describe\n    .only\n    .concurrent(\"foo\", () => {})\n*/"),
			invalid("/*\n  test.for<Row>`\n    value\n    ${1}\n  `(\"$value\", ({ value }) => {})\n*/"),
		},
	)
}
