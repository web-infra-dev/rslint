# prefer-global/url-search-params

## Rule details

Enforces consistent use of the global `URLSearchParams` or the `URLSearchParams` export from `url` and `node:url`. Module references include `require()`, `process.getBuiltinModule()`, and static imports and re-exports.

## Options

This rule accepts one string option:

- `"always"` (default): use the global `URLSearchParams`.
- `"never"`: use `URLSearchParams` from the module.

Enable Node.js globals in the configuration:

```javascript
import { globals } from "@rslint/core";

export default [
  {
    plugins: ["node"],
    languageOptions: { globals: globals.node },
    rules: {
      "node/prefer-global/url-search-params": ["error", "always"],
    },
  },
];
```

### always

Examples of **incorrect** code:

```javascript
const { URLSearchParams } = require("url");
const params = new URLSearchParams();

import { URLSearchParams as Params } from "node:url";
```

Examples of **correct** code:

```javascript
const params = new URLSearchParams();
```

### never

Examples of **incorrect** code:

```javascript
const params = new URLSearchParams();
```

Examples of **correct** code:

```javascript
const { URLSearchParams } = require("node:url");
const params = new URLSearchParams();
```

Global references respect configured globals and local bindings. For example, a function parameter named `URLSearchParams` is not reported with `"never"`.

This rule has no automatic fix or suggestions.

## Original documentation

- [eslint-plugin-n: prefer-global/url-search-params](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url-search-params.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/url-search-params.js)
