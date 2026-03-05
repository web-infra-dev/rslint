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

## Original Documentation

- [typescript-eslint use-unknown-in-catch-callback-variable](https://typescript-eslint.io/rules/use-unknown-in-catch-callback-variable)
