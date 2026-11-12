# no-array-sort-for-min-max

## Rule Details

Use `Math.min()` or `Math.max()` to find a numeric extreme in one pass. Sorting the whole array takes `O(n log n)` work just to read one item.

The rule checks `sort()` and `toSorted()` with simple numeric comparators such as `(a, b) => a - b` or `(a, b) => b - a`. It reports reads using `[0]`, `.at(0)`, or `.at(-1)`.

Examples of **incorrect** code:

```javascript
const minByIndex = array.sort((a, b) => a - b)[0];
const minByAt = array.toSorted((a, b) => a - b).at(0);
const maxByAt = array.sort((a, b) => a - b).at(-1);
const maxByIndex = array.toSorted((a, b) => b - a)[0];
```

Examples of **correct** code:

```javascript
const minimum = Math.min(...array);
const maximum = Math.max(...array);
const sorted = array.toSorted((a, b) => a - b);
const first = array.toSorted((a, b) => a.value - b.value)[0];
```

The rule offers editor suggestions. Review them before applying because `sort()` mutates the original array, and the replacement can behave differently with empty arrays, sparse arrays, `NaN`, or arrays too large to spread into a function call.

## Differences from upstream

Rslint reports the sort and withholds the editor suggestion when `Math` resolves to a local binding or is set to `'off'` in `languageOptions.globals`.

```javascript
function minimum(Math) {
  return [3, 1, 2].toSorted((a, b) => a - b)[0];
}
```

## Original Documentation

- [eslint-plugin-unicorn: no-array-sort-for-min-max](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-array-sort-for-min-max.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-array-sort-for-min-max.js)
