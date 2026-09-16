# no-extraneous-import

Disallow imports of installed packages that are not declared in `package.json`.

A transitive dependency can be available locally but disappear when another
dependency changes. Declare packages that your code imports directly.

## Rule details

This rule checks static imports, re-exports, and dynamic `import()` calls with
literal sources. It accepts the current package's name and packages listed in
`dependencies`, `devDependencies`, `peerDependencies`, or `optionalDependencies`.
For workspace packages, it also accepts those names from the nearest ancestor
whose `workspaces` patterns include the package. Both workspace arrays and the
`{ "packages": [...] }` form are supported; negative patterns exclude packages.

For example, with `declared` in `dependencies` and an installed but undeclared
package named `transitive`, these imports are incorrect:

```js
import value from 'transitive';
export { value } from 'transitive';
const module = import('transitive');
```

These imports are correct:

```js
import value from 'declared';
import fs from 'node:fs';
import local from './local.js';
```

Type-only declarations can also be satisfied by their corresponding `@types`
dependency. For example, declaring `@types/example` permits
`import type { Value } from 'example'` and `export type { Value } from 'example'`.
An inline type specifier such as `import { type Value } from 'example'` still
requires the runtime package, matching upstream.

Missing packages, Node builtins, relative paths, URL imports, and TypeScript
`paths` aliases are not reported. `require()` and TypeScript `import()` type
expressions are outside this rule. The rule does not provide fixes or suggestions.

## Options

```js
import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    rules: {
      'node/no-extraneous-import': [
        'error',
        {
          allowModules: [],
          resolvePaths: [],
          resolverConfig: { modules: ['node_modules'] },
        },
      ],
    },
  },
]);
```

- `allowModules`: package names permitted without a dependency declaration.
  Scoped names are supported; allowing a package also allows its subdirectories.
- `resolvePaths`: additional base directories for package resolution, relative
  to the working directory. The importing file's directory is also searched.
- `resolverConfig.modules`: a module directory string or an array of directories
  to search. The default is `['node_modules']`. Names such as `bower_components`
  are searched up the directory tree; absolute directories are also supported.
  An empty string uses the default; an empty array disables module lookup.
- `convertPath`: accepts the upstream object or array syntax. This option has
  no effect on this rule in upstream v18.3.0 or in rslint.

Options take precedence over shared `settings.node` values. An explicitly empty
list overrides the shared list. For compatibility, legacy `settings.n` values
are also accepted and take precedence over `settings.node` when both are present.

Shared `settings.node.tryExtensions` controls implicit file extensions. Its
default is `.js`, `.json`, `.node`, `.mjs`, and `.cjs`. TypeScript files also use
the nearest tsconfig's extension and path settings. Package entry points with
explicit extensions do not depend on `tryExtensions`.

Shared `typescriptExtensionMap` and `tsconfigPath` can override TypeScript's
extension substitution, using the same formats as upstream.

Supported `resolverConfig` properties are `modules`, `alias`, `extensions`,
`extensionAlias`, `conditionNames`, `mainFields`, `mainFiles`, and `aliasFields`.

`resolverConfig.mainFields` selects package entry fields in order, for example
`['browser', 'module', 'main']`. `mainFiles` selects directory entry filenames,
such as `['api', 'index']`. `aliasFields: ['browser']` applies package mappings
including `false` to ignore a target. Field names can be nested arrays, such as
`[['build', 'main'], 'main']`. Empty entry lists disable that lookup.

## Differences from upstream

Other `resolverConfig` properties, including `fallback`, `symlinks`, and
`fullySpecified`, are ignored. For example, `fallback: { virtual: './shim.js' }`
does not redirect an unresolved `virtual` request. Use `alias` if the redirect
should apply to every matching request.

Package entry names containing literal backslashes are not resolved on POSIX;
use `/` for portable directory separators. On Windows, rslint accepts relative
paths such as `require('.\\entry.js')`; upstream can treat these as package names instead.

When object-form aliases overlap, rslint tries their names in sorted order;
upstream uses declaration order. Use an alias array to specify priority, such as
`[{ name: 'pkg/entry', alias: './entry.js' }, { name: 'pkg', alias: './fallback' }]`.

With `workspaces: ['packages/{1..3}']`, rslint treats `packages/1`, `packages/2`
and `packages/3` as workspace members; eslint-plugin-n instead matches the literal
directory `packages/1..3`. An installed dependency declared only at the workspace
root is therefore accepted by rslint in `packages/2`, but reported by
eslint-plugin-n. Use `packages/{1,2,3}` to include the same packages in both tools.

With `workspaces: ['packages/[^a]*']`, eslint-plugin-n includes `packages/app`
but rslint excludes it. An installed dependency declared only at the workspace
root is therefore reported by rslint in that child package, but accepted by
eslint-plugin-n. Use `packages/[!a]*` to exclude package names starting with `a`
in both tools.

Disabling a wildcard alias affects only matching requests. For example,
`alias: { 'pkg/*': false }` disables resolution of `pkg/sub`, but rslint still
resolves `pkg` and unrelated packages. Upstream can ignore those other requests
as well. Use exact alias names when identical behavior is required.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-extraneous-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-import.js)
- [Supported resolver settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md#resolverconfig)
