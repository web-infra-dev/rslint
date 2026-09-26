# no-typeof-undefined

Disallow comparing `undefined` using `typeof`.

Prefer a direct comparison:

```js
// Incorrect
if (typeof value === 'undefined') {}

// Correct
if (value === undefined) {}
```

By default, references that resolve to globals are ignored because removing
`typeof` can turn an existence check into a `ReferenceError`. Set
`checkGlobalVariables: true` to report those cases as suggestions instead of
autofixes.

This port follows eslint-plugin-unicorn v75.0.0, including converting
`==`/`!=` to strict equality, ASI-sensitive replacements, and multiline return/throw handling.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-typeof-undefined.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-typeof-undefined.js)
