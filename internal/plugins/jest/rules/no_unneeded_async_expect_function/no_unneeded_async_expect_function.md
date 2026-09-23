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

rslint still reports the wrapper but does not fix it when the unwrapped call
would not mean the same thing: when the awaited call contains another `await`
of the wrapper's own, such as `await run(await load())`; when the wrapper has
parameters, type parameters or a function-expression name; when a function
expression wrapper reads `this`, `arguments` or `new.target`; when `expect()`
has an explicit type argument; or when a comment inside the wrapper sits
outside the awaited call and would be deleted.

## Original Documentation

- [eslint-plugin-jest: no-unneeded-async-expect-function](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-unneeded-async-expect-function.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-unneeded-async-expect-function.ts)
