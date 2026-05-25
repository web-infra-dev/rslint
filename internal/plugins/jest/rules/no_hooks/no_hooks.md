# no-hooks

## Rule Details

Disallow Jest lifecycle hooks (`beforeEach`, `afterEach`, `beforeAll`, `afterAll`). This rule helps enforce tests that are isolated and explicit, instead of relying on shared setup/teardown behavior that can make test order and failures harder to reason about.

Examples of **incorrect** code for this rule:

```javascript
beforeEach(() => {
  setupDatabase();
});

afterAll(() => {
  cleanup();
});
```

Examples of **correct** code for this rule:

```javascript
test("works with explicit setup", () => {
  const db = createTestDatabase();
  expect(runWith(db)).toBe(true);
});

describe("suite", () => {
  test("case", () => {
    expect(1 + 1).toBe(2);
  });
});
```

## Options

- First argument (optional): object with `allow`
  - `allow`: array of hook names that are allowed. Supported values: `beforeEach`, `afterEach`, `beforeAll`, `afterAll`.

## Original Documentation

- [jest/no-hooks](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-hooks.md)
