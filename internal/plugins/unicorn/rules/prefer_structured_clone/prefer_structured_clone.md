# prefer-structured-clone

Prefer `structuredClone` when creating deep clones.

The rule recognizes:

- `JSON.parse(JSON.stringify(value))`
- `_.cloneDeep(value)`
- `lodash.cloneDeep(value)`
- additional configured function paths

The rule reports suggestions only; it does not autofix automatically.

```js
// Incorrect
const clone = JSON.parse(JSON.stringify(value));
const other = _.cloneDeep(value);

// Correct
const clone = structuredClone(value);
```

Custom functions can be configured through the `functions` option.

This port follows eslint-plugin-unicorn v75.0.0 and preserves authored
parentheses, comments, and trailing commas in suggestions.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-structured-clone.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-structured-clone.js)
