# no-disabled-tests

## Rule Details

Disallow disabled or incomplete Jest tests. This rule reports skipped suites/tests via `.skip` and `x*` aliases, disallows `pending()` in test bodies, and flags `it()` / `test()` calls that omit the callback function (except `test.todo(...)`). It helps prevent accidentally committing tests that are skipped or not actually executed.

Examples of **incorrect** code for this rule:

```javascript
describe.skip('foo', () => {});
it.skip('foo', () => {});
test.skip('foo', () => {});

describe['skip']('bar', () => {});
it['skip']('bar', () => {});
test['skip']('bar', () => {});

xdescribe('foo', () => {});
xit('foo', () => {});
xtest('foo', () => {});

it('bar');
test('bar');

it('foo', () => {
  pending();
});
```

Examples of **correct** code for this rule:

```javascript
describe('foo', () => {});
it('foo', () => {});
test('foo', () => {});

describe.only('bar', () => {});
it.only('bar', () => {});
test.only('bar', () => {});
```

## Limitations

The plugin looks at the literal function names within test code, so will not catch more complex examples of disabled tests, such as:

```javascript
const testSkip = test.skip;
testSkip('skipped test', () => {});

const myTest = test;
myTest('does not have function body');
```

## Original Documentation

- [jest/no-disabled-tests](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-disabled-tests.md)
