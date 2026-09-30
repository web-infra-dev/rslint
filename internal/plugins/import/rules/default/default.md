# default

## Rule Details

This rule reports a default import when the imported module does not provide a default export.

Examples of **incorrect** code for this rule:

```javascript
// ./bar.js
export const bar = 1;

// ./foo.js
import bar from "./bar";
```

Examples of **correct** code for this rule:

```javascript
// ./bar.js
export default 1;

// ./foo.js
import bar from "./bar";
```

Modules that cannot be resolved, are ignored, or are not ES modules are not reported by this rule.

## Differences from upstream

- `esModuleInterop` does not supply a missing ES module default. Given
  `export const value = 1` in `values.mjs`, rslint reports
  `import value from './values.mjs'` even with `esModuleInterop: true`;
  upstream v2.32.0 allows it. Import `{ value }` or add a default export.
  With NodeNext, native ES imports of CommonJS modules still receive
  `module.exports` as their default, regardless of the interop setting.

## Original Documentation

- [eslint-plugin-import: default](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/default.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/default.js)
