# consistent-this

## Rule Details

It is often necessary to capture the current execution context in order to make it available subsequently, for example in a callback. This rule enforces two things about variables with the designated alias names for `this`:

- If a variable with a designated name is declared, it must be either initialized (in the declaration) or assigned (in the same scope as the declaration) the value `this`.
- If a variable is initialized or assigned the value `this`, the name of the variable must be a designated alias.

Examples of **incorrect** code for this rule with the default `"that"` option:

```javascript
let that = 42;

let self = this;

that = 42;

self = this;
```

Examples of **correct** code for this rule with the default `"that"` option:

```javascript
let that = this;

const self = 42;

let foo;

that = this;

foo.bar = this;
```

Examples of **incorrect** code for this rule with the default `"that"` option, if the variable is not initialized:

```javascript
let that;
function f() {
    that = this;
}
```

Examples of **correct** code for this rule with the default `"that"` option, if the variable is not initialized:

```javascript
let that;
that = this;
```

```javascript
let foo = 42, that;
that = this;
```

## Options

This rule has one or more string options: designated alias names for `this` (default `"that"`).

Examples of **incorrect** code for this rule with `"self", "vm"`:

```json
{ "consistent-this": ["error", "self", "vm"] }
```

```javascript
let self = this;
let vm = 42;
```

## Differences from ESLint

- rslint always treats a nested block (`if`, `for`, `try`, a bare `{}`, ...) as a different scope from its enclosing function, so an alias declared in the function and assigned `this` only inside such a block is still reported. ESLint's behavior here depends on the configured ECMAScript version: under `ecmaVersion: 5` it does not create a separate scope for the block, so the same code is accepted.

## Original Documentation

- [ESLint: consistent-this](https://eslint.org/docs/latest/rules/consistent-this)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/consistent-this.js)
