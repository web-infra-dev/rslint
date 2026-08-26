# no-inner-declarations

## Rule Details

Disallow variable and/or function declarations outside the root of a program,
function body, or class static block body. With the `"both"` option, this also
applies to `var` declarations in `for`, `for-in`, and `for-of` headers. It does
not apply to `let`, `const`, `using`, or `await using`, which are block-scoped.

By default, nested function declarations are allowed only when the code is in
strict mode and `languageOptions.ecmaVersion` is `2015` or newer. Use the
`blockScopedFunctions` option to disallow them in those contexts too.

Examples of **incorrect** code for this rule with the default options in a
non-strict script:

```javascript
if (test) {
  function doSomething() {}
}
```

Examples of **incorrect** code for this rule with `{ blockScopedFunctions: "disallow" }`:

```json
{
  "no-inner-declarations": [
    "error",
    "functions",
    { "blockScopedFunctions": "disallow" }
  ]
}
```

```javascript
"use strict";

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

"use strict";

if (test) {
  function doSomething() {}
}

export function foo() {}
```

Examples of **incorrect** code for this rule with the `"both"` option:

```json
{ "no-inner-declarations": ["error", "both"] }
```

```javascript
if (test) {
  var x = 1;
}

function doSomething() {
  if (test) {
    var x = 1;
  }
}

for (var i = 0; i < items.length; i++) {}

for (var key in object) {}

for (var value of values) {}
```

Examples of **correct** code for this rule with the `"both"` option:

```json
{ "no-inner-declarations": ["error", "both"] }
```

```javascript
var x = 1;

function doSomething() {
  var y = 2;
}

if (test) {
  let x = 1;
}

for (const value of values) {}
```

## Options

- `"functions"` (default): Only disallows `function` declarations in nested blocks (when `blockScopedFunctions` is `"disallow"`).
- `"both"`: Disallows both `function` declarations (when `blockScopedFunctions` is `"disallow"`) and `var` declarations in nested blocks.
- `{ blockScopedFunctions: "allow" | "disallow" }` (default `"allow"`): With
  `"allow"`, nested function declarations are permitted only in strict code
  when `languageOptions.ecmaVersion` is `2015` or newer. With `"disallow"`, they
  are always checked.

## Original Documentation

- [ESLint: no-inner-declarations](https://eslint.org/docs/latest/rules/no-inner-declarations)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-inner-declarations.js)
