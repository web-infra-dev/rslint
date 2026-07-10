# require-array-join-separator

## Rule Details

Enforces passing an explicit separator to `Array#join()` instead of relying on
the default comma separator.

Examples of **incorrect** code for this rule:

```javascript
const string = array.join();
const arrayLikeString = Array.prototype.join.call(arrayLike);
```

Examples of **correct** code for this rule:

```javascript
const string = array.join(',');
const arrayLikeString = Array.prototype.join.call(arrayLike, '');
```

## Original Documentation

- [eslint-plugin-unicorn require-array-join-separator](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/require-array-join-separator.md)
