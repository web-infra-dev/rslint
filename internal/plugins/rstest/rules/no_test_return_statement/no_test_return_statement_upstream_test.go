// Package no_test_return_statement_test migrates the upstream suites for
// rstest/no-test-return-statement. Rstest-specific behavior lives in
// no_test_return_statement_extras_test.go.
//
// The direct counterpart is vitest-dev/eslint-plugin-vitest@v1.6.27
// tests/no-test-return-statement.test.ts, kept verbatim. The
// jest-community/eslint-plugin-jest@v29.16.0 suite adds the parameterized
// cases below; its remaining cases duplicate the Vitest ones.
//
// Dispositions that are not a plain Keep:
//
//   - `it.each()(...)` and `it.only.each()(...)` pass no cases. Rstest throws
//     while registering such a test, so the cases are adapted to
//     `it.each([1])(...)`, the smallest table that registers one.
package no_test_return_statement_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_test_return_statement"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noTestReturnStatement(line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noTestReturnStatement",
		Message:   "Return statements are not allowed in tests",
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

func TestNoTestReturnStatementUpstream(t *testing.T) {
	runNoTestReturnStatement(
		t,
		[]rule_tester.ValidTestCase{
			{Code: `it("noop", function () {});`},
			{Code: `test("noop", () => {});`},
			{Code: `test("one", () => expect(1).toBe(1));`},
			{Code: `test("empty")`},
			{Code: `it("one", myTest);
    function myTest() {
      expect(1).toBe(1);
    }`},
			{Code: `it("one", () => expect(1).toBe(1));
       function myHelper() {}`},
			// eslint-plugin-jest
			{Code: `test("one", () => {
  expect(1).toBe(1);
});`},
			{Code: `it("one", function () {
  expect(1).toBe(1);
});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("one", () => {
      return expect(1).toBe(1);
       });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 7, 32)},
			},
			{
				Code: `it("one", function () {
      return expect(1).toBe(1);
       });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 7, 32)},
			},
			{
				Code: `it.skip("one", function () {
      return expect(1).toBe(1);
       });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 7, 32)},
			},
			{
				Code: `it("one", myTest);
     function myTest () {
       return expect(1).toBe(1);
     }`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(3, 8, 33)},
			},
			// eslint-plugin-jest
			{
				Code:   "it.each``(\"one\", function () {\n  return expect(1).toBe(1);\n});",
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 3, 28)},
			},
			{
				Code: `it.each([1])("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 3, 28)},
			},
			{
				Code:   "it.only.each``(\"one\", function () {\n  return expect(1).toBe(1);\n});",
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 3, 28)},
			},
			{
				Code: `it.only.each([1])("one", function () {
  return expect(1).toBe(1);
});`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 3, 28)},
			},
		},
	)
}
