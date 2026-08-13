# no-label-var

## Rule Details

This rule aims to create clearer code by disallowing the bad practice of creating a label that shares a name with a variable that is in scope.

Examples of **incorrect** code for this rule:

```javascript
var x = foo;
function bar() {
x:
  for (;;) {
    break x;
  }
}
```

Examples of **correct** code for this rule:

```javascript
// The variable that has the same name as the label is not in scope.

function foo() {
  var q = t;
}

function bar() {
q:
  for (;;) {
    break q;
  }
}
```

## Options

This rule has no options.

## Differences from ESLint

- On files without type information, only declarations written in the file
  plus [configured globals](/config/#languageoptionsglobals)
  (`languageOptions.globals` / `/* global foo */`) are checked; clashes with
  built-in globals (`Promise`, `Array`, …) are not reported in that case. The
  `globals` catalog exported by `@rslint/core` can supply runtime-specific
  declarations.

## Original Documentation

- [ESLint: no-label-var](https://eslint.org/docs/latest/rules/no-label-var)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-label-var.js)
