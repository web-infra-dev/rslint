// TestPreferMockReturnShorthandUpstream migrates every valid and invalid case
// from @vitest/eslint-plugin@v1.6.27 tests/prefer-mock-return-shorthand.test.ts,
// with the mock utilities object written as Rstest spells it. Rstest call
// shapes, tsgo edit shapes and branch lock-ins live in the extras suite.
//
// One upstream group is deliberately reversed: a callback returning
// `Promise.reject(...)` is not reported here at all. See the rule source for
// why, and the extras suite for the Rstest behavior that decides it.
package prefer_mock_return_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMockReturnShorthandUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferMockReturnShorthandRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe()`},
			{Code: `it()`},
			{Code: `describe.skip()`},
			{Code: `it.skip()`},
			{Code: `test()`},
			{Code: `test.skip()`},
			{Code: `var appliedOnly = describe.only; appliedOnly.apply(describe)`},
			{Code: `var calledOnly = it.only; calledOnly.call(it)`},
			{Code: `it.each()()`},
			{Code: "it.each`table`()"},
			{Code: `test.each()()`},
			{Code: "test.each`table`()"},
			{Code: `test.concurrent()`},
			{Code: `rs.fn().mockReturnValue(42)`},
			{Code: `rs.fn(() => Promise.resolve(42))`},
			{Code: `rs.fn(() => 42)`},
			{Code: `rs.fn(() => ({}))`},
			{Code: `aVariable.mockImplementation`},
			{Code: `aVariable.mockImplementation()`},
			{Code: `rs.fn().mockImplementation(async () => 1);`},
			{Code: `rs.fn().mockImplementation(async function () {});`},
			{Code: `rs.fn().mockImplementation(async function () {
  return 42;
});`},
			{Code: `aVariable.mockImplementation(() => {
  if (true) {
    return 1;
  }

  return 2;
});`},
			{Code: `aVariable.mockImplementation(() => value++)`},
			{Code: `aVariable.mockImplementationOnce(() => --value)`},
			{Code: `const aValue = 0;
aVariable.mockImplementation(() => {
  return aValue++;
});`},
			{Code: `aVariable.mockImplementation(() => {
  aValue += 1;

  return aValue;
});`},
			{Code: `aVariable.mockImplementation(() => {
  aValue++;

  return aValue;
});`},
			{Code: `aVariable.mockReturnValue()`},
			{Code: `aVariable.mockReturnValue(1)`},
			{Code: `aVariable.mockReturnValue("hello world")`},
			{Code: `rs.spyOn(Thingy, 'method').mockImplementation(param => param * 2);`},
			{Code: `rs.spyOn(Thingy, 'method').mockImplementation(param => true ? param : 0);`},
			{Code: `aVariable.mockImplementation(() => {
  const value = new Date();

  return Promise.resolve(value);
});`},
			{Code: `aVariable.mockImplementation(() => {
  throw new Error('oh noes!');
});`},
			{Code: `aVariable.mockImplementation(() => { /* do something */ });`},
			{Code: `aVariable.mockImplementation(() => {
  const x = 1;

  console.log(x + 2);
});`},
			{Code: `aVariable.mockReturnValue(Promise.all([1, 2, 3]));`},
			{Code: `let currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);

currentX++;`},
			{Code: `let currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `let currentX = 0;
currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `var currentX = 0;
currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `var currentX = 0;
var currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `let doSomething = () => {};

