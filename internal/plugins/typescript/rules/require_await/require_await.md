# require-await

## Rule Details

Disallow async functions which have no `await` expression. Asynchronous functions that do not use `await` might not need to be asynchronous, and may be the unintentional result of refactoring. They also cause a performance penalty by wrapping the function return value in an extra `Promise`.

This rule extends the base ESLint `require-await` rule with TypeScript-specific support: it considers returning a thenable value from an async function as equivalent to using `await`, and also handles `for-await-of` loops, `yield*` on async iterables, and `using`/`await using` declarations.

Examples of **incorrect** code for this rule:

```typescript
async function foo() {
  return 'value';
}
async function bar() {
  doSomethingSync();
}
```

Examples of **correct** code for this rule:

```typescript
async function foo() {
  await doSomethingAsync();
}
async function bar() {
  return await fetchData();
}
function baz() {
  return 'value';
}
```

## Original Documentation

- [typescript-eslint require-await](https://typescript-eslint.io/rules/require-await)
