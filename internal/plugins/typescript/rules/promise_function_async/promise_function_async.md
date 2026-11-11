# promise-function-async

## Rule Details

Require any function or method that returns a Promise to be marked async. Ensures that each function is only capable of either returning a rejected promise or throwing an `Error` object. In contrast, non-async Promise-returning functions are technically capable of either. Code that handles both rejected promises and thrown errors simultaneously is often overly complex and hard to maintain.

Examples of **incorrect** code for this rule:

```typescript
function foo(): Promise<string> {
  return Promise.resolve('value');
}
const bar = (): Promise<number> => Promise.resolve(42);
class Baz {
  method(): Promise<void> {
    return Promise.resolve();
  }
}
```

Examples of **correct** code for this rule:

```typescript
async function foo(): Promise<string> {
  return 'value';
}
const bar = async (): Promise<number> => 42;
class Baz {
  async method(): Promise<void> {}
}
```

## Differences from upstream

Autofixes preserve token separation when adding `async` after a keyword. For
example, `return()=>Promise.resolve(1)` becomes
`return async ()=>Promise.resolve(1)`, rather than merging `return` and `async`.
The same safeguard applies after `throw` and `yield`, including generic arrows.

## Original Documentation

- [typescript-eslint: promise-function-async](https://typescript-eslint.io/rules/promise-function-async)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.29.1/packages/eslint-plugin/src/rules/promise-function-async.ts)
