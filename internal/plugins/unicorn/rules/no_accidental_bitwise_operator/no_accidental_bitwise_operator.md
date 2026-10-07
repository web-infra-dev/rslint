# no-accidental-bitwise-operator

## Rule Details

Disallow high-signal uses of bitwise operators where a logical operator was likely intended.

The rule reports three typo patterns:

- `object & object.property`, which is usually meant to short-circuit as `object && object.property`.
- `value | nonNumericFallback`, such as `options | {}`, which is usually meant to be `options || {}`.
- `value |= nonNumericFallback`, such as `input |= ''`, which is usually meant to be `input ||= ''`.

Examples of **incorrect** code:

```javascript
if (object & object.property) {}
options = options | {};
input |= '';
```

Examples of **correct** code:

```javascript
if (object && object.property) {}
options = options || {};
input ||= '';

const masked = flags & MASK;
const truncated = value | 0;
```

Because replacing a bitwise operator changes runtime behavior, the rule offers an editor suggestion instead of an automatic fix.

Unlike ESLint's broad `no-bitwise` rule, this rule intentionally leaves ordinary bitwise arithmetic alone.

## Original Documentation

- [eslint-plugin-unicorn: no-accidental-bitwise-operator](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-accidental-bitwise-operator.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-accidental-bitwise-operator.js)
