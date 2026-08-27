# Require `var` declarations at the top of their scope (`vars-on-top`)

The `vars-on-top` rule requires `var` declarations to appear before executable
statements in the program or in a function body. Directive strings and imports
may appear before the declarations. Declarations inside nested blocks and loop
headers are reported; declarations at the start of a class static block are
also allowed.

## Rule Details

Examples of **incorrect** code:

```js
function example() {
  doSomething();
  var value = 1;
}
```

Examples of **correct** code:

```js
function example() {
  "use strict";
  var value = 1;
  doSomething(value);
}
```

## Differences from ESLint

rslint applies the same rule behavior to JavaScript and TypeScript syntax. As
with ESLint, the rule has no options and reports the complete declaration.
