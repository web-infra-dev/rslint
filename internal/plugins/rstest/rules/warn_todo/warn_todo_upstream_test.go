package warn_todo_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/warn_todo"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const warnTodoMessage = "The use of `.todo` is not recommended."

func upstreamError(line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "warnTodo",
		Message:   warnTodoMessage,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

// Upstream: @vitest/eslint-plugin v1.6.27 tests/warn-todo.test.ts.
//
// Two upstream cases move because Rstest's TestOptions has no `todo` field
// (rstest 2d7652e6 packages/core/src/types/api.ts:53-78): the invalid
// `test("foo", { todo: true }, fn)` case becomes valid here, and the already
// valid `{ todo: false }` case stays valid for the same reason.
func TestWarnTodoUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&warn_todo.WarnTodoRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("foo", function () {})`},
			{Code: `it("foo", function () {})`},
			{Code: `it.concurrent("foo", function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: `test("foo", { todo: false }, function () {})`},
			{Code: `test.concurrent("foo", function () {})`},
			{Code: `describe.only("foo", function () {})`},
			{Code: `it.only("foo", function () {})`},
			{Code: `it.each()("foo", function () {})`},
			// Rstest divergence: an options object cannot carry `todo`.
			{Code: `test("foo", { todo: true }, function () {})`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `describe.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 10, 14)},
			},
			{
				Code:   `it.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 4, 8)},
			},
			{
				Code:   `test.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 6, 10)},
			},
			{
				Code:   `describe.todo.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 10, 14)},
			},
			{
				Code:   `it.todo.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 4, 8)},
			},
			{
				Code:   `test.todo.each([])("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 6, 10)},
			},
			{
				Code:   `describe.only.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 15, 19)},
			},
			{
				Code:   `it.only.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 9, 13)},
			},
			{
				Code:   `test.only.todo("foo", function () {})`,
				Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 11, 15)},
			},
		},
	)
}
