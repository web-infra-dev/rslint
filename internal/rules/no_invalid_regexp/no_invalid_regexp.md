# no-invalid-regexp

## Rule Details

Disallows invalid regular expression strings in `RegExp` constructors. This rule validates patterns and flags passed to `new RegExp(pattern, flags)` and `RegExp(pattern, flags)` when the arguments are string literals. Only references to the global `RegExp` constructor are checked; shadowed bindings and disabled globals are ignored.

Valid flags are: `d`, `g`, `i`, `m`, `s`, `u`, `v`, `y`.

Examples of **incorrect** code for this rule:

```javascript
RegExp('.', 'z'); // invalid flag 'z'
new RegExp('.', 'gg'); // duplicate flag 'g'
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
new RegExp(pattern, 'g'); // valid flags; pattern is unknown
new RegExp('.', flags); // pattern is valid for every possible flag mode
```

## Options

### `allowConstructorFlags`

An array of additional flags to allow in RegExp constructors. For example, to allow the `a` and `z` flags:

```json
{ "allowConstructorFlags": ["a", "z"] }
```

## Pattern Validation

Patterns are checked using ECMAScript syntax independently of the matching engine. This includes Unicode property escapes, `v`-flag set notation and string properties, Unicode capture group names, duplicate named groups in separate alternatives, and inline modifiers.

When flags are unknown, a pattern is reported only if it is invalid in all three modes: non-Unicode, `u`, and `v`. Unknown patterns still receive flag validation when flags are string literals.

This rule provides no autofixes or suggestions.

## Original Documentation

- [ESLint: no-invalid-regexp](https://eslint.org/docs/latest/rules/no-invalid-regexp)
- [Source code](https://github.com/eslint/eslint/blob/v10.14.0/lib/rules/no-invalid-regexp.js)
