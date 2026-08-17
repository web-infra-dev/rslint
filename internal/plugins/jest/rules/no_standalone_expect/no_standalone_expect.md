# no-standalone-expect

## Rule Details

Disallow using `expect` outside of `it` or `test` blocks. This rule reports `expect` calls that sit directly in a `describe` block, at module scope, or in other places where Jest will not run them as part of a test case. That helps catch assertions that look meaningful but never execute.

`expect` inside a helper function is allowed, even when the helper is defined outside the `it`/`test` callback, because the assertion still runs when the helper is invoked from a test. Static `expect` APIs such as `expect.any()` and `expect.extend()` at module scope are also allowed.

Examples of **incorrect** code for this rule:

```javascript
describe('a test', () => {
  expect(1).toBe(1);
});

describe('a test', () => {
  it('an it', () => {
    expect(1).toBe(1);
  });

  expect(1).toBe(1);
});

expect(1).toBe(1);

expect.hasAssertions();
```

Examples of **correct** code for this rule:

```javascript
describe('a test', () => {
  it('an it', () => {
    expect(1).toBe(1);
  });
});

describe('a test', () => {
  const helper = () => {
    expect(1).toBe(1);
  };

  it('an it', () => {
    helper();
  });
});

expect.any(String);
expect.extend({});
```

## Options

- First argument (optional): object with `additionalTestBlockFunctions`
  - `additionalTestBlockFunctions`: array of function names that should also be treated as test blocks (for example `each.test`).

## Differences from ESLint

rslint treats method, constructor, getter, and setter bodies as helper function
scopes. An `expect` inside one of those bodies is therefore allowed, just like
an assertion inside a function declaration or arrow-function helper. The
upstream rule reports method and accessor bodies.

rslint also balances the test scope opened by every recognized registration.
As a result, a standalone assertion after a chained registration such as
`test.only('case', () => {})` is still reported. The upstream rule can leave the
registration scope open and miss that assertion.

Every `expect.<modifier>.<matcher>()` chain is treated as a static value
constructor and allowed outside a test block. That covers the asymmetric matcher
constructors jest supports, such as `expect.not.stringContaining('value')`, but
also chains that assert rather than build a value, such as
`expect.resolves.toBe(1)` and `expect.rejects.toThrow()`. The upstream rule
allows a one-member chain only and reports all of these. `expect.assertions()`
and `expect.hasAssertions()` are reported here as well.

## Original Documentation

- [eslint-plugin-jest: no-standalone-expect](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-standalone-expect.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-standalone-expect.ts)
