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

			// Type arguments are only recognized after each/for, so prose
			// containing angle brackets is not mistaken for a test call.
			{Code: `// test <input> (see docs)`},
			{Code: `// it <-> (x)`},
			{Code: `// describe <T> (generic explanation)`},

			// Factory calls do not register a test or suite until the returned
			// API/registrar is invoked.
			{Code: `// test.runIf(condition)`},
			{Code: `// test.skipIf(condition).only`},
			{Code: `// test.each(rows)`},
			{Code: "// test.for`value | expected`"},
			{Code: `// describe.each<Row>(rows)`},
			{Code: `// test.extend(fixtures)`},
			{Code: `// test.extend(fixtures).only`},

			// These chains are not part of the current Rstest API. In
			// particular, each/for return a plain registrar rather than an API
			// carrying trailing modifiers.
			{Code: `// test.each(rows).only("foo", () => {})`},
			{Code: `// describe.for(rows).concurrent("foo", () => {})`},
			{Code: `// test.only.extend(fixtures)("foo", () => {})`},
			{Code: `// describe.extend(fixtures)("foo", () => {})`},
			{Code: `// describe.fails("foo", () => {})`},
			{Code: `// test.unknown("foo", () => {})`},
			{Code: `// test[modifier]("foo", () => {})`},

			// Documentation and prose are not disabled tests.
			{Code: `/** test("documented example", () => {}) */`},
			{Code: "// ```ts\n// test(\"documented example\", () => {})\n// ```"},
			{Code: `// test("foo") should be preferred in examples`},
			{Code: "/* `test(\"foo\", () => {})` */"},
			{Code: `// test (see docs)`},
			{Code: `// test (foo bar)`},
			{Code: `// describe (documentation example)`},
			{
				Code:     `// test("generic", () => <T>(value: T) => value)`,
				FileName: "file.tsx",
			},
			{
				Code:     `// test("invalid JSX here", () => <div />)`,
				FileName: "file.ts",
			},
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
			invalid("/*\n * test(\"foo\", () => {})\n */"),
			invalid("/*\n\t* describe\n\t*   .only(\"foo\", () => {})\n */"),
			invalid("/*\n * test.each`\n * value | expected\n * ${1} | ${2}\n * `(\"foo\", () => {})\n */"),

			// Conditional factories only become registrations after the
			// returned API is invoked.
			invalid(`// test.runIf(condition).only("foo", () => {})`),
			invalid(`// describe.only.skipIf(condition)("foo", () => {})`),
			invalid(`// test["runIf"](condition)["concurrent"]("foo", () => {})`),

			// extend creates another test API. A bare extend call above is
			// valid, while invoking the returned API defines a test.
			invalid(`// test.extend(fixtures)("foo", () => {})`),
			invalid(`// it.extend<Fixtures>(fixtures).only("foo", () => {})`),
			invalid(`// test.extend(a).extend(b)("foo", () => {})`),
			invalid(`// test.extend(fixtures).each(rows)("foo", () => {})`),

			// Transparent syntax around a canonical Rstest root remains a
			// commented-out registration.
			invalid(`// ((test))("foo", () => {})`),
			invalid(`// test?.("foo", () => {})`),
			invalid(`// test?.only("foo", () => {})`),
			invalid("// test[`only`](\"foo\", () => {})"),
			invalid(`// ;describe("foo", () => {})`),

			// Consecutive line comments are reconstructed before parsing, so
			// chains and tagged templates may span physical comments.
			invalid("// test\n//   .only\n//   .concurrent(\"foo\", () => {})"),
			invalid("// test.each<Row>`\n// value | expected\n// ${1} | ${2}\n// `(\"foo\", () => {})"),
			invalid("// test.extend({\n//   value: 1,\n// }).only(\"foo\", () => {})"),
			invalid("// test\r//   .only(\"foo\", () => {})"),
			invalid("// test\u2028//   .only(\"foo\", () => {})"),
			invalid("// test\u2029//   .only(\"foo\", () => {})"),

			// Commented code is parsed with the original file's script kind.
			{
				Code:     `/* test("renders", () => <div />) */`,
				FileName: "file.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code:     `// test("generic", () => <T>(value: T) => value)`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: "// test(\"first\", () => {})\n// test(\"second\", () => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			invalid("/*\n * test(\"first\", () => {})\n * test(\"second\", () => {})\n */"),
			invalid(`// test("first", () => {}); cleanup()`),
			{
				Code: "// setup only\n// test(\"foo\", () => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			{
				Code: "// setup only\r\n// describe(\"foo\", () => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 2, Column: 1},
				},
			},
			{
				Code: "// test(\"valid\", () => {})\n// test (see docs)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
			{
				Code: "// test(\"valid\", () => {})\n// unrelated invalid prose here",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "commentedTests", Line: 1, Column: 1},
				},
			},
		},
	)
}
