// TestPreferMockReturnShorthandUpstream migrates every valid and invalid case
// from eslint-plugin-jest@v29.16.1
// src/rules/__tests__/prefer-mock-return-shorthand.test.ts. tsgo edit shapes and
// branch lock-ins live in the Rstest rule's extras suite, which exercises the
// same shared engine.
package prefer_mock_return_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
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
			{Code: `jest.fn().mockReturnValue(42)`},
			{Code: `jest.fn(() => Promise.resolve(42))`},
			{Code: `jest.fn(() => 42)`},
			{Code: `jest.fn(() => ({}))`},
			{Code: `aVariable.mockImplementation`},
			{Code: `aVariable.mockImplementation()`},
			{Code: `jest.fn().mockImplementation(async () => 1);`},
			{Code: `jest.fn().mockImplementation(async function () {});`},
			{Code: `jest.fn().mockImplementation(async function () {
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
			{Code: `jest.spyOn(Thingy, 'method').mockImplementation(param => param * 2);`},
			{Code: `jest.spyOn(Thingy, 'method').mockImplementation(param => true ? param : 0);`},
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
jest.spyOn(X, getCount).mockImplementation(() => currentX);

// stuff happens

currentX++;

// more stuff happens`},
			{Code: `let currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `let currentX = 0;
currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `var currentX = 0;
currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `var currentX = 0;
var currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`},
			{Code: `let doSomething = () => {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`},
			{Code: `let currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => {
  currentX += 1;

  return currentX;
});`},
			{Code: `const currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => {
  console.log('returning', currentX);

  return currentX;
});`},
			{Code: `let value = 1;

jest.fn().mockImplementation(() => ({ value }));`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => [value]);`},
			{Code: `var value = 1;

aVariable.mockImplementation(() => [0, value, 2]);`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => value + 1);`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => 1 - value);`},
			{Code: `var value = 1;

aVariable.mockImplementation(() => {
  return { value: value + 1 };
});`},
			{Code: `var value = 1;

aVariable.mockImplementation(() => value * value + 1);
aVariable.mockImplementation(() => 1 + value / 2);
aVariable.mockImplementation(() => (1 + value) / 2);
aVariable.mockImplementation(() => {
  return { value: value + 1 };
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(function () {
  return { items: [value] };
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: { value },
  }
});`},
			{Code: `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [{
    type: 'object',
    with: [1, 2, value],
  }]
});`},
			{Code: `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ];
});`},
			{Code: `let value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 3]},
    {type: 'object', with: [1, value]}
  ];
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  }
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => value & 1);
aVariable.mockImplementation(() => value | 1);
aVariable.mockImplementation(() => 1 & value);
aVariable.mockImplementation(() => 1 | value);`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => !value);
aVariable.mockImplementation(() => ~value);
aVariable.mockImplementation(() => typeof value);`},
			{Code: `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx & my);
