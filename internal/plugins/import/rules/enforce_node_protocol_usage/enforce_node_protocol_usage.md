# enforce-node-protocol-usage

## Rule Details

Enforce consistent use of the `node:` prefix for Node.js builtin modules.
The rule checks import declarations, named re-exports, dynamic imports and
non-optional calls to `require` with exactly one string literal argument.
It automatically adds or removes the prefix, preserving the surrounding quotes.

Template literals, `require.resolve`, TypeScript import types and import-equals
declarations are not checked. Like upstream, the rule also ignores `export *`
and `export * as name` declarations.

## Options

One string option is required; there is no default:

- `"always"`: require the prefix when both the bare and prefixed module exist.
- `"never"`: remove the prefix when the bare module exists. Modules available
  only with the prefix, such as `node:test`, keep it.

### `always`

Examples of **incorrect** code for this rule with the `"always"` option:

```javascript
import fs from 'fs';
export { promises } from 'fs';
const fileSystem = require('fs/promises');
```

Examples of **correct** code for this rule with the `"always"` option:

```javascript
import fs from 'node:fs';
export { promises } from 'node:fs';
const fileSystem = require('node:fs/promises');
import * as test from 'node:test';
```

### `never`

Examples of **incorrect** code for this rule with the `"never"` option:

```javascript
import fs from 'node:fs';
export { promises } from 'node:fs';
const fileSystem = require('node:fs/promises');
```

Examples of **correct** code for this rule with the `"never"` option:

```javascript
import fs from 'fs';
export { promises } from 'fs';
const fileSystem = require('fs/promises');
import * as test from 'node:test';
```

## Settings

Set `import/node-version` to the Node.js version your project targets. This
controls which modules support the prefix:

```javascript
export default [
  {
    plugins: ['import'],
    settings: { 'import/node-version': '16.0.0' },
    rules: { 'import/enforce-node-protocol-usage': ['error', 'always'] },
  },
];
```

## Differences from upstream

- **Invalid settings produce lint diagnostics.** For example,
  `settings: { 'import/node-version': 'bad' }` reports a diagnostic at each
  checked module reference and applies no fixes from this rule. These diagnostics
  follow the rule's configured severity and respect disable comments. Upstream
  throws an exception instead.
- **The default Node.js version is `22.0.0`.** ESLint uses the Node.js version
  running it. For example, `import 'fs'` with `"always"` is reported by default
  in rslint, but not by ESLint running on Node 14.17. Set `import/node-version`
  explicitly to use the same target in both tools.
- **Version numbers have an upper limit.** Each part of `import/node-version`
  must be between `0` and `4294967295`. For example, rslint rejects
  `4294967296.0.0`, which upstream accepts. This does not affect released Node.js
  versions.
- **Fixes preserve the module when its prefix contains escapes.** With `"never"`,
  rslint fixes `import 'node\u{3a}fs'` to `import 'fs'`. Upstream changes the
  imported module to `u{3a}fs` instead.

## Original Documentation

- [eslint-plugin-import: enforce-node-protocol-usage](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/enforce-node-protocol-usage.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/enforce-node-protocol-usage.js)
