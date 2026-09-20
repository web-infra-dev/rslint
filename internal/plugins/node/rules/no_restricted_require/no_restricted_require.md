# no-restricted-require

Disallow specified modules in `require()` and `require.resolve()` calls.

## Rule details

This rule reports the module argument when it matches a configured restriction.
It follows aliases, destructured `resolve` methods, optional calls, and access
through global objects. Local bindings that shadow `require` are ignored.

Constant expressions such as string concatenation and template literals are
supported. An identifier argument such as `require(moduleName)` is not evaluated,
even when it has a constant initializer. Imports and TypeScript
`import value = require('module')` declarations are outside this rule.

No modules are restricted by default. This rule provides no fixes or suggestions.

## Options

```js
import path from 'node:path';
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    languageOptions: { globals: globals.node },
    rules: {
      'node/no-restricted-require': [
        'error',
        [
          'fs',
          'node:fs',
          {
            name: ['lodash/*', '!lodash/pick'],
            message: 'Use the public entry point.',
          },
          {
            name: path.resolve('server', '**'),
            message: 'Keep server code out of the client.',
          },
        ],
      ],
    },
  },
]);
```

Incorrect with this configuration:

```js
const fs = require('fs');
require.resolve('node:fs');
const load = require;
load('lodash/omit');
```

Correct:

```js
const crypto = require('node:crypto');
const lodash = require('lodash');
const pick = require('lodash/pick');
```

- Each restriction is a string or an object with `name` and an optional `message`.
  `name` accepts a string or an array of patterns.
- `*` matches text except `/`; `**` as a complete `/`-separated segment matches
  across directories. `?`, brackets, braces, and extended glob groups are literal.
- Patterns within one `name` array apply in order. A leading `!` removes an
  earlier match; a later positive pattern can add it again. An initial `!(` is
  literal. Negations do not cancel restrictions in another object.
- The first matching restriction supplies the diagnostic and custom message.
- Names match exactly, including module paths and the `node:` prefix. Restricting
  `fs` does not restrict `node:fs`; restricting `lodash` does not restrict
  `lodash/pick`.
- Loader parameters after the first `!` are removed before matching. Query
  strings and fragments remain part of the name.

Absolute patterns match resolved file paths. CommonJS resolution supports
directory entry points and the `require` export condition. Missing local modules
use their lexical absolute path; unresolved package names have no path to match.
Use `node:path` to construct absolute patterns with the host's separators.
Patterns are matched as written, including literal backslashes on Windows.

Resolution uses `settings.node.resolvePaths`, `settings.node.tryExtensions`,
and the `modules`, `alias`, `fallback`, `fullySpecified`, `extensions`, `extensionAlias`, `conditionNames`,
`mainFields`, `mainFiles`, and `aliasFields` properties of `settings.node.resolverConfig`. TypeScript files also use tsconfig
path aliases and extension settings. Explicit `alias` and `extensionAlias`
settings replace those TypeScript mappings, including when set to empty objects.

`resolverConfig.mainFields` selects package entry fields in order, for example
`['browser', 'module', 'main']`. `mainFiles` selects directory entry filenames,
such as `['api', 'index']`. `aliasFields: ['browser']` applies package mappings
including `false` to ignore a target. Field names can be nested arrays, such as
`[['build', 'main'], 'main']`. Empty entry lists disable that lookup.

`resolverConfig.fallback` redirects unresolved requests, for example
`{ fallback: { virtual: './shim.js' } }`. Existing targets take precedence.
Object-form `alias` and `fallback` entries use declaration order, so place a
specific alias before a broader prefix when both can match.
`fullySpecified: true` disables extension and directory-entry guessing for
ordinary requests: `./helper.js` can resolve while `./helper` cannot. Package
entry points and alias redirects retain their normal lookup behavior.

## Differences from upstream

Unlisted `resolverConfig` properties are ignored. For example,
`symlinks: false` does not preserve a symlink path; checks still use its real target.

Package entry names containing literal backslashes are not resolved on POSIX;
use `/` for portable directory separators. On Windows, rslint accepts relative
paths such as `require('.\\entry.js')`; upstream can treat these as package names instead.

Disabling a wildcard alias affects only matching requests. For example,
`alias: { 'pkg/*': false }` disables resolution of `pkg/sub`, but rslint still
resolves `pkg` and unrelated packages. Upstream can ignore those other requests
as well. Use exact alias names when identical behavior is required.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-require.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-require.js)
