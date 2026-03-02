# no-debugger

## Rule Details

Disallows the use of `debugger` statements. The `debugger` statement is used to tell the JavaScript runtime to pause execution and open a debugging session. These statements should be removed before deploying code to production, as they can halt execution and are only useful during development.

Examples of **incorrect** code for this rule:

```javascript
function check(value) {
  debugger;
  return value > 0;
}
```

Examples of **correct** code for this rule:

```javascript
function check(value) {
  return value > 0;
}
```

## Original Documentation

- [ESLint no-debugger](https://eslint.org/docs/latest/rules/no-debugger)
