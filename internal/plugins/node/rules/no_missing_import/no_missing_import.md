# no-missing-import

Disallow imports and re-exports whose modules cannot be found.

The rule checks static `import`, `export ... from`, and dynamic `import()` with
literal arguments. It checks installed packages and their `exports` maps as
well as local files. Declaring a dependency in `package.json` does not make an
uninstalled package available.

## Examples

Incorrect:

```js
import missing from './missing.js';
export { value } from 'missing-package';
const module = import('./missing.js');
```

Correct, when `existing.js` and `installed-package` exist:

```js
import existing from './existing.js';
export { value } from 'installed-package';
import fs from 'node:fs';
```

Relative directory imports such as `import './directory'` are reported even
when the directory contains `index.js` or a package entry point. Import that
file explicitly. Bare package imports can use package entry points.

Runtime imports of Node builtins, `data:` URLs, and HTTP or HTTPS URLs are
accepted. Expressions such as `import(variable)` and template literals are
outside this rule's checks. The rule provides no fixes or suggestions.

## Options

```js
import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    rules: {
      'node/no-missing-import': [
        'error',
        {
          allowModules: ['electron'],
          resolvePaths: [],
          ignoreTypeImport: false,
        },
      ],
    },
  },
]);
```

- `allowModules`: package roots to accept without resolving them, including
  scoped packages and `virtual:` modules. Allowing a package also allows its
  internal paths, such as `electron/main`. The default is an empty list.
- `resolvePaths`: additional base directories, searched before the importing
  file's directory. Relative entries use the working directory, or `settings.cwd`
  when configured. The default is an empty list.
- `tryExtensions`: extensions to try for extensionless imports. The default is
  `['.js', '.json', '.node', '.mjs', '.cjs']`. An empty list disables extension
  guessing; explicit filenames can still resolve.
- `resolverConfig.modules`: a module directory string or array, such as
  `'node_modules'` or `['custom_modules', 'node_modules']`. The default is
  `['node_modules']`. Relative names are searched up the directory tree;
  absolute directories are also accepted. An empty array disables package
  lookup without disabling relative file imports.
- `ignoreTypeImport`: skips whole `import type` declarations when `true`.
  The default is `false`. Type re-exports and individual `type` specifiers in a
  value import are still checked, matching upstream.
- `typescriptExtensionMap`: overrides TypeScript extension substitution.
  Accepts `[sourceExtension, emittedExtension]` pairs, or `preserve`, `react`,
  `react-jsx`, `react-jsxdev`, and `react-native`. An empty array disables
  substitution.
- `tsconfigPath`: selects a TypeScript config for extension substitution.
  The nearest `tsconfig.json` supplies the default mapping and path aliases.
  Relative values use the working directory, independently of `settings.cwd`.
  An explicit extension map takes precedence over `tsconfigPath`.

Except for `ignoreTypeImport`, these options can also be supplied through
`settings.node`. Rule options take precedence over shared settings. Legacy
`settings.n` is accepted with the same precedence as upstream. Explicit empty
lists override shared lists.

In TypeScript files, the default mapping substitutes `.ts` for `.js`, `.mts`
for `.mjs`, `.cts` for `.cjs`, and `.tsx` for `.jsx` in preserve mode or `.js`
in React modes. With `allowImportingTsExtensions`, the default extension list
also includes `.ts`, `.mts`, and `.cts`, and emitted extensions are not
substituted. Type-only imports also activate the `types` export condition.

## Differences from upstream

Only `resolverConfig.modules` affects module lookup. Other `resolverConfig`
options, including `alias`, are ignored. For example, mapping `virtual` to
`./local.js` through `resolverConfig.alias` lets upstream accept
`import 'virtual'` when that file exists; rslint reports it as missing.
Use TypeScript `compilerOptions.paths` for local aliases, or `allowModules`
for intentionally unresolved modules.

Some failures in `package.json#imports` have different wording. If `#redirect`
maps to a private package export, rslint reports `Can't resolve '#redirect'`
with the importing directory; upstream explains which export is unavailable.
An invalid target array such as `[null, './entry.js']` also produces a normal
missing-import message instead of upstream's `Invalid value used as weak map key`.
Both cases are reported by both linters.

Module names containing an unpaired Unicode surrogate, such as
`import('\uD800')`, may display replacement characters in rslint's message.
Use well-formed Unicode module names for consistent resolution and diagnostics.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-missing-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-missing-import.js)
- [Shared settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md)
