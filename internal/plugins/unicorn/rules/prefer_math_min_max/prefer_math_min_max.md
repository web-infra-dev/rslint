# prefer-math-min-max

## Rule details

Prefer `Math.min()` and `Math.max()` for ternaries that compare two values and
return one of those same values.

The rule skips comparisons involving known BigInt values, `new Date()` operands,
and bindings with non-number TypeScript annotations or initializers such as
string literals.

Examples of **incorrect** code for this rule

```javascript
height > 50 ? 50 : height;
height < 50 ? 50 : height;
```

Examples of **correct** code for this rule

```javascript
Math.min(height, 50);
Math.max(height, 50);
```

## Fixes and suggestions

The rule offers editor suggestions because the replacement can change the
result for `NaN`, signed zero, or values that need numeric coercion. When `a`
is `NaN`, `a > b ? a : b` returns `b` while `Math.max(a, b)` returns `NaN`.

You still get a diagnostic when the replacement would remove a comment or
change how often an operand with side effects runs.

## Options

This rule has no options.

## Differences from upstream

Rslint reports without a suggestion when `Math` is shadowed or disabled in
`languageOptions.globals`.

```javascript
function cap(Math, height) {
  return height > 50 ? 50 : height;
}
```

Rslint also reports without a suggestion for `delete` expressions and classes
with side effects in an `extends` expression, field initializer, or static
block. For example, replacing this ternary could change how many times the
property is deleted.

```javascript
const capped = (delete cache.key, height) > 50 ? 50 : (delete cache.key, height);
```

When linting a file without a TypeScript project, Rslint also skips comparisons
that use the result of a callback typed as `() => bigint` or a function whose
return type is `bigint & unknown` or a signed BigInt literal type like `-1n`.
Passing these values to `Math.min()` or `Math.max()` throws a `TypeError`.

```typescript
declare const read: (() => bigint);
const value = read();
value > 0 ? value : 0;

function readLimit(): bigint & unknown {
  return 1n;
}
const limit = readLimit();
limit > 0 ? limit : 0;
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-min-max.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-math-min-max.js)
