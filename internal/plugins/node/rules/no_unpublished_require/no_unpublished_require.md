# no-unpublished-require

Disallow requiring unpublished files and development dependencies from published files.

## Rule details

The rule checks `require()` and `require.resolve()`, including aliases,
destructured methods, optional calls, and constant arguments. It uses the
nearest valid `package.json`, its `files` list, and applicable `.npmignore` or
`.gitignore` files to decide which files are published. Unpublished source
files and, by default, private packages are skipped.

In a published file, a call is reported if its target is outside the package
or excluded from publication, or its package is listed only in
`devDependencies`. Packages also listed in `dependencies`, `peerDependencies`,
or `optionalDependencies` are accepted. Development dependencies need not be
installed to be reported. Workspace ancestors do not supply production
dependencies for this check.

For a package with `files: ['lib']` and `devDependencies: { dev: '*' }`, these
calls in `lib/index.js` are **incorrect**:

```js
const helper = require('../test/helper.js');
const tool = require('dev');
const filename = require.resolve('dev/subpath');
```

These calls are **correct** when `lib/helper.js` exists:

```js
const helper = require('./helper');
const fs = require('node:fs');
```

CommonJS resolution checks directory entries and package `main` fields, with
the `node` and `require` export conditions. Local bindings that shadow
`require` and dynamic arguments such as `require(packageName)` are ignored.
Ordinary imports and TypeScript `import value = require('package')`
declarations are outside this rule. There are no fixes or suggestions.

Use `node/no-missing-require` to check missing targets and
`node/no-extraneous-require` to check undeclared packages.

## Options

```js
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    languageOptions: { globals: globals.node },
    rules: {
      'node/no-unpublished-require': ['error', {
        allowModules: [],
        ignorePrivate: true,
      }],
    },
  },
]);
```

- `allowModules` accepts package roots, including scoped and `virtual:` names.
  Allowing `dev` also allows `dev/subpath`. It does not exempt local files.
- `ignorePrivate` defaults to `true`. Set it to `false` to check packages with
  `private: true`, for example before deployment.
- `convertPath` maps both the requiring file and local targets from source
  paths to published paths, relative to the package directory. It accepts an
  object such as `{ 'src/**': ['^src/(.*)\\.ts$', 'lib/$1.js'] }`, or an ordered
  array of `{ include, exclude?, replace }` entries. The first matching entry
  wins. Omit this option instead of using `null`, which the upstream schema
  also rejects.
- `resolvePaths` supplies extra resolution base directories before the
  requiring file's directory. Relative values use the working directory or
  `settings.cwd` when configured.
- `tryExtensions` defaults to `['.js', '.json', '.node', '.mjs', '.cjs']`.
  An empty array disables extension guessing.
- `resolverConfig` supports `modules`, `alias`, `fallback`, `fullySpecified`,
  `extensions`, `extensionAlias`, `conditionNames`, `mainFields`, `mainFiles`, and
  `aliasFields`, as described for [no-missing-require](/rules/node/no-missing-require).

Except for `ignorePrivate`, these options can also be supplied through
`settings.node`. Rule options take precedence; legacy `settings.n` takes
precedence over `settings.node`. Explicit empty lists override shared lists.
TypeScript extension mappings and path aliases affect local target resolution.
A package alias does not exempt a development dependency.

## Differences from upstream

Compared with `eslint-plugin-n` v18.3.0:

- Subdirectory ignore files affect publication; upstream only uses the package
  root's ignore file.
- `README.js` is treated as published even with `files: []`. Included files such
  as `lib/..hidden.js` are also treated as published.
- On case-insensitive filesystems, `main` matches filenames regardless of case;
  upstream can skip checking those files.
- In `files`, `[!b]` excludes `b`, `\*` matches a literal `*`, and whitespace is
  significant. Upstream may select different files for these patterns.
- Overlapping object-form `convertPath`, `resolverConfig.alias`, and
  `resolverConfig.fallback` entries use alphabetical priority instead of
  declaration order. Use arrays to set priority.
- Absolute `convertPath` replacements stay absolute; targets outside the package
  are reported.
- An invalid `convertPath` regex in shared settings skips the check instead of
  failing lint.
- `resolverConfig` options not listed above, including `symlinks`, are not supported.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-require.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-require.js)
