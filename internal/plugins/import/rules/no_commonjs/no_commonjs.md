# no-commonjs

## Rule Details

Disallow CommonJS imports and exports when migrating to ES modules.

The rule reports `require()` calls with one string literal or a template literal
without substitutions in the module's variable scope. Calls inside functions,
dynamic arguments, and calls with multiple arguments are allowed. It also reports
`module.exports` and `exports.*` member access, including reads.

Examples of **incorrect** code for this rule:

```javascript
const path = require('node:path');
module.exports = { path };
exports.name = 'example';
```

Examples of **correct** code for this rule:

```javascript
import path from 'node:path';
export { path };
export const name = 'example';
```

The rule does not provide automatic fixes or suggestions.

## Options

| Option                    | Default | Behavior                                                                   |
| ------------------------- | ------- | -------------------------------------------------------------------------- |
| `allowRequire`            | `false` | Allow all `require()` calls.                                               |
| `allowConditionalRequire` | `true`  | Allow calls inside conditionals or `try` statements.                       |
| `allowPrimitiveModules`   | `false` | Allow `module.exports` assignments without an object literal on the right. |

### `allowRequire`

With `{ "allowRequire": true }`, this is allowed:

```javascript
const path = require('node:path');
```

CommonJS exports are still reported.

### `allowConditionalRequire`

By default, the following calls are allowed. Set
`{ "allowConditionalRequire": false }` to report them:

```javascript
const optional = enabled && require('optional');

if (enabled) {
  require('optional');
}

try {
  require('optional');
} catch (error) {}
```

Conditional expressions (`condition ? a : b`) and logical expressions (`&&`,
`||`, `??`) also qualify. Loop and switch statements do not.

### `allowPrimitiveModules`

With `{ "allowPrimitiveModules": true }`, these assignments are allowed:

```javascript
module.exports = 'value';
module.exports = function create() {};
```

Object literal assignments and `exports.*` are still reported:

```javascript
module.exports = { name: 'example' };
exports.create = function create() {};
```

The legacy string option `"allow-primitive-modules"` enables the same behavior.

## Original Documentation

- [eslint-plugin-import: no-commonjs](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-commonjs.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-commonjs.js)
