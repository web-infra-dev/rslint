# prefer-date-now

## Rule Details

Prefer `Date.now()` when reading the current timestamp. This avoids creating
a `Date` object only to convert it into the number of milliseconds since the
Unix Epoch.

Examples of **incorrect** code for this rule:

```javascript
const timestamp = new Date().getTime();
const value = new Date().valueOf();
const number = Number(new Date());
const coerced = +new Date();
const elapsed = new Date() - start;
```

Examples of **correct** code for this rule:

```javascript
const timestamp = Date.now();
const value = Date.now();
const number = Date.now();
const coerced = Date.now();
const elapsed = Date.now() - start;
```

The rule also checks unary minus, `BigInt(new Date())`, and the numeric
operators `-`, `*`, `/`, `%`, and `**`, including their assignment forms.
Automatic fixes preserve the surrounding conversion or arithmetic operation.

Only argument-free `new Date()` expressions are checked. Date construction
with arguments, optional method calls, computed method names, addition, and
bitwise operations are left unchanged. The rule is syntax-based: a locally
declared constructor named `Date` is treated the same as the built-in.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: prefer-date-now](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-date-now.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-date-now.js)
