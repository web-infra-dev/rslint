# no-missing-require

Disallow `require()` and `require.resolve()` calls whose targets cannot be found.

The rule checks constant module arguments, including calls through aliases,
destructured methods, optional calls, and Node's global object. Local bindings
that shadow `require` are ignored. It reports the argument without offering
fixes or suggestions.

## Examples

Incorrect when the targets do not exist:

```js
const value = require('./missing.js');
const filename = require.resolve('missing-package');
const load = require;
load('missing-package');
```

Correct when `existing.js` and `installed-package` exist:

```js
const value = require('./existing');
const packageValue = require('installed-package');
const fs = require('node:fs');
```

CommonJS directory entries, package `main` fields, and `exports` maps are
checked. The default export conditions are `node` and `require`. Declaring a
dependency in `package.json` does not make an uninstalled package available.
Node builtins are accepted; `data:`, HTTP, and HTTPS URLs still require a
resolvable target.

Constant expressions such as `require('pack' + 'age')` and template literals
are checked. Identifier arguments such as `require(packageName)` are ignored,
even when the identifier has a constant initializer. Imports and TypeScript
`import value = require('package')` declarations are outside this rule.

## Options

```js
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    languageOptions: { globals: globals.node },
    rules: {
      'node/no-missing-require': [
        'error',
        {
          allowModules: [],
          resolvePaths: [],
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      ],
    },
  },
]);
```

- `allowModules`: package roots accepted without resolution. Allowing
  `electron`, `@scope/package`, or `virtual:module` also allows their subdirectories.
- `resolvePaths`: additional base directories searched before the requiring
  file's directory. Relative entries use the working directory, or
  `settings.cwd` when configured.
- `tryExtensions`: extensions tried for extensionless targets. The JavaScript
  defaults are shown above. An empty list disables extension guessing;
  explicit filenames can still resolve.
- `resolverConfig`: overrides module resolution. Supported properties are
  `modules`, `alias`, `fallback`, `fullySpecified`, `extensions`,
  `extensionAlias`, `conditionNames`, `mainFields`, `mainFiles`, and `aliasFields`. For example,
  `{ modules: ['custom_modules', 'node_modules'] }` adds a module directory,
  and `{ alias: { virtual: './shim.js' } }` redirects a request. Directory
  defaults are `mainFields: ['main']` and `mainFiles: ['index']`.
- `typescriptExtensionMap`: overrides TypeScript extension substitution with
  `[sourceExtension, emittedExtension]` pairs, or `preserve`, `react`,
  `react-jsx`, `react-jsxdev`, or `react-native`. An empty array disables
  substitution.
- `tsconfigPath`: selects the TypeScript config used for extension mapping.
  Relative paths use the working directory. An explicit extension map takes
  precedence. Path aliases still come from the nearest `tsconfig.json` and
  must resolve to existing targets.

Options can also be supplied through `settings.node`. Rule options take
precedence, including explicit empty lists.

For TypeScript files, the default mapping substitutes `.ts` for `.js`, `.mts`
for `.mjs`, `.cts` for `.cjs`, and `.tsx` for `.jsx` in preserve mode or `.js`
in React modes. `allowImportingTsExtensions` uses source extensions directly.

`resolverConfig.fallback` redirects unresolved requests, for example
`{ fallback: { virtual: './shim.js' } }`. Existing targets take precedence.
`fullySpecified: true` disables extension and directory-entry guessing for
ordinary requests: `./helper.js` can resolve while `./helper` cannot. Package
entry points and alias redirects retain their normal lookup behavior.

## Differences from upstream

- `resolverConfig` options not listed above, including `symlinks`, are not supported.
- Overlapping object-form `alias` and `fallback` entries use alphabetical priority
  instead of declaration order. Use arrays to set priority.
- Error messages for invalid `package.json#imports` mappings and circular aliases
  may differ.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-missing-require.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-missing-require.js)
- [Shared settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md)
