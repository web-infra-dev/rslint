# no-useless-error-capture-stack-trace

## Rule Details

Report redundant `Error.captureStackTrace()` calls inside constructors of
classes that directly extend a built-in error constructor. The first argument
must be `this`, and the second must identify the same class, `new.target`, or
`this.constructor`.

Example of **incorrect** code:

```javascript
class CustomError extends Error {
  constructor(message) {
    super(message);
    Error.captureStackTrace(this, CustomError);
  }
}
```

Example of **correct** code:

```javascript
class CustomError extends Error {
  constructor(message) {
    super(message);
  }
}
```

Nested arrow functions retain the constructor's `this` binding. Ordinary
nested functions, unrelated methods, static blocks, and nested classes are
checked in their own scope rather than treated as the outer constructor.
Locally shadowed or disabled error globals are not matched.

An optional call such as `Error.captureStackTrace?.(this, new.target)` is also
checked. Optional member access and computed method names are not matched.

## Automatic Fixes

A standalone call used as an expression statement directly inside a block can
be removed. Calls used as return values, arrow-expression bodies, or unbraced
control-flow bodies are reported without an automatic fix.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-error-capture-stack-trace](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-useless-error-capture-stack-trace.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-useless-error-capture-stack-trace.js)
