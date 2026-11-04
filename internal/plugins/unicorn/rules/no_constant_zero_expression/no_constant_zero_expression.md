# no-constant-zero-expression

## Rule Details

Disallow multiplication or bitwise AND with the numeric literal `0`, and
division with `0` as the numerator. These operations often indicate a typo
or leftover debugging code. Other zero literal forms, such as `0.0` and
`0x0`, are also checked.

Examples of **incorrect** code:

```javascript
const total = price * 0;
const flags = mask & 0;
const ratio = 0 / count;
```

Examples of **correct** code:

```javascript
const total = 0;
const flags = 0;
const ratio = 0;
```

## Suggestions

The rule suggests replacing an expression with `0` only when it can safely
evaluate the whole expression to positive zero and the expression contains no
comments. Negative zero, `NaN`, unknown values, side effects and mutable
member reads prevent the suggestion. The rule does not apply automatic fixes.

Expressions such as `x * -0`, `x * 0n`, and `x * (0 as number)` are ignored.
Division by zero (`x / 0`) and other operators are outside this rule's scope.

## Options

This rule has no options.

## Differences from upstream

Suggestions preserve spacing next to keywords and TypeScript operators. For
example, `return(5)*0` becomes `return 0`, and `5*(0)as number` becomes
`0 as number`.

rslint withholds a suggestion when it cannot establish the value of an operand
discarded by a comma expression or `void`. For example, `(1n / 0n, 5) * 0`
and `(void missing) & 0` are reported without a suggestion, preserving their
`RangeError` and `ReferenceError`. Known operands such as `(1, 5) * 0` and
`(void 5) & 0` still receive suggestions.

## Original Documentation

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-constant-zero-expression.md)
- [Upstream rule source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-constant-zero-expression.js)
