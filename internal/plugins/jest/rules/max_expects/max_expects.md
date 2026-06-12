# max-expects

## Rule Details

Enforce a maximum number of `expect()` calls in a test body. As more assertions are added, a test is more likely to mix multiple objectives. This rule reports when a single test callback exceeds the configured limit.

The rule counts top-level `expect()` calls inside each `test` or `it` callback (including `async` callbacks and forms such as `test.each` and `it.each` that the Jest integration recognizes). The counter resets when entering a new test case. Nested `expect()` calls used as matchers (for example `expect.any(Boolean)` inside `toEqual`) and static `expect` APIs such as `expect.hasAssertions()` are not counted.

`expect` calls inside nested functions within a test (for example a helper arrow function defined in the callback) are counted toward that test's limit. `expect` calls in standalone helper functions defined outside the test callback are not attributed to the test body.

Examples of **incorrect** code for this rule (with the default `{ "max": 5 }`):

```javascript
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

it('should not pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

describe('test', () => {
  test('should not pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});
```

Examples of **correct** code for this rule (with the default `{ "max": 5 }`):

```javascript
test('should pass');

test('should pass', () => {});

test.skip('should pass', () => {});

test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

test('should pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toEqual(expect.any(Boolean));
});

function myHelper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}

test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  myHelper();
});
```

## Options

- First argument (optional): object with `max`
  - `max`: maximum allowed `expect()` calls per test callback. Default is `5`.

Examples of **correct** code with `{ "max": 10 }`:

```javascript
test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});
```

Examples of **incorrect** code with `{ "max": 1 }`:

```javascript
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});
```

## Original Documentation

- [jest/max-expects](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/max-expects.md)