rs.spyOn(X, getCount).mockImplementation(() => doSomething);`},
			{Code: `let currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => {
  currentX += 1;

  return currentX;
});`},
			{Code: `const currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => {
  console.log('returning', currentX);

  return currentX;
});`},
			// Upstream reports and fixes each of these four; this rule does not.
			// Moving `Promise.reject(...)` into `mockReturnValue` builds the
			// rejected promise when the mock is configured rather than when it is
			// called, so a mock that is never called leaves an unhandled rejection
			// that Rstest reports as a run-level error.
			{Code: `rs.fn().mockImplementation(() => Promise.reject(13))`},
			{Code: `rs.fn().mockImplementation(() => {
  return Promise.reject(13);
})`},
			{Code: `aVariable.mockImplementation(() => Promise.reject(13))`},
			{Code: `aVariable.mockImplementation(() => {
  return Promise.reject(13);
})`},
			{Code: `aVariable
  .mockImplementationOnce(() => Promise.reject(42))
  .mockReturnValueOnce(Promise.reject(42))`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `rs.fn().mockImplementation(() => "hello sunshine")`,
				Output: []string{`rs.fn().mockReturnValue("hello sunshine")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => null)`,
				Output: []string{`rs.fn().mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(() => {
  return null;
})`,
				Output: []string{`rs.fn().mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => null)`,
				Output: []string{`aVariable.mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return null;
})`,
				Output: []string{`aVariable.mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => 0)`,
				Output: []string{`rs.fn().mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(() => {
  return 0;
})`,
				Output: []string{`rs.fn().mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => 0)`,
				Output: []string{`aVariable.mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return 0;
})`,
				Output: []string{`aVariable.mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`rs.fn().mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(() => {
  return Promise.resolve(42);
})`,
				Output: []string{`rs.fn().mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`aVariable.mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return Promise.resolve(42);
})`,
				Output: []string{`aVariable.mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => [])`,
				Output: []string{`rs.fn().mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(() => {
  return [];
})`,
				Output: []string{`rs.fn().mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => [])`,
				Output: []string{`aVariable.mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return [];
})`,
				Output: []string{`aVariable.mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => ({}))`,
				Output: []string{`rs.fn().mockReturnValue({})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => x)`,
				Output: []string{`rs.fn().mockReturnValue(x)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => true ? x : y)`,
				Output: []string{`rs.fn().mockReturnValue(true ? x : y)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(function () {
  return "hello world";
})`,
				Output: []string{`rs.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => "hello world")`,
				Output: []string{`rs.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `rs.fn().mockImplementation(() => {
  return "hello world";
})`,
				Output: []string{`rs.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => "hello world")`,
				Output: []string{`aVariable.mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return "hello world";
})`,
				Output: []string{`aVariable.mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockImplementationOnce(() => "hello world")`,
				Output: []string{`rs.fn().mockReturnValueOnce("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 1, Column: 9},
				},
			},
			{
				Code:   `aVariable.mockImplementationOnce(() => "hello world")`,
				Output: []string{`aVariable.mockReturnValueOnce("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => ({
  target: 'world',
  message: 'hello'
}))`,
				Output: []string{`aVariable.mockReturnValue({
  target: 'world',
  message: 'hello'
})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable
  .mockImplementation(() => 42)
  .mockImplementation(async () => 42)
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValue("hello world")`,
				Output: []string{`aVariable
  .mockReturnValue(42)
  .mockImplementation(async () => 42)
  .mockReturnValue(Promise.resolve(42))
  .mockReturnValue("hello world")`},
				// A member chain nests outermost-call-first, so the rule visits the
				// last link before the first one. The linter sorts a completed
				// diagnostic set by position before emitting it; the rule tester
				// compares in report order.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 4},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 2, Column: 4},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => [], xyz)`,
				Output: []string{`rs.fn().mockReturnValue([], xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.spyOn(fs, "readFile").mockImplementation(() => new Error("oh noes!"))`,
				Output: []string{`rs.spyOn(fs, "readFile").mockReturnValue(new Error("oh noes!"))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 26},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return Promise.resolve(value)
    .then(value => value + 1);
});`,
				Output: []string{`aVariable.mockReturnValue(Promise.resolve(value)
    .then(value => value + 1));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return Promise.all([1, 2, 3]);
});`,
				Output: []string{`aVariable.mockReturnValue(Promise.all([1, 2, 3]));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `const currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`,
				Output: []string{`const currentX = 0;
rs.spyOn(X, getCount).mockReturnValue(currentX);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 2, Column: 23},
				},
			},
			{
				Code: `import { currentX } from './elsewhere';
rs.spyOn(X, getCount).mockImplementation(() => currentX);`,
				Output: []string{`import { currentX } from './elsewhere';
rs.spyOn(X, getCount).mockReturnValue(currentX);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 2, Column: 23},
				},
			},
			{
				Code: `const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    rs.spyOn(X, getCount).mockImplementation(() => currentX);
  });
});`,
				Output: []string{`const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    rs.spyOn(X, getCount).mockReturnValue(currentX);
  });
});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 27},
				},
			},
			{
				Code: `function doSomething() {};

rs.spyOn(X, getCount).mockImplementation(() => doSomething);`,
				Output: []string{`function doSomething() {};

rs.spyOn(X, getCount).mockReturnValue(doSomething);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 23},
				},
			},
			{
				Code: `const doSomething = () => {};

rs.spyOn(X, getCount).mockImplementation(() => doSomething);`,
				Output: []string{`const doSomething = () => {};

rs.spyOn(X, getCount).mockReturnValue(doSomething);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 23},
				},
			},
		},
	)
}
