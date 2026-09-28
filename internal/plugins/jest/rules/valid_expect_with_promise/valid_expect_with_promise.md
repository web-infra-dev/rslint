# valid-expect-with-promise

## Rule Details

Require that `resolves` and `rejects` modifiers are present (and only) for promise-like types.

When working with promises, you must remember to use `resolves` and `rejects` to assert on the value returned (or thrown) by the promise, rather than the promise itself. Inversely, while Jest does not prevent you from using `resolves` and `rejects` on non-promise values, it is not necessary.

This rule requires type information and reports when:

- an `expect` is given a promise-like value but without `resolves` or `rejects`
- an `expect` is not given a promise-like value, but is used with `resolves` or `rejects`

Examples of **incorrect** code for this rule:

```ts
expect('hello world').resolves.toBe('hello sunshine');

expect(new Promise(r => r(0))).toThrow('oh noes!');
```

Examples of **correct** code for this rule:

```ts
expect('hello world').toBe('hello sunshine');

expect(new Promise(r => r(0))).rejects.toThrow('oh noes!');
```

## Options

- First argument (optional): object with `checkThenables`
  - `checkThenables`: also treat values whose `then` method accepts both a fulfillment and a rejection callback as promises, not just `Promise` itself. Default is `false`.

A [thenable](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise#thenables) is an object with a `then` method, such as a `Promise`, TypeScript's built-in `PromiseLike` interface, or any custom object that has a `.then()`. This option is useful when working with older `Promise` polyfills instead of the native `Promise` class, or when the global `Promise` is not declared by TypeScript's default libraries.

Examples of **incorrect** code with `{ "checkThenables": true }`:

```json
{ "jest/valid-expect-with-promise": ["error", { "checkThenables": true }] }
```

```ts
declare function createPromiseLike(): PromiseLike<string>;

expect(createPromiseLike()).toBe('hello sunshine');

interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

expect(createMyThenable()).toBe('hello sunshine');
```

Examples of **correct** code with `{ "checkThenables": true }`:

```json
{ "jest/valid-expect-with-promise": ["error", { "checkThenables": true }] }
```

```ts
declare function createPromiseLike(): PromiseLike<string>;

await expect(createPromiseLike()).resolves.toBe('hello sunshine');

interface MyThenable {
  then(onFulfilled: () => void, onRejected: () => void): MyThenable;
}

declare function createMyThenable(): MyThenable;

await expect(createMyThenable()).resolves.toBe('hello sunshine');
```

## Original Documentation

- [eslint-plugin-jest: valid-expect-with-promise](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/valid-expect-with-promise.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/valid-expect-with-promise.ts)
