# no-array-from-fill

## Rule Details

Disallow calling `.fill()` after `Array.from({ length: … })`. Prefer the
mapping function argument of `Array.from()` when creating a fixed-length array
with generated values.

Examples of **incorrect** code for this rule:

```javascript
Array.from({ length: 3 }).fill().map((_, index) => index);
Array.from({ length: 3 }).fill({});
```

Examples of **correct** code for this rule:

```javascript
Array.from({ length: 3 }, (_, index) => index);
Array.from({ length: 3 }, () => ({}));
```

Using a mapping function also avoids unintentionally sharing one mutable value
between every array element.

## Original Documentation

- [eslint-plugin-unicorn: no-array-from-fill](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-array-from-fill.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-array-from-fill.js)
