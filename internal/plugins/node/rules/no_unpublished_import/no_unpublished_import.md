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
  `aliasFields`, as described for [no-missing-import](../no_missing_import/no_missing_import.md).

Except for `ignoreTypeImport` and `ignorePrivate`, these options can also be
supplied through `settings.node`. Rule options take precedence over shared
settings; legacy `settings.n` takes precedence over `settings.node`. Explicit
empty lists override shared lists. TypeScript extension mappings and path
aliases affect which local file is checked. A package alias does not exempt
a development dependency from this rule.

## Differences from upstream

Compared with `eslint-plugin-n` v18.3.0:

- Publication checks respect subdirectory ignore files. For example, a
  `lib/.npmignore` entry for `helper.js` makes an import of `lib/helper.js`
  report; upstream only reads the root ignore file.
- Root metadata such as `README.js` is always published, so its development
  imports are checked even with `files: []`. Conversely, an included file such
  as `lib/..hidden.js` is not mistaken for a path outside the package.
- On a filesystem that ignores filename case, the package's `main` entry also
  ignores case. For example, `main: 'INDEX.js'` keeps `index.js` published even
  with `files: []`; upstream can skip checking that file's imports.
- In `files` patterns, `[!b]` excludes `b` and `\*` matches a literal `*`.
  Whitespace inside an entry remains part of the pattern. Upstream can select
  different files for these patterns; prefer explicit filenames when the
  published file list must match both tools.
- Overlapping object-form `convertPath` patterns are tried in alphabetical
  order instead of declaration order. Use the ordered array form when a file
  matches several mappings. Absolute replacements stay absolute, so targets
  outside the package are reported; use relative replacements for package
  files. An invalid conversion regex in shared settings skips the check
  instead of failing lint.
- Unlisted `resolverConfig` properties are ignored. For example,
  `symlinks: false` does not preserve a symlink path; checks still use its real
  target. Overlapping object-form `alias` and `fallback` entries use alphabetical
  priority; use an array to specify the intended order. See
  [no-missing-import](../no_missing_import/no_missing_import.md#differences-from-upstream)
  for further unusual filename and alias cases.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-import.js)
