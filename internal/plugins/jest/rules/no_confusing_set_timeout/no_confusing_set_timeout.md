# no-confusing-set-timeout

## Rule Details

Disallow confusing usages of `jest.setTimeout`. In a single test file Jest applies only the **last** `jest.setTimeout` call that runs **before** any tests execute; later calls and calls inside suites or cases do not change the timeout the way many authors expect. This rule flags patterns that look file- or suite-specific but are misleading.

rslint walks each Jest API call site (including `jest` imported from `@jest/globals` and renamed bindings such as `Jest.setTimeout`). For every `jest.setTimeout` call it may report:

- **`globalSetTimeout`**: the call is not at module/global top level (for example inside a `describe` / `test` / `it` callback, a `beforeEach` body, a block statement, or a class).
- **`orderSetTimeout`**: another Jest API (`describe`, `test`, `it`, hooks, `expect`, and so on) appears **earlier** in the same file.
- **`multipleSetTimeouts`**: `jest.setTimeout` is invoked more than once in the file (only the last pre-test call matters to Jest).

Plain `setTimeout` and `window.setTimeout` are not checked.

Examples of **incorrect** code for this rule:

```javascript
describe('test foo', () => {
  jest.setTimeout(1000);
  it('test-description', () => {
    // test logic
  });
});

describe('test bar', () => {
  it('test-description', () => {
    jest.setTimeout(1000);
    // test logic
  });
});

test('foo-bar', () => {
  jest.setTimeout(1000);
});

describe('unit test', () => {
  beforeEach(() => {
    jest.setTimeout(1000);
  });
});

jest.setTimeout(1000);
describe('suite', () => {
  it('case', () => {});
});
jest.setTimeout(800);

jest.setTimeout(800);
jest.setTimeout(900);

import { jest } from '@jest/globals';
{
  jest.setTimeout(800);
}
```

Examples of **correct** code for this rule:

```javascript
jest.setTimeout(500);
test('test test', () => {
  // do some stuff
});
```

```javascript
jest.setTimeout(1000);
describe('test bar bar', () => {
  it('test-description', () => {
    // test logic
  });
});
```

```javascript
jest.setTimeout(1000);
window.setTimeout(60000);
setTimeout(1000);
```

## Original Documentation

- [jest/no-confusing-set-timeout](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-confusing-set-timeout.md)
