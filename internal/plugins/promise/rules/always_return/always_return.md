# always-return

Require returning inside each `then()` to create readable and reusable Promise chains.

## Rule Details

This rule checks inline callbacks passed as the first argument to `.then()`. A
block-bodied callback must explicitly `return`, `throw`, call `process.exit()`,
or call `process.abort()` on every path. Expression-bodied arrow callbacks are
allowed because their expression is implicitly returned.

Examples of **incorrect** code for this rule:

```javascript
promise.then(function (value) {})

promise.then((value) => {
  doSomething(value)
})

promise.then((value) => {
  if (value) {
    return value
  }
})
```

Examples of **correct** code for this rule:

```javascript
promise.then((value) => value * 2)

promise.then(function (value) {
  return value * 2
})

promise.then((value) => {
  if (!value) {
    throw new Error('missing value')
  }
  return value
})
```

## Options

### `ignoreLastCallback`

Pass `{ "ignoreLastCallback": true }` to allow the last `.then()` callback in a
promise chain to omit a `return`. This is useful when the last callback only
performs side effects. Default is `false`.

```json
{ "promise/always-return": ["error", { "ignoreLastCallback": true }] }
```

```javascript
promise.then((value) => {
  console.log(value)
})

promise
  .then((value) => {
    console.log(value) // still incorrect: not the last callback
  })
  .then((value) => {
    console.log(value) // correct with ignoreLastCallback
  })
```

### `ignoreAssignmentVariable`

Pass `{ "ignoreAssignmentVariable": ["globalThis"] }` to allow the last
`.then()` callback to omit a `return` when it assigns to one of the configured
root variables. Default is `["globalThis"]`.

```json
{ "promise/always-return": ["error", { "ignoreAssignmentVariable": ["globalThis", "window"] }] }
```

```javascript
promise.then((value) => {
  globalThis.result = value
})

promise.then((value) => {
  window.result = value
})
```

## Differences from ESLint

None known.

## Original Documentation

- [eslint-plugin-promise: always-return](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/always-return.md)
