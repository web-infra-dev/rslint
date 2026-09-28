# prefer-array-last-methods

Prefer last-oriented array methods over reversing an array and then calling the corresponding forward method.

The rule recognizes:

- `.reverse().find()` and `.toReversed().find()` → `.findLast()`
- `.reverse().findIndex()` and `.toReversed().findIndex()` → `.findLastIndex()`
- `.reverse().indexOf()` and `.toReversed().indexOf()` → `.lastIndexOf()`
- `.reverse().reduce()` and `.toReversed().reduce()` → `.reduceRight()`

```js
// Incorrect
const result = array.reverse().find(isUnicorn);

// Correct
const result = array.findLast(isUnicorn);
```

The rule reports a diagnostic and offers an editor suggestion rather than an automatic fix because replacing reversal can change observable mutation, sparse-array, callback-index, or index-return semantics. Suggestions are withheld when comments occur inside the matched call.

This port follows eslint-plugin-unicorn v75.0.0.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-array-last-methods.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-array-last-methods.js)
