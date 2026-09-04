# no-null

## Rule Details

Disallow the `null` literal and encourage using `undefined` as the single value
for missing data.

Examples of **incorrect** code for this rule:

```javascript
let value = null;

if (value == null) {}
```

Examples of **correct** code for this rule:

```javascript
let value;

if (value == undefined) {}

const dictionary = Object.create(null);
```

Strict equality comparisons are ignored by default:

```javascript
if (value === null) {}
```

## Options

This rule accepts an object with two boolean options:

- `checkArguments` (default: `true`) checks `null` when it is a direct function
  call or constructor argument. Set it to `false` when APIs require `null`
  arguments.
- `checkStrictEquality` (default: `false`) checks strict equality and inequality
  comparisons against `null`.

The rule follows Unicorn's built-in exceptions for `Object.create(null)`,
`useRef(null)`, `React.useRef(null)`, and `null` as the second argument of a
two-argument `insertBefore` call.

## Original Documentation

- [eslint-plugin-unicorn: no-null](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-null.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-null.js)
