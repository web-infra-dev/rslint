# use-unknown-in-catch-callback-variable

## Rule Details

Enforce typing the rejection callback parameter of `.catch()` and `.then()` as `unknown`. Similar to how catch clause variables should be typed as `unknown` (since any value can be thrown), Promise rejection callback parameters should also be typed as `unknown`. This prevents unsafe property accesses and assumptions about the shape of the rejection reason.

The rule checks `.catch(callback)` and `.then(onFulfilled, onRejected)` calls on thenable types and flags the rejection callback parameter when it is not typed as `unknown`.

Examples of **incorrect** code for this rule:

```typescript
promise.catch(err => {});
promise.catch((err: Error) => {});
promise.then(undefined, err => {});
promise.then(undefined, (err: string) => {});
```

Examples of **correct** code for this rule:

```typescript
promise.catch((err: unknown) => {});
promise.then(undefined, (err: unknown) => {});
promise.catch((...args: [unknown]) => {});
```

## Differences from Upstream

Rslint avoids evaluating array and string spreads with the default iterator when
the corresponding prototype may have been modified. typescript-eslint v8.67.0
can still use the default iterator and mistake a fulfillment callback for a
rejection callback in this situation:

```typescript
Array.prototype[Symbol.iterator] = function* () {
  yield 'then';
  return undefined;
};
const names = [...['catch']];
const key = names[0];
Promise.resolve(1)[key as 'then'](
  (value: number) => value + 1,
  (error: unknown) => {},
);
```

The call executes `then`, so Rslint does not suggest changing `value` to `unknown`.
The same safeguard applies to a modified `String.prototype[Symbol.iterator]`.

## Original Documentation

- [typescript-eslint: use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/use-unknown-in-catch-callback-variable.ts)
