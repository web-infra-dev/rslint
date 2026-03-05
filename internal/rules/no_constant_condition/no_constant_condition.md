# no-constant-condition

## Rule Details

Disallows constant expressions in conditions of `if`, `while`, `do-while`, `for` statements, and ternary expressions. A constant condition is one that always evaluates to the same truthy or falsy value, such as a literal, an always-truthy expression like an object literal, or a compile-time computable expression. This usually indicates a programmer error.

By default, `while (true)` loops are allowed (via the `"allExceptWhileTrue"` option) since they are a common pattern for intentional infinite loops with a `break` inside.

Examples of **incorrect** code for this rule:

```javascript
if (true) {
}

if ('hello') {
}

while (1) {}

for (; false; ) {}

var result = 0 ? a : b;
```

Examples of **correct** code for this rule:

```javascript
if (x === 0) {
}

while (true) {} // allowed by default

while (condition) {}

for (; i < 10; i++) {}

var result = x ? a : b;
```

## Original Documentation

- [ESLint no-constant-condition](https://eslint.org/docs/latest/rules/no-constant-condition)
