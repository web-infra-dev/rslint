# promise/no-return-in-finally

Disallow return statements inside a callback passed to `.finally()`, since nothing would consume what's returned.

## Rule Details

Examples of **incorrect** code for this rule:

```javascript
myPromise.finally(function (val) {
  return val
})

myPromise.finally(() => {
  return 2
})
```

Examples of **correct** code for this rule:

```javascript
myPromise.finally(function (val) {
  console.log('value:', val)
})

myPromise.finally(() => {})
```

## Original Documentation

https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-return-in-finally.md

## ESLint Parity and Compatibility

To align with the behavior of `eslint-plugin-promise`'s `no-return-in-finally` rule, the following behaviors are intentionally preserved:

1. **Nested Return Statements**: ESLint's rule only checks top-level statements directly inside the finally callback's block body. It ignores nested return statements (e.g., inside an `if` statement block, a `try/catch` block, or any nested block). rslint mirrors this behavior and only flags top-level returns directly under the main function body block.
2. **Implicit Arrow Returns**: ESLint does not flag implicit returns in arrow functions without a block body (e.g., `promise.finally(() => 2)`). rslint also does not flag these implicit returns.

