package prefer_todo_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_todo"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTodoRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_todo.PreferTodoRule,
		[]rule_tester.ValidTestCase{
			{Code: `test()`},
			{Code: `test.concurrent()`},
			{Code: `test.todo("i need to write this test");`},
			{Code: `test(obj)`},
			{Code: `test.concurrent(obj)`},
			{Code: `fit("foo")`},
			{Code: `fit.concurrent("foo")`},
			{Code: `xit("foo")`},
			{Code: `test("foo", 1)`},
			{Code: `test("stub", () => expect(1).toBe(1));`},
			{Code: `test.concurrent("stub", () => expect(1).toBe(1));`},
			{Code: `supportsDone && params.length < test.length
  ? done => test(...params, done)
  : () => test(...params);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `test("i need to write this test");`,
				Output: []string{`test.todo("i need to write this test");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unimplementedTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `test("i need to write this test",);`,
				Output: []string{`test.todo("i need to write this test",);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unimplementedTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   "test(`i need to write this test`);",
				Output: []string{"test.todo(`i need to write this test`);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unimplementedTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `it("foo", function () {})`,
				Output: []string{`it.todo("foo")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `it("foo", () => {})`,
				Output: []string{`it.todo("foo")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `test.skip("i need to write this test", () => {});`,
				Output: []string{`test.todo("i need to write this test");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `test.skip("i need to write this test", function() {});`,
				Output: []string{`test.todo("i need to write this test");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   `test["skip"]("i need to write this test", function() {});`,
				Output: []string{`test['todo']("i need to write this test");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code:   "test[`skip`](\"i need to write this test\", function() {});",
				Output: []string{`test['todo']("i need to write this test");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 1, Column: 1},
				},
			},
			{
				Code: `if (true) {
    test.skip("i need to write this test", () => {});
}`,
				Output: []string{`if (true) {
    test.todo("i need to write this test");
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emptyTest", Line: 2, Column: 5},
				},
			},
		},
	)
}
