# no-useless-fallback-in-spread

## Rule Details

Disallow an empty object fallback in an object spread. Spreading `undefined`,
`null`, or another falsy value into an object adds no properties, so the
fallback is unnecessary.

Examples of **incorrect** code:

```javascript
const object = { ...(foo || {}) };
const other = { ...(foo ?? {}) };
```

Examples of **correct** code:

```javascript
const object = { ...foo };
const conditional = { ...(foo && {}) };
const array = [...(foo || [])];
```

## Automatic Fixes

Fixes preserve the left operand and its comments. When removing the fallback
would delete a comment, the rule reports the empty object without offering a fix.

Like upstream, this rule assumes falsy values have no enumerable properties.
Browser code that may spread `document.all` should disable the rule for that
expression: `document.all` is falsy but can have enumerable properties, so
removing its fallback can change the result.

## Options

This rule has no options. It is enabled at `error` severity in
`unicornPlugin.configs.recommended`.

## Differences from upstream

Fixes keep the parentheses required by JavaScript JSDoc `@type` assertions:

```javascript
const object = {.../** @type {object} */ (value || {})};
// Fixed:
const object = {.../** @type {object} */ (value)};
```

This preserves the assertion when `checkJs` or `// @ts-check` is enabled.
Unicorn v79.0.0 removes these parentheses, which can introduce a spread-type
error when `value` has type `unknown`.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-fallback-in-spread](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-fallback-in-spread.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-fallback-in-spread.js)
