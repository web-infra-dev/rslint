# prefer-global/process

Enforce consistent use of the global `process` variable or the `process` module.

## Options

This rule accepts one string option:

- `"always"` (default): use the global `process` variable.
- `"never"`: import or require the `process` module.

The rule recognizes both `process` and `node:process` module names, including
ESM imports and re-exports, `require()`, and `process.getBuiltinModule()`.
Global references respect local bindings and configured globals.
Enable Node.js globals in the configuration:

```js
import { globals } from '@rslint/core';

export default [
  {
    plugins: ['node'],
    languageOptions: { globals: globals.node },
    rules: {
      'node/prefer-global/process': ['error', 'always'],
    },
  },
];
```

### always

Incorrect:

```js
const process = require('process');
import proc from 'node:process';
```

Correct:

```js
process.exit(0);
```

### never

Incorrect:

```js
process.exit(0);
```

Correct:

```js
const process = require('process');
process.exit(0);
```

This rule does not provide automatic fixes or suggestions.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/process.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/process.js)
