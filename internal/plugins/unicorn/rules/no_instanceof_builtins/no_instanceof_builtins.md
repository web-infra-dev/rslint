# no-instanceof-builtins

## Rule Details

Disallow `instanceof` with built-in objects. `instanceof` can be unreliable across realms, so safer checks such as `typeof`, `Array.isArray()`, `Object.prototype.toString.call()`, or dedicated type helpers are preferred.

This rule can automatically fix `Array`, `Function`, and `Error` checks when `useErrorIsError` is enabled. Primitive wrapper checks such as `String`, `Number`, and `Boolean` provide suggestions instead of automatic fixes because converting object-wrapper values to `typeof` checks can change behavior.

Examples of **incorrect** code for this rule:

```javascript
foo instanceof String;
foo instanceof Number;
foo instanceof Array;
foo instanceof Function;
```

Examples of **correct** code for this rule:

```javascript
typeof foo === "string";
typeof foo === "number";
Array.isArray(foo);
typeof foo === "function";
```

Examples of **incorrect** code for this rule with `{ "strategy": "strict" }`:

```json
{ "unicorn/no-instanceof-builtins": ["error", { "strategy": "strict" }] }
```

```javascript
foo instanceof Map;
foo instanceof Error;
foo instanceof Date;
```

Examples of **incorrect** code for this rule with `{ "include": ["HTMLElement"] }`:

```json
{ "unicorn/no-instanceof-builtins": ["error", { "include": ["HTMLElement"] }] }
```

```javascript
foo instanceof HTMLElement;
```

Examples of **correct** code for this rule with `{ "exclude": ["String"] }`:

```json
{ "unicorn/no-instanceof-builtins": ["error", { "exclude": ["String"] }] }
```

```javascript
foo instanceof String;
```

Examples of **incorrect** code for this rule with `{ "strategy": "strict", "useErrorIsError": true }`:

```json
{
  "unicorn/no-instanceof-builtins": [
    "error",
    { "strategy": "strict", "useErrorIsError": true }
  ]
}
```

```javascript
foo instanceof Error;
```

## Options

### `strategy`

Type: `"loose" | "strict"`  
Default: `"loose"`

- `"loose"` reports primitive wrapper constructors, `Function`, `Array`, and constructors listed in `include`.
- `"strict"` also reports built-in constructors such as `Error`, `Map`, `Set`, `Date`, typed arrays, `Object`, and `RegExp`.

### `include`

Type: `string[]`  
Default: `[]`

Additional constructor names to report.

### `exclude`

Type: `string[]`  
Default: `[]`

Constructor names to ignore. This takes precedence over `strategy`, `include`, and `useErrorIsError`.

### `useErrorIsError`

Type: `boolean`  
Default: `false`

When enabled, `Error` checks are reported with an autofix to `Error.isError(value)`.

## Original Documentation

- [eslint-plugin-unicorn: no-instanceof-builtins](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-instanceof-builtins.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/no-instanceof-builtins.js)
