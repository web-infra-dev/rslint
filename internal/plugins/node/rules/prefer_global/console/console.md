# prefer-global/console

Enforce consistent use of the global `console` variable or the `console` module.

## Options

- `"always"` (default): use the global `console` variable.
- `"never"`: load `console` from a module instead of using the global variable.

```javascript
export default [{
  plugins: ['node'],
  rules: {
    'node/prefer-global/console': ['error', 'always'],
  },
}];
```

With `"always"`, these module references are **incorrect**:

```javascript
const console = require('console');
import importedConsole from 'node:console';
const logger = process.getBuiltinModule('console');
```

Use the global variable instead:

```javascript
console.log('hello');
```

With `"never"`, using the global variable is **incorrect**:

```javascript
console.log('hello');
```

Load the module instead:

```javascript
const console = require('node:console');
console.log('hello');
```

The rule recognizes CommonJS loads, `process.getBuiltinModule()`, static imports
and re-exports, including the `node:` prefix. Local bindings that shadow globals
are respected. The rule provides no automatic fixes or suggestions.

## Original documentation

- [eslint-plugin-n: prefer-global/console](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/console.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/console.js)
