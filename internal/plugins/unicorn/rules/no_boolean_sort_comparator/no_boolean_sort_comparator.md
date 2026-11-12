# no-boolean-sort-comparator

## Rule Details

Sorting comparators must return a negative number, zero, or a positive number.
A boolean comparator can only produce `0` or `1` after conversion to a number,
so it cannot place the first value before the second.

The rule checks `sort()` and `toSorted()` on arrays and typed arrays. It
recognizes concise callbacks, functions with a single return statement,
explicit boolean return types, and references to known boolean functions.

Examples of **incorrect** code

```javascript
array.sort((a, b) => a > b);
array.sort((a, b) => a.score >= b.score);
strings.toSorted((a, b) => a === b);
```

Examples of **correct** code

```javascript
array.sort((a, b) => a - b);
array.sort((a, b) => a.score - b.score);
strings.toSorted((a, b) => a.localeCompare(b));
```

## Suggestions

For simple comparisons between the two parameters or matching properties, the
rule can suggest subtraction in the editor. Review the suggested order before
applying it. The rule has no automatic fix.

## Options

This rule has no options.

## Differences from upstream

rslint withholds the subtraction suggestion when it recognizes a string or
BigInt operand. For example, `new BigInt64Array(2).sort((a, b) => a > b)`
still produces a diagnostic. Use `localeCompare()` for strings. For BigInts,
return a number with `a < b ? -1 : a > b ? 1 : 0`.

rslint also withholds the suggestion if converting the comparator to an arrow
would change what its `this`, `arguments`, or function name refers to, or if the
rewrite would discard a return type assertion or leave a boolean JSDoc return
type in place. Update the comparator body and any return type together.

## Original Documentation

- [eslint-plugin-unicorn: no-boolean-sort-comparator](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-boolean-sort-comparator.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-boolean-sort-comparator.js)
