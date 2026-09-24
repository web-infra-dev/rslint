# prefer-expect-assertions

## Rule Details

Ensure every test has either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression. Assertions inside callbacks, loops, and promise handlers may never run, and a test whose assertions are all skipped still passes; declaring the count makes Jest fail it.

The rule also reports `expect.hasAssertions()` called with arguments and `expect.assertions()` called with anything other than a single integer number literal. A declaration that a `beforeEach` or `afterEach` callback makes on every run covers every test of the same `describe` block and its nested blocks: it must be a top-level statement of the callback, outside those same skippable positions, that no earlier `return` or `throw` can skip. Only tests whose callback is a function written at the registration are checked.

Examples of **incorrect** code for this rule:

```js
test('my test', () => {
  expect.assertions('1');
  expect(someThing()).toEqual('foo');
});

test('my test', () => {
  expect(someThing()).toEqual('foo');
});
```

Examples of **correct** code for this rule:

```js
test('my test', () => {
  expect.assertions(1);
  expect(someThing()).toEqual('foo');
});

test('my test', () => {
  expect.hasAssertions();
  expect(someThing()).toEqual('foo');
});

describe('users', () => {
  beforeEach(() => {
    expect.hasAssertions();
  });

  test('responds ok', () => {
    client.get('/user', (response) => {
      expect(response.status).toBe(200);
    });
  });
});
```

## Options

```json
{
  "jest/prefer-expect-assertions": [
    "warn",
    { "onlyFunctionsWithAsyncKeyword": true }
  ]
}
```

- `onlyFunctionsWithAsyncKeyword` (default `false`): only check tests whose callback uses the `async` keyword.
- `onlyFunctionsWithExpectInLoop` (default `false`): only check tests that call `expect` inside a native loop of the test callback.
- `onlyFunctionsWithExpectInCallback` (default `false`): only check tests that call `expect` inside a function nested in the test callback.

With none of these enabled, every test is checked. When any is enabled, a test is checked if it matches at least one of them.

## Differences from upstream

Hooks cover tests by suite rather than by source order. Jest collects every hook of a `describe` block before running its tests, so a hook declared below a test still applies to it:

```javascript
describe('users', () => {
  // Not reported here; upstream reports it.
  test('responds ok', () => {
    client.get('/user', (response) => expect(response.status).toBe(200));
  });

  beforeEach(() => {
    expect.hasAssertions();
  });
});
```

A hook counts only for calls its callback makes directly, but that callback may also be a function declared in the same file or `expect.hasAssertions` passed by reference, as in `beforeEach(expect.hasAssertions)`. A call in a timer or another nested function started by the hook no longer counts, since it may run after Jest has checked the test, and neither does one the hook skips on some runs. `expect.assertions(n)` in a hook covers tests as `expect.hasAssertions()` does.

The first expression must belong to the test callback itself and always run: a declaration in the branch of an `if`, on one side of `&&`, `||`, `??`, or `?:`, after `?.`, or in a destructuring default does not count, while directives such as `'use strict'` before it are allowed. A matcher that happens to be named `assertions`, as in `expect(value).assertions(1)`, is not a declaration.

For the options, a loop or nested function counts only when it sits between the `expect` call and the test callback, so a test registered inside a loop is not treated as asserting in one. `while` and `do...while` loops count as loops, and methods count as nested functions. A test that satisfies the rule no longer exempts the next test when an option skips the first one.

## Original Documentation

- [eslint-plugin-jest: prefer-expect-assertions](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/prefer-expect-assertions.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/prefer-expect-assertions.ts)
