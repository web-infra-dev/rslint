# node/no-extraneous-import

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

## Differences from upstream

Only `resolverConfig.modules` changes how this rule finds imported packages in
rslint. Other `resolverConfig` options, such as `alias`, are ignored. For example,
suppose `virtual` is neither declared nor installed, but `local.js` exists. With
`resolverConfig: { alias: { virtual: './local.js' } }`, eslint-plugin-n reports
`import 'virtual'` as an undeclared dependency; rslint ignores the alias and
reports nothing.
For local aliases in TypeScript files, use `compilerOptions.paths`. To allow an
undeclared package, add its name to `allowModules`.

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

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-extraneous-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-import.js)
- [Supported resolver settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md#resolverconfig)
