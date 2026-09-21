# prefer-global/url

Enforces consistent use of the global `URL` or the `URL` export from Node.js's
`url` module.

## Options

- `"always"` (default): use the global `URL`.
- `"never"`: use `URL` from `url` or `node:url`.

Configure Node.js globals when enabling the rule:

```javascript
import { globals } from "@rslint/core";

export default [
  {
    plugins: ["node"],
    languageOptions: { globals: globals.node },
    rules: {
      "node/prefer-global/url": ["error", "always"],
    },
  },
];
```

## Rule details

With `"always"`, these module references are **incorrect**:

```javascript
const { URL } = require("node:url");
const url = process.getBuiltinModule("url");
new url.URL("https://example.com");
```

Use the global instead:

```javascript
const url = new URL("https://example.com");
```

With `"never"`, the global example is **incorrect**. Import the constructor:

```javascript
import { URL } from "node:url";
const url = new URL("https://example.com");
```

The rule follows static imports, re-exports, destructuring and aliases.
It respects local bindings and configured globals. It does not check
`URLSearchParams` or provide automatic fixes or suggestions.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/url.js)
