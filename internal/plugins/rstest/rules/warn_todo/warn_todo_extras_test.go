package warn_todo_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/warn_todo"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func extrasError(line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "warnTodo",
		Message:   warnTodoMessage,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

// N/A: fix boundaries and edit demand — the rule reports only, with no autofix
// and no suggestion, so there is no builder to exercise.
//
// N/A: malformed and recovery-parsed sources — the rule tester refuses a
// fixture with syntactic errors, so an unterminated literal or an unclosed call
// cannot be expressed as a case here.
//
// N/A: private identifiers — `#todo` is only legal inside a class body, where
// no Rstest registration head can be reached.
//
// N/A: expect shapes and TestContext expect — the rule never parses an expect
// call; its whole target set is `test` / `describe` registrations.
func TestWarnTodoExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&warn_todo.WarnTodoRule,
		[]rule_tester.ValidTestCase{
			// Registrations without `.todo`.
			{Code: `test("case", () => {}); it("case", () => {}); describe("suite", () => {});`},
			{Code: `test.skip("case", () => {}); test.only("case", () => {}); test.fails("case", () => {});`},
			{Code: `test.concurrent("case", () => {}); test.sequential("case", () => {});`},
			{Code: `test.runIf(cond)("case", () => {}); test.skipIf(cond)("case", () => {});`},
			{Code: `test.each([1])("case", () => {}); test.for([1])("case", () => {});`},
			{Code: "test.each`value\n${1}`(\"case\", () => {});"},
			{Code: `describe.skip("suite", () => {}); describe.each([1])("suite", () => {});`},
			{Code: `test("case", { timeout: 100 }, () => {}); test("case", () => {}, 100);`},
			// Rstest's TestOptions has no `todo` field, so an options object
			// never registers a todo.
			{Code: `test("case", { todo: true }, () => {});`},
			{Code: `describe("suite", { todo: true }, () => {});`},
			// Hooks take no modifiers.
			{Code: `beforeEach(() => {}); beforeAll(() => {}); afterEach(() => {}); afterAll(() => {});`},
			// `.todo` that is never called is not a registration.
			{Code: `const pending = test.todo;`},
			{Code: `console.log(test.todo);`},
			// Non-Rstest owners of the same member name.
			{Code: `import { test } from "vitest"; test.todo("case");`},
			{Code: `import { it } from "node:test"; it.todo("case");`},
			{Code: `import { test } from "@jest/globals"; test.todo("case");`},
			{Code: `import { customTest } from "./test-utils"; customTest.todo("case");`},
			{Code: `todo("case");`},
			{Code: `board.todo("case");`},
			// @rstest/playwright does not export `it`.
			{Code: `import { it } from "@rstest/playwright"; it.todo("case");`},
			// Local shadows.
			{Code: `const test = createRunner(); test.todo("case");`},
			{Code: `function run(test) { test.todo("case"); }`},
			{Code: `{ let describe = makeSuite(); describe.todo("suite"); }`},
			// A computed member the parser cannot read as `todo`.
			{Code: `test[name]("case", () => {});`},
			{Code: "test[`to${suffix}`](\"case\", () => {});"},
			{Code: `test[0]("case", () => {});`},
			// A declaration without an initializer must not be read as a require.
			{Code: `declare const pending: any; pending("case");`},
		},
		[]rule_tester.InvalidTestCase{
			// Dimension 4: transparent wrappers and optional chaining.
			{
				Code:   `(test).todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 8, 12)},
			},
			{
				Code:   `((test)).todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 10, 14)},
			},
			{
				Code:   `test?.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 7, 11)},
			},
			{
				Code:   `test.todo?.("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			{
				Code:   `test!.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 7, 11)},
			},
			{
				Code:   `(test as typeof test).todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 23, 27)},
			},
			{
				Code:   `(test satisfies typeof test).todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 30, 34)},
			},
			{
				Code:   `(<typeof test>test).todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 21, 25)},
			},
			// A todo registration with no title is still a todo registration.
			{
				Code:   `test.todo();`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			// Accessor spellings.
			{
				Code:   `test['todo']("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 12)},
			},
			{
				Code:   "test[`todo`](\"case\");",
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 12)},
			},
			{
				Code:   `describe["todo"]("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 10, 16)},
			},
			// Chained modifiers around `.todo`.
			{
				Code:   `test.concurrent.todo("case"); test.todo.concurrent("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 17, 21), extrasError(1, 36, 40)},
			},
			{
				Code:   `test.todo.for([1])("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			{
				Code:   "test.todo.each`value\n${1}`(\"case\");",
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			{
				Code:   `describe.todo.for([1])("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 10, 14)},
			},
			// A chained `test`/`describe` stays chainable in Rstest, so a
			// conditional modifier or a second `.todo` after `.todo` is still a
			// todo registration.
			{
				Code:   `test.todo.runIf(cond)("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			{
				Code:   `test.todo.skipIf(cond)("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			{
				Code:   `describe.todo.runIf(cond)("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 10, 14)},
			},
			// Two `.todo` accessors register one todo, so one report is issued,
			// anchored on the first accessor.
			{
				Code:   `test.todo.todo("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 6, 10)},
			},
			// A member chain broken across lines, with a comment in it.
			{
				Code: `test
  // pending work
  .todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(3, 4, 8)},
			},
			// API sources.
			{
				Code: `import { test as check } from "@rstest/core";
check.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 7, 11)},
			},
			{
				Code: `const { describe: suite } = require("@rstest/core");
suite.todo("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 7, 11)},
			},
			{
				Code: `import * as rstest from "@rstest/core";
rstest.test.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 13, 17)},
			},
			{
				Code: `const rstest = require("@rstest/core");
rstest.describe.todo("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 17, 21)},
			},
			{
				Code:   `import.meta.rstest.test.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(1, 25, 29)},
			},
			{
				Code: `const { it } = import.meta.rstest;
it.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 4, 8)},
			},
			{
				Code: `import { describe } from "@rstest/playwright";
describe.todo("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 10, 14)},
			},
			// Same-file aliases. A plain alias keeps its own call-site accessor.
			{
				Code: `const check = test;
check.todo("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 7, 11)},
			},
			// An alias that already carries `.todo` has no call-site accessor,
			// so the diagnostic anchors on the head identifier instead.
			{
				Code: `const pending = test.todo;
pending("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 1, 8)},
			},
			{
				Code: `const base = test;
const pending = base.todo;
pending.each([1])("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(3, 1, 8)},
			},
			{
				Code: `const { todo: pending } = test;
pending("case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 1, 8)},
			},
			{
				Code: `const { todo } = test;
todo("another case");`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(2, 1, 5)},
			},
			{
				Code: `import * as rstest from "@rstest/core";
const pendingSuite = rstest.describe.todo;
pendingSuite("suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{extrasError(3, 1, 13)},
			},
		},
	)
}
