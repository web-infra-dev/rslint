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
			// A logical assignment stores the chain, only conditionally, so the
			// binding is real and the value it may leave in place survives.
			{Code: `test("case", async () => { let pending; pending ||= promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending; pending ??= promise.then(value => assert.equal(value, 1)); await pending; });`},
			{Code: `test("case", async () => { let pending = other; pending &&= promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending; await (pending ||= promise.then(value => expect(value).toBe(1))); });`},
			{Code: `test("case", async () => { let pending = promise.then(value => expect(value).toBe(1)); pending ||= other; await pending; });`},
			// Every operator that carries the chain into the binding binds it,
			// the same way the value walks up to a direct await.
			{Code: `test("case", async () => { let pending = fallback || promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending = fallback ?? promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending = ready && promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending = condition ? promise.then(value => expect(value).toBe(1)) : other; await pending; });`},
			{Code: `test("case", async () => { let pending = (setup(), promise.then(value => expect(value).toBe(1))); await pending; });`},
			{Code: `test("case", async () => { let pending = other; pending = pending || promise.then(value => expect(value).toBe(1)); await pending; });`},
			// The longhand of a preserving logical assignment keeps the promise
			// the binding already holds, because a promise is truthy.
			{Code: `test("case", async () => { let pending = promise.then(value => assert.equal(value, 1)); pending = pending || other; await pending; });`},
			// An array in a binding is consumed element-wise.
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await Promise.all(pending); });`},
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; for await (const settled of pending) {} });`},
			{Code: `test("case", async () => { for await (const settled of [promise.then(value => expect(value).toBe(1))]) {} });`},
			// A spread carries every element of the binding into the literal,
			// and an element access reaches one of them, so both keep an
			// element-wise consumption of the binding.
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await Promise.all([...pending]); });`},
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await Promise.all([...pending, other()]); });`},
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; for await (const settled of [...pending]) {} });`},
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await pending[0]; });`},
			// Every chain in the literal binds to the same identifier, and one
			// element-wise consumption of the binding consumes all of them.
			{Code: `test("case", async () => { const pending = [first.then(value => expect(value).toBe(1)), second.then(value => expect(value).toBe(2))]; await Promise.all(pending); });`},
			{Code: `test("case", async () => { const pending = [first.then(value => expect(value).toBe(1)), second.then(value => expect(value).toBe(2))]; for await (const settled of pending) {} });`},
			{Code: `test("case", async () => { const pending = [first.then(value => expect(value).toBe(1)), second.then(value => expect(value).toBe(2))]; await Promise.all([...pending]); });`},
			{Code: `test("case", async () => { const pending = [first.then(value => expect(value).toBe(1)), second.then(value => expect(value).toBe(2)), third.then(value => expect(value).toBe(3))]; await Promise.all(pending); });`},
			// A `for await` over a literal holding the binding consumes it, just
			// as `Promise.all` in that position does.
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); for await (const settled of [pending]) {} });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); if (!ready) { throw new Error("no"); } await pending; });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { setup(); await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { throw new Error("no"); } finally { cleanup(); } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { await pending; } finally { cleanup(); } });`},
			{Code: `try { test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); await pending; }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.resolve(pending); }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([pending]); }); } catch (error) {}`},
			{Code: `import assert from "node:assert"; test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
			{Code: `import { assert } from "chai"; test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
			{Code: `const assert = createAssertionLibrary(); test("case", () => { promise.then(value => assert.equal(value, 1)); });`},
			{Code: `function handler() { queue.then(handler); } test("case", () => promise.then(handler));`},
			{Code: `function first() { queue.then(second); } function second() { queue.then(first); } test("case", () => promise.then(first));`},
			{Code: `function handler(value) { expect(value).toBe(1); return queue.then(handler); } test("case", () => promise.then(handler));`},
			{Code: `let handler; test("case", () => promise.then(handler));`},
			{Code: `declare const handler: () => void; test("case", () => promise.then(handler));`},
			{Code: `let handler: (value: unknown) => void; beforeEach(() => { handler = value => expect(value).toBe(1); }); test("case", () => promise.then(handler));`},
		},
		[]rule_tester.InvalidTestCase{
			// A key behind parentheses is a chain the identifier table cannot
			// answer, so it also covers the source prescan's fallback.
			{
				Code: `test("case", () => { promise[("then")](value => expect(value).toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Message:   "This promise should either be returned or awaited to ensure the assertions in its chain are called",
					Line:      1,
					Column:    22,
				}},
			},
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
				Code: `function leaf(value) { assert.equal(value, 1); } function shared() { return queue.then(leaf); } test("one", () => { promise.then(shared); }); test("two", () => { promise.then(shared); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectInFloatingPromise"},
					{MessageId: "expectInFloatingPromise"},
				},
			},
			{
				Code: `function leaf(value) { assert.equal(value, 1); } function shared() { queue.then(leaf); } test("one", () => promise.then(shared)); test("two", () => promise.then(shared));`,
				Errors: []rule_tester.InvalidTestCaseError{
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
				Code: `test("case", async () => { let pending; pending ||= promise.then(value => expect(value).toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    41,
					EndColumn: 97,
				}},
			},
			// A chain stored by a logical assignment floats from the
			// assignment, and one handed to another call from its statement.
			{
				Code: `test("case", async () => { obj.pending ||= promise.then(value => expect(value).toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    28,
					EndColumn: 88,
				}},
			},
			{
				Code: `test("case", () => { list.push(promise.then(value => expect(value).toBe(1))); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    22,
					EndColumn: 78,
				}},
			},
			{
				Code: `test("case", () => { helper(promise.then(value => expect(value).toBe(1))); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    22,
					EndColumn: 75,
				}},
			},
			// Awaiting an array does not await its elements.
			{
				Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			// `p = p && other` always replaces the promise p holds.
			{
				Code: `test("case", async () => { let pending = promise.then(value => expect(value).toBe(1)); pending = pending && other; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			// A promise is truthy, so `&&=` always replaces the one p holds.
			{
				Code: `test("case", async () => { let pending = promise.then(value => expect(value).toBe(1)); pending &&= other; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			// A compound arithmetic assignment does not store the promise.
			{
				Code: `test("case", async () => { let count = 0; count += promise.then(value => assert.equal(value, 1)); await count; });`,
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
			{
				Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { throw new Error("caught"); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => assert.equal(value, 1)); try { throw new Error("suppressed"); } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			// A write only preserves the promise a binding holds when every
			// path through the right-hand side evaluates to it.
			{
				// a conditional drops the promise on the branch that is not the binding.
				Code: `test("case", async () => { let pending = promise.then(value => assert.equal(value, 1)); pending = condition ? pending : other; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				// a truthy left operand of `||` replaces the promise.
				Code: `test("case", async () => { let pending = promise.then(value => assert.equal(value, 1)); pending = other || pending; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				// a non-nullish left operand of `??` replaces the promise.
				Code: `test("case", async () => { let pending = promise.then(value => assert.equal(value, 1)); pending = other ?? pending; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				// a falsy left operand of `&&` becomes the value.
				Code: `test("case", async () => { let pending = promise.then(value => assert.equal(value, 1)); pending = ready && pending; await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
		},
	)
}
