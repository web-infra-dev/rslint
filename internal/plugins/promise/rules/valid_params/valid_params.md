# valid-params

Enforce the proper number of arguments passed to Promise functions (`promise/valid-params`).

Calling a Promise function with the incorrect number of arguments can lead to unexpected behavior or hard to spot bugs.

## Rule Details

This rule reports calls to Promise functions with an invalid number of arguments:

- `Promise.resolve()` and `Promise.reject()` require 0 or 1 arguments.
- `Promise.all()`, `Promise.race()`, `Promise.allSettled()`, and `Promise.any()` require 1 argument.
- `.then()` requires 1 or 2 arguments.
- `.catch()` and `.finally()` require 1 argument.

Examples of incorrect code for this rule:

```js
Promise.all()
Promise.race(a, b)
Promise.resolve(a, b)
Promise.reject(error, extra)
promise.then()
promise.then(a, b, c)
promise.catch()
promise.finally(a, b)
```

Examples of correct code for this rule:

```js
Promise.all([p1, p2])
Promise.race(iterable)
Promise.resolve(value)
Promise.reject(error)
promise.then(onFulfilled)
promise.then(onFulfilled, onRejected)
promise.catch(onRejected)
promise.finally(onFinally)
```

## Options

### `exclude`

An array of method names to skip. For example, projects using Bluebird-style filtered catches can exclude `catch`:

    {
      "promise/valid-params": ["warn", { "exclude": ["catch"] }]
    }

## Differences from ESLint

None known.

## Original Documentation

- [eslint-plugin-promise: valid-params](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/valid-params.md)
