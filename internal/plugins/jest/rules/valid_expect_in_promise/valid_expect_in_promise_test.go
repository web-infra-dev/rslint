package valid_expect_in_promise_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_expect_in_promise"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidExpectInPromiseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_expect_in_promise.ValidExpectInPromiseRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("case", async () => { await promise.then(value => expect(value).toBe(1)); });`},
			{Code: `test("case", () => promise.then(value => expect(value).toBe(1)));`},
			{Code: `test("case", () => { try { return promise.then(value => expect(value).toBe(1)); } catch (error) {} });`},
			{Code: `test("case", async () => { try { return promise.then(value => expect(value).toBe(1)); } catch (error) {} });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); other = 1; await pending; });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.resolve(pending); });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([pending]); });`},
			{Code: `test("case", () => { const [pending] = [promise.then(value => expect(value).toBe(1))]; return pending; });`},
			{Code: `test("case", () => { const { pending } = { pending: promise.then(value => expect(value).toBe(1)) }; return pending; });`},
			{Code: `test("case", () => { let pending; [pending] = [promise.then(value => expect(value).toBe(1))]; return pending; });`},
			{Code: `test("case", () => { let pending; ({ pending } = { pending: promise.then(value => expect(value).toBe(1)) }); return pending; });`},
			// A logical assignment stores the chain, only conditionally, so the
			// binding is real and the value it may leave in place survives.
			{Code: `test("case", async () => { let pending; pending ||= promise.then(value => expect(value).toBe(1)); await pending; });`},
			{Code: `test("case", async () => { let pending; pending ??= promise.then(value => expect(value).toBe(1)); await pending; });`},
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
			{Code: `test("case", async () => { let pending = promise.then(value => expect(value).toBe(1)); pending = pending || other; await pending; });`},
			// An array in a binding is consumed element-wise.
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; await Promise.all(pending); });`},
			{Code: `test("case", async () => { const pending = [promise.then(value => expect(value).toBe(1))]; for await (const settled of pending) {} });`},
			{Code: `test("case", async () => { for await (const settled of [promise.then(value => expect(value).toBe(1))]) {} });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); expect(pending).resolves.toBeUndefined(); });`},
			{Code: `test("case", () => expect(promise.then(value => expect(value).toBe(1))).resolves.toBeUndefined());`},
			{Code: `test("case", () => condition ? promise.then(value => expect(value).toBe(1)) : Promise.resolve());`},
			{Code: `test("case", () => condition && promise.then(value => expect(value).toBe(1)));`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); if (condition) await pending; else return pending; });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); if (!ready) { throw new Error("no"); } await pending; });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { setup(); await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { throw new Error("no"); } finally { cleanup(); } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } finally { cleanup(); } });`},
			{Code: `try { test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); await pending; }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.resolve(pending); }); } catch (error) {}`},
			{Code: `try { test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([pending]); }); } catch (error) {}`},
			{Code: `test("done", done => { promise.then(value => { expect(value).toBe(1); done(); }); });`},
			{Code: `function callback() { const pending = promise.then(value => expect(value).toBe(1)); return pending; } test("case", callback);`},
			{Code: `function handler(value) { expect(value).toBe(1); } test("one", () => promise.then(handler)); test("two", () => promise.then(handler));`},
			{Code: `function handler() { queue.then(handler); } test("case", () => promise.then(handler));`},
			{Code: `function first() { queue.then(second); } function second() { queue.then(first); } test("case", () => promise.then(first));`},
			{Code: `function handler(value) { expect(value).toBe(1); return queue.then(handler); } test("case", () => promise.then(handler));`},
			{Code: `let handler; test("case", () => promise.then(handler));`},
			{Code: `declare const handler: () => void; test("case", () => promise.then(handler));`},
			{Code: `let handler: (value: unknown) => void; beforeEach(() => { handler = value => expect(value).toBe(1); }); test("case", () => promise.then(handler));`},
			{Code: `test("done", done => { test("nested", () => {}); promise.then(value => { expect(value).toBe(1); done(); }); });`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("case", () => { promise.then(value => expect(value).toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    22,
					EndColumn: 67,
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.reject(pending); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    28,
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.allSettled([pending]); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    28,
				}},
			},
			{
				Code: `function callback() { promise.then(value => expect(value).toBe(1)); } test("case", callback);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    23,
				}},
			},
			{
				Code: `test("case", async () => { await first.then(() => { second.then(value => expect(value).toBe(1)); }); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    53,
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); if (condition) await pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    34,
				}},
			},
			{
				Code: `test("case", () => { if (condition) { promise.then(value => expect(value).toBe(1)); } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
					Line:      1,
					Column:    39,
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([[pending]]); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const [pending] = [promise.then(value => expect(value).toBe(1))]; log(pending); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { let pending; [pending] = [promise.then(value => expect(value).toBe(1))]; log(pending); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { let pending; ({ pending } = { pending: promise.then(value => expect(value).toBe(1)) }); log(pending); });`,
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
				Code: `test("case", async () => { let count = 0; count += promise.then(value => expect(value).toBe(1)); await count; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { promise.then(value => expect(value).toBe(1)).then(value => expect(value).toBe(2)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `function handler(value) { expect(value).toBe(1); } test("one", () => { promise.then(handler); }); test("two", () => { promise.then(handler); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectInFloatingPromise"},
					{MessageId: "expectInFloatingPromise"},
				},
			},
			{
				Code: `function leaf(value) { expect(value).toBe(1); } function shared() { return queue.then(leaf); } test("one", () => { promise.then(shared); }); test("two", () => { promise.then(shared); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectInFloatingPromise"},
					{MessageId: "expectInFloatingPromise"},
				},
			},
			{
				Code: `function leaf(value) { expect(value).toBe(1); } function shared() { queue.then(leaf); } test("one", () => promise.then(shared)); test("two", () => promise.then(shared));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectInFloatingPromise"},
				},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); while (condition) { await pending; } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return condition ? pending : other; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return condition && pending; });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { throw new Error("caught"); } catch {} });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
			{
				Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { throw new Error("suppressed"); } finally { return; } });`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectInFloatingPromise",
				}},
			},
		},
	)
}
