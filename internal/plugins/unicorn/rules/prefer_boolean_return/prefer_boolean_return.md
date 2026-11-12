# prefer-boolean-return

## Rule details

Prefer returning a boolean expression directly when an `if` and its `else` or
the immediately following `return` return opposite boolean literals.

The rule keeps consecutive guard clauses that return the same value together.

**Incorrect code**

```javascript
function isValid(value) {
  if (value > 0) {
    return true;
  }

  return false;
}

function hasItems(items) {
  if (items.length) {
    return true;
  }

  return false;
}

function isInvalid(value) {
  if (value > 0) {
    return false;
  }

  return true;
}
```

**Correct code**

```javascript
function isValid(value) {
  return value > 0;
}

function hasItems(items) {
  return Boolean(items.length);
}

function isInvalid(value) {
  return !(value > 0);
}
```

## Automatic fixes

The fix returns conditions that already evaluate to booleans directly. It uses
`Boolean(...)` to coerce other positive conditions and `!` for inverted
conditions. A report has no fix when the edit would discard a comment or when
the replacement needs `Boolean(...)` and a local value shadows the global.

## Options

This rule has no options.

## Differences from upstream

- rslint keeps the diagnostic and withholds a fix that relies on `Boolean` when
  the global is disabled in the lint configuration. A TypeScript `import type`
  that introduces the name `Boolean` also prevents inserting a call, since
  TypeScript can reject the new call.
- Inside a JavaScript `with` block, the object can provide its own `Boolean` or
  `Array`. rslint only applies fixes that can preserve a boolean return without
  relying on those built-ins.

An enclosing variable or parameter named `Boolean` prevents inserting
`Boolean(...)` even when an inner TypeScript type alias has the same name.
Calling the parameter in this example could return a string.

```typescript
function check(value: unknown, Boolean: (value: unknown) => string) {
  {
    type Boolean = string;
    if (value) return true;
    return false;
  }
}
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-boolean-return.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-boolean-return.js)
