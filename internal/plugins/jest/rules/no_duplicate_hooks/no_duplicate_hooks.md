# no-duplicate-hooks

## Rule Details

Disallow duplicate Jest lifecycle hooks **in the same scope**. Each `describe` block (including `describe.skip`, `describe.each`, and tagged `describe.each`) opens a new scope; registering the same hook name twice (`beforeEach`, `afterEach`, `beforeAll`, or `afterAll`) reports the **second and later** calls.

- **Same scope**: hooks directly in a `describe` callback, or anywhere still inside that `describe` while it is active (including inside nested `test` / `it` bodies).
- **Separate scopes**: nested `describe` blocks; sibling `describe` blocks; file top level (hooks outside any `describe` share one scope).
- **Allowed**: one of each hook type in the same block; the same hook name again in a child or sibling `describe`.
- **Imports**: `@jest/globals` hooks and renamed bindings (e.g. `afterEach as somethingElse`) count toward the same hook name.

Examples of **incorrect** code for this rule:

```javascript
describe('foo', () => {
  beforeEach(() => {
    // some setup
  });
  beforeEach(() => {
    // some setup
  });
  test('foo_test', () => {
    // some test
  });
});

// Nested describe scenario
describe('foo', () => {
  beforeEach(() => {
    // some setup
  });
  test('foo_test', () => {
    // some test
  });
  describe('bar', () => {
    test('bar_test', () => {
      afterAll(() => {
        // some teardown
      });
      afterAll(() => {
        // some teardown
      });
    });
  });
});
```

Examples of **correct** code for this rule:

```javascript
describe('foo', () => {
  beforeEach(() => {
    // some setup
  });
  test('foo_test', () => {
    // some test
  });
});

// Nested describe scenario
describe('foo', () => {
  beforeEach(() => {
    // some setup
  });
  test('foo_test', () => {
    // some test
  });
  describe('bar', () => {
    test('bar_test', () => {
      beforeEach(() => {
        // some setup
      });
    });
  });
});
```

## Original Documentation

- [jest/no-duplicate-hooks](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-duplicate-hooks.md)
