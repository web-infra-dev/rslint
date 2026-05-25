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

This rule accepts a single object option with the following properties:

### `allow`

- Type: `string[]` (subset of `"|"`, `"&"`, `"^"`, `"<<"`, `">>"`, `">>>"`, `"|="`, `"&="`, `"^="`, `"<<="`, `">>="`, `">>>="`, `"~"`)
- Default: `[]`

Whitelists the listed bitwise operators as exceptions. Only operators exactly matching a string in this list are allowed; all others still report.

Example configuration:

```json
{ "no-bitwise": ["error", { "allow": ["~"] }] }
```

Examples of **correct** code with the above configuration:

```javascript
~[1, 2, 3].indexOf(1) === -1;
```

### `int32Hint`

- Type: `boolean`
- Default: `false`

When `true`, permits the `x | 0` idiom commonly used to coerce a number to a 32-bit integer. Only `|` with a literal `0` on the right-hand side is allowed — `0 | x`, `x & 0`, `x | 1`, `x | -0`, and `x | 0n` still report.

Example configuration:

```json
{ "no-bitwise": ["error", { "int32Hint": true }] }
```

Examples of **correct** code with the above configuration:

```javascript
var b = a | 0;
```

## Original Documentation

- [ESLint rule: no-bitwise](https://eslint.org/docs/latest/rules/no-bitwise)
