# no-magic-array-flat-depth

## Rule Details

Disallow a magic number as the `depth` argument in `Array#flat(…)`.

`Array#flat()` is often used to flatten deeply nested arrays, but using a
hard-coded depth like `flat(2)` or `flat(99)` is usually a sign the caller
doesn't know the array's structure. Prefer `flat(Infinity)` for "fully flatten"
or restructure the data so the depth is implicit.

Examples of **incorrect** code for this rule:

```javascript
const array = [1, [2, [3]]];
array.flat(2);
```

Examples of **correct** code for this rule:

```javascript
const array = [1, [2, [3]]];
array.flat(Infinity);
```

```javascript
const array = [1, 2, 3];
array.flat();
```

## Original Documentation

- [eslint-plugin-unicorn: no-magic-array-flat-depth](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-magic-array-flat-depth.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-magic-array-flat-depth.js)
