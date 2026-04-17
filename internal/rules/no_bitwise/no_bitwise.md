# no-bitwise

## Rule Details

The use of bitwise operators in JavaScript is very rare and often `&` or `|` is simply a mistyped `&&` or `||`, which will lead to unexpected behavior.

This rule disallows bitwise operators.

Examples of **incorrect** code for this rule:

```javascript
var x = y | z;

var x = y & z;

var x = y ^ z;

var x = ~z;

var x = y << z;

var x = y >> z;

var x = y >>> z;

x |= y;

x &= y;

x ^= y;

x <<= y;

x >>= y;

x >>>= y;
```

Examples of **correct** code for this rule:

```javascript
var x = y || z;

var x = y && z;

var x = y > z;

var x = y < z;

x += y;
```

## Options

This rule supports the following options:

- `allow` (`string[]`): Allows a list of bitwise operators to be used as exceptions.
- `int32Hint` (`boolean`): Allows the use of bitwise OR in `|0` pattern for type casting.

### allow

Examples of **correct** code for this rule with the `{ "allow": ["~"] }` option:

```javascript
/*eslint no-bitwise: ["error", { "allow": ["~"] }] */

~[1, 2, 3].indexOf(1) === -1;
```

### int32Hint

Examples of **correct** code for this rule with the `{ "int32Hint": true }` option:

```javascript
/*eslint no-bitwise: ["error", { "int32Hint": true }] */

var b = a | 0;
```

## Original Documentation

- [ESLint rule: no-bitwise](https://eslint.org/docs/latest/rules/no-bitwise)
