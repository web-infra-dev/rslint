# no-export

## Rule Details

Disallow exporting from files that contain Jest tests or suites. If a file has at least one `test` or `describe` (including equivalent aliases and chained forms), this rule reports any export in that file.

Exporting from a test file is risky because importing that file runs its tests again in every consumer. That can duplicate test runs, slow down suites, and make failures harder to trace. Move shared helpers into a separate non-test module instead.

This rule checks:

- ES module exports (`export const`, `export default`, `export =`, and other `export` forms)
- CommonJS-style assignments rooted at `module.exports` (including `module["exports"]` and deeply nested properties)

Locally declared variables or parameters named `module` are not treated as the CommonJS global. Files that export but contain no Jest tests or suites are allowed.

Examples of **incorrect** code for this rule:

```javascript
export function myHelper() {}

module.exports = function () {};

module.exports = {
  something: 'that should be moved to a non-test file',
};

describe('a test', () => {
  expect(1).toBe(1);
});
```

Examples of **correct** code for this rule:

```javascript
function myHelper() {}

const myThing = {
  something: 'that can live here',
};

describe('a test', () => {
  expect(1).toBe(1);
});
```

## When Not To Use It

Do not enable this rule on files that are not Jest test files. For shared test utilities that must export helpers, either disable the rule for those files or keep helpers in a dedicated module that does not define tests.

## Original Documentation

- [jest/no-export](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-export.md)
