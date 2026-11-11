# no-useless-recursion

## Rule details

Disallows directly returning a call to the same named function. A loop can
express simple tail recursion without growing the call stack.

The rule ignores optional calls, calls to shadowed names, and generator
functions. It also skips returns inside `try`, `catch`, or `finally`, or while a
resource declared with `using` or `await using` is still in scope.

Examples of **incorrect** code for this rule

```javascript
function foo(bar) {
  if (bar.baz) {
    return foo(bar.baz);
  }

  return bar;
}
```

Examples of **correct** code for this rule

```javascript
function foo(bar) {
  while (bar.baz) {
    bar = bar.baz;
  }

  return bar;
}
```

```javascript
function foo(bar) {
  if (Array.isArray(bar)) {
    return bar.map(baz => foo(baz));
  }

  return bar;
}
```

## Options

This rule has no options or automatic fixes.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-recursion.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-recursion.js)
