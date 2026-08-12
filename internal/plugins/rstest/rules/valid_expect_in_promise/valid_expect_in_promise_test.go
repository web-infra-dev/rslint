package valid_expect_in_promise_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect_in_promise"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidExpectInPromiseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_expect_in_promise.ValidExpectInPromiseRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("case", async (context) => { await promise.then(value => context.expect(value).toBe(1)); });`},
			{Code: `test("case", (context) => promise.then(value => context.expect(value).toBe(1)));`},
			{Code: `test("case", () => { try { return promise.then(value => expect(value).toBe(1)); } catch (error) {} });`},
			{Code: `test("case", async () => { try { return promise.then(value => expect(value).toBe(1)); } catch (error) {} });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); await pending; });`},
			{Code: `test("case", () => { const pending = promise.then(value => assert.equal(value, 1)); expect(pending).resolves.toBeUndefined(); });`},
			{Code: `test("case", () => expect(promise.then(value => assert.equal(value, 1))).resolves.toBeUndefined());`},
			{Code: `test("case", () => condition ? promise.then(value => assert.equal(value, 1)) : Promise.resolve());`},
			{Code: `test("case", () => condition && promise.then(value => assert.equal(value, 1)));`},
			{Code: `test("case", () => { const [pending] = [promise.then(value => assert.equal(value, 1))]; return pending; });`},
			{Code: `test("case", () => { const { pending } = { pending: promise.then(value => assert.equal(value, 1)) }; return pending; });`},
			{Code: `test("case", () => { let pending; [pending] = [promise.then(value => assert.equal(value, 1))]; return pending; });`},
			{Code: `test("case", () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); return pending; });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } finally { cleanup(); } });`},
			{Code: `try { test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); await pending; }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.resolve(pending); }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([pending]); }); } catch (error) {}`},
			{Code: `import assert from "node:assert"; test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
			{Code: `import { assert } from "chai"; test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
			{Code: `const assert = createAssertionLibrary(); test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("case", (context) => { promise.then(value => context.expect(value).toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Message:   "This promise should either be returned or awaited to ensure the assertions in its chain are called",
					Line:      1,
					Column:    29,
				}},
			},
			{
				Code: `test.for([1])("case", (row, context) => { promise.then(value => context.expect(value).toBe(row)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    43,
				}},
			},
			{
				Code: `test("case", () => { promise.then(value => assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    22,
				}},
			},
			{
				Code: `import { assert, test } from "@rstest/core"; test("case", () => { promise.then(value => assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    67,
				}},
			},
			{
				Code: `import * as rstest from "@rstest/core"; rstest.test("case", () => { promise.then(value => rstest.assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    69,
				}},
			},
			{
				Code: `import.meta.rstest.test("case", () => { promise.then(value => import.meta.rstest.assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    41,
				}},
			},
			{
				Code: `const { assert, test } = import.meta.rstest; test("case", () => { promise.then(value => assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    67,
				}},
			},
			{
				Code: `const api = import.meta.rstest; api.test("case", () => { promise.then(value => api.assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    58,
				}},
			},
			{
				Code: `import { expect, test } from "@rstest/playwright"; test("case", async ({ page }) => { page.goto("url").then(() => expect(page).toBeDefined()); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    87,
				}},
			},
			{
				Code: `test("case", () => { promise.then(value => assert(value)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `import * as rstest from "@rstest/core"; rstest.test("case", () => { promise.then(value => rstest.assert(value)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { promise.then(value => expect(value).to.be.ok).then(value => assert.equal(value, 1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `function handler(value) { assert.equal(value, 1); } test("one", () => { promise.then(handler); }); test("two", () => { promise.then(handler); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectInFloatingPromise"},
					{MessageId: "expectInFloatingPromise"},
				},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => assert.equal(value, 1)); return Promise.all([[pending]]); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { let pending; [pending] = [promise.then(value => assert.equal(value, 1))]; log(pending); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); log(pending); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => assert.equal(value, 1)); return condition ? pending : other; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => assert.equal(value, 1)); return condition && pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
		},
	)
}
