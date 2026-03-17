# no-inner-declarations

## Rule Details

Disallow variable and/or function declarations in nested blocks. Function declarations and, optionally, variable declarations (`var`) should only appear at the root level of a program or the body of a function. This does not apply to `let` or `const`, which are block-scoped by design.

Examples of **incorrect** code for this rule with the default `"functions"` option:

```javascript
if (test) {
  function doSomething() {}
}

while (test) {
  function doSomething() {}
}
```

Examples of **correct** code for this rule with the default `"functions"` option:

```javascript
function doSomething() {}

function doSomethingElse() {
  function doAnotherThing() {}
}
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

- `"functions"` (default): Only disallows `function` declarations in nested blocks.
- `"both"`: Disallows both `function` declarations and `var` declarations in nested blocks.

## Original Documentation

https://eslint.org/docs/latest/rules/no-inner-declarations
