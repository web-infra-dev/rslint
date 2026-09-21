# prefer-global/text-encoder

## Rule details

Enforces consistent use of the global `TextEncoder` or the `TextEncoder` export from `util` and `node:util`. Module references include `require()`, `process.getBuiltinModule()`, and static imports and re-exports.

## Options

This rule accepts one string option:

- `"always"` (default): use the global `TextEncoder`.
- `"never"`: use `TextEncoder` from the module.

Enable Node.js globals in the configuration:

```javascript
import { globals } from "@rslint/core";

export default [
  {
    plugins: ["node"],
    languageOptions: { globals: globals.node },
    rules: {
      "node/prefer-global/text-encoder": ["error", "always"],
    },
  },
];
```

### always

Examples of **incorrect** code:

```javascript
const { TextEncoder } = require("util");
const encoder = new TextEncoder();

import { TextEncoder as Encoder } from "node:util";
```

Examples of **correct** code:

```javascript
const encoder = new TextEncoder();
```

### never

Examples of **incorrect** code:

```javascript
const encoder = new TextEncoder();
```

Examples of **correct** code:

```javascript
const { TextEncoder } = require("node:util");
const encoder = new TextEncoder();
```

Global references respect configured globals and local bindings. For example, a function parameter named `TextEncoder` is not reported with `"never"`.

This rule has no automatic fix or suggestions.

## Original documentation

- [eslint-plugin-n: prefer-global/text-encoder](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/text-encoder.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/text-encoder.js)
