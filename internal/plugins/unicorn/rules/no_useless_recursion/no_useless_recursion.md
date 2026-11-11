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

## Differences from upstream

Rslint skips calls to a function declaration's name when that name is
reassigned in the same file. A saved reference can still call the original
function after a later reassignment.

```javascript
function walk() {
  walk = () => 1;
  return walk();
}
```

In scripts, writing a known property of `globalThis` can also replace a
top-level function with that name.

```javascript
function walk() {
  globalThis.walk = () => 1;
  return walk();
}
```

In non-strict scripts, a function declared in a block can make a different
function available under the same name outside the block. Rslint skips calls
that could use that function.

```javascript
function walk(n) {
  if (true) {
    function walk() { return 1; }
  }
  if (n) return walk(n - 1);
  return 0;
}
```

In scripts, Rslint skips a call when a `with` statement can supply the same name
from an object.

```javascript
function walk() {
  with ({ walk: () => 1 }) {
    return walk();
  }
}
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-recursion.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-recursion.js)
