# no-unneeded-async-expect-function

## Rule Details

Disallow wrapping an expected promise in an unnecessary async function when
using Jest promise assertions.

Jest promise assertions can receive the promise directly:
`await expect(doSomethingAsync()).rejects.toThrow()` or
`await expect(doSomethingAsync()).resolves.toBe(value)`. Wrapping that call in
`async () => { await doSomethingAsync(); }` is more verbose and makes the test
harder to read without changing the assertion.

This rule reports `expect()` calls whose first argument is an async function
with a single awaited call expression. It is fixable: the async wrapper is
replaced with the awaited call. Renamed `expect` bindings imported from
`@jest/globals` are also recognized.

Examples of **incorrect** code for this rule:

```js
it('wrong1', async () => {
  await expect(async () => {
    await doSomethingAsync();
  }).rejects.toThrow();
});

it('wrong2', async () => {
  await expect(async function () {
    await doSomethingAsync();
  }).rejects.toThrow();
});
```

Examples of **correct** code for this rule:

```js
it('right1', async () => {
  await expect(doSomethingAsync()).rejects.toThrow();
});
```

## Differences from ESLint

rslint also fixes equivalent concise arrow functions and parenthesized async
function arguments, such as `expect(async () => await doSomethingAsync())` and
`expect((async () => { await doSomethingAsync(); }))`. These shapes are handled
as the same safe unwrap because tsgo preserves them explicitly in the AST.

## Original Documentation

- [jest/no-unneeded-async-expect-function](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-unneeded-async-expect-function.md)
