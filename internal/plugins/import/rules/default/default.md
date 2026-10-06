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

Modules that cannot be resolved or are ignored are not reported. Like upstream,
this rule also skips files that fail its module text check or contain neither
an import/export declaration nor a runtime dynamic import. Plain CommonJS
exports and compiler-forced module markers alone do not establish an export map.

With `esModuleInterop: true` explicitly enabled, local named exports also provide
a default for this rule. Named re-exports and `export *` alone do not. For example,
a default import from `export const value = 1` is accepted with this option.

The upstream text check can skip a file containing only `import './setup.mjs';`
or a spaced dynamic `import ('./setup.mjs')`. This rule preserves that behavior.

## Original Documentation

- [eslint-plugin-import: default](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/default.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/default.js)
