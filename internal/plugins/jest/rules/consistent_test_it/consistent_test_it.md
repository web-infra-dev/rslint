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

The diagnostics are the same as upstream. The fix differs in three ways.

When a test call has more than one member after its name, the fix keeps all of them. Upstream drops every member except the last:

```js
/* rslint jest/consistent-test-it: ["error", { "fn": "test" }] */
it.only.each([1])('foo', (n) => {});
// rslint:   test.only.each([1])('foo', (n) => {});
// upstream: test.each([1])('foo', (n) => {});
```

When the call comes from an `@jest/globals` import, the fix also imports the new name, or reuses an existing import of it. Upstream calls the new name without importing it, which fails when `injectGlobals` is `false`:

```js
/* rslint jest/consistent-test-it: ["error", { "fn": "test" }] */
import { it } from '@jest/globals';
it('foo');
// rslint:   import { test, it } from '@jest/globals'; test('foo');
// upstream: import { it } from '@jest/globals'; test('foo');
```

No fix is offered when the new name would not reach the Jest API: when a local variable, parameter or other declaration with that name is in scope, when the file already declares it some other way (such as an import from another module), when the call comes from a `let` or `var` binding of a CommonJS `require`, which may have been reassigned, or when it comes from any other `require` binding and the file does not already import the new name. Upstream calls whatever that name refers to:

```js
/* rslint jest/consistent-test-it: ["error", { "fn": "test" }] */
function run() {
  const test = () => {};
  it('foo'); // reported, not fixed; upstream fixes it to call the local `test`
}
```

## Original Documentation

- [eslint-plugin-jest: consistent-test-it](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/consistent-test-it.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/consistent-test-it.ts)
