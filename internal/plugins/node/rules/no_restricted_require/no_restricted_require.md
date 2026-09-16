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
`settings.node.resolverConfig.modules`, and TypeScript path aliases and extension
settings.

## Differences from upstream

When restricting absolute file paths, aliases configured with
`settings.node.resolverConfig.alias` are not followed. If `virtual` points to a
restricted `/project/api.js`, upstream reports `require('virtual')`; rslint does
not report it unless `virtual` also resolves as an installed module. Restrict
`'virtual'` by name or configure the alias in `tsconfig.json` instead. Of the
`resolverConfig` options, only `modules` is supported.

Write restricted module names as strings when constructing them with BigInt
arithmetic. With `['42']` restricted, upstream reports `require(40n + 2n)` and
rslint does not. Use `require('42')` to ensure the restriction is checked;
`require(42n)` is also checked.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-require.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-require.js)
