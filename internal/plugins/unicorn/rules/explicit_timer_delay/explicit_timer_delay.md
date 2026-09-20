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
without arguments, and a spread first argument are not reported.

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

## Differences from upstream

Calls with a spread first argument are left unchanged in both modes. In
`"never"` mode, upstream v75.0.0 can mistake a later zero for the delay even
when the spread already supplies the callback and delay:

```javascript
const args = [callback, 1000];
setTimeout(...args, 0);
```

Here `0` is passed to the callback. Removing it changes behavior, so this rule
does not report the call. Pass the callback and delay explicitly to make their
positions clear and enable checking.

## Original Documentation

- [eslint-plugin-unicorn: explicit-timer-delay](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/explicit-timer-delay.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/explicit-timer-delay.js)
