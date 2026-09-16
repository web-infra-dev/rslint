# no-restricted-import

Disallow specified modules in imports and re-exports.

The rule checks static imports, `export ... from`, and dynamic `import()` with
literal arguments, including Node builtins and TypeScript type-only imports.
It does not check `require()`, import types such as `type T = import('pkg').T`,
or dynamic imports with variables or template literals. It provides no fixes
or suggestions.

## Options

Pass an array of restricted module names or objects with `name` and an optional
`message`. With no options, the rule reports nothing.

```js
import { resolve } from 'node:path';
import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    rules: {
      'node/no-restricted-import': [
        'error',
        [
          'fs',
          'node:fs',
          {
            name: ['lodash/*', '!lodash/pick'],
            message: 'Import from lodash or use lodash/pick.',
          },
          {
            name: resolve(import.meta.dirname, 'server', '**'),
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
import fs from 'fs';
export * from 'node:fs';
const omit = import('lodash/omit');
```

Correct:

```js
import crypto from 'node:crypto';
import lodash from 'lodash';
import pick from 'lodash/pick';
```

- `name` accepts a string or an array of patterns. `*` matches text except `/`;
  `**` as a complete `/`-separated segment matches across directories. `?`, brackets,
  braces, and extended glob groups are literal text.
- Patterns in one `name` array apply in order. A leading `!` removes an earlier
  match; a later positive pattern can add it again. An initial `!(` is literal.
  Negations do not cancel restrictions in another object.
- The first matching restriction supplies the diagnostic and optional message.
- Names match exactly, including module paths and the `node:` prefix. Restricting
  `fs` does not restrict `node:fs`; restricting `lodash` does not restrict
  `lodash/pick`.
- Loader parameters after the first `!` are removed before matching. Query
  strings and fragments remain part of the name.

Absolute patterns match resolved file paths, including relative imports and
installed packages. Use `node:path` to build absolute patterns with the current
platform's separators, as in the example above. Patterns are matched as written;
backslashes are literal characters, including on Windows.

Unresolved relative imports use their lexical absolute path; unresolved package
names have no file path to match.

Resolution uses the Node plugin's shared `settings.node` options, including
`resolvePaths`, `tryExtensions`, and `resolverConfig.modules`, and TypeScript
path aliases. Legacy `settings.n` is also recognized, as in the other Node rules.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-import.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-import.js)
- [Shared settings](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/shared-settings.md)
