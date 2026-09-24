package no_test_return_statement_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoTestReturnStatementExtras covers Rstest call shapes, API provenance and
// callback resolution that the upstream suites do not exercise.
func TestNoTestReturnStatementExtras(t *testing.T) {
	runNoTestReturnStatement(
		t,
		[]rule_tester.ValidTestCase{
			// Only statements directly in the callback body are checked; an early
			// guard nested in another statement ends the test deliberately.
			{Code: `test('guarded', () => {
  if (!supported) return;
  expect(1).toBe(1);
});`},
			// A nested function's return belongs to that function.
			{Code: `test('nested', () => {
  const load = () => { return 1; };
  expect(load()).toBe(1);
});`},
			// A hook may return a cleanup function, and suites are not tests.
			{Code: `beforeEach(() => { return () => cleanup(); });`},
			{Code: `describe('suite', () => { return; });`},
			// Options without a callback.
			{Code: `test('pending', { timeout: 100 });`},
			// A named callback that is also called directly may need its return
			// value there, so it is not reported.
			{Code: `it('one', myTest);
function myTest() { return expect(1).toBe(1); }
const value = myTest();`},
			// Exported callbacks may be used as helpers elsewhere.
			{Code: `export function myTest() { return expect(1).toBe(1); }
it('one', myTest);`},
			{Code: `function myTest() { return expect(1).toBe(1); }
it('one', myTest);
export { myTest };`},
			{Code: `export default function myTest() { return expect(1).toBe(1); }
it('one', myTest);`},
			// A reassigned binding may hold a different function when the test
			// runs.
			{Code: `let myTest = () => { return 1; };
myTest = () => {};
it('one', myTest);`},
			// A parameter shadowing the function name is a different binding.
			{Code: `function myTest() { return 1; }
function register(myTest) { it('one', myTest); }`},
			// Local and foreign test functions are not Rstest's.
			{Code: `const test = (name, fn) => fn();
test('local', () => { return 1; });`},
			{Code: `import { test } from 'vitest';
test('foreign', () => { return 1; });`},
			// A callback imported from another module has no body in this file.
			{Code: `import { myTest } from './helpers';
it('one', myTest);`},
		},
		[]rule_tester.InvalidTestCase{
			// `(name, options, fn)` puts the callback third.
			{
				Code:   `test('options', { timeout: 100 }, () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 43, 52)},
			},
			// `(name, fn, timeout)` keeps it second.
			{
				Code:   `test('timeout', () => { return 1; }, 100);`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 25, 34)},
			},
			// A bare return is still a return statement.
			{
				Code:   `test('bare', () => { return; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 22, 29)},
			},
			// async callbacks are checked the same way.
			{
				Code:   `test('async', async () => { return load(); });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 29, 43)},
			},
			// Only the first direct return is reported.
			{
				Code: `test('twice', () => {
  return 1;
  return 2;
});`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 3, 12)},
			},
			// Parentheses and TypeScript assertions around the callback erase at
			// runtime, so Rstest runs the function they wrap.
			{
				Code:   `test('asserted', (() => { return 1; }) as () => number);`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 27, 36)},
			},
			// A const-bound function is resolved like a function declaration.
			{
				Code: `const myTest = () => { return expect(1).toBe(1); };
it('one', myTest);`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 24, 49)},
			},
			// A callback shared by several tests is reported once, including when
			// it is declared after its uses.
			{
				Code: `it('one', myTest);
it.for([1])('two', myTest);
function myTest() {
  return expect(1).toBe(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(4, 3, 28)},
			},
			// A `.todo` registration still receives the function it is given.
			{
				Code:   `test.todo('later', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 28, 37)},
			},
			// Modifiers, fixtures and each/for layers.
			{
				Code:   `test.concurrent.fails('modifiers', async () => { return load(); });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 50, 64)},
			},
			{
				Code: `const withUser = test.extend({ user: async ({}, use) => { await use('ann'); } });
withUser('fixture', ({ user }) => { return user; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 37, 49)},
			},
			{
				Code:   `test.for([1])('for', (value, context) => { return value; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 44, 57)},
			},
			// Imports, namespaces, CommonJS and import.meta.rstest.
			{
				Code: `import { test as check } from '@rstest/core';
check('renamed', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 26, 35)},
			},
			{
				Code: `import * as rs from '@rstest/core';
rs.test('namespace', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 30, 39)},
			},
			{
				Code: `const { it } = require('@rstest/core');
it('commonjs', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 24, 33)},
			},
			{
				Code: `import { test } from 'rstack/test';
test('rstack', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 24, 33)},
			},
			{
				Code:   `import.meta.rstest.test('in-source', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 46, 55)},
			},
			{
				Code: `import { test } from '@rstest/playwright';
test('page', async ({ page }) => { return page.goto('/'); });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(2, 36, 58)},
			},
			// Non-ASCII text before the return keeps UTF-16 columns.
			{
				Code:   `test('日本', () => { return 1; });`,
				Errors: []rule_tester.InvalidTestCaseError{noTestReturnStatement(1, 20, 29)},
			},
		},
	)
}
