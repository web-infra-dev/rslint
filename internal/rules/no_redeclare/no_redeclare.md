# no-redeclare

## Rule Details

This rule disallows declaring the same variable more than once in the same scope.

Examples of **incorrect** code for this rule:

```javascript
var a = 3;
var a = 10;
```

```javascript
function a() {}
function a() {}
```

Examples of **correct** code for this rule:

```javascript
var a = 3;
var b = function () {
  var a = 10;
};
```

```javascript
if (foo) {
  let a = 1;
} else {
  let a = 2;
}
```

## Options

### `builtinGlobals` (default: `true`)

When `true`, this rule reports redeclarations of ECMAScript built-in globals.
Configured `languageOptions.globals` also participate as built-ins. Active
`/* global */` directives participate as declarations in either mode; a final
`:off` setting removes that inline global.

```json
{ "no-redeclare": ["error", { "builtinGlobals": true }] }
```

```javascript
var Object = 0;
```

Set `builtinGlobals` to `false` to allow redeclaring built-in global names.

```json
{ "no-redeclare": ["error", { "builtinGlobals": false }] }
```

```javascript
var Object = 0;
```

## Differences from ESLint

- rslint determines module files from top-level `import` / `export` syntax, not a separate `sourceType` override.

## Original Documentation

[https://eslint.org/docs/latest/rules/no-redeclare](https://eslint.org/docs/latest/rules/no-redeclare)
