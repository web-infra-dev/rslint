# no-deprecated-api

## Rule Details

Disallow the use of deprecated Node.js APIs. The rule flags:

- Deprecated builtin **modules** (`require('sys')`, `import domain from 'domain'`).
- Deprecated **members** of builtin modules, reached through `require()`, ESM `import`, or `process.getBuiltinModule()` — including intermediate variables and destructuring (`var b = require('buffer'); new b.Buffer()`).
- Deprecated Node.js **globals** (`new Buffer()`, `Intl.v8BreakIterator`, `process.binding`).

`node:`-prefixed specifiers (`require('node:buffer')`) are treated the same as the bare form.

Examples of **incorrect** code for this rule:

```javascript
require("fs").exists;
new Buffer();
require("url").parse;
import domain from "domain";
var b = require("buffer");
new b.Buffer();
```

Examples of **correct** code for this rule:

```javascript
require("fs").existsSync;
Buffer.alloc(10);
new (require("url").URL)("http://example.com");
var http = require("http");
http.request();
```

## Options

```json
{
  "n/no-deprecated-api": [
    "error",
    { "version": ">=16.0.0", "ignoreModuleItems": [], "ignoreGlobalItems": [] }
  ]
}
```

- `version` (string) — the target Node.js version range. Replacement suggestions in the message are filtered to APIs already available on that version. Defaults to `>=16.0.0`.
- `ignoreModuleItems` (string[]) — module-API names to allow, e.g. `["buffer.Buffer()", "fs.exists"]`.
- `ignoreGlobalItems` (string[]) — global names to allow, e.g. `["new Buffer()", "Buffer()"]`.

Examples of **correct** code with `{ "ignoreModuleItems": ["fs.exists"] }`:

```json
{ "n/no-deprecated-api": ["error", { "ignoreModuleItems": ["fs.exists"] }] }
```

```javascript
require("fs").exists;
```

## Differences from ESLint

- The target Node.js version comes from the `version` option or `settings.n.version` / `settings.node.version`, defaulting to `>=16.0.0`. Unlike eslint-plugin-n, it is not read from `package.json` (`engines.node` / `devEngines.runtime`) — set it explicitly when you need a non-default target.

## Original Documentation

- [n/no-deprecated-api](https://github.com/eslint-community/eslint-plugin-n/blob/HEAD/docs/rules/no-deprecated-api.md)
