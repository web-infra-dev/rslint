# prefer-global/text-decoder

Enforces consistent use of the global `TextDecoder` or the `TextDecoder` export from Node.js's `util` module.

## Options

- `"always"` (default): use the global `TextDecoder`.
- `"never"`: use `TextDecoder` from `util` or `node:util`.

The rule recognizes CommonJS `require`, `process.getBuiltinModule`, and ES module imports and re-exports. It follows aliases and static property access, and respects shadowed or disabled globals. It does not provide automatic fixes or suggestions.

### always

Examples of **incorrect** code:

```javascript
const { TextDecoder } = require("node:util");
const decoder = new TextDecoder();
```

Examples of **correct** code:

```javascript
const decoder = new TextDecoder();
```

### never

Examples of **incorrect** code:

```javascript
const decoder = new TextDecoder();
```

Examples of **correct** code:

```javascript
import { TextDecoder } from "node:util";
const decoder = new TextDecoder();
```

## Original Documentation

- [eslint-plugin-n: prefer-global/text-decoder](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/text-decoder.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/text-decoder.js)
