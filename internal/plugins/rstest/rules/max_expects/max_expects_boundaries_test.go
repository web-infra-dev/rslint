package max_expects_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxExpectsCountingBoundaries pins which bodies open a count. A callback
// the parser cannot resolve is recognised by position, wherever inside the
// registration's callback argument it sits, and a lifecycle hook body counts
// like a test body. The cases the rule declines to count are listed as valid so
// the boundary of the positional recognition stays visible.
func TestMaxExpectsCountingBoundaries(t *testing.T) {
	runMaxExpectsRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			// A callback reached through a member expression is not attributed
			// to the registration: the function is written outside it and
			// nothing at the registration ties the two together.
			{
				Code: `
const suite = { run: () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
} };
test('a', suite.run);`,
				Options: max1Option,
			},
			{
				Code: `
class Suite {
  run = () => {
    expect(1).toBe(1);
    expect(2).toBe(2);
  };
}
test('a', new Suite().run);`,
				Options: max1Option,
			},
			// An object literal argument is the options overload, so a function
			// held in one of its properties is not the callback.
			{
				Code: `
test('a', { retry: 2 }, () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- An unresolved callback counts wherever it sits inside the
			// registration's callback argument, not only as a direct argument
			// of the wrapping call ----
			{
				Code: `
test('a', new Wrapper(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
test('a', wrap([() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}]));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
test('a', cond ? () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
} : null);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
test('a', wrap({ cb: () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
} }));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
test('a', wrap(withDb({ run: () => {
  expect(1).toBe(1);
  expect(2).toBe(2);
} })));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			// ---- A hook body carries its own count ----
			{
				Code: `
beforeEach(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
afterAll(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			{
				Code: `
describe('d', () => {
  beforeEach(() => {
    expect(1).toBe(1);
    expect(2).toBe(2);
  });
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 5, 22),
				},
			},
			{
				Code: `
beforeEach(wrap(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3, 20),
				},
			},
			// ---- Any function-valued expression between two assertions is a
			// boundary: it takes its own count and hands the enclosing count
			// back. Upstream resets to zero instead and stops reporting the
			// assertions that follow ----
			{
				Code: `
test('a', () => {
  expect(1).toBe(1);
  test('b', () => {});
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 3, 20),
				},
			},
			{
				Code: `
test('a', () => {
  expect(1).toBe(1);
  const o = { m() {} };
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 3, 20),
				},
			},
			{
				Code: `
test('a', () => {
  expect(1).toBe(1);
  new Foo(() => {});
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 3, 20),
				},
			},
			// A hook body and the test bodies around it are separate counts.
			{
				Code: `
describe('d', () => {
  beforeEach(() => {
    expect(1).toBe(1);
    expect(2).toBe(2);
  });

  test('a', () => {
    expect(3).toBe(3);
    expect(4).toBe(4);
  });
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 5, 22),
					exceededMaxError(2, 1, 10, 5, 22),
				},
			},
		},
	)
}
