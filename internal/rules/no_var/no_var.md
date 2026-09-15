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

- [ESLint: no-var](https://eslint.org/docs/latest/rules/no-var)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/no-var.js)

## Differences from ESLint

Rslint leaves some declarations unfixed when replacing `var` with `let` would
introduce a temporal dead zone, even if ESLint offers a fix. This includes a
function called during the variable's initialization that reads that variable:

```javascript
var value = read();
function read() {
  return value;
}
```

The same protection applies to immediately invoked callbacks, constructors,
getters and loop binding defaults that read an uninitialized variable. Stored
callbacks and methods that run after initialization can still be fixed.
