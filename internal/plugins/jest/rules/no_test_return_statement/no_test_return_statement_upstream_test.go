package no_test_return_statement_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_return_statement"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noReturnValue(line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noReturnValue",
		Message:   "Jest tests should not return a value",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func runNoTestReturnStatement(
	t *testing.T,
	valid []rule_tester.ValidTestCase,
	invalid []rule_tester.InvalidTestCase,
) {
	t.Helper()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_test_return_statement.NoTestReturnStatementRule,
		valid,
		invalid,
	)
}

// TestNoTestReturnStatementUpstream migrates the full valid/invalid suite from
// jest-community/eslint-plugin-jest@v29.16.0
// src/rules/__tests__/no-test-return-statement.test.ts, plus the examples in
// docs/rules/no-test-return-statement.md. Upstream's dedent strips the common
// indentation, so each case is written at the same columns.
func TestNoTestReturnStatementUpstream(t *testing.T) {
	runNoTestReturnStatement(
		t,
		[]rule_tester.ValidTestCase{
			{Code: `it("noop", function () {});`},
			{Code: `test("noop", () => {});`},
			{Code: `test("one", () => expect(1).toBe(1));`},
			{Code: `test("empty")`},
			{Code: `test("one", () => {
  expect(1).toBe(1);
});`},
			{Code: `it("one", function () {
  expect(1).toBe(1);
});`},
			{Code: `it("one", myTest);
function myTest() {
  expect(1).toBe(1);
}`},
			{Code: `it("one", () => expect(1).toBe(1));
function myHelper() {}`},
			// docs/rules/no-test-return-statement.md
			{Code: `it('returning a promise', async () => {
  await new Promise(res => setTimeout(res, 100));
  expect(1).toBe(1);
});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("one", () => {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it.skip("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code:   "it.each``(\"one\", function () {\n  return expect(1).toBe(1);\n});",
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it.each()("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code:   "it.only.each``(\"one\", function () {\n  return expect(1).toBe(1);\n});",
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it.only.each()("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it("one", myTest);
function myTest () {
  return expect(1).toBe(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(3, 3, 28)},
			},
			// docs/rules/no-test-return-statement.md
			{
				Code: `test('return an expect', () => {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 28)},
			},
			{
				Code: `it('returning a promise', function () {
  return new Promise(res => setTimeout(res, 100)).then(() => expect(1).toBe(1));
});`,
				Errors: []rule_tester.InvalidTestCaseError{noReturnValue(2, 3, 81)},
			},
		},
	)
}
