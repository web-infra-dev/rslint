# callback-return

Requires a `return` statement with callback calls to help prevent invoking a callback more than once.

## Rule Details

This rule identifies callbacks by name. Within a function, a matching call must be inside a `return`, be the final callback expression in the function body, or immediately precede a `return` at the end of its enclosing block. Arrow functions with expression bodies return implicitly.

A callback expression is a standalone call or the direct right operand of a binary or logical expression, such as `cb && cb()`.

Examples of **incorrect** code:

```javascript
function read(err, callback) {
  if (err) {
    callback(err);
  }
  callback();
}
```

Examples of **correct** code:

```javascript
function read(err, callback) {
  if (err) {
    return callback(err);
  }
  callback();
}

function finish(err, cb) {
  if (err) {
    cb(err);
    return;
  }
  cb();
}
```

## Options

The single option is an array of callback names. It defaults to `["callback", "cb", "next"]`. A supplied array replaces those defaults; an empty array disables matching.

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/callback-return": ["error", ["done", "send.error", "send.success"]],
    },
  },
];
```

Names match the callee's exact source text. Object methods are supported when their receiver chain starts with an identifier. Spaces, comments, and computed property syntax are significant: `send.error` and `send["error"]` are different names.

## Differences from upstream

In rslint, enable the `node` plugin and configure `node/callback-return`. The upstream rule is named `n/callback-return`.

## Known Limitations

The rule does not track callback references passed to other functions, connect calls across nested functions, or infer that mutually exclusive branches invoke a callback only once. Calls outside functions are ignored. Optional calls such as `cb?.()` are not accepted as a final callback expression; use `return cb?.()` when returning from the function.

This rule has no automatic fix or suggestions.

## Original Documentation

- [eslint-plugin-n: callback-return](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/callback-return.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/callback-return.js)
