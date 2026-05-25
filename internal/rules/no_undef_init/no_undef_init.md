# no-undef-init

## Rule Details

Disallow initializing variables to `undefined`.

In JavaScript, a variable that is declared and not initialized to any value automatically gets the value of `undefined`. It's therefore unnecessary to initialize a variable to `undefined`.

This rule aims to eliminate `var` and `let` variable declarations that initialize to `undefined`.

Examples of **incorrect** code for this rule:

```javascript
var foo = undefined;
let bar = undefined;
```

Examples of **correct** code for this rule:

```javascript
var foo;
let bar;
const baz = undefined;
```

## Differences from ESLint

The autofix preserves TypeScript type annotations and definite assignment tokens. ESLint's fix removes from the end of the variable name to the end of the declarator, which in TypeScript would also remove any type annotation:

```typescript
// ESLint autofix: let a: string = undefined; → let a;  (type annotation lost)
// rslint autofix: let a: string = undefined; → let a: string;  (type annotation preserved)
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-undef-init
