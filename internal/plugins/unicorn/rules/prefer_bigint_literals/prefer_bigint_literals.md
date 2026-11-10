# prefer-bigint-literals

## Rule details

Prefer a bigint literal when the argument to the built-in `BigInt()` function
is an integer number or a string that can be converted to a bigint. The rule
checks quoted strings, templates without substitutions, and numeric literals
with unary `+` or `-`.

Examples of **incorrect** code for this rule

```javascript
const bigint = BigInt(1);
const large = BigInt('9007199254740991');
const hex = BigInt('0xFF');
```

Examples of **correct** code for this rule

```javascript
const bigint = 1n;
const large = 9007199254740991n;
const hex = 0xFFn;
const computed = BigInt(getSomeNumber());
const astronomicalNumber = 999999999999999999999999n;
```

The rule offers editor suggestions for calls that contain comments or an
explicit `+` sign. It also offers a suggestion when a numeric argument needs a
different spelling as a bigint literal, as in `BigInt(1e2)` becoming `100n`.

## Options

This rule has no options.

## Differences from upstream

- Rslint skips calls where a local binding or a `with` statement can supply
  another `BigInt` function. It also skips calls with TypeScript type arguments
  or when `languageOptions.globals` disables `BigInt`.
- Keep the call when the configured ECMAScript version predates ES2020. Rslint
  also skips TypeScript files with a build target older than ES2020.
- `BigInt(1000000000000000100)` suggests `1000000000000000128n` to preserve the
  value of the call after JavaScript rounds the numeric argument.
- Fixes add spaces where compact code would otherwise run together. For example,
  `BigInt(1)in object` becomes `1n in object`.
- Rslint rejects the legacy spellings in `BigInt(0777)` and `BigInt(0888)` even
  in scripts. Write `511n` and `888n` for those values.

## Original documentation

- [eslint-plugin-unicorn: prefer-bigint-literals](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-bigint-literals.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-bigint-literals.js)
