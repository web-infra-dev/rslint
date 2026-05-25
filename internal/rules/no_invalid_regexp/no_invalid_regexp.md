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

## Known Limitations

The following ECMAScript regex features are not yet fully supported in pattern validation:

- Unicode property long names (`\p{Letter}`) and `Script=` syntax (`\p{Script=Latin}`)
- `v`-flag set notation (`[A--B]`, `[A&&B]`, `[A--[0-9]]`)
- Surrogate pair named capture groups (`(?<\ud835\udc9c>.)`)
- Invalid escape detection in unicode mode (`\a` with `u` flag)
- `v`-flag specific parsing (`[[]` with `v` flag)
- Duplicate named capture groups outside alternatives (`(?<k>a)(?<k>b)`)
- Inline modifier validation (`(?ii:foo)`, `(?-:foo)`, `(?-u:foo)`)

Flag validation (invalid flags, duplicate flags, `u`/`v` conflict) and `allowConstructorFlags` are fully aligned with ESLint.

## Original Documentation

- [ESLint no-invalid-regexp](https://eslint.org/docs/latest/rules/no-invalid-regexp)
