# no-standalone-expect

## Rule Details

Disallow using `expect` outside of `it` or `test` blocks. This rule reports `expect` calls that sit directly in a `describe` block, at module scope, or in other places where Rstest will not run them as part of a test case. This helps catch assertions that look meaningful but never execute as part of a test.

`expect` inside a helper function is allowed, even when the helper is defined outside the `it`/`test` callback, because the assertion still runs when the helper is invoked from a test. Static `expect` APIs such as `expect.any()` and `expect.extend()` at module scope are also allowed.

Only the callback body of a registration counts as the test case. A test's name, its options object and any other argument are evaluated while Rstest collects the file, so an assertion written there runs before any case starts and is reported — including on a registration that takes no callback at all, such as `test.todo(expect(1).toBe(1))`. A tagged-template registration is the exception: `` test.each`...`(name, callback) `` counts as a test block in its entirety, so an assertion in an interpolation or in an argument of the call the tag returns is allowed.

Examples of **incorrect** code for this rule:

```javascript
describe('a test', () => {
  expect(1).toBe(1);
});

describe('a test', () => {
  test('a test case', () => {
    expect(1).toBe(1);
  });

  expect(1).toBe(1);
});

expect(1).toBe(1);
expect.hasAssertions();

test.todo(expect(1).toBe(1));

test('a test case', { timeout: expect(1).toBe(1) }, () => {});
```

Examples of **correct** code for this rule:

```javascript
describe('a test', () => {
  test('a test case', () => {
    expect(1).toBe(1);
  });
});

describe('a test', () => {
  const helper = () => {
    expect(1).toBe(1);
  };

  test('a test case', () => {
    helper();
  });
});

expect.any(String);
expect.extend({});
```

## Options

- First argument (optional): object with `additionalTestBlockFunctions`
  - `additionalTestBlockFunctions`: array of function names that should also be treated as test blocks (for example `each.test`).
