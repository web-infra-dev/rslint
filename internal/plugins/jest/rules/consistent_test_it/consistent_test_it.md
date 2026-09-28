# consistent-test-it

## Rule Details

Jest allows you to choose how you want to define your tests, using the `it` or the `test` keywords, with multiple permutations for each:

- **it:** `it`, `xit`, `fit`, `it.only`, `it.skip`.
- **test:** `test`, `xtest`, `test.only`, `test.skip`.

This rule gives you control over the usage of these keywords in your codebase. By default, it requires `test` for top-level tests and `it` for tests nested within a `describe`. It is fixable.

## Options

This rule accepts an object with the following properties:

- `fn` (`"it"` or `"test"`): decides whether to use `test` or `it`.
- `withinDescribe` (`"it"` or `"test"`): decides whether to use `test` or `it` within a `describe` scope. When omitted, it takes the value of `fn`, or `"it"` if `fn` is not set either.

```json
{
  "jest/consistent-test-it": ["error", { "fn": "test" }]
}
```

```js
test('foo'); // valid
test.only('foo'); // valid

it('foo'); // invalid
it.only('foo'); // invalid
```

```json
{
  "jest/consistent-test-it": ["error", { "fn": "it" }]
}
```

```js
it('foo'); // valid
it.only('foo'); // valid

test('foo'); // invalid
test.only('foo'); // invalid
```

```json
{
  "jest/consistent-test-it": ["error", { "fn": "it", "withinDescribe": "test" }]
}
```

```js
it('foo'); // valid
describe('foo', function () {
  test('bar'); // valid
});

test('foo'); // invalid
describe('foo', function () {
  it('bar'); // invalid
});
```

With the default configuration:

```js
test('foo'); // valid
describe('foo', function () {
  it('bar'); // valid
});

it('foo'); // invalid
describe('foo', function () {
  test('bar'); // invalid
});
```

## Differences from upstream

When a test call has more than one member after its name, the fix keeps all of them. Upstream drops every member except the last:

```js
/* rslint jest/consistent-test-it: ["error", { "fn": "test" }] */
it.only.each([1])('foo', (n) => {});
// rslint:   test.only.each([1])('foo', (n) => {});
// upstream: test.each([1])('foo', (n) => {});
```

The diagnostics are the same as upstream.

## Original Documentation

- [eslint-plugin-jest: consistent-test-it](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/consistent-test-it.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/consistent-test-it.ts)
