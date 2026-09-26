# prefer-flat-math-min-max

Prefer flat `Math.min()` and `Math.max()` calls over nesting the same method.

Both functions accept any number of arguments, so same-method nesting is unnecessary.

```js
// Incorrect
const biggest = Math.max(Math.max(a, b), c);
const smallest = Math.min(a, Math.min(b, c));

// Correct
const biggest = Math.max(a, b, c);
const smallest = Math.min(a, b, c);

// Mixed methods are intentionally preserved.
const clamped = Math.max(Math.min(value, upper), lower);
```

The rule is automatically fixable when the matched call contains no comments. Nested calls are flattened recursively while different `Math.min`/`Math.max` methods remain intact.

This port follows eslint-plugin-unicorn v75.0.0.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-flat-math-min-max.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-flat-math-min-max.js)
