# no-commented-out-tests

## Rule Details

Disallow commenting out Jest tests. Reviewers often skim past comments, so disabled cases can sit in the tree indefinitely. Prefer removing dead tests, extracting helpers, or using `.skip` / `test.todo` when you need an explicit, auditable signal. This is the comment-side complement to `jest/no-disabled-tests`, which reports `skip` / `only` / `todo` on real call sites instead of commented-out text.

rslint rebuilds each comment, joining adjacent `//` lines, and parses it as code using the file's language (TypeScript or TSX). It reports the **whole** comment when a statement in it is a complete `test`, `it` or `describe` registration, with the message “Do not comment out tests”. Recognized forms are the plain call, the legacy `xit` / `xtest` / `fit` / `xdescribe` / `fdescribe` aliases, member chains such as `.skip`, `.only`, `.concurrent`, `.failing` and `['skip']`, and `.each` tables (array or tagged template, with optional type arguments) once they are called with a title.

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

Only calls whose callee is written with the Jest names are recognized. Indirect or renamed patterns are not flagged, for example:

```javascript
// const testSkip = test.skip;
// testSkip('skipped test', () => {});

// const myTest = test;
// myTest('does not have function body');
```

## Differences from upstream

eslint-plugin-jest matches each comment line against a regular expression. rslint parses the commented code instead, which changes these edge cases:

- Text that only looks like a call is no longer reported, such as `// it (see docs)`, `// test("foo") should be preferred`, or a call cut short by prose such as `// test(` or a lone `// it('foo', () => {` opener that is never closed in the comment.
- `// test.each(rows)` without the call that registers a test is not reported, since nothing is registered yet.
- `/** ... */` documentation comments and commented-out Markdown code fences are not reported.
- More real registrations are reported: chains such as `test.only.each(rows)(...)`, `.each` tagged templates, `*`-led lines inside a block comment, and a leading `;` or `?.` call.
- Only real Jest globals are roots. A name the regular expression matched by accident, such as an `f`-prefixed `test`, is no longer treated as one.
- A member that is not a known Jest modifier is still accepted directly after the root, as upstream does, e.g. `// test.someNewMethod()`.

## Original Documentation

- [eslint-plugin-jest: no-commented-out-tests](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-commented-out-tests.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-commented-out-tests.ts)
