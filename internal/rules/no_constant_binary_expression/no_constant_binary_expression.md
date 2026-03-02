# no-constant-binary-expression

## Rule Details

Disallows expressions where the operation is guaranteed to always produce the same result, indicating a likely logic error. This includes comparisons against newly constructed objects (which can never be referentially equal to anything), constant short-circuit expressions where the left-hand side determines the result regardless of the right-hand side, and comparisons that always evaluate to the same boolean value.

Examples of **incorrect** code for this rule:

```javascript
if (x === []) {}             // always false, new array is never === to anything

const value = x ?? "default" || y; // constant ?? on left when left is non-nullish

if (x === true && "foo") {}  // constant && with literal left side

if ({} === {}) {}            // two new objects are never equal
```

Examples of **correct** code for this rule:

```javascript
if (x === someVar) {
}

const value = x ?? y;

if (x && y) {
}

if (x === null) {
}
```

## Original Documentation

- [ESLint no-constant-binary-expression](https://eslint.org/docs/latest/rules/no-constant-binary-expression)
