# no-commented-out-tests

## Rule Details

Disallow commenting out Jest tests. Reviewers often skim past comments, so disabled cases can sit in the tree indefinitely. Prefer removing dead tests, extracting helpers, or using `.skip` / `test.todo` when you need an explicit, auditable signal. This is the comment-side complement to `jest/no-disabled-tests`, which reports `skip` / `only` / `todo` on real call sites instead of commented-out text.

rslint walks each comment body line by line: the slice after `//`, or the text inside `/* … */`. If **any** line matches the eslint-plugin-jest-style heuristic—optional `x` or `f` prefix (`xit`, `fit`, …), then `test`, `it`, or `describe`, optional dot- or bracket-member chains (e.g. `.skip`, `.only`, `.concurrent`, `['skip']`), then optional whitespace and `(`—it reports the **entire** comment range with the message “Do not comment out tests”.

Examples of **incorrect** code for this rule:

```javascript
// describe('foo', () => {});
// it('foo', () => {});
// test('foo', () => {});

// describe.skip('foo', () => {});
// it.skip('foo', () => {});
// test.skip('foo', () => {});

// describe['skip']('bar', () => {});
// it['skip']('bar', () => {});
// test['skip']('bar', () => {});

// xdescribe('foo', () => {});
// xit('foo', () => {});
// xtest('foo', () => {});

// it.only('foo', () => {});
// it.concurrent('foo', () => {});
// fit('foo', () => {});

/*
describe('foo', () => {});
*/
```

Examples of **correct** code for this rule:

```javascript
describe('foo', () => {});
it('foo', () => {});
test('foo', () => {});

describe.only('bar', () => {});
it.only('bar', () => {});
test.only('bar', () => {});

// foo('bar', () => {});

// latest(dates)
```

## Limitations

Matching is based on the **literal** shape of test API names inside comment text, not on full parsing of the commented code. It will not flag indirect or renamed patterns, for example:

```javascript
// const testSkip = test.skip;
// testSkip('skipped test', () => {});

// const myTest = test;
// myTest('does not have function body');
```

Because the heuristic treats any `test` / `it` / `describe`-like call opening inside a comment as suspicious, a comment that merely **mentions** that shape (for example documenting an API) can be reported; prefer rephrasing such comments or using examples that do not mirror a call form.

## Original Documentation

- [jest/no-commented-out-tests](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-commented-out-tests.md)
