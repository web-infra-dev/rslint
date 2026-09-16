# global-require

## Rule Details

Requires `require()` calls to appear at the top level of a module. Calls inside functions, blocks, loops, and other nested statements are reported. Calls to a locally declared `require` are ignored.

Top-level variable declarations, assignments, member access, calls, and conditional expressions are allowed. Other enclosing expressions, such as arrays, objects, logical expressions, and optional chains, are reported even at the top level.

Examples of **incorrect** code for this rule:

```javascript
function loadModule(name) {
  return require(name);
}

if (DEBUG) {
  require("debug");
}
```

Examples of **correct** code for this rule:

```javascript
const fs = require("fs");
const logger = DEBUG ? require("dev-logger") : require("logger");
require("./setup").initialize();

function useLoader(require) {
  return require("custom-module");
}
```

This rule has no options and does not provide automatic fixes or suggestions.

## Differences from upstream

Enable this rule as `node/global-require` with the `node` plugin in rslint. The upstream configuration uses `n/global-require`.

## Original Documentation

- [eslint-plugin-n: global-require](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/global-require.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/global-require.js)
