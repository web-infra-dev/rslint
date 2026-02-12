# await-thenable

## Rule Details

Disallow awaiting a value that is not a Thenable (Promise-like). Using `await` on a non-Promise value is almost always a programmer error and has no effect at runtime, since `await` on a non-Thenable value simply returns it immediately.

This rule also checks `for await...of` loops for non-async iterables and `await using` declarations for non-async disposable values.

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
```

## Original Documentation

- [typescript-eslint await-thenable](https://typescript-eslint.io/rules/await-thenable)
