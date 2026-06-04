# always-return

## Rule Details

Require returning inside each `then()` to create readable and reusable Promise chains.

Inside a `.then()` you must `return` a new promise or value, or `throw` an error.
Callbacks that call `process.exit()` or `process.abort()` are also considered as terminating.

Examples of **incorrect** code for this rule:

```javascript
hey.then(function(val) {
  doSomething(val)
})

hey.then(function(val) {
  if (val) {
    return val
  }
})
```

Examples of **correct** code for this rule:

```javascript
hey.then(function(val) {
  return val * 2
})

hey.then(function(val) {
  if (!val) {
    throw new Error('no val')
  }
  return val
})

hey.then(x => x * 2)
```

Examples of **correct** code for this rule with `{ "ignoreLastCallback": true }`:

```json
{ "promise/always-return": ["error", { "ignoreLastCallback": true }] }
```

```javascript
// last in a chain — result is not used further
hey.then(function(val) {
  doSomething(val)
})

hey
  .then(function(val) { return transform(val) })
  .then(function(val) { doSomething(val) })
```

Examples of **correct** code for this rule with `{ "ignoreAssignmentVariable": ["globalThis", "window"] }`:

```json
{ "promise/always-return": ["error", { "ignoreAssignmentVariable": ["globalThis", "window"] }] }
```

```javascript
hey.then(function(val) {
  globalThis.result = val
})

hey.then(function(val) {
  window.result = val
})
```

## Original Documentation

https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/always-return.md
