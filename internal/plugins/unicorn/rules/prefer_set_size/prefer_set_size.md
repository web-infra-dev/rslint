# prefer-set-size

Prefer using `Set#size` instead of `Array#length`.

Converting a `Set` into an array solely to read its length performs unnecessary
work. Use its `size` property directly.

## Examples

```js
// Incorrect
const uniqueCount = [...new Set(userIds)].length;

// Correct
const uniqueCount = new Set(userIds).size;
```

```js
// Incorrect
const count = Array.from(items).length;

// Correct
const count = items.size;
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-set-size](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-set-size.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-set-size.js)
