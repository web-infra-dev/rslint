# no-xor-as-exponentiation

## Rule Details

In JavaScript, `^` is the [bitwise XOR](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Bitwise_XOR) operator, not exponentiation. Developers coming from languages like Lua, Julia, R, or MATLAB, or from math notation, often expect `^` to mean "to the power of", so `2 ^ 32` silently evaluates to `34` instead of `4294967296`. The actual exponentiation operator is [`**`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Exponentiation).

This rule flags `^` between two decimal integer literals, which is almost always this mistake. Hexadecimal, octal, and binary literals (such as `0xFF ^ 8`) and any non-literal operands (such as `flags ^ MASK`) are ignored, since those are far more likely to be intentional bitwise XOR.

Examples of **incorrect** code for this rule:

```javascript
const bytes = 2 ^ 10; // 8, not 1024
const cube = 3 ^ 3; // 0, not 27
```

Examples of **correct** code for this rule:

```javascript
const bytes = 2 ** 10;
const cube = 3 ** 3;
```
