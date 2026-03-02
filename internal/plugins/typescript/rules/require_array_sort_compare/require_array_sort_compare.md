# require-array-sort-compare

## Rule Details

Require `Array#sort` and `Array#toSorted` calls to always provide a `compareFunction`. When called without a compare function, `Array#sort()` and `Array#toSorted()` convert all non-undefined elements to strings and then compare them using their UTF-16 code unit values. This can lead to surprising sort orders, especially for arrays of numbers (e.g., `[1, 10, 2]` instead of `[1, 2, 10]`).

By default, string arrays are ignored since the default sort behavior is appropriate for them.

Examples of **incorrect** code for this rule:

```typescript
const numbers = [3, 1, 2];
numbers.sort();

const mixed = [1, 'a', 2];
mixed.sort();
```

Examples of **correct** code for this rule:

```typescript
const numbers = [3, 1, 2];
numbers.sort((a, b) => a - b);

const strings = ['c', 'a', 'b'];
strings.sort(); // OK, string arrays are ignored by default
```

## Original Documentation

- [typescript-eslint require-array-sort-compare](https://typescript-eslint.io/rules/require-array-sort-compare)
