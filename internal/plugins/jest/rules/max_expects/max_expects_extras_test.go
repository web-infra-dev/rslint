// TestMaxExpectsExtras pins which bodies own an assertion count. Each test or
// hook callback, and each function detached from its call site, counts on its
// own, and leaving a nested function restores the enclosing count. Code
// outside any test or hook callback is not counted.
package max_expects_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/max_expects"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var max1Option = []any{map[string]any{"max": 1}}

func TestMaxExpectsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_expects.MaxExpectsRule,
		[]rule_tester.ValidTestCase{
			// A detached helper counts on its own, not toward the test.
			{
				Code: `test('a', () => {
  expect(1).toBe(1);
  const helper = () => { expect(2).toBe(2); };
});`,
				Options: max1Option,
			},
			// Sibling hooks do not share a count.
			{
				Code: `beforeEach(() => { expect(1).toBe(1); });
afterEach(() => { expect(2).toBe(2); });`,
				Options: max1Option,
			},
			// Code outside every test and hook callback is not counted.
			{
				Code: `expect(1).toBe(1);
expect(2).toBe(2);`,
				Options: max1Option,
			},
			{
				Code: `describe('d', () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
			},
			{
				Code: `describe('d', () => {
  test('a', () => { expect(1).toBe(1); });
  expect(2).toBe(2);
  expect(3).toBe(3);
});`,
				Options: max1Option,
			},
			{
				Code: `beforeEach(() => { expect(1).toBe(1); });
expect(2).toBe(2);`,
				Options: max1Option,
			},
			{
				Code: `const helper = () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
};`,
				Options: max1Option,
			},
			// Each branch of a conditional is a separate candidate callback.
			{
				Code: `test('a', flag
  ? () => { expect(1).toBe(1); }
  : () => { expect(2).toBe(2); }
);`,
				Options: max1Option,
			},
			{
				Code: `beforeEach(flag
  ? () => { expect(1).toBe(1); }
  : () => { expect(2).toBe(2); }
);`,
				Options: max1Option,
			},
			// `test.each(table, fn)` registers no test, so its function is not a
			// test body.
			{
				Code: `test.each([1, 2], () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
			},
		},
		[]rule_tester.InvalidTestCase{
			// Leaving a detached helper restores the enclosing test's count.
			{
				Code: `test('a', () => {
  expect(1).toBe(1);
  const helper = () => { expect(2).toBe(2); };
  expect(3).toBe(3);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 4, Column: 3},
				},
			},
			{
				Code: `beforeEach(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 3, Column: 3},
				},
			},
			// A callback passed by name is still the test body.
			{
				Code: `const body = () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
};
test('a', body);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 3, Column: 3},
				},
			},
			{
				Code: `test('a', () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}, 1000);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceededMaxAssertion", Line: 3, Column: 3},
				},
			},
		},
	)
}
