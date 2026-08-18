package require_local_test_context_for_concurrent_snapshots_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_local_test_context_for_concurrent_snapshots"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireLocalTestContextForConcurrentSnapshotsRule(t *testing.T) {
	expectedError := rule_tester.InvalidTestCaseError{
		MessageId: "requireLocalTestContext",
		Message:   "Use local Test Context instead",
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_local_test_context_for_concurrent_snapshots.RequireLocalTestContextForConcurrentSnapshotsRule,
		[]rule_tester.ValidTestCase{
			{Code: `it("x", () => expect(1).toMatchSnapshot());`},
			{Code: `it.concurrent("x", () => expect(true).toBe(true));`},
			{Code: `it.concurrent("x", ({ expect }) => expect(1).toMatchSnapshot());`},
			{Code: `it.concurrent("x", ({ expect }) => expect(1).toMatchInlineSnapshot("1"));`},
			{Code: `it.concurrent("x", ctx => ctx.expect(1).toMatchSnapshot());`},
			{Code: `describe.concurrent("s", () => it("x", ({ expect }) => expect(1).toMatchSnapshot()));`},
			{Code: `describe("s", () => it("x", context => context.expect(1).toMatchInlineSnapshot()));`},
			{Code: `describe("s", () => it("x", context => expect(1).toMatchInlineSnapshot()));`},
			{Code: `describe.concurrent("s", () => it.sequential("x", () => expect(1).toMatchSnapshot()));`},
			{Code: `test.concurrent("x", ({ expect: check }) => check(1).toMatchSnapshot());`},
			{Code: `test.concurrent.for(rows)("x", (row, ctx) => ctx.expect(row).toMatchSnapshot());`},
			{Code: `import { test, expect } from '@playwright/test'; test.concurrent("x", () => expect(1).toMatchSnapshot());`},
			{Code: `import { test, expect } from 'vitest'; test.concurrent("x", () => expect(1).toMatchSnapshot());`},
			{Code: `const expect = createAssertionLibrary(); test.concurrent("x", () => expect(1).toMatchSnapshot());`},
			// Declared outside every registration callback, so the rule cannot
			// prove it runs concurrently.
			{Code: `function helper() { expect(1).toMatchSnapshot(); } test.concurrent("x", () => helper());`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `it.concurrent("x", () => expect(true).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => expect(true).toMatchInlineSnapshot("true"));`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => expect(true).toMatchFileSnapshot("./o.html"));`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => expect(foo()).toThrowErrorMatchingSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => expect(foo()).toThrowErrorMatchingInlineSnapshot("bar"));`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `describe.concurrent("s", () => it("x", () => expect(true).toMatchSnapshot()));`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `describe.sequential("s", () => it.concurrent("x", () => expect(true).toMatchSnapshot()));`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", context => expect(true).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => { expect(1).toMatchSnapshot(); expect(2).toMatchSnapshot(); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError}},
			{Code: `const t = test.concurrent; t("x", () => expect(1).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `import.meta.rstest.test.concurrent("x", () => import.meta.rstest.expect(1).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `import { test, expect } from '@rstest/playwright'; test.concurrent("x", () => expect(1).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `test.concurrent.each(rows)("x", () => expect(1).toMatchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `test.concurrent("x", callback); function callback() { expect(1).toMatchSnapshot(); }`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `describe.concurrent("s", suite); function suite() { test("x", callback); } function callback() { expect(1).toMatchSnapshot(); }`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `test.sequential("a", callback); test.concurrent("b", callback); function callback() { expect(1).toMatchSnapshot(); }`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// Rstest supports this Chai alias even though the Vitest rule omits it.
			{Code: `test.concurrent("x", () => expect(1).matchSnapshot());`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// Assertions nested inside closures, hooks and helpers declared in a
			// concurrent body still run as part of the concurrent test.
			{Code: `test.concurrent("x", () => { const helper = () => expect(1).toMatchSnapshot(); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", () => { items.forEach(item => { expect(item).toMatchSnapshot(); }); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `it.concurrent("x", async () => { await Promise.all(items.map(async item => { expect(item).toMatchSnapshot(); })); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `describe.concurrent("s", () => { beforeEach(() => { expect(1).toMatchSnapshot(); }); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `describe.concurrent("s", () => { function helper() { expect(1).toMatchSnapshot(); } it("t", () => helper()); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// A test registered from inside a loop callback still inherits the
			// enclosing concurrent describe.
			{Code: `describe.concurrent("s", () => { rows.forEach(row => { it(row, () => { expect(row).toMatchSnapshot(); }); }); });`, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		},
	)
}
