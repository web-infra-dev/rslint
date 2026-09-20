# no-zero-fractions

## Rule Details

Remove trailing zero fractions and dangling decimal points from numeric literals. For example, `1.0` becomes `1`, `1.010` becomes `1.01`, and `1.` becomes `1`. Numeric separators and exponent spelling are preserved outside the removed fraction.

The rule provides an automatic fix. It preserves required parentheses, keyword spacing, and statement boundaries when changing a numeric member receiver.

Examples of **incorrect** code:

```javascript
const value = 123_456.000_000;
const scale = 123.00e20;
const text = 1.00.toString();
```

Examples of **correct** code:

```javascript
const value = 123_456;
const scale = 123e20;
const text = (1).toString();
```

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-zero-fractions](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-zero-fractions.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-zero-fractions.js)
