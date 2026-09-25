# exports-last

## Rule Details

Require all ES module exports to appear after other top-level statements. Each
export followed by a non-export statement is reported. This includes default
exports, named declarations, re-exports, and TypeScript type exports.

Examples of **incorrect** code for this rule:

```javascript
export const enabled = true;
const name = 'example';
```

Examples of **correct** code for this rule:

```javascript
const name = 'example';

export const enabled = true;
export { name };
```

Only file-level ordering is checked. Statements inside exported functions,
classes, or TypeScript namespaces do not affect the order. Comments after an
export are allowed, but an extra semicolon forming an empty statement counts
as a non-export statement. For example, `export default function () {};` is
reported. Remove the trailing semicolon to write `export default function () {}`.

CommonJS assignments, TypeScript `export =`, and `export as namespace` are
treated as non-export statements. They are not reported themselves, but an ES
export before them is reported.

The rule does not provide automatic fixes or suggestions.

## Options

This rule has no options.

## When Not To Use It

Disable this rule if exports may appear throughout a file or if you want to
enforce ordering for CommonJS exports.

## Original Documentation

- [eslint-plugin-import: exports-last](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/exports-last.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/exports-last.js)
