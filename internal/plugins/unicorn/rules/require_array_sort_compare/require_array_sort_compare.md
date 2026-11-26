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

The rule provides two editor suggestions: a numeric comparator and a string `localeCompare` comparator. The comparator replaces an `undefined` argument, including its parentheses, or is inserted before the closing parenthesis, so comments elsewhere in the call are preserved, for example `array.sort(/* keep */)` becomes `array.sort(/* keep */(a, b) => a - b)`. Suggestions are withheld only when replacing the argument would remove a comment, such as `array.sort((/* keep */ undefined))`.

Typed arrays are intentionally not reported because `TypedArray#sort()` already sorts numerically.

## Differences from upstream

Upstream skips every instance of a custom class. Rslint still reports an instance of a class that extends `Array`, because its `sort()` and `toSorted()` compare elements as strings too:

```javascript
class Numbers extends Array {}

new Numbers(10, 2, 1).sort(); // reported by rslint, not by upstream
```

The class chain must reach `Array` through classes declared in the same file. TypeScript files with type information also follow the checker's base types.

## Original Documentation

- [eslint-plugin-unicorn: require-array-sort-compare](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/docs/rules/require-array-sort-compare.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/require-array-sort-compare.js)
