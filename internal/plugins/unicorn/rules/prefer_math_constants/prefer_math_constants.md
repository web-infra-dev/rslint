# prefer-math-constants

## Rule details

Prefer named `Math` constants when a decimal literal matches a rounded or
truncated value. The rule checks `Math.PI`, `Math.E`, `Math.LN2`, `Math.LN10`,
`Math.LOG2E`, `Math.LOG10E`, `Math.SQRT2`, and `Math.SQRT1_2`.

The minimum number of fractional digits is two for `Math.PI`, three for
`Math.E`, and four for the other constants. The rule counts the written digits
and ignores numeric separators. It skips exponential notation, non-computed
property names, and TypeScript literal types.

Replacing an approximation can change a calculation, so the rule offers an
editor suggestion for the user to apply.

### Incorrect

```javascript
const area = 3.14 * radius ** 2;
const growth = amount * 2.718;
const diagonal = side * 1.4142;
```

### Correct

```javascript
const area = Math.PI * radius ** 2;
const growth = amount * Math.E;
const diagonal = side * Math.SQRT2;
```

## Options

This rule has no options.

## Differences from upstream

- Rslint adds whitespace when a suggestion would join adjacent tokens. For
  example, `return.6931` becomes `return Math.LN2`.
- If a local binding shadows `Math` or `languageOptions.globals` disables it,
  Rslint reports the literal and withholds the suggestion.
- Rslint withholds suggestions in places where TypeScript requires a literal
  or constant expression. This includes const assertions such as `3.14 as const`,
  enum initializers that determine an implicit member value or belong to a
  `const enum`, declarations such as `declare const angle = 3.14`, and computed
  property names in interfaces.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-constants.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-constants.js)
