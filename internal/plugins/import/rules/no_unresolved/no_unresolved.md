# no-unresolved

## Rule Details

Report imports whose module cannot be resolved. The rule checks ES imports,
re-exports, and dynamic `import()` calls with string literals. It ignores explicit
`import type` and `export type` declarations.

Examples of **incorrect** code for this rule, when the target does not exist:

```javascript
import value from './missing.js';
export { value } from './missing.js';
import('./missing.js');
```

Examples of **correct** code for this rule, when `existing.js` exists:

```javascript
import fs from 'node:fs';
import value from './existing.js';
```

The rule does not provide automatic fixes or suggestions.

## Options

| Option | Default | Behavior |
| --- | --- | --- |
| `commonjs` | `false` | Check single-argument `require('module')` calls. |
| `amd` | `false` | Check dependency arrays in two-argument `require` and `define` calls. |
| `esmodule` | `true` | Check ES imports, exports, and dynamic imports. |
| `ignore` | None | Skip specifiers matching any JavaScript regular expression in this list. |
| `caseSensitive` | `true` | On a filesystem that ignores case, check the spelling of resolved paths below the working directory. |
| `caseSensitiveStrict` | `false` | Also check the working directory and its ancestors, even if `caseSensitive` is disabled. |

The `ignore` option is separate from `settings['import/ignore']`: a module can
exist even when other import rules cannot parse its contents.

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-unresolved': ['error', {
        commonjs: true,
        amd: true,
        ignore: ['\\.img$'],
      }],
    },
  },
];
```

## Resolution settings

The default `node` resolver checks runtime files, including JSON and native
modules. It tries `.mjs`, `.js`, `.json`, and `.node` extensions. For packages,
it tries the `module`, `jsnext:main`, and `main` fields in that order, then falls
back to index files.
It does not treat an installed `@types` package as a runtime module.
`import/core-modules` exempts exact module names, such as `electron`.

`import/resolver` accepts `node`, an object containing Node options, or an array
of resolvers tried in order. Node options include `extensions`, `paths`,
`moduleDirectory`, and `preserveSymlinks`. The legacy `import/resolve` setting
supplies Node options when `import/resolver` is absent.

```javascript
export default [
  {
    plugins: ['import'],
    settings: {
      'import/resolver': { node: { extensions: ['.js', '.jsx'] } },
      'import/core-modules': ['electron'],
    },
    rules: { 'import/no-unresolved': 'error' },
  },
];
```

## Differences from upstream

- Supported resolvers are `node` and `typescript`, also accepted as
  `eslint-import-resolver-node` and `eslint-import-resolver-typescript`.
  For example, `settings['import/resolver'] = 'webpack'` reports a resolver error
  and unresolved imports. For bundler aliases, use `typescript` with matching
  `paths` in `tsconfig.json`, or exclude those imports with the `ignore` option.
- The `typescript` resolver follows the TypeScript project selected by rslint.
  Its `project`, `alwaysTryTypes`, and other resolver-specific options have no
  effect. For example, `{ typescript: { project: 'tsconfig.app.json' } }` does not
  select that config; use `languageOptions.parserOptions.project` instead.
- If an `import/resolver` object contains several resolver names, they are tried
  alphabetically. Use an array, such as `['typescript', 'node']`, to choose the
  order explicitly.
- Babel's experimental `export value from './module'` syntax is not supported.
  Use `export { default as value } from './module'` instead.

## Original Documentation

- [eslint-plugin-import: no-unresolved](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-unresolved.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unresolved.js)
