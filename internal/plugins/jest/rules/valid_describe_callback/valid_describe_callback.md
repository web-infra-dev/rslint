# valid-describe-callback

## Rule Details

Enforce valid `describe()` callback usage in Jest. A `describe` block should include both a suite name and a callback function, the callback must not be `async`, should not take parameters (except in `describe.each(...)`), and should not return a value.

Examples of **incorrect** code for this rule:

```javascript
describe("suite");

describe("suite", "not a function");

describe("suite", async () => {
  await setup();
});

describe("suite", (done) => {
  done();
});

describe("suite", () => test("case", () => {}));
```

Examples of **correct** code for this rule:

```javascript
describe("suite", () => {
  test("case", () => {
    expect(true).toBe(true);
  });
});

describe.each([1, 2, 3])("value %s", (value) => {
  test("is truthy", () => {
    expect(value).toBeTruthy();
  });
});
```

## Original Documentation

- [jest/valid-describe-callback](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/valid-describe-callback.md)
