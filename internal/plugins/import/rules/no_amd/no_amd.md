# no-amd

## Rule Details

Disallow AMD `require` and `define` calls in ES module scope. This rule helps migrate AMD modules to ES imports.

A call is reported when its callee is the identifier `require` or `define`, it has exactly two arguments, and its first argument is an array literal. The second argument can be any expression.

Calls inside nested scopes, including functions and blocks, are allowed. CommonJS `require("module")` calls and files configured with `sourceType: "script"` or `"commonjs"` are also allowed. Local bindings named `require` or `define` do not exempt calls in module scope.

Examples of **incorrect** code for this rule:

```javascript
define(["a", "b"], function (a, b) {});
require(["a"], callback);
```

Examples of **correct** code for this rule:

```javascript
import a from "a";
import b from "b";

const c = require("c");
```

## Options

This rule has no options and provides no automatic fixes or suggestions.

## Original Documentation

- [eslint-plugin-import: no-amd](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-amd.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-amd.js)
