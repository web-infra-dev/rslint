# prefer-promise-reject-errors

## Rule Details

Require using Error objects as Promise rejection reasons. When rejecting a Promise, it is best practice to reject with an `Error` object, because `Error` objects store a stack trace, making it much easier to debug by determining where the error came from. This rule reports calls to `Promise.reject()` and the `reject` parameter in `new Promise((resolve, reject) => ...)` when they are called with a non-Error value.

Examples of **incorrect** code for this rule:

```typescript
Promise.reject('error');
Promise.reject(0);
new Promise((resolve, reject) => reject('error'));
```

Examples of **correct** code for this rule:

```typescript
Promise.reject(new Error('error'));
new Promise((resolve, reject) => reject(new Error('error')));
Promise.reject(unknownVariable);
```

## Original Documentation

- [typescript-eslint prefer-promise-reject-errors](https://typescript-eslint.io/rules/prefer-promise-reject-errors)
