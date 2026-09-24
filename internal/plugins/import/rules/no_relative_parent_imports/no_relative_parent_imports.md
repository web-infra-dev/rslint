# no-relative-parent-imports

## Rule Details

Disallow imports whose resolved file is in a parent directory. The rule checks
static imports, re-exports and dynamic `import()` calls, including type-only
imports and exports. Unresolved paths and external packages are ignored.

Paths are checked after resolution, so `./../main.js` and aliases that resolve
to a parent file can also be reported.

Examples of **incorrect** code in `lib/example.js`:

```javascript
import main from '../main.js';
export { value } from '../shared.js';
const shared = import('../shared.js');
```

Examples of **correct** code:

```javascript
import sibling from './sibling.js';
import child from './child/index.js';
import dependency from 'dependency';
```

Move the importing file, pass the dependency as a function argument, or expose
it as a package. The rule does not provide automatic fixes or suggestions.

## Options

| Option | Default | Behavior |
| --- | --- | --- |
| `commonjs` | `false` | Check single-argument `require('module')` calls. |
| `amd` | `false` | Check dependency arrays in two-argument `require` and `define` calls. |
| `esmodule` | `true` | Check ES imports, re-exports and dynamic imports. |
| `ignore` | None | Skip module paths matching any JavaScript regular expression in this list. |

If provided, `ignore` must contain at least one pattern, with no duplicates.
It is separate from `settings['import/ignore']`, which does not exempt imports
from this rule.

This configuration also checks CommonJS imports and allows parent imports from
`generated`:

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-relative-parent-imports': ['error', {
        commonjs: true,
        ignore: ['^\\.\\./generated/'],
      }],
    },
  },
];
```

## Resolution settings

Resolution uses `import/resolver` settings. The default Node resolver supports
JavaScript and JSON paths; use the `typescript` resolver for TypeScript paths
and aliases. `import/core-modules`, `import/internal-regex` and
`import/external-module-folders` also affect classification.

## Differences from upstream

- Supported resolvers are `node` and `typescript`, including their full names
  `eslint-import-resolver-node` and `eslint-import-resolver-typescript`.
  For example, `settings['import/resolver'] = 'webpack'` reports a resolver error.
  For bundler aliases, use `typescript` with matching `paths` in `tsconfig.json`,
  or exclude those imports with the `ignore` option.
- The `typescript` resolver follows the TypeScript project selected by rslint.
  Resolver-specific options such as `project` and `alwaysTryTypes` have no effect.
  To select `tsconfig.app.json`, use `languageOptions.parserOptions.project`
  instead of `{ typescript: { project: 'tsconfig.app.json' } }`.
- If an `import/resolver` object contains several resolver names, they are tried
  alphabetically. Use an array, such as `['typescript', 'node']`, to choose the
  order explicitly.
- Babel's experimental `export value from './module'` syntax is not supported.
  Use `export { default as value } from './module'` instead.

## Original Documentation

- [eslint-plugin-import: no-relative-parent-imports](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-parent-imports.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-parent-imports.js)
