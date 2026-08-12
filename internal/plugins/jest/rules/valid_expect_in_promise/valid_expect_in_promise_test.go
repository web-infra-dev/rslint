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
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); other = 1; await pending; });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.resolve(pending); });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); return Promise.all([pending]); });`},
			{Code: `test("case", () => { const [pending] = [promise.then(value => expect(value).toBe(1))]; return pending; });`},
			{Code: `test("case", () => { const { pending } = { pending: promise.then(value => expect(value).toBe(1)) }; return pending; });`},
			{Code: `test("case", () => { let pending; [pending] = [promise.then(value => expect(value).toBe(1))]; return pending; });`},
			{Code: `test("case", () => { let pending; ({ pending } = { pending: promise.then(value => expect(value).toBe(1)) }); return pending; });`},
			{Code: `test("case", () => { const pending = promise.then(value => expect(value).toBe(1)); expect(pending).resolves.toBeUndefined(); });`},
			{Code: `test("case", () => expect(promise.then(value => expect(value).toBe(1))).resolves.toBeUndefined());`},
			{Code: `test("case", () => condition ? promise.then(value => expect(value).toBe(1)) : Promise.resolve());`},
			{Code: `test("case", () => condition && promise.then(value => expect(value).toBe(1)));`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); if (condition) await pending; else return pending; });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } catch (error) { throw error; } });`},
			{Code: `test("case", async () => { const pending = promise.then(value => expect(value).toBe(1)); try { await pending; } finally { cleanup(); } });`},
			{Code: `test("done", done => { promise.then(value => { expect(value).toBe(1); done(); }); });`},
			{Code: `function callback() { const pending = promise.then(value => expect(value).toBe(1)); return pending; } test("case", callback);`},
			{Code: `function handler(value) { expect(value).toBe(1); } test("one", () => promise.then(handler)); test("two", () => promise.then(handler));`},
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
		},
	)
}
