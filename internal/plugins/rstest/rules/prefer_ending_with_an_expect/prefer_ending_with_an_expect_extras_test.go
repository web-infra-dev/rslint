// TestPreferEndingWithAnExpectExtras covers Rstest provenance, the call shapes
// and assertion forms Rstest adds on top of the upstream jest rule, and ts-go
// edge shapes the upstream suite does not reach. The migrated upstream behavior
// lives in prefer_ending_with_an_expect_upstream_test.go.
package prefer_ending_with_an_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEndingWithAnExpectExtras(t *testing.T) {
	runPreferEndingWithAnExpectRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: foreign APIs, shadows, and type-only bindings ----
			{Code: `import { it } from 'vitest'; it('places the order', () => { checkout(); });`},
			{Code: `import { test } from '@jest/globals'; test('places the order', () => { checkout(); });`},
			{Code: `function run(test: Function) { test('places the order', () => { checkout(); }); }`},
			{Code: `import type { test as scenario } from '@rstest/core'; scenario('places the order', () => { checkout(); });`},

			// ---- Dimension 2: registrations that run no callback ----
			{Code: `describe('checkout', () => { resetCart(); });`},
			{Code: `beforeEach(() => { resetCart(); });`},
			{Code: `test.todo('supports gift cards');`},
			// `.todo` ignores the function it is handed, so requiring an
			// assertion at its end would report code that never runs.
			{Code: `test.todo('supports gift cards', () => { resetCart(); });`},
			{Code: `const todoTest = test.todo;
todoTest('supports gift cards', () => { resetCart(); });`},
			// A callback passed by reference is left alone.
			{Code: `test('places the order', run); function run() { resetCart(); }`},
			{Code: `test.extend({ cart: async ({}, use) => use(createCart()) });`},

			// ---- Dimension 3: both Rstest call shapes ----
			{Code: `test('places the order', { timeout: 100 }, () => { expect(checkout()).toBe('ok'); });`},
			{Code: `test('places the order', () => { expect(checkout()).toBe('ok'); }, 100);`},
			{Code: `test.each([[1]])('places order %i', { timeout: 100 }, count => { expect(checkout(count)).toBe('ok'); });`},
			{Code: `test.for([{ currency: 'USD' }])('$currency cart', (row, context) => { context.expect(row).toBeDefined(); });`},
			{Code: `test.concurrent('places the order', () => { expect(checkout()).toBe('ok'); });`},

			// ---- Dimension 4: assertion forms the callee-text patterns cannot match ----
			{Code: `test('places the order', context => { context.expect(checkout()).toBe('ok'); });`},
			{Code: `test('places the order', ({ expect }) => { expect(checkout()).toBe('ok'); });`},
			{Code: `import * as rstest from '@rstest/core';
rstest.test('places the order', () => { rstest.expect(checkout()).toBe('ok'); });`},
			{Code: `import { test, expect as check } from '@rstest/core';
test('places the order', () => { check(checkout()).toBe('ok'); });`},
			{Code: `if (import.meta.rstest) {
  import.meta.rstest.test('places the order', () => {
    import.meta.rstest.expect(checkout()).toBe('ok');
  });
}`},
			// A computed key that folds to a constant names the same API as the
			// plain key, so the destructured local still asserts.
			{Code: `const { ['expect']: check } = import.meta.rstest;
test('places the order', () => { check(checkout()).toBe('ok'); });`},
			{Code: `const { [` + "`expect`" + `]: check } = import.meta.rstest;
test('places the order', () => { check(checkout()).toBe('ok'); });`},
			// Chai's `assert` is an Rstest global, so it asserts by default.
			{Code: `test('places the order', () => { assert.equal(checkout(), 'ok'); });`},
			{Code: `test('places the order', () => { expect.soft(checkout()).toBe('ok'); });`},
			{Code: `test('places the order', async () => { await expect.poll(() => checkout()).toBe('ok'); });`},
			{Code: `const cartTest = test.extend({ cart: async ({}, use) => use(createCart()) });
cartTest('places the order', ({ cart, expect }) => { expect(cart).toBeDefined(); });`},

			// ---- Dimension 5: Playwright ----
			{Code: `import { test, expect } from '@rstest/playwright';
test('places the order', async ({ page }) => { await expect(page).toHaveTitle('Cart'); });`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: provenance ----
			{
				Code: `import { test as scenario } from '@rstest/core';
scenario('places the order', () => { checkout(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(2, 1, 9)},
			},
			{
				Code: `import { test } from 'rstack/test';
test('places the order', () => { checkout(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(2, 1, 5)},
			},
			{
				Code:   `import.meta.rstest.test('places the order', () => { checkout(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 24)},
			},

			// ---- Dimension 3: the options overload is checked, not skipped ----
			{
				Code:   `test('places the order', { timeout: 100 }, () => { checkout(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:   `test('places the order', () => { checkout(); }, 100);`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:   `test.concurrent('places the order', () => { checkout(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 16)},
			},

			// ---- Dimension 4: an assertion that is not the last statement ----
			{
				Code:   `test('places the order', context => { context.expect(checkout()).toBe('ok'); resetCart(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:   `test('places the order', ({ expect }) => { expect(checkout()).toBe('ok'); resetCart(); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code: `const cartTest = test.extend({ cart: async ({}, use) => use(createCart()) });
cartTest('places the order', ({ cart }) => { checkout(cart); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(2, 1, 9)},
			},
			// A return statement is not a call expression, so upstream reports it
			// even though the assertion is the value being returned.
			{
				Code:   `test('places the order', () => { return expect(checkout()).toBe('ok'); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},

			// ---- Dimension 5: Playwright ----
			{
				Code: `import { test } from '@rstest/playwright';
test('places the order', async ({ page }) => { await page.click('#checkout'); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(2, 1, 5)},
			},
		},
	)
}
