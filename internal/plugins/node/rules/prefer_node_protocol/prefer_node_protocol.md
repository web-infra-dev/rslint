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
As upstream, the rule matches call syntax even if `require` or `process` is
shadowed by a local variable.

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

Version ranges containing a major, minor or patch number greater than
`4294967295` are treated as invalid, so the next configured version source is
used. For example, with `{ version: "<=4294967296" }` and no other version
configuration, `import "fs"` is reported and fixed to `import "node:fs"`;
upstream leaves it unchanged. This only affects ranges containing version
numbers far beyond released Node.js versions.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-node-protocol.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-node-protocol.js)
