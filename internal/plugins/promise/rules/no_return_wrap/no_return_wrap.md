# no-return-wrap

Disallow wrapping values in `Promise.resolve` or `Promise.reject` when not needed.

## Rule Details

Inside a `.then()`, `.catch()`, or `.finally()` callback, returning a raw value
already resolves the promise chain with that value, and throwing already rejects
it. This rule reports callbacks that return `Promise.resolve(...)` or
`Promise.reject(...)` instead of using the direct value or `throw`.

Examples of **incorrect** code for this rule:

```javascript
promise.then(function (value) {
  return Promise.resolve(value * 2)
})

promise.catch(function (error) {
  return Promise.reject(error)
})

promise.then(() => Promise.resolve(42))
```

Examples of **correct** code for this rule:

```javascript
promise.then(function (value) {
  return value * 2
})

promise.catch(function (error) {
  throw error
})

promise.then(() => 42)
```

## Options

### `allowReject`

Pass `{ "allowReject": true }` to allow returning `Promise.reject(...)` from a
promise callback. `Promise.resolve(...)` remains disallowed.

```json
{ "promise/no-return-wrap": ["error", { "allowReject": true }] }
```

```javascript
promise.catch(function (error) {
  return Promise.reject(error)
})
```

## Differences from ESLint

None known.

## Original Documentation

- [eslint-plugin-promise: no-return-wrap](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-return-wrap.md)
