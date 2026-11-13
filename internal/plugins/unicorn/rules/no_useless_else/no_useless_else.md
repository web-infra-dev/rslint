# no-useless-else

## Rule Details

Disallow an `else` after an `if` branch that always exits. This includes branches that end in `return`, `throw`, `break`, `continue`, or a call to the global `process.exit()`. The rule also recognizes nested conditionals, switches, loops, and `try` statements when execution cannot continue after the first branch.

Examples of **incorrect** code:

```javascript
function save(item) {
  if (!item) {
    return;
  } else {
    write(item);
  }
}

if (failed) {
  throw new Error();
} else if (ready) {
  start();
}
```

Examples of **correct** code:

```javascript
function save(item) {
  if (!item) {
    return;
  }
  write(item);
}

if (failed) {
  throw new Error();
}
if (ready) {
  start();
}
```

The fix moves the `else` body after the `if`. The rule reports without a fix when moving the body could change declaration scope, comment placement, multiline literal values, or automatic semicolon insertion.

## Options

This rule has no options. Enable it with `"unicorn/no-useless-else": "error"`.

## Differences from upstream

- A call to `process.exit()` is treated as an ordinary call when the `process` global is configured as `"off"` or a surrounding `with` statement can supply that name. Unicorn may report an `else` in these cases.
- If a `finally` block overrides a pending `break` or `continue`, the rule follows the destination that execution actually takes. This can keep an `else` that Unicorn reports, or report one when the `finally` block exits.
- The rule keeps an `else` if a `catch` could let execution continue after a JSX factory call, a destructured catch binding, or a read of `this` in a derived constructor. Reads of `this` are treated conservatively even after a call to `super()`.
- In a `try` block, more complex assignments and optional chains can keep an `else` that Unicorn reports. For example, the rule can prove that `value = 1` is safe when `value` is a local parameter, but it may keep the `else` after `value = other`.
- When an `else` body contains a lint directive such as `eslint-disable-next-line`, the rule reports it without an autofix. Keep the directive next to its intended statement if you move the body manually.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-else](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-else.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-else.js)
