# prefer-node-protocol

Requires the `node:` prefix when importing Node.js built-in modules.

## Rule details

The rule checks static imports, re-exports, dynamic imports, `require()`, and
`process.getBuiltinModule()`, including `globalThis.process.getBuiltinModule()`.
Automatic fixes insert `node:` while preserving quotes, escapes and comments.

Examples of **incorrect** code:

```javascript
import fs from "fs";
export { promises } from "fs";
const path = require("path");
const util = process.getBuiltinModule("util");
```

Examples of **correct** code:

```javascript
import fs from "node:fs";
export { promises } from "node:fs";
const path = require("node:path");
const util = process.getBuiltinModule("node:util");
```

Relative paths, third-party packages, template literals, `require.resolve()`,
and optional `require?.()` calls are ignored. Names available only with the
prefix, such as `node:test`, do not make the bare name `test` a built-in module.
Calls on locally defined `require`, `process`, or `globalThis` are ignored
unless they are recognized Node bindings. Direct imports or `require()` loads
of `process` and `require` functions initialized by an imported
`module.createRequire()` are recognized; arbitrary local aliases and custom
wrappers are not followed.

## Options

The optional object accepts `version`, a Node.js version range:

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/prefer-node-protocol": ["error", { version: ">=16.0.0" }],
    },
  },
];
```

The target range comes from the first valid value in this order:

1. The rule's `version` option.
2. `settings.n.version`, then `settings.node.version`.
3. The nearest package's `engines.node`.
4. Its `devEngines.runtime` entry named `node`.
5. `>=16.0.0`.

Imports and re-exports are checked only when the entire target range supports
the prefix: `^12.20.0 || >=14.13.1`. `require()` additionally needs
`^14.18.0 || >=16.0.0`. For example, `>=14.18.0` includes Node 15, so it enables
checks for imports but not for `require()`. `process.getBuiltinModule()` is
always checked because that API implies support for the prefix.

## Differences from upstream

- Custom local bindings are left unchanged. For example,
  `function require(name) { return name; } require("fs")` is not reported,
  because adding the prefix would change the returned value. Upstream fixes
  calls based on their spelling even when the name refers to a custom function.
- Contradictory alternatives do not affect a version range. Both
  `>=16 || >20 <16` and `>20 <16 || >=16` enable fixes for imports and
  `require()`, because `>20 <16` contains no versions. Upstream disables these
  checks for the first ordering.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-node-protocol.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-node-protocol.js)
