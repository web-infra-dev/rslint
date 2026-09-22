# prefer-type-error

## Rule Details

Require `TypeError` instead of the generic `Error` when an `if` condition is exclusively checking a value's type and the branch contains only that throw.

The rule recognizes `typeof`, non-error `instanceof` checks, the global `isNaN()` and `isFinite()` functions, and common `.isType()`-style helpers used by JavaScript libraries. Existence checks such as `typeof window !== 'undefined'` are not treated as type checks.

Examples of **incorrect** code:

```javascript
if (Array.isArray(value)) {
  throw new Error('Expected a non-array value.');
}
```

Examples of **correct** code:

```javascript
if (Array.isArray(value)) {
  throw new TypeError('Expected a non-array value.');
}
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Differences from upstream

Rslint resolves built-in names before offering the rewrite. A locally shadowed `Error` is not reported, and bare `isNaN` or `isFinite` calls are recognized only when they resolve to the environment globals. If `TypeError` is unavailable or locally shadowed, the diagnostic is still reported but the autofix is withheld rather than rewriting to a different binding.

These checks avoid binding-changing autofixes while preserving the upstream behavior for normal global built-ins.

## Original Documentation

- [eslint-plugin-unicorn: prefer-type-error](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-type-error.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-type-error.js)
