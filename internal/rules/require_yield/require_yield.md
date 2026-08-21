# require-yield

## Rule Details

This rule generates warnings for generator functions that do not have the `yield` keyword.

Examples of **incorrect** code for this rule:

```javascript
function* foo() {
  return 10;
}
```

Examples of **correct** code for this rule:

```javascript
function* foo() {
  yield 5;
  return 10;
}

function foo() {
  return 10;
}

// This rule does not warn on empty generator functions.
function* foo() {}
```

## When Not To Use It

If you don't want to notify generator functions that have no `yield` expression, then it's safe to disable this rule.

## Original Documentation

- [ESLint: require-yield](https://eslint.org/docs/latest/rules/require-yield)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/require-yield.js)
