# prefer-math-trunc

## Rule details

Prefer `Math.trunc()` to remove the fractional part of a number.

This rule reports `x | 0`, `x >> 0`, `x << 0`, `x ^ 0`, their compound assignments,
and `~~x`. It also reports `parseInt(String(x), 10)` and `Number.parseInt(String(x), 10)`
when the calls refer to the global built-ins.

Examples of **incorrect** code for this rule

```javascript
value | 0;
value >> 0;
value << 0;
value ^ 0;
value |= 0;
~~value;
parseInt(String(value), 10);
Number.parseInt(String(value), 10);
```

Examples of **correct** code for this rule

```javascript
Math.trunc(value);
value = Math.trunc(value);
value | 3;
~value;
```

The rule offers editor suggestions only. Bitwise operators convert numbers to
32-bit integers and can wrap large values. Parsing a string can truncate at the
exponent marker when `String()` uses exponential notation.

The rule withholds suggestions when the replacement would remove a comment or
when a compound assignment would repeat a target with side effects.

## Options

This rule has no options.

## Differences from upstream

Rslint reports the truncation without a suggestion when a local binding shadows
`Math` or `languageOptions.globals` disables it. The suggestion requires the
built-in `Math` object.

```javascript
function truncate(Math, value) {
  return ~~value;
}
```

Rslint also omits the suggestion for a compound assignment whose target calls a
tagged template. The rewrite would call the tag twice. To rewrite this example,
compute the index once and use the result in the assignment.

```javascript
values[key`index`] |= 0;
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-trunc.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-trunc.js)
