# no-internal-modules

Prevent imports from reaching into other modules' implementation files. The rule
checks imports, reexports, string-literal dynamic imports and direct CommonJS
`require()` calls, including type-only imports and exports.

With no options, a resolved import with more than one non-scope path component
is reported. Package entrypoints such as `package` and `@scope/package`, sibling
files such as `./file`, and unresolved imports are allowed. Node builtins and
absolute imports are excluded unless `import/internal-regex` classifies them as
internal.

```javascript
// Incorrect, when these modules resolve.
import helper from 'package/internal/helper';
export * from './feature/internal';
const submodule = require('./feature/internal');

// Correct.
import helper from 'package';
import local from './local';
```

## Options

Use either `allow` or `forbid`; both accept arrays of minimatch glob patterns.
They cannot be combined. Omitted options and `{ allow: [] }` prohibit resolved
internal-module imports. `{ forbid: [] }` permits all imports.

`allow` permits matching internal-module paths:

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-internal-modules': ['error', {
        allow: ['**/actions/*', 'source-map-support/*'],
      }],
    },
  },
];
```

`forbid` reports only matching paths, including package entrypoints and
unresolved imports:

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-internal-modules': ['error', {
        forbid: ['**/actions/*', 'source-map-support/*'],
      }],
    },
  },
];
```

Matching uses the normalized import path, with and without a leading slash,
then the resolved absolute path. Backslashes become forward slashes, empty and
`.` components are removed, and `..` removes the preceding component. Patterns
support braces, character classes, extended globs and negation. Wildcards do
not match dotfiles by default. As in upstream's `minimatch.makeRe()`, `app/**/**`
matches `app/a/b` but does not match `app/a`.

## Resolution settings

The default resolver uses Node-style JavaScript and JSON resolution. Set
`settings['import/resolver']` to `typescript` for TypeScript modules and
`tsconfig.json` path aliases. The shared import settings, including
`import/core-modules` and `import/internal-regex`, affect module classification.

## Differences from upstream

- Native rules cannot execute custom JavaScript resolvers such as `webpack`.
  Unsupported resolvers produce a resolver error. Use `typescript` with
  matching `paths` for aliases. Select projects with
  `languageOptions.parserOptions.project`; options inside a TypeScript resolver
  configuration do not select projects or enable `alwaysTryTypes`.
- Multiple resolvers configured as an object are tried alphabetically. Use an
  array, such as `['typescript', 'node']`, to preserve the intended priority.
- Babel's experimental `export value from './module'` syntax is unsupported.
  Use `export { default as value } from './module'` instead.
- Empty patterns and patterns beginning with `#` match nothing. Upstream can
  throw when testing the non-regexp value returned for these patterns.
- Glob backslashes escape characters on every platform, following the shared
  matcher. Upstream minimatch 3 treats them as separators on Windows.

This rule offers no autofixes or suggestions.

## Original Documentation

- [eslint-plugin-import documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-internal-modules.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-internal-modules.js)
