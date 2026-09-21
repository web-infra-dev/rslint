// TestRequireTopLevelDescribeExtras covers Rstest provenance, Rstest-only test
// and suite APIs, callback ownership, and ts-go edge shapes outside the
// upstream suites. The migrated upstream behavior lives in
// require_top_level_describe_upstream_test.go.
package require_top_level_describe_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireTopLevelDescribeExtras(t *testing.T) {
	runRequireTopLevelDescribeRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: foreign APIs, shadows, and type-only bindings ----
			{Code: `import { it } from 'vitest'; it('places the order', () => {});`},
			{Code: `import { beforeEach } from '@jest/globals'; beforeEach(() => {});`},
			{Code: `function run(it: Function) { it('places the order', () => {}); }`},
			{Code: `import type { test as scenario } from '@rstest/core'; scenario('places the order', () => {});`},
			{Code: `const beforeAll = createHookRegistry(); beforeAll(() => {});`},

			// ---- Dimension 2: factories and non-registrations ----
			{Code: `import { test } from '@rstest/core'; test.extend({});`},
			{Code: `test.for(); test.runIf(true); test.skipIf(false); describe.each([1]); describe.for([1]);`},
			{Code: `subject.test('places the order', () => {}); new test('places the order', () => {});`},
			{Code: `import * as rstest from '@rstest/core'; const api = 'test'; rstest[api]('places the order', () => {});`},
			// onTestFinished and onTestFailed run inside a test body, so they
			// are not suite hooks and carry no suite requirement.
			{Code: `import { onTestFinished, onTestFailed } from '@rstest/core';
onTestFinished(() => {});
onTestFailed(() => {});`},

			// ---- Dimension 3: callback ownership keeps suite membership ----
			{Code: `function suite() { it('places the order', () => {}); }
describe('checkout', suite);`},
			{Code: `const suite = () => { beforeEach(() => resetCart()); };
describe('checkout', suite);`},
			{Code: `function inner() { it('places the order', () => {}); }
function outer() { describe('signed-in customer', inner); }
describe('checkout', outer);`},
			// Mutually recursive suite callbacks must terminate rather than
			// recurse; neither registration escapes the outer suite.
			{Code: `function first() { describe('signed-in customer', second); }
function second() { describe('saved payment method', first); it('places the order', () => {}); }
describe('checkout', first);`},
			// A suite registered through a named callback is not top level, so
			// it does not consume the top-level budget.
			{
				Code: `function nested() { describe('signed-in customer', () => {}); }
describe('checkout', nested);`,
				Options: maxDescribesOption(1),
			},

			// ---- Dimension 4: Playwright suites ----
			{Code: `import { test } from '@rstest/playwright';
test.describe('checkout', () => {
  test('places the order', () => {});
  test.beforeEach(() => {});
});`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: named, renamed, namespace, CommonJS, rstack/test, import.meta ----
			{
				Code: `import { test as scenario } from '@rstest/core';
scenario('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(2, 1)},
			},
			{
				Code: `import * as rstest from '@rstest/core';
rstest.test('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(2, 1)},
			},
			{
				Code: `const { afterAll: cleanup } = require('@rstest/core');
cleanup(() => closeDatabase());`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(2, 1)},
			},
			{
				Code: `import { test } from 'rstack/test';
test('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(2, 1)},
			},
			{
				Code:   `import.meta.rstest.beforeAll(() => connectDatabase());`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(1, 1)},
			},

			// ---- Dimension 2: Rstest-only test modifiers and parameterization ----
			{
				Code:   `test.concurrent('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `test.sequential.fails('rejects an empty cart', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `test.for([{ currency: 'USD' }])('$currency cart', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `test.runIf(enabled)('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `test.todo('supports gift cards');`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			// A fixture factory alone registers nothing, but invoking its
			// result registers a test case like any other.
			{
				Code: `import { test } from '@rstest/core';
const cartTest = test.extend({ cart: async ({}, use) => use(createCart()) });
cartTest('places the order', ({ cart }) => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(3, 1)},
			},
			{
				Code:   `test.extend({ cart: async ({}, use) => use(createCart()) })('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},

			// ---- Dimension 3: ownership resolves suites, not every callback ----
			{
				Code: `function body() { it('places the order', () => {}); }
it('runs the shared body', body);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedTestCaseError(1, 19),
					unexpectedTestCaseError(2, 1),
				},
			},
			// A callback reached only through a local write is not attributed
			// to the suite, so the registration inside it stays top level.
			{
				Code: `let body = () => { it('places the order', () => {}); };
body = () => {};
describe('checkout', body);`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 20)},
			},

			// ---- Dimension 4: Playwright top-level registrations and hooks ----
			{
				Code: `import { test } from '@rstest/playwright';
test('places the order', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(2, 1)},
			},
			{
				Code: `import { test } from '@rstest/playwright';
test.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(2, 1)},
			},

			// ---- Dimension 5: top-level budget with ownership and ts-go shapes ----
			{
				Code: `describe?.('checkout', () => {});
describe('returns', () => {});`,
				Options: maxDescribesOption(1),
				Errors:  []rule_tester.InvalidTestCaseError{tooManyDescribesError(1, 2, 1)},
			},
			{
				Code: `function nested() { describe('signed-in customer', () => {}); }
describe('checkout', nested);
describe('returns', () => {});`,
				Options: maxDescribesOption(1),
				Errors:  []rule_tester.InvalidTestCaseError{tooManyDescribesError(1, 3, 1)},
			},
		},
	)
}
