# no-extraneous-require

Disallow loading installed packages that are not declared in `package.json`.

A transitive dependency can disappear when another dependency changes. Declare
the packages your code loads directly, even when they are already installed.

## Rule details

This rule checks `require()` and `require.resolve()`, including aliases,
destructured methods, optional calls, and access through Node's global object.
It reports the module argument when its constant value names an installed,
undeclared package. Local bindings that shadow `require` are ignored.

For example, if `declared` is a dependency and `transitive` is installed but
undeclared, these calls are incorrect:

```js
const value = require('transitive');
const filename = require.resolve('transitive');
const load = require;
load('transitive');
```

These calls are correct:

```js
const value = require('declared');
const fs = require('node:fs');
const local = require('./local.js');
```

The package's own name and all four dependency fields are accepted:
`dependencies`, `devDependencies`, `peerDependencies`, and
`optionalDependencies`. Workspace members also inherit those names from the
nearest matching ancestor workspace. Workspace arrays and the
`{ "packages": [...] }` form are supported, including negative patterns.

Missing packages, builtins, relative paths, and TypeScript `paths` aliases are
ignored. Resolution uses CommonJS export conditions. An `@types` dependency
alone does not permit loading its runtime package. Imports and TypeScript
`import value = require('package')` declarations are outside this rule.

Constant expressions such as concatenation and template literals are supported.
An identifier argument such as `require(packageName)` is not evaluated, even
when it has a constant initializer. The rule provides no fixes or suggestions.

## Options

```js
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    languageOptions: { globals: globals.node },
    rules: {
      'node/no-extraneous-require': [
        'error',
        {
          allowModules: [],
          resolvePaths: [],
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          resolverConfig: { modules: ['node_modules'] },
        },
      ],
    },
  },
]);
```

- `allowModules`: package names accepted without a declaration. Scoped names
  are supported, and allowing a package also allows its subdirectories.
- `resolvePaths`: additional directories from which to resolve packages,
  relative to the working directory. The source file's directory is also used.
- `tryExtensions`: implicit file extensions. The JavaScript default is shown
  above. An empty array disables extension probing; explicit package entry
  points still resolve. TypeScript files also respect the nearest tsconfig's
  extension settings and the shared `typescriptExtensionMap` and `tsconfigPath`.
- `resolverConfig.modules`: a module directory string or an array of directories.
  The default is `['node_modules']`. Relative names are searched up the directory
  tree; absolute directories are also supported. An empty array disables lookup.
- `convertPath`: accepts the upstream object or array syntax. It has no effect
  on this rule in upstream v18.3.0 or in rslint.

These options can also be configured in `settings.node`. Rule options take
precedence, including explicitly empty arrays. Legacy `settings.n` is accepted
for compatibility and takes precedence over `settings.node` when both are set.

## Differences from upstream

The supported `resolverConfig` properties are `modules`, `alias`, `extensions`,
`extensionAlias`, and `conditionNames`. Other properties are ignored. For example,
`mainFiles: ['entry']` does not change directory lookup to `entry.js`; use an
explicit file path instead.

When object-form aliases overlap, rslint tries their names in sorted order;
upstream uses declaration order. Use an alias array to specify priority, such as
`[{ name: 'pkg/entry', alias: './entry.js' }, { name: 'pkg', alias: './fallback' }]`.

With `workspaces: ['packages/{1..3}']`, rslint includes `packages/1`, `packages/2`,
and `packages/3`; upstream matches the literal directory `packages/1..3`.
Consequently, a dependency declared only at the workspace root is accepted by
rslint in `packages/2`, but reported upstream. Use `packages/{1,2,3}` for the
same results in both tools.

With `workspaces: ['packages/[^a]*']`, upstream includes `packages/app`, while
rslint excludes it. A dependency declared only at the workspace root is therefore
reported by rslint in that child package but accepted upstream. Use
`packages/[!a]*` to exclude names starting with `a` in both tools.

Disabling a wildcard alias affects only matching requests. For example,
`alias: { 'pkg/*': false }` disables resolution of `pkg/sub`, but rslint still
resolves `pkg` and unrelated packages. Upstream can ignore those other requests
as well. Use exact alias names when identical behavior is required.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-extraneous-require.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-require.js)
- [Shared settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md)
