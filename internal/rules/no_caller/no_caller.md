# no-caller

## Rule Details

Disallows the use of `arguments.caller` and `arguments.callee`. The use of `arguments.caller` and `arguments.callee` make several code optimizations impossible. They have been deprecated in future versions of JavaScript and their use is forbidden in ECMAScript 5 strict mode.

Examples of **incorrect** code for this rule:

```javascript
function foo(n) {
  if (n <= 0) {
    return;
  }
  arguments.callee(n - 1);
}

[1, 2, 3, 4, 5].map(function (n) {
  return !(n > 1) ? 1 : arguments.callee(n - 1) * n;
});
```

Examples of **correct** code for this rule:

```javascript
function foo(n) {
  if (n <= 0) {
    return;
  }
  foo(n - 1);
}

[1, 2, 3, 4, 5].map(function factorial(n) {
  return !(n > 1) ? 1 : factorial(n - 1) * n;
});
```

## Original Documentation

- [ESLint no-caller](https://eslint.org/docs/latest/rules/no-caller)
