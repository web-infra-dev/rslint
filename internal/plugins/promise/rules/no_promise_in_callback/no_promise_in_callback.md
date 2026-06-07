# no-promise-in-callback

Disallow using promises inside callbacks.

## Rule Details

Promises and callbacks are different ways to handle asynchronous code. Mixing the
styles inside the same callback can make error handling unclear: callback errors
are usually passed through the callback argument, while promise errors must be
handled through the promise chain.

This rule reports promise-like calls inside callback-like functions whose first
parameter is named `err` or `error`.

Examples of **incorrect** code for this rule:

```javascript
doSomething((err, value) => {
  if (err) {
    console.error(err)
  } else {
    doSomethingElse(value).then(console.log)
  }
})
```

```javascript
function handler(err) {
  Promise.resolve(err)
}
```

Examples of **correct** code for this rule:

```javascript
promisify(doSomething)()
  .then(doSomethingElse)
  .then(console.log)
  .catch(console.error)
```

```javascript
doSomething((err, value) => {
  return doSomethingElse(value).then(console.log)
})
```

Promise callbacks passed to `.then()` or `.catch()` are not treated as callback
containers for this rule.

## Options

### `exemptDeclarations`

Pass `{ "exemptDeclarations": true }` to exempt function declarations. Defaults
to `false`.

```json
{ "promise/no-promise-in-callback": ["warn", { "exemptDeclarations": true }] }
```

With this option enabled, this code is valid:

```javascript
function handler(err) {
  Promise.resolve(err)
}
```

Function expressions and arrow functions are still checked when this option is
enabled.

## Differences from ESLint

None known.

## Original Documentation

- [eslint-plugin-promise: no-promise-in-callback](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-promise-in-callback.md)
