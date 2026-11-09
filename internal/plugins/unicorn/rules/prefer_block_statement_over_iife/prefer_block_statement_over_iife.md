# prefer-block-statement-over-iife

## Rule Details

Prefer a block statement when an immediately invoked function expression (IIFE)
only creates a scope for local declarations.

The rule checks standalone calls to anonymous functions and arrow functions
with block bodies. Async functions, generators, and IIFEs with arguments,
parameters, or a `return` statement are allowed. The rule also allows functions
that need their own scope for `var`, directives such as `'use strict'`, or direct
`eval`, and ordinary functions that use their own `this`, `arguments`, or
`new.target`.

**Incorrect**

```javascript
(() => {
  const value = getValue();
  run(value);
})();
```

**Correct**

```javascript
{
  const value = getValue();
  run(value);
}

const value = (() => {
  return getValue();
})();
```

## Automatic Fixes

The fix replaces the call statement with the function's block. Comments inside
the block are preserved. A comment in the surrounding IIFE syntax leaves the
diagnostic available for manual review.

## Options

This rule has no options.

## Differences from upstream

rslint allows some IIFEs that Unicorn v79.0.0 reports when a block could change
what the code can access.

- In CommonJS, `(() => { function load() {} load(); })();` keeps `load` private
  to the function. In non-strict code, a replacement block could expose the
  name outside that block.
- A callback such as `() => eval('arguments.length')` can read the enclosing
  IIFE's arguments. A JSX tag such as `<this.View />` or `<arguments.View />`, or
  a TypeScript parameter decorator such as `@decorate(this)`, can also rely on
  the enclosing IIFE's bindings.

In the following cases, rslint reports the IIFE for manual review.

- A generic IIFE such as `(<T>() => { run<T>(); })<string>();` would lose access
  to `T` if it became a block.
- In a TypeScript project targeting ES5, moving a function declaration into a
  block can cause a compile error. An IIFE such as
  `(() => { function load() {} load(); })();` keeps that declaration inside a
  function.

## Original Documentation

- [eslint-plugin-unicorn: prefer-block-statement-over-iife](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-block-statement-over-iife.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-block-statement-over-iife.js)
