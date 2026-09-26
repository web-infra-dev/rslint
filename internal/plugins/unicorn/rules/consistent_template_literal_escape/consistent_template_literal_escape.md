# consistent-template-literal-escape

Enforce one consistent way to escape a literal `${` sequence inside template literals.

Use `\${` by escaping the dollar sign rather than escaping the opening brace.

```js
// Incorrect
const template = `$\{variableName}`;
const redundant = `\$\{variableName}`;

// Correct
const template = `\${variableName}`;
```

Tagged templates are ignored because tags can observe the raw template text.

The rule is automatically fixable and preserves the surrounding template delimiters and interpolation boundaries.

This port follows eslint-plugin-unicorn v75.0.0.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/consistent-template-literal-escape.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/consistent-template-literal-escape.js)
