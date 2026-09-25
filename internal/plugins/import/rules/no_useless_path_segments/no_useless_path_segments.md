# no-useless-path-segments

## Rule Details

Disallow unnecessary path segments in relative imports, re-exports, and dynamic
imports. The rule can also check CommonJS `require()` calls.

Paths are normalized when the original and shorter spelling resolve to the same
module, or when neither resolves. Resolved paths that leave and re-enter the
current directory can also be shortened. Automatic fixes use double quotes.

Examples of **incorrect** code for this rule:

```javascript
import value from './utils//value.js';
export * from './utils/./index.js';
const module = import('./utils/');
```

Examples of **correct** code for this rule:

```javascript
import value from './utils/value.js';
export * from './utils/index.js';
const module = import('./utils');
```

## Options

```json
{
  "import/no-useless-path-segments": [
    "error",
    { "commonjs": false, "noUselessIndex": false }
  ]
}
```

### commonjs

Set `commonjs` to `true` to check `require('./utils//value')`. It is disabled by
default. Non-literal arguments, template literals, and `require.resolve()` are
not checked.

### noUselessIndex

Set `noUselessIndex` to `true` to remove a trailing `/index` or `/index` with a
configured extension. It is disabled by default.

```javascript
// With noUselessIndex: true
import './utils/index.js'; // becomes "./utils"
import './index'; // becomes "."
```

If `utils.js` exists alongside `utils/index.js`, the replacement is `"./utils/"`
to preserve the directory import. Extensions come from `import/extensions`
(default: `.js`, `.mjs`, `.cjs`) plus `import/parsers`, independently of resolver
extensions.

As upstream does, this option removes `index` even for unresolved paths and does
not check whether a directory's `package.json` selects a different entry point.

## Differences from upstream

- Supported resolvers are `node` and `typescript`, also accepted as
  `eslint-import-resolver-node` and `eslint-import-resolver-typescript`.
  For example, `settings['import/resolver'] = 'webpack'` reports a resolver error.
  For bundler aliases, use `typescript` with matching `paths` in `tsconfig.json`.
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

- [eslint-plugin-import: no-useless-path-segments](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-useless-path-segments.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-useless-path-segments.js)
