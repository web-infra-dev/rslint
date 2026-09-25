# prefer-default-export

## Rule Details

Prefer a default export when a file exports a single name. With
`target: 'any'`, require a default export whenever the file has named exports.

Examples of **incorrect** code with the default options:

```javascript
export const value = 1;
```

```javascript
export { value } from './values.js';
```

Examples of **correct** code with the default options:

```javascript
export default 1;
```

```javascript
export const first = 1;
export const second = 2;
```

A default export, a star export (including `export * as name`), or an exported
type alias or interface exempts the entire file. Star exports do not inspect
the other module. Type-only export lists such as `export type { Value }` are
counted as named exports, unlike `export type Value = number`.

The rule reports the last named export when the file does not meet the selected
policy. It provides no automatic fixes or suggestions. Files without ES module
exports, including files using only CommonJS assignments, are allowed.

## Options

The rule accepts an object with a `target` property:

- `"single"` (default): require a default export when there is exactly one
  named export.
- `"any"`: require a default export when there is at least one named export.
  The file exemptions above still apply.

For example, with `["error", { "target": "any" }]`, this code is **incorrect**:

```javascript
export const first = 1;
export const second = 2;
```

This code is **correct** with `target: 'any'`:

```javascript
export const first = 1;
export default 2;
```

## When Not To Use It

Disable this rule if your project prefers named exports even when a file has
only one export.

## Differences from upstream

Compared with eslint-plugin-import v2.32.0:

- **Babel re-export shortcuts are unsupported.** Replace
  `export default from './values.js'` with
  `export { default } from './values.js'`, which satisfies either `target`.
  Replace `export value, { other } from './values.js'` with
  `export { default as value, other } from './values.js'`. This exports two
  named values, so it satisfies `target: 'single'` but still needs a default
  export with `target: 'any'`, unless another export exempts the file.

## Original Documentation

- [eslint-plugin-import: prefer-default-export](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/prefer-default-export.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/prefer-default-export.js)
