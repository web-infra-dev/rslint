# no-new-require

Disallow using `require` as a constructor.

## Rule Details

`new require('app-header')` constructs `require` itself. To construct the value
exported by a module, call `require` first and then use `new` on its result.

Examples of **incorrect** code for this rule:

```javascript
var appHeader = new require('app-header');
```

Examples of **correct** code for this rule:

```javascript
var AppHeader = require('app-header');
var appHeader = new AppHeader();

var anotherHeader = new (require('app-header'))();
```

This rule checks the identifier's name, including locally declared variables
named `require`. Member access such as `new loader.require()` is allowed.

## Options

This rule has no options and does not provide automatic fixes or suggestions.

## Original Documentation

- [eslint-plugin-n: no-new-require](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-new-require.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-new-require.js)
