package valid_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_expect.ValidExpectRule,
		[]rule_tester.ValidTestCase{
			// Basic well-formed assertions.
			{Code: `expect(value).toBe(1);`},
			{Code: `expect(value).not.toBe(1);`},
			{Code: `expect("something").toEqual("else");`},

			// Static members carry no assertion factory: no arg/await checks.
			{Code: `expect.hasAssertions();`},
			{Code: `expect.assertions(1);`},
			{Code: `expect.any(Number);`},
			{Code: `expect.stringContaining("x");`},

			// vitest maxArgs allowance: message second argument and poll/element
			// options are not excess arguments.
			{Code: `expect(value, "message").toBe(1);`},
			{Code: "expect(value, `msg ${x}`).toBe(1);"},
			{Code: `test("t", async () => { await expect.poll(() => value, { timeout: 1 }).toBe(1); });`},
			{Code: `test("t", async () => { await expect.element(locator, { timeout: 1 }).toBeVisible(); });`},

			// Chai property matchers are valid without a call.
			{Code: `expect(value).to.be.ok;`},
			{Code: `expect(value).to.be.a("string");`},
			{Code: `expect(spy).to.have.been.called;`},

			// Async assertions awaited or returned.
			{Code: `test("t", async () => { await expect(p).resolves.toBe(1); });`},
			{Code: `test("t", () => { return expect(p).rejects.toThrow(); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.true; });`},
			{Code: `test("t", () => { return expect(p).resolves.to.be.true; });`},
			{Code: `test("t", () => expect(p).resolves.to.be.true);`},
			// A Chai chain may carry several matchers; the await/return check
			// looks at the outermost expression, not the first matcher.
			// Parentheses are nodes in the TypeScript AST but not in ESTree, so
			// the await/return check must look past them to match upstream.
			{Code: `test("t", async () => { await (expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", () => { return (expect(p).resolves.toBe(1)); });`},
			{Code: `test("t", async () => { await ((expect(p).resolves.to.be.true)); });`},
			{Code: `test("t", async () => { await Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`},
			{Code: `test("t", async () => { await (expect(p).resolves.toBe(1)).then(() => {}); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			{Code: `test("t", async () => { await expect(p).resolves.to.be.an("object").that.is.ok; });`},
			{Code: `test("t", () => { return expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			{Code: `expect(value).not.toBe(1);`},
			{Code: `test("t", async () => { await expect.poll(() => value).toBe(1); });`},
			// poll/element are not treated as async by valid-expect (matching
			// vitest: only resolves/rejects modifiers and asyncMatchers require
			// await), so a poll chain without await is not reported here.
			{Code: `test("t", () => { expect.poll(() => value).toBe(1); });`},
			{Code: `expect(1, 2).toBe(1);`, Options: []any{map[string]any{"maxArgs": 2}}},
			{
				Code:    `test("t", async () => { await expect(p).toResolveEventually(); });`,
				Options: []any{map[string]any{"asyncMatchers": []any{"toResolveEventually"}}},
			},

			// Provenance: non-Rstest expect and local shadow are ignored.
			{Code: `import { expect } from "vitest"; expect(1);`},
			{Code: `const expect = () => {}; expect(1);`},
		},
		[]rule_tester.InvalidTestCase{
			// --- arg count ---
			{
				Code: `expect().toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs"},
				},
			},
			{
				Code: `expect(1, 2, 3).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs"},
				},
			},
			{
				Code: `expect(value, 123).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyArgs"},
				},
			},
			// jest would report tooManyArgs here; rstest allows the message arg,
			// so this appears ONLY in the valid list above. A three-arg call still
			// reports because the allowance is a single trailing message.

			// --- broken chains (reason from the expect parser) ---
			{
				Code: `expect(value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect(value).toBe;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotCalled"},
				},
			},
			{
				Code: `expect(value).resolves;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `expect(value).foo.toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "modifierUnknown"},
				},
			},

			// --- async must be awaited ---
			{
				Code: `test("t", async () => { expect(p).resolves.toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.toBe(1); });`},
			},
			{
				Code: `test("t", async () => { expect(p).rejects.toThrow(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).rejects.toThrow(); });`},
			},
			{
				Code: `test("t", async () => { expect(p).resolves.to.be.true; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.to.be.true; });`},
			},
			{
				// A parenthesized assertion that nothing awaits still reports,
				// and the fix goes inside the parentheses.
				Code: `test("t", async () => { (expect(p).resolves.toBe(1)); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { (await expect(p).resolves.toBe(1)); });`},
			},
			{
				Code:    `test("t", () => { return (expect(p).resolves.toBe(1)); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await (expect(p).resolves.toBe(1)); });`},
			},
			{
				// A multi-matcher chain still reports when it is not awaited,
				// and the fix covers the whole chain.
				Code: `test("t", async () => { expect(p).resolves.to.be.a("string").that.contains("x"); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.to.be.a("string").that.contains("x"); });`},
			},
			{
				Code: `test("t", async () => { const all = Promise.all([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { const all = await Promise.all([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)]); });`},
			},
			{
				Code: `test("t", async () => { Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all([(expect(p).resolves.toBe(1)), expect(q).resolves.toBe(2)]); });`},
			},
			{
				Code: `test("t", async () => { Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)])); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "promisesWithAsyncAssertionsMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await Promise.all(([expect(p).resolves.toBe(1), expect(q).resolves.toBe(2)])); });`},
			},
			{
				Code:    `test("t", () => { return expect(p).resolves.toBe(1); });`,
				Options: []any{map[string]any{"alwaysAwait": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).resolves.toBe(1); });`},
			},
			{
				Code:    `expect(value).toBe(1);`,
				Options: []any{map[string]any{"minArgs": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notEnoughArgs"},
				},
			},
			{
				Code:    `test("t", () => { expect(p).toResolveEventually(); });`,
				Options: []any{map[string]any{"asyncMatchers": []any{"toResolveEventually"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "asyncMustBeAwaited"},
				},
				Output: []string{`test("t", async () => { await expect(p).toResolveEventually(); });`},
			},

			// --- provenance (resolved by the expect parser) ---
			{
				Code: `import { expect as check } from "@rstest/core"; check(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
			{
				Code: `test("t", ({ expect }) => { expect(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matcherNotFound"},
				},
			},
		},
	)
}
