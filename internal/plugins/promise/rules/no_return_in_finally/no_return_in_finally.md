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
