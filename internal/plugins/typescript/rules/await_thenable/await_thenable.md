# await-thenable

## Rule Details

Disallow awaiting a value that is not a Thenable (Promise-like). Using `await` on a non-Promise value is almost always a programmer error and has no effect at runtime, since `await` on a non-Thenable value simply returns it immediately.

This rule also checks `for await...of` loops for non-async iterables and `await using` declarations for non-async disposable values.

It checks inputs to `Promise.all`, `Promise.allSettled`, `Promise.any`, and `Promise.race`, even when the call is not awaited. It reports array literal elements that are always non-Thenable, and arrays, tuples, or other iterable references whose value types can contain non-Thenables.

Examples of **incorrect** code for this rule:

```typescript
async function foo() {
  await 42;
}

async function bar(x: number) {
  await x;
}

async function baz(arr: number[]) {
  for await (const item of arr) {
  }
}

Promise.all([Promise.resolve(1), 42]);
```

Examples of **correct** code for this rule:

```typescript
async function foo() {
  await Promise.resolve(42);
}

async function bar(x: Promise<number>) {
  await x;
}

async function baz(iter: AsyncIterable<number>) {
  for await (const item of iter) {
  }
}

Promise.all([Promise.resolve(1), Promise.resolve(42)]);
```

## Differences from upstream

Rslint can resolve a computed aggregator name read directly from a constant object, such as `const names = Object.freeze({ aggregate: 'all' }); Promise[names.aggregate]([1]);`, and reports the non-Thenable element. With typescript-eslint v8.71.0 and eslint-utils v4.10.1, upstream's static evaluator recursively revisits this computed access while checking object mutations and gives up, so it does not report this example.

## Original Documentation

- [typescript-eslint: await-thenable](https://typescript-eslint.io/rules/await-thenable)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.71.0/packages/eslint-plugin/src/rules/await-thenable.ts)
