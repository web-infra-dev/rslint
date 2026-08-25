# no-array-fill-with-reference-type

## Rule Details

Disallow using reference values as `Array#fill()` values. `Array#fill()` uses
the same value for every element, so mutating one filled object also affects all
other elements.

Examples of **incorrect** code for this rule:

```javascript
new Array(3).fill({});

const value = new Map();
array.fill(value);
```

Examples of **correct** code for this rule:

```javascript
Array.from({ length: 3 }, () => ({}));
array.fill(0);
```

A receiver known not to be an array is ignored. This includes typed arrays,
whose `fill()` method coerces the value to a number. Unknown receivers are still
checked.

## Original Documentation

- [eslint-plugin-unicorn: no-array-fill-with-reference-type](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-array-fill-with-reference-type.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-array-fill-with-reference-type.js)
