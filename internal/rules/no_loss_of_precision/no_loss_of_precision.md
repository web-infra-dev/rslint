# no-loss-of-precision

Disallow literal numbers that lose precision.

## Rule Details

This rule prevents the use of number literals that lose precision when converted to a JavaScript `Number` due to 64-bit floating-point rounding.

JavaScript Numbers are represented as 64-bit floating-point values according to the IEEE 754 standard. This means they can only accurately represent integers up to `Number.MAX_SAFE_INTEGER` (2^53 - 1 = 9007199254740991). Numbers exceeding this limit, or decimal numbers with too many significant digits, will lose precision at runtime.

Examples of **incorrect** code for this rule:

```javascript
// Integers exceeding MAX_SAFE_INTEGER
var x = 9007199254740993;

// Very large integers with precision loss
var x = 5123000000000000000000000000001;

// Decimals with too many significant digits
var x = 1.0000000000000000000000123;

// Scientific notation causing precision loss
var x = 9.007199254740993e15;

// Binary, octal, hex exceeding safe limits
var x = 0x20000000000001;
var x = 0o400000000000000001;
var x = 0b100000000000000000000000000000000000000000000000000001;

// Numbers that become Infinity
var x = 2e999;
```

Examples of **correct** code for this rule:

```javascript
// Safe integers
var x = 12345;
var x = 9007199254740991; // MAX_SAFE_INTEGER

// Safe decimals
var x = 123.456;
var x = 0.00000000000000000000000123;

// Safe scientific notation
var x = 123e34;

// Safe binary, octal, hex
var x = 0x1fffffffffffff;
var x = 0o377777777777777777;
var x = 0b11111111111111111111111111111111111111111111111111111;

// Using numeric separators (ES2021)
var x = 9007_1992547409_91;
```

## Options

This rule has no options.

## When Not To Use It

If you don't mind the precision loss in certain numeric literals, you can disable this rule.

## Original Documentation

[ESLint no-loss-of-precision](https://eslint.org/docs/latest/rules/no-loss-of-precision)
