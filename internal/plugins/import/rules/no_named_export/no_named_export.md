# no-named-export

## Rule Details

Disallow named ES module exports, including exported declarations, named
re-exports, star exports, and TypeScript type exports. Each offending export
statement produces one diagnostic, regardless of how many names it exports.

Examples of **incorrect** code for this rule:

```javascript
export const value = 1;
export { other } from './other.js';
export * from './values.js';
```

Examples of **correct** code for this rule:

```javascript
const value = 1;
export { value as default };
```

Default exports and aliases exported as `default` or `"default"` are allowed.
Empty exports (`export {}`) and namespace re-exports, including
`export * as default from './values.js'`, are reported.

CommonJS assignments are allowed. Files configured with `sourceType: 'script'`
or `sourceType: 'commonjs'` are ignored. The rule provides no automatic fixes
or suggestions.

## Options

This rule has no options.

## When Not To Use It

Disable this rule if your project allows or prefers named exports.

## Differences from upstream

Compared with eslint-plugin-import v2.32.0:

- **Babel re-export shortcuts are unsupported.** Replace
  `export default from './values.js'` with
  `export { default } from './values.js'`, which this rule allows. Named
  shortcuts such as `export value from './values.js'` and
  `export value, { other } from './values.js'` are also unsupported. Their
  standard equivalents, `export { default as value } from './values.js'` and
  `export { default as value, other } from './values.js'`, are named exports
  and are reported.

## Original Documentation

- [eslint-plugin-import: no-named-export](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-export.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-export.js)
