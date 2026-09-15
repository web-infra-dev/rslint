// TestMaxNestedDescribeExtras covers Rstest provenance, Rstest-only suite APIs,
// ts-go edge shapes, and regressions outside the upstream suites. The migrated
// upstream behavior lives in max_nested_describe_upstream_test.go.
package max_nested_describe_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxNestedDescribeExtras(t *testing.T) {
	runMaxNestedDescribeRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			// ---- Real-user: @vitest/eslint-plugin issue #23 ----
			{Code: `describe('currency', () => {
  describe('constructor', () => {});
  describe('normalize currency name', () => {});
  describe('get canonical name and multiplier', () => {});
  describe('to string', () => {});
});`, Options: maxOption(2)},

			// ---- Dimension 1: framework provenance and shadowing ----
			{Code: `import { describe } from 'vitest'; describe('one', () => { describe('two', () => {}); });`, Options: maxOption(1)},
			{Code: `import { describe } from '@jest/globals'; describe('one', () => { describe('two', () => {}); });`, Options: maxOption(1)},
			{Code: `suite('one', () => { suite('two', () => {}); });`, Options: maxOption(1)},
			{Code: `function run(describe: Function) { describe('one', () => { describe('two', () => {}); }); }`, Options: maxOption(1)},
			{Code: `import type { describe as suite } from '@rstest/core'; suite('one', () => {});`, Options: maxOption(0)},

			// ---- Dimension 2: non-registration and factory layers ----
			{Code: `describe.each([1]); describe.for([1]); describe.runIf(true); describe.skipIf(false);`, Options: maxOption(0)},
			{Code: `obj.describe('one', () => {}); new describe('two', () => {});`, Options: maxOption(0)},

			// ---- Dimension 4: computed dynamic key and malformed chains ----
			{Code: `const method = 'describe'; import.meta.rstest[method]('one', () => {});`, Options: maxOption(0)},
			{Code: `describe.unknown('one', () => {});`, Options: maxOption(0)},

			// ---- Callback ownership: mutable or reassigned bindings are not stable ----
			{Code: `let body = () => { describe('stale inner', () => {}); }; body = () => {}; describe('outer', body);`, Options: maxOption(1)},
			{Code: `var body = () => { describe('mutable inner', () => {}); }; describe('outer', body);`, Options: maxOption(1)},
			{Code: `function body() { describe('stale inner', () => {}); } body = () => {}; describe('outer', body);`, Options: maxOption(1)},
			{Code: `function body() { describe('captured inner', () => {}); } describe('outer', body); body = () => {};`, Options: maxOption(1)},
			{Code: `function body() { describe('stale inner', () => {}); } body ||= () => {}; describe('outer', body);`, Options: maxOption(1)},
			{Code: `function body() { describe('stale inner', () => {}); } (body satisfies (() => void)) = () => {}; describe('outer', body);`, Options: maxOption(1)},
			{Code: `const holder = { body: () => {} }; let body = () => { describe('stale inner', () => {}); }; ({ body } = holder); describe('outer', body);`, Options: maxOption(1)},
			{Code: `const holder = { body: () => {} }; function body() { describe('stale inner', () => {}); } ({ body } = holder); describe('outer', body);`, Options: maxOption(1)},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: optional call still invokes the defined Rstest global ----
			{Code: `describe?.('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 1)}},
			// ---- Rstest callback ownership: named suite callbacks preserve runtime nesting ----
			{Code: `describe('one', suiteBody); function suiteBody() { describe('two', () => {}); }`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 52)}},
			{Code: `const suiteBody = () => { describe('two', () => {}); }; describe('one', suiteBody);`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 27)}},
			{Code: `function shared() { describe('child', () => {}); } describe('one', shared); describe('other', () => { describe('inner', shared); });`, Options: maxOption(2), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(3, 2, 1, 21)}},
			{Code: `describe('one', wrap(() => { describe('two', () => {}); }));`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 30)}},
			{Code: `function register() { const body = () => { describe('inner', () => {}); }; describe('outer', body); } register();`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 44)}},
			{Code: `{ const body = () => { describe('inner', () => {}); }; describe('outer', body); }`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 24)}},
			{Code: `class C { static { const body = () => { describe('inner', () => {}); }; describe('outer', body); } }`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 41)}},

			// ---- Dimension 1: named, renamed, namespace, CommonJS, and re-export imports ----
			{Code: `import { describe as suite } from '@rstest/core'; suite('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 51)}},
			{Code: `import * as core from '@rstest/core'; core.describe('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 39)}},
			{Code: `const { describe: suite } = require('@rstest/core'); suite('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 54)}},
			{Code: `import { describe } from 'rstack/test'; describe('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 41)}},
			{Code: `const suite = describe; suite('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 25)}},
			{Code: `import.meta.rstest.describe('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 1)}},

			// ---- Dimension 2: Rstest-only modifiers and factories ----
			{Code: `describe.concurrent('one', () => { describe.sequential.todo('two'); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 36)}},
			{Code: `describe.runIf(true)('one', () => { describe.skipIf(false)('two', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 37)}},
			{Code: `describe('one', () => { describe.for([[1]])('%s', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 25)}},

			// ---- Dimension 1: import.meta and const aliases preserve semantics ----
			{Code: `const { describe: suite } = import.meta.rstest; suite('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 49)}},
			{Code: `import { describe } from '@rstest/core'; const nested = describe.only; nested('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 72)}},

			// ---- Dimension 1: Playwright suite registrations ----
			{Code: `import { test } from '@rstest/playwright'; test.describe('one', () => { test.describe.only('two', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 73)}},
			{Code: `import { test as base } from '@rstest/playwright'; const test = base.extend({}); test.describe('one', () => {});`, Options: maxOption(0), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(1, 0, 1, 82)}},

			// ---- Dimension 4: parentheses and static element access ----
			{Code: `(describe)('one', () => { (describe)['only']('two', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 27)}},
			{Code: `const suite = describe as typeof describe; suite('one', () => { (describe!)('two', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 65)}},
			{Code: "describe('one', () => { describe[`for`]`value\n${1}\n`('$value', () => {}); });", Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 25)}},

			// ---- Dimension 3: options overload and callback syntax do not change depth ----
			{Code: `describe('one', { timeout: 10 }, async function () { describe('two', () => {}); });`, Options: maxOption(1), Errors: []rule_tester.InvalidTestCaseError{exceededDepthError(2, 1, 1, 54)}},
		},
	)
}
