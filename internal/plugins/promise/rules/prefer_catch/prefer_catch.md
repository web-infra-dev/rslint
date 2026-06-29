# prefer-catch

Prefer `.catch()` over `.then(null, fn)` or `.then(a, b)` for error handling.

## Rule Details

A `.then()` call with two arguments can obscure that an error handler is present.
The second argument also only catches rejections from earlier in the chain — not
from the first argument — so `.catch()` is both clearer and semantically preferred.

Examples of **incorrect** code for this rule:

```javascript
hey.then(fn1, fn2)
hey.then(null, fn2)
hey.then(undefined, fn2)
```

Examples of **correct** code for this rule:

```javascript
prom.then(fn)
prom.catch(handleErr).then(handle)
prom.catch(handleErr)
```

## Differences from ESLint

Fixed an upstream autofix bug: `x.then(a, b, c)` now fixes to `x.catch(b).then(a, c)` instead of `x.catch(b).then(ac)`.

## Original Documentation

- [eslint-plugin-promise: prefer-catch](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/prefer-catch.md)
