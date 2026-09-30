# require-array-sort-compare

## Rule Details

Require an explicit compare function when calling `Array#sort()` or `Array#toSorted()`.

Without a comparator, ordinary arrays convert elements to strings before comparing them. That can produce surprising numeric ordering.

Examples of **incorrect** code:

```javascript
numbers.sort();
numbers.toSorted();
```

Examples of **correct** code:

```javascript
numbers.sort((a, b) => a - b);
names.toSorted((a, b) => a.localeCompare(b));
```

The rule provides two editor suggestions: a numeric comparator and a string `localeCompare` comparator. Suggestions are withheld when the call contains comments so an edit cannot discard or move authored commentary.

Typed arrays are intentionally not reported because `TypedArray#sort()` already sorts numerically.

## Original Documentation

- [eslint-plugin-unicorn: require-array-sort-compare](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/require-array-sort-compare.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/require-array-sort-compare.js)
