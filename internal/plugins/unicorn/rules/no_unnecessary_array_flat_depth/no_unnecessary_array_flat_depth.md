# no-unnecessary-array-flat-depth

## Rule Details

Disallow passing `1` as the `depth` argument to `Array#flat()`.

The default depth of `Array#flat()` is already `1`, so writing the argument
explicitly is redundant.

Examples of **incorrect** code for this rule:

```javascript
const nested = [1, [2, 3]];
nested.flat(1);
```

Examples of **correct** code for this rule:

```javascript
const nested = [1, [2, 3]];
nested.flat();
```

Depths other than the numeric literal `1` are left unchanged:

```javascript
nested.flat(2);
nested.flat(Infinity);
nested.flat(depth);
```

## Original Documentation

- [eslint-plugin-unicorn: no-unnecessary-array-flat-depth](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-unnecessary-array-flat-depth.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-unnecessary-array-flat-depth.js)
