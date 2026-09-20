# prefer-global/buffer

Enforces consistent use of the global `Buffer` or the `Buffer` export from the Node.js `buffer` module.

## Options

```js
import { globals } from '@rslint/core';

export default [{
  plugins: ['node'],
  languageOptions: { globals: globals.node },
  rules: {
    'node/prefer-global/buffer': ['error', 'always'],
  },
}];
```

- `"always"` (default) requires the global `Buffer`.
- `"never"` requires the module export instead of the global `Buffer`.

The rule recognizes `buffer` and `node:buffer`, including CommonJS, ES module imports and `process.getBuiltinModule()`. It follows aliases and destructuring and respects local bindings that shadow globals. It does not provide automatic fixes or suggestions.

## Examples

With `"always"`, these uses are incorrect:

```js
const { Buffer } = require('node:buffer');
import { Buffer as NodeBuffer } from 'buffer';
```

Use the global instead:

```js
const bytes = Buffer.alloc(16);
```

With `"never"`, the global use above is incorrect. Import `Buffer` instead:

```js
import { Buffer } from 'node:buffer';
const bytes = Buffer.alloc(16);
```

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/buffer.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/buffer.js)
