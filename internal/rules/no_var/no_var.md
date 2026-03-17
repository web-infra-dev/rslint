# no-var

## Rule Details

Requires `let` or `const` instead of `var`. ECMAScript 6 introduced `let` and `const` as alternatives to `var` for variable declarations. `let` and `const` provide block scoping, which helps avoid common issues caused by the function scoping of `var`.

Examples of **incorrect** code for this rule:

```javascript
var x = 'y';
var CONFIG = {};
```

Examples of **correct** code for this rule:

```javascript
let x = 'y';
const CONFIG = {};
```

## Original Documentation

- [ESLint no-var](https://eslint.org/docs/latest/rules/no-var)
