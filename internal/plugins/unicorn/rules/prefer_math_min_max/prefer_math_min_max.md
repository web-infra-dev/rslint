# prefer-math-min-max

## Rule details

Prefer `Math.min()` and `Math.max()` for ternaries that compare two values and
return one of those same values.

The rule skips comparisons involving known BigInt values, `new Date()` operands,
and bindings with non-number TypeScript annotations or initializers such as
string literals.

The replacement is an editor suggestion because the two forms can differ for
`NaN`, signed zero, and values that need coercion. For example, when `a` is
`NaN`, `a > b ? a : b` returns `b` while `Math.max(a, b)` returns `NaN`.
The rule keeps the diagnostic when a suggestion would discard comments or
change how often an operand with side effects is evaluated.

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

## Options

This rule has no options.

## Differences from upstream

The diagnostic has no suggestion when a local variable shadows `Math` or
`languageOptions.globals` disables it.

```javascript
function cap(Math, height) {
  return height > 50 ? 50 : height;
}
```

The diagnostic also has no suggestion when an operand uses `delete` or
contains a class with a call in its `extends` expression, a field initializer,
or a static block.

```javascript
const capped = (delete cache.key, height) > 50 ? 50 : (delete cache.key, height);
```

Rslint skips comparisons with values returned by a callback typed as
`() => bigint` or a function with a signed BigInt literal return type such
as `-1n`, even in files outside a TypeScript project. It also skips return
types of `bigint & unknown`, which simplify to `bigint`. Passing these
values to `Math.min()` or `Math.max()` throws a `TypeError`.

```typescript
declare const read: () => bigint;
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
