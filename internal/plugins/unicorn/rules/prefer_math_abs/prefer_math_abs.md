# prefer-math-abs

## Rule details

Prefer `Math.abs()` for manual absolute value expressions and checks that
compare a value against both a limit and its negation. The rule recognizes
negation written as `-value`, `0 - value`, or multiplication by `-1`. It
skips operands it recognizes as BigInt because `Math.abs()` throws on BigInt.

The rule offers editor suggestions. Review them if your code depends on signed
zero or non-number operands. For example, with `value = -0`, the expression
`value < 0 ? -value : value` yields `-0` and `Math.abs(value)` yields `0`.
`Math.abs()` also coerces its operand to a number. Suggestions are withheld
when the rewrite would remove comments or change how often a member is read.

### Incorrect

```javascript
const absolute = value < 0 ? -value : value;
const magnitude = value <= 0 ? 0 - value : value;
const distance = value > 0 ? value : -value;

if (number > MAXIMUM || number < -MAXIMUM) {
}
```

### Correct

```javascript
const absolute = Math.abs(value);

if (Math.abs(number) > MAXIMUM) {
}
```

## Options

This rule has no options.

## Differences from upstream

- Rslint reports the expression and withholds the suggestion if a local
  binding shadows `Math` or `languageOptions.globals` disables it.
- Rslint skips additional cases where an operand is known to be a BigInt. For
  example, it skips a check on a property inferred to be `bigint` from
  `const state = { offset: 1n }` or a value cast with JSDoc
  `/** @type {bigint} */`.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-abs.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-abs.js)
