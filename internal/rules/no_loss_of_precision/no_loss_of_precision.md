# no-loss-of-precision

## Rule Details

Disallow number literals that lose precision at runtime when converted to a JavaScript `Number`.

JavaScript numbers are stored as double-precision floating-point values. If a number literal contains more significant digits than the runtime value can preserve, the literal may evaluate to a different number than the source text suggests.

Examples of **incorrect** code for this rule:

```javascript
const a = 9007199254740993;
const b = 5123000000000000000000000000001;
const c = 1230000000000000000000000.0;
const d = .1230000000000000000000000;
const e = 0x20000000000001;
const f = 0x2_000000000_0001;
```

Examples of **correct** code for this rule:

```javascript
const a = 12345;
const b = 123.456;
const c = 123e34;
const d = 12300000000000000000000000;
const e = 0x1fffffffffffff;
const f = 9007199254740991;
const g = 9007_1992547409_91;
```

## Options

This rule has no options.

## Original Documentation

- [ESLint no-loss-of-precision](https://eslint.org/docs/latest/rules/no-loss-of-precision)
