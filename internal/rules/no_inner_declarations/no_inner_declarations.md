# no-inner-declarations

## Rule Details

Disallow variable and/or function declarations in nested blocks. Function declarations and, optionally, variable declarations (`var`) should only appear at the root level of a program or the body of a function. This does not apply to `let` or `const`, which are block-scoped by design.

By default, function declarations in blocks are allowed because they are valid in ES2015+ strict mode (block-scoped). Use the `blockScopedFunctions` option to disallow them.

Examples of **incorrect** code for this rule with `{ blockScopedFunctions: "disallow" }`:

```javascript
if (test) {
  function doSomething() {}
}

while (test) {
  function doSomething() {}
}
```

Examples of **correct** code for this rule with the default options:

```javascript
function doSomething() {}

function doSomethingElse() {
  function doAnotherThing() {}
}

// Block-scoped functions are allowed by default
if (test) {
  function doSomething() {}
}

export function foo() {}
```

Examples of **incorrect** code for this rule with the `"both"` option:

```javascript
if (test) {
  var x = 1;
}

function doSomething() {
  if (test) {
    var x = 1;
  }
}
```

Examples of **correct** code for this rule with the `"both"` option:

```javascript
var x = 1;

function doSomething() {
  var y = 2;
}

if (test) {
  let x = 1;
}
```

## Options

- `"functions"` (default): Only disallows `function` declarations in nested blocks (when `blockScopedFunctions` is `"disallow"`).
- `"both"`: Disallows both `function` declarations (when `blockScopedFunctions` is `"disallow"`) and `var` declarations in nested blocks.
- `{ blockScopedFunctions: "allow" | "disallow" }` (default `"allow"`): Controls whether function declarations inside blocks are reported. When `"allow"` (default), block-scoped function declarations are permitted as they are valid in ES2015+ strict mode.

## Original Documentation

https://eslint.org/docs/latest/rules/no-inner-declarations
