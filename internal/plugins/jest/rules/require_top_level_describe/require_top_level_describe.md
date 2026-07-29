# require-top-level-describe

## Rule Details

Require test cases (`test` / `it`) and hooks (`beforeAll`, `beforeEach`, `afterEach`, `afterAll`) to be inside a top-level `describe` block. Optionally limit how many top-level `describe` blocks a file may have via `maxNumberOfTopLevelDescribes`.

Examples of **incorrect** code for this rule:

```js
test('my test', () => {});
describe('test suite', () => {
  it('test', () => {});
});

beforeAll(() => {});
describe('test suite', () => {});
```

Examples of **correct** code for this rule:

```js
describe('test suite', () => {
  test('my test', () => {});
});

describe('test suite', () => {
  beforeEach(() => {});
  it('my test', () => {});
});
```

With `{ "maxNumberOfTopLevelDescribes": 1 }`:

```js
// incorrect — second top-level describe
describe('one', () => {});
describe('two', () => {});

// correct
describe('one', () => {
  describe('nested', () => {});
});
```

## Original Documentation

- [jest/require-top-level-describe](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/require-top-level-describe.md)
