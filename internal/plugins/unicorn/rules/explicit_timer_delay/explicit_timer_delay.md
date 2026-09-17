# explicit-timer-delay

## Rule Details

Require an explicit delay for global `setTimeout()` and `setInterval()` calls.
The rule also recognizes calls through `window`, `globalThis`, `global`, and
`self`, when the global is enabled and not shadowed by a local declaration.

Examples of **incorrect** code with the default options:

```javascript
setTimeout(callback);
globalThis.setInterval(callback);
```

Examples of **correct** code:

```javascript
setTimeout(callback, 0);
globalThis.setInterval(callback, 1000);
```

Optional calls, computed method names, locally declared timer functions, calls
without arguments, and a lone spread argument are not reported.

## Options

Type: `"always" | "never"`

Default: `"always"`

`"always"` requires a delay when a timer has exactly one non-spread argument.
Its automatic fix adds `, 0` after the complete callback argument, preserving
parentheses and any existing trailing comma.

`"never"` disallows an explicit zero delay when there are exactly two
arguments. Non-zero delays and calls with additional callback arguments remain
allowed. Signed and parenthesized zero literals are recognized.

```javascript
{
  'unicorn/explicit-timer-delay': ['error', 'never']
}
```

With `"never"`, `setTimeout(callback, 0)` becomes `setTimeout(callback)`.

## Original Documentation

- [eslint-plugin-unicorn: explicit-timer-delay](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/explicit-timer-delay.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/explicit-timer-delay.js)
