# no-relative-parent-imports

## Rule Details

Disallow imports whose resolved file is in a parent directory. The rule checks
static imports, re-exports and dynamic `import()` calls, including type-only
imports and exports. Unresolved paths and external packages are ignored.

Paths are checked after resolution, so `./../main.js` and aliases that resolve
to a parent file can also be reported.

Examples of **incorrect** code for this rule in `lib/example.js`:

```javascript
import main from '../main.js';
export { value } from '../shared.js';
const shared = import('../shared.js');
```

Examples of **correct** code for this rule:

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

Set `import/resolver` to `node` or `typescript`. Their full names,
`eslint-import-resolver-node` and `eslint-import-resolver-typescript`, are also
accepted. The default Node resolver supports JavaScript and JSON paths; use
`typescript` for TypeScript paths and aliases. `import/core-modules`,
`import/internal-regex` and `import/external-module-folders` also affect which
imports are reported.

## Differences from upstream

- If your ESLint configuration uses `webpack` or another custom resolver,
  rslint reports a resolver error for imports checked by this rule. Use `node`,
  or use `typescript` with matching `paths` in `tsconfig.json` for aliases.
  The rule's `ignore` option can exclude imports you do not want to check.
- To select a TypeScript project, use `languageOptions.parserOptions.project`.
  Setting `import/resolver` to
  `{ typescript: { project: 'tsconfig.app.json' } }` does not select that project
  in rslint. Other TypeScript resolver options, including `alwaysTryTypes`, also
  have no effect, so imports may resolve differently from ESLint.
- When configuring multiple resolvers, use an array such as
  `['typescript', 'node']` to set their priority. Object configurations are tried
  alphabetically in rslint: `{ typescript: {}, node: {} }` tries `node` first,
  whereas ESLint tries `typescript` first. This can change the resolved file
  and whether the import is reported.
- Babel's experimental `export value from './module'` syntax is not supported.
  Use `export { default as value } from './module'` instead.

## Original Documentation

- [eslint-plugin-import: no-relative-parent-imports](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-parent-imports.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-parent-imports.js)
