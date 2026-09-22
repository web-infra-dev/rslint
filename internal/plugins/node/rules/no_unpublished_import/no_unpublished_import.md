# no-unpublished-import

Disallow imports of unpublished files and development dependencies from published files.

## Rule details

The rule checks static imports, re-exports, and dynamic `import()` calls with
literal arguments. It uses the nearest `package.json`, its `files` list, and
applicable `.npmignore` or `.gitignore` files to determine what is published.
Imports from unpublished files are ignored.

In a published file, an import is reported when its target is outside the
package or excluded from publication, or its package is listed only in
`devDependencies`. A dependency also listed in `dependencies`, `peerDependencies`,
or `optionalDependencies` is accepted. A development dependency need not be
installed to be reported. Workspace ancestors do not supply production
dependencies for this check.

Given a package with `files: ['lib']` and `devDependencies: { dev: '*' }`, these
imports in `lib/index.js` are **incorrect**:

```js
import helper from '../test/helper.js';
export { build } from 'dev';
const tool = import('dev');
```

These imports are **correct** when `lib/helper.js` exists:

```js
import helper from './helper.js';
import fs from 'node:fs';
```

Use `node/no-missing-import` to check whether targets exist and
`node/no-extraneous-import` to check undeclared packages. This rule provides
no fixes or suggestions.

## Options

```js
export default [
  {
    plugins: ['node'],
    rules: {
      'node/no-unpublished-import': ['error', {
        allowModules: [],
        ignoreTypeImport: false,
        ignorePrivate: true,
      }],
    },
  },
];
```

- `allowModules` accepts package roots, including scoped and `virtual:` names.
  Allowing `dev` also allows `dev/subpath`. It does not exempt local files.
- `ignoreTypeImport` defaults to `false`. Setting it to `true` skips whole
  `import type` declarations. Type re-exports and individual `type` specifiers
  in value imports are still checked.
- `ignorePrivate` defaults to `true`, skipping packages with `private: true`.
  Set it to `false` to check those packages, for example before deployment.
- `convertPath` maps both the importing file and local targets from source
  paths to published paths, relative to the package directory. It accepts an
  object such as `{ 'src/**': ['^src/(.*)\\.ts$', 'lib/$1.js'] }`, or an ordered
  array of `{ include, exclude?, replace }` entries. The first matching entry
  wins. Omit the option instead of using `null`, which the upstream schema
  also rejects.
- `resolvePaths` supplies extra resolution base directories before the
  importing file's directory. Relative values use the working directory or
  `settings.cwd` when configured.
- `tryExtensions` defaults to `['.js', '.json', '.node', '.mjs', '.cjs']`.
  An empty array disables extension guessing.
- `resolverConfig` supports `modules`, `alias`, `fallback`, `fullySpecified`,
  `extensions`, `extensionAlias`, `conditionNames`, `mainFields`, `mainFiles`, and
  `aliasFields`, as described for [no-missing-import](/rules/node/no-missing-import).

Except for `ignoreTypeImport` and `ignorePrivate`, these options can also be
supplied through `settings.node`. Rule options take precedence over shared
settings; legacy `settings.n` takes precedence over `settings.node`. Explicit
empty lists override shared lists. TypeScript extension mappings and path
aliases affect which local file is checked. A package alias does not exempt
a development dependency from this rule.

## Differences from upstream

Compared with `eslint-plugin-n` v18.3.0:

- Subdirectory ignore files affect which files are considered published; upstream
  only uses the package root's ignore file.
- `README.js` is treated as published even with `files: []`. Included files such
  as `lib/..hidden.js` are also treated as published.
- On case-insensitive filesystems, `main` matches filenames regardless of case;
  upstream can skip checking those files' imports.
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

See [no-missing-import](/rules/node/no-missing-import#differences-from-upstream)
for shared module-resolution differences.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-import.js)
