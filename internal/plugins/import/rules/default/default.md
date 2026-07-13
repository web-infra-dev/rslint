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

## Original Documentation

- [eslint-plugin-import/default](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/default.md)