aVariable.mockImplementation(() => my | mx);`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => value || 0);
aVariable.mockImplementation(() => 1 && value);
aVariable.mockImplementation(() => 1 ?? value);
aVariable.mockImplementation(() => 1 ?? (value && 0));`},
			{Code: `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => my ?? mx);
aVariable.mockImplementation(() => mx ?? (7 && my));`},
			{Code: `let value = [1];

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, ...value],
      },
    },
  }]
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, ...[value]],
      },
    },
  }]
});`},
			{Code: `let obj = {};

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    ...obj,
  }
});`},
			{Code: `let value = 1;

aVariable.mockImplementation(function () {
  function mx() {
    return value;
  }
  return mx();
});`},
			{Code: `let value = 1;

jest.fn().mockImplementation(() => new Mx(value));
jest.fn().mockImplementation(() => new Mx(() => value));
jest.fn().mockImplementation(() => new Mx(() => { return value }));`},
			{Code: `let value = 1;

jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx?.(value));
jest.fn().mockImplementation(() => mx(value).my());
jest.fn().mockImplementation(() => mx(value).my);
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx?.my(value));
jest.fn().mockImplementation(() => mx?.my?.(value));
jest.fn().mockImplementation(() => mx.my?.(value));
jest.fn().mockImplementation(() => mx().my(value));
jest.fn().mockImplementation(() => mx()?.my(value));
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx(value).my(value));
jest.fn().mockImplementation(() => mx?.(value)?.my?.(value));
jest.fn().mockImplementation(() => new Mx().add(value));
jest.fn().mockImplementation(() => {
  return mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }])
});`},
			{Code: `let propName = 'world';

aVariable.mockImplementation(() => mx[propName]());
aVariable.mockImplementation(() => mx[propName]);
aVariable.mockImplementation(() => ({ [propName]: 1 }));`},
			{Code: `const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockImplementation(() => x ? value : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : value ? true : false);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockImplementation(() => true ? true : true ? false : value);

aVariable.mockImplementation(function() {
  if (x) {
    return value;
  } else {
    return 0;
  }
});`},
			// Upstream reports and fixes each of these four; this rule does not.
			// Moving `Promise.reject(...)` into `mockReturnValue` builds the
			// rejected promise when the mock is configured rather than when it is
			// called, so a mock that is never called leaves an unhandled rejection
			// that the test runner reports even though every test passed.
			{Code: `jest.fn().mockImplementation(() => Promise.reject(13))`},
			{Code: `jest.fn().mockImplementation(() => {
  return Promise.reject(13);
})`},
			{Code: `aVariable.mockImplementation(() => Promise.reject(13))`},
			{Code: `aVariable.mockImplementation(() => {
  return Promise.reject(13);
})`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `jest.fn().mockImplementation(() => "hello sunshine")`,
				Output: []string{`jest.fn().mockReturnValue("hello sunshine")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => null)`,
				Output: []string{`jest.fn().mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {
  return null;
})`,
				Output: []string{`jest.fn().mockReturnValue(null)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
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
				Code:   `jest.fn().mockImplementation(() => 0)`,
				Output: []string{`jest.fn().mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {
  return 0;
})`,
				Output: []string{`jest.fn().mockReturnValue(0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
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
				Code:   `jest.fn().mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`jest.fn().mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {
  return Promise.resolve(42);
})`,
				Output: []string{`jest.fn().mockReturnValue(Promise.resolve(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
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
				Code:   `jest.fn().mockImplementation(() => [])`,
				Output: []string{`jest.fn().mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {
  return [];
})`,
				Output: []string{`jest.fn().mockReturnValue([])`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
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
				Code:   `jest.fn().mockImplementation(() => ({}))`,
				Output: []string{`jest.fn().mockReturnValue({})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => x)`,
				Output: []string{`jest.fn().mockReturnValue(x)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => true ? x : y)`,
				Output: []string{`jest.fn().mockReturnValue(true ? x : y)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(function () {
  return "hello world";
})`,
				Output: []string{`jest.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => "hello world")`,
				Output: []string{`jest.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {
  return "hello world";
})`,
				Output: []string{`jest.fn().mockReturnValue("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
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
				Code:   `jest.fn().mockImplementationOnce(() => "hello world")`,
				Output: []string{`jest.fn().mockReturnValueOnce("hello world")`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 1, Column: 11},
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
				Code: `aVariable
  .mockImplementationOnce(() => Promise.reject(42))
  .mockImplementation(() => "hello sunshine")
  .mockReturnValueOnce(Promise.reject(42))`,
				// Upstream also rewrites the first link; see the reversed group in
				// the valid cases above for why the rejected promise stays where
				// it is built.
				Output: []string{`aVariable
  .mockImplementationOnce(() => Promise.reject(42))
  .mockReturnValue("hello sunshine")
  .mockReturnValueOnce(Promise.reject(42))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 4},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => [], xyz)`,
				Output: []string{`jest.fn().mockReturnValue([], xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.spyOn(fs, "readFile").mockImplementation(() => new Error("oh noes!"))`,
				Output: []string{`jest.spyOn(fs, "readFile").mockReturnValue(new Error("oh noes!"))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 1, Column: 28},
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
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
				Output: []string{`const currentX = 0;
jest.spyOn(X, getCount).mockReturnValue(currentX);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 2, Column: 25},
				},
			},
			{
				Code: `import { currentX } from './elsewhere';
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
				Output: []string{`import { currentX } from './elsewhere';
jest.spyOn(X, getCount).mockReturnValue(currentX);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 2, Column: 25},
				},
			},
			{
				Code: `const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    jest.spyOn(X, getCount).mockImplementation(() => currentX);
  });
});`,
				Output: []string{`const currentX = 0;

describe('some tests', () => {
  it('works', () => {
    jest.spyOn(X, getCount).mockReturnValue(currentX);
  });
});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 29},
				},
			},
			{
				Code: `function doSomething() {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`,
				Output: []string{`function doSomething() {};

jest.spyOn(X, getCount).mockReturnValue(doSomething);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 25},
				},
			},
			{
				Code: `const doSomething = () => {};

jest.spyOn(X, getCount).mockImplementation(() => doSomething);`,
				Output: []string{`const doSomething = () => {};

jest.spyOn(X, getCount).mockReturnValue(doSomething);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 25},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => [value]);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue([value]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => [0, value, 2]);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue([0, value, 2]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => [0,, value, 2]);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue([0,, value, 2]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementation(() => ({ value }));`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValue({ value });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => ({ items: [value] }));`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue({ items: [value] });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: { value },
  }
});`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue({
    type: 'object',
    with: { value },
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const vX = 1;
let vY = 1;

getPoint.mockImplementation(() => vX + vY);
getPoint.mockImplementation(() => {
  return { x: vX, y: 1 }
});`,
				Output: []string{`const vX = 1;
let vY = 1;

getPoint.mockImplementation(() => vX + vY);
getPoint.mockReturnValue({ x: vX, y: 1 });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 10},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => value & 0);
aVariable.mockImplementation(() => 0 & value);
aVariable.mockImplementation(() => value | 1);
aVariable.mockImplementation(() => 1 | value);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue(value & 0);
aVariable.mockReturnValue(0 & value);
aVariable.mockReturnValue(value | 1);
aVariable.mockReturnValue(1 | value);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 6, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => ~value);
aVariable.mockImplementation(() => !value);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue(~value);
aVariable.mockReturnValue(!value);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => value + 1);
aVariable.mockImplementation(() => 1 + value);
aVariable.mockImplementation(() => value * value + 1);
aVariable.mockImplementation(() => 1 + value / 2);
aVariable.mockImplementation(() => (1 + value) / 2);`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue(value + 1);
aVariable.mockReturnValue(1 + value);
aVariable.mockReturnValue(value * value + 1);
aVariable.mockReturnValue(1 + value / 2);
aVariable.mockReturnValue((1 + value) / 2);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 6, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 7, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    with: [1, 2, value],
  }
});`,
				Output: []string{`const value = 1;

aVariable.mockReturnValue({
    type: 'object',
    with: [1, 2, value],
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const obj = {};

aVariable.mockImplementation(() => {
  return {
    type: 'object',
    ...obj,
  }
});`,
				Output: []string{`const obj = {};

aVariable.mockReturnValue({
    type: 'object',
    ...obj,
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ];
});`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValueOnce([
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, value]}
  ]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementationOnce(() => {
  return [
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, 0 + value]}
  ];
});`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValueOnce([
    1,
    {type: 'object', with: [1, 2, 3]},
    {type: 'object', with: [1, 2, 0 + value]}
  ]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementationOnce(() => {
  return {
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  }
});`,
				Output: []string{`const value = 1;

aVariable.mockReturnValueOnce({
    type: 'object',
    with: {
      inner: {
        value,
      },
    },
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

aVariable.mockImplementationOnce(() => {
  return {
    type: 'object',
    with: {
      inner: {
        ...{ value },
      },
    },
  }
});`,
				Output: []string{`const value = 1;

aVariable.mockReturnValueOnce({
    type: 'object',
    with: {
      inner: {
        ...{ value },
      },
    },
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValueOnce", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementation(() => {
  return {
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }
});`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValue({
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementation(() => {
  return [{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]
});`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValue([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
				},
			},
			{
				Code: `const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockImplementation(() => mx || 0);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => mx ?? (7 && my));
aVariable.mockImplementation(() => mx ?? (7 && 0));`,
				Output: []string{`const mx = 1
let my = 2;

aVariable.mockImplementation(() => mx || my);
aVariable.mockReturnValue(mx || 0);
aVariable.mockImplementation(() => my && mx);
aVariable.mockImplementation(() => mx ?? (7 && my));
aVariable.mockReturnValue(mx ?? (7 && 0));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 8, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementation(() => new Mx(value));
jest.fn().mockImplementation(() => new Mx(() => value));
jest.fn().mockImplementation(() => new Mx(() => { return value }));`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValue(new Mx(value));
jest.fn().mockReturnValue(new Mx(() => value));
jest.fn().mockReturnValue(new Mx(() => { return value }));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
				},
			},
			{
				Code: `const value = 1;

jest.fn().mockImplementation(() => mx(value));
jest.fn().mockImplementation(() => mx?.(value));
jest.fn().mockImplementation(() => mx().my());
jest.fn().mockImplementation(() => mx().my);
jest.fn().mockImplementation(() => mx.my());
jest.fn().mockImplementation(() => mx?.my());
jest.fn().mockImplementation(() => mx.my);
jest.fn().mockImplementation(() => mx(value).my());
jest.fn().mockImplementation(() => mx(value)?.my());
jest.fn().mockImplementation(() => mx(value).my);
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx().my(value));
jest.fn().mockImplementation(() => mx.my(value));
jest.fn().mockImplementation(() => mx.my?.(value));
jest.fn().mockImplementation(() => mx(value).my(value));
jest.fn().mockImplementation(() => mx?.(value)?.my?.(value));
jest.fn().mockImplementation(() => new Mx().add(value));
jest.fn().mockImplementation(() => {
  return mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }])
});`,
				Output: []string{`const value = 1;

jest.fn().mockReturnValue(mx(value));
jest.fn().mockReturnValue(mx?.(value));
jest.fn().mockReturnValue(mx().my());
jest.fn().mockReturnValue(mx().my);
jest.fn().mockReturnValue(mx.my());
jest.fn().mockReturnValue(mx?.my());
jest.fn().mockReturnValue(mx.my);
jest.fn().mockReturnValue(mx(value).my());
jest.fn().mockReturnValue(mx(value)?.my());
jest.fn().mockReturnValue(mx(value).my);
jest.fn().mockReturnValue(mx.my(value));
jest.fn().mockReturnValue(mx().my(value));
jest.fn().mockReturnValue(mx.my(value));
jest.fn().mockReturnValue(mx.my?.(value));
jest.fn().mockReturnValue(mx(value).my(value));
jest.fn().mockReturnValue(mx?.(value)?.my?.(value));
jest.fn().mockReturnValue(new Mx().add(value));
jest.fn().mockReturnValue(mx([{
    type: 'object',
    with: {
      inner: {
        items: [1, 2, value],
      },
    },
  }]));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 6, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 7, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 8, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 9, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 10, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 11, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 12, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 13, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 14, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 15, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 16, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 17, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 18, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 19, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 20, Column: 11},
				},
			},
			{
				Code: `const propName = 'world';

aVariable.mockImplementation(() => mx[propName]());
aVariable.mockImplementation(() => mx[propName]);
aVariable.mockImplementation(() => ({ [propName]: 1 }));`,
				Output: []string{`const propName = 'world';

aVariable.mockReturnValue(mx[propName]());
aVariable.mockReturnValue(mx[propName]);
aVariable.mockReturnValue({ [propName]: 1 });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 3, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 4, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
				},
			},
			{
				Code: `const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockImplementation(() => x ? true : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockImplementation(() => true ? true : true ? x : false);
aVariable.mockImplementation(() => true ? true : true ? true : false);`,
				Output: []string{`const x = true;
let value = 1;

aVariable.mockImplementation(() => value ? true : false);
aVariable.mockReturnValue(x ? true : false);
aVariable.mockImplementation(() => x ? true : value);
aVariable.mockImplementation(() => true ? true : value);
aVariable.mockImplementation(() => true ? true : true ? value : false);
aVariable.mockReturnValue(true ? true : true ? x : false);
aVariable.mockReturnValue(true ? true : true ? true : false);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 5, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 9, Column: 11},
					{MessageId: "useMockShorthand", Message: "Prefer mockReturnValue", Line: 10, Column: 11},
				},
			},
		},
	)
}
