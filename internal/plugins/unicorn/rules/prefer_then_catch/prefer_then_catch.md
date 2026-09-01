# prefer-then-catch

## Rule Details

Prefer `.then(...).catch(...)` over passing a rejection handler to `.then()`.
A `.catch` handler attached after `.then` is reached both for rejections
forwarded by `.then` and for rejections thrown from the fulfillment handler,
while a second argument to `.then` is only called when the upstream promise
itself rejects.

The rule reports `.then(fulfilled, rejected)` calls when both handlers are
non-nullish and the receiver type's `.catch` is callable. It offers a
suggestion that rewrites the call to `.then(fulfilled).catch(rejected)` when
the rejection handler can safely be moved (identifiers and function expressions
qualify; calls with possible side effects do not).

Examples of **incorrect** code for this rule:

```javascript
promise.then(onFulfilled, onRejected);
```

Examples of **correct** code for this rule:

```javascript
promise.then(onFulfilled).catch(onRejected);
```

When the rejection handler is a call with potential side effects, no
suggestion is offered but the diagnostic still fires:

```javascript
promise.then(onFulfilled, createRejectionHandler());
// Diagnostic still fires; user must rewrite manually.
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-then-catch](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-then-catch.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-then-catch.js)