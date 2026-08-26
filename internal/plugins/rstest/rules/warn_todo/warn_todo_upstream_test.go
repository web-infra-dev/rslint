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
// Three divergences from upstream, all of them deliberate; the cases that show
// them live in warn_todo_extras_test.go, except where noted.
//
//  1. Rstest's TestOptions has no `todo` field (rstest 2d7652e6
//     packages/core/src/types/api.ts:53-78), so `todo` written there is an
//     ordinary unknown property with no effect on the runner. Two upstream
//     cases move accordingly, both of them in this file: the invalid
//     `test("foo", { todo: true }, fn)` case becomes valid, and the already
//     valid `{ todo: false }` case stays valid for the same reason.
//  2. A computed `todo` accessor is reported. Upstream matches the member only
//     when it is an identifier (src/rules/warn-todo.ts), so `test['todo']("x")`
//     and its template-literal spelling go unreported there. Both register the
//     same todo, so both are reported here.
//  3. A modifier chained after `.todo` keeps the registration. Upstream
//     validates the whole chain against a fixed allowlist
//     (src/utils/valid-vitest-fn-call-chains.ts), which holds no entry for
//     `todo.runIf`, `todo.skipIf` or `todo.todo`, so those calls parse as
//     nothing at all upstream. Rstest's `test` and `describe` return a
//     chainable API, so each of those forms still registers a todo and is
//     reported.
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
