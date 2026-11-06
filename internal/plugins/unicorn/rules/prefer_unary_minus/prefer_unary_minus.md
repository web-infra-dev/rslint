# prefer-unary-minus

## Rule Details

Prefer the unary minus operator over multiplying or dividing by `-1`.

The rule reports `value * -1`, `-1 * value`, and `value / -1`.
Division in the opposite direction (`-1 / value`) and expressions where both
operands are `-1` are left alone. A BigInt multiplier or divisor such as `-1n`
does not match the numeric literal `-1`.

Examples of **incorrect** code for this rule:

```javascript
const inverted = value * -1;
const flipped = -1 * someNumber;
const divided = value / -1;
```

Examples of **correct** code for this rule:

```javascript
const inverted = -value;
const flipped = -someNumber;
const divided = -value;
```

## Fixes

The fix negates the other operand, adding parentheses and spacing where needed.
Comments inside that operand are preserved. If the fix would remove a comment
around the operator or `-1`, the rule reports the expression without a fix.

## Options

This rule has no options.

## Differences from upstream

To lint a TypeScript type assertion followed by `**`, put parentheses around
the assertion. For example, write `(value * -1 as any) ** 2`. The
unparenthesized form `value * -1 as any ** 2` produces a syntax error instead
of this rule's diagnostic. The parenthesized form is checked and fixed normally.

## Original Documentation

- [eslint-plugin-unicorn: prefer-unary-minus](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-unary-minus.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-unary-minus.js)
