# no-octal

## Rule Details

Disallows octal literals — integer numerals whose source form starts with a leading
zero followed by another digit (e.g. `071`, `00`, `08`). Octal numeric literals were
deprecated in ECMAScript 5 and are forbidden in strict mode; the leading-zero notation
has also been a long-standing source of confusion (`08` is decimal 8, but `017` is
decimal 15). Prefer the explicit `0o...` octal notation introduced in ES2015.

Examples of **incorrect** code for this rule:

```javascript
const num = 071;
const result = 5 + 07;
const leadingDigit = 08;
const leadingDecimal = 09.1;
```

Examples of **correct** code for this rule:

```javascript
const num = "071";
const hex = 0x1234;
const binary = 0b101;
const modernOctal = 0o17;
const zero = 0;
const decimal = 0.1;
```

## Options

This rule has no options.

## Original Documentation

https://eslint.org/docs/latest/rules/no-octal
