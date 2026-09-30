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

The rule reports every matching nested built-in `Math.min()`/`Math.max()` call. It only autofixes when the flattened arguments are statically primitive and side-effect-free, because flattening can otherwise reorder JavaScript numeric coercion. Non-flattened argument text, including required parentheses, is preserved. Calls through a locally shadowed `Math` binding are ignored.

This intentionally makes autofixing more conservative than eslint-plugin-unicorn v75.0.0 while preserving the same diagnostic intent.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-flat-math-min-max.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-flat-math-min-max.js)
