# no-invalid-regexp

## Rule Details

Disallows invalid regular expression strings in `RegExp` constructors. This rule validates patterns and flags passed to `new RegExp(pattern, flags)` and `RegExp(pattern, flags)` when the arguments are string literals.

Valid flags are: `d`, `g`, `i`, `m`, `s`, `u`, `v`, `y`.

Examples of **incorrect** code for this rule:

```javascript
RegExp('.', 'z'); // invalid flag 'z'
new RegExp('.', 'aa'); // duplicate flag 'a'
RegExp('.', 'uv'); // 'u' and 'v' flags are mutually exclusive
RegExp('['); // unterminated character class
RegExp('('); // unterminated group
new RegExp('\\'); // trailing backslash
```

Examples of **correct** code for this rule:

```javascript
RegExp('.');
new RegExp('.', 'im');
new RegExp('.', 'gmi');
new RegExp(pattern, 'g'); // non-literal pattern, skipped
new RegExp('.', flags); // non-literal flags, skipped
```

## Options

### `allowConstructorFlags`

An array or string of additional flags to allow in RegExp constructors. For example, to allow the `a` and `z` flags:

```json
{ "allowConstructorFlags": "az" }
```

## Original Documentation

- [ESLint no-invalid-regexp](https://eslint.org/docs/latest/rules/no-invalid-regexp)
