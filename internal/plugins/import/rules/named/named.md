# named

## Rule Details

Require named imports and re-exports to exist in the referenced module. The rule follows explicit re-exports and `export *` declarations, reporting a broken explicit re-export chain when necessary.

Given `values.js`:

```javascript
export const value = 1;
```

Examples of **incorrect** code for this rule:

```javascript
import { missing } from './values.js';
export { missing as renamed } from './values.js';
```

Examples of **correct** code for this rule:

```javascript
import { value as local } from './values.js';
export { value as renamed } from './values.js';
```

Unresolved, ignored and non-ES modules are skipped. Declaration-level type imports and exports, and inline type import specifiers, are skipped. Like upstream, inline `export { type Missing } from './values.js'` is checked. Default imports and namespace imports are outside this rule's scope.

## Options

This rule accepts an options object with the following default:

```json
{ "commonjs": false }
```

Set `commonjs` to `true` to also check identifier keys in destructured `require` calls:

```javascript
// With { commonjs: true }:
const { missing } = require('./values.js'); // Reported.
const { value: local } = require('./values.js'); // Allowed.
```

This option checks ES module exports consumed through `require`; it does not infer CommonJS exports. String keys and rest elements are skipped. The rule has no automatic fixes or suggestions.

## Differences from upstream

- Imports follow the project's TypeScript module resolution settings. ESLint
  resolver plugins and their custom package entry fields are not supported.
  For example, a package with a `types` entry is checked against that declaration
  file, even if its runtime entry is CommonJS. Configure TypeScript's `paths`
  for a different entry point, or use `import/ignore` to exclude a package.
- When both `languageOptions.parserOptions.project` and `projectService` are
  disabled, only imported files also selected for linting can be checked.
  For example, linting `app.ts` alone does not check names exported by `lib.ts`.
  Enable either project option to check dependencies without linting them directly.
- Flow `typeof` imports and Babel's experimental `export name from './module'`
  syntax are unsupported. Use TypeScript type imports and standard
  `export { default as name } from './module'` syntax instead.
- Syntax errors in imported files are not reported by this rule. For example,
  importing a file containing `return; export {};` produces a parse-error report
  upstream, but no report from this rule. Check the imported file's syntax separately.
- Given `export * as ns from './values.js'`, only `ns` is exported. rslint reports
  an import of `value` from that barrel; upstream v2.32.0 also exposes namespace
  members as named exports. Import `ns` and access `ns.value` instead.
- Quoted re-export names are recognized. For example,
  `export { value as 'quoted' } from './values.js'` allows
  `import { quoted } from './barrel.js'` in rslint, while upstream v2.32.0 reports
  it as missing. Removing the quotes makes both tools agree.
- `esModuleInterop` does not supply a missing ES module default. Given
  `export const value = 1` in `values.mjs`, rslint reports
  `import { default as value } from './values.mjs'` even with
  `esModuleInterop: true`; upstream v2.32.0 allows it. Import `{ value }` or add
  a default export. With NodeNext, native ES imports of CommonJS modules still
  receive `module.exports` as their default, regardless of the interop setting.

## Original Documentation

- [eslint-plugin-import: named](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/named.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/named.js)
