# prefer-exponentiation-operator

## Rule Details

This rule disallows calls to `Math.pow` and suggests using the `**` operator
instead.

Examples of **incorrect** code for this rule:

```javascript
const foo = Math.pow(2, 8);
const bar = Math.pow(a, b);
let baz = Math.pow(a + b, c + d);
let quux = Math.pow(-1, n);
```

Examples of **correct** code for this rule:

```javascript
const foo = 2 ** 8;
const bar = a ** b;
let baz = (a + b) ** (c + d);
let quux = (-1) ** n;
```

## When Not To Use It

Do not enable this rule if your runtime target does not support the
exponentiation operator (`**`).

## Original Documentation

- [ESLint prefer-exponentiation-operator](https://eslint.org/docs/latest/rules/prefer-exponentiation-operator)
