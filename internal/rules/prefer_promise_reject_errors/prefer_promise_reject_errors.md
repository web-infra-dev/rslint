# prefer-promise-reject-errors

## Rule Details

This rule requires that Promises are rejected with `Error` objects only. Rejecting with non-`Error` values (such as strings or numbers) makes debugging harder because the rejection reason does not carry a stack trace.

The rule flags two patterns:

- `Promise.reject(value)` calls where `value` cannot be an `Error`.
- `new Promise((resolve, reject) => { ... })` executors where the second parameter is invoked with a value that cannot be an `Error`.

Examples of **incorrect** code for this rule:

```javascript
Promise.reject("something bad happened");
Promise.reject(5);
Promise.reject();

new Promise(function (resolve, reject) {
  reject("something bad happened");
});

new Promise(function (resolve, reject) {
  reject();
});
```

Examples of **correct** code for this rule:

```javascript
Promise.reject(new Error("something bad happened"));
Promise.reject(new TypeError("something bad happened"));

new Promise(function (resolve, reject) {
  reject(new Error("something bad happened"));
});

const foo = getUnknownValue();
Promise.reject(foo);
```

## Options

This rule accepts an options object with the following property:

- `allowEmptyReject` (`boolean`, default `false`) — when `true`, allows calls to `Promise.reject()` and the executor's reject callback with no arguments.

Examples of **correct** code for this rule with `{ "allowEmptyReject": true }`:

```json
{ "prefer-promise-reject-errors": ["error", { "allowEmptyReject": true }] }
```

```javascript
Promise.reject();

new Promise(function (resolve, reject) {
  reject();
});
```

## Original Documentation

- [ESLint: prefer-promise-reject-errors](https://eslint.org/docs/latest/rules/prefer-promise-reject-errors)
