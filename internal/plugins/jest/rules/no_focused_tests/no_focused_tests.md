# no-focused-tests

## Rule Details

Disallow focused Jest tests and suites. This rule reports usages like `.only` and focused aliases such as `fdescribe` / `fit`, because they cause only part of the test suite to run and can accidentally hide failing tests in CI or local verification.

Examples of **incorrect** code for this rule:

```javascript
describe.only('foo', () => {});
it.only('foo', () => {});
describe['only']('bar', () => {});
it['only']('bar', () => {});
test.only('foo', () => {});
test['only']('bar', () => {});
fdescribe('foo', () => {});
fit('foo', () => {});
fit.each`
  table
`();
```

Examples of **correct** code for this rule:

```javascript
describe('foo', () => {});
it('foo', () => {});
describe.skip('bar', () => {});
it.skip('bar', () => {});
test('foo', () => {});
test.skip('bar', () => {});
it.each()();
it.each`
  table
`();
test.each()();
test.each`
  table
`();
```

## Original Documentation

- [jest/no-focused-tests](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-focused-tests.md)
