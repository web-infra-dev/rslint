# no-unnecessary-slice-end

## Rule Details

Remove a redundant `.length`, `Infinity`, or `Number.POSITIVE_INFINITY` end argument from a two-argument `slice` call. Arrays, strings, and typed arrays are supported.

The length must refer to the same receiver. Optional member access is supported, while optional calls, computed method names, and spread arguments are not changed.

Examples of **incorrect** code:

```javascript
const copy = array.slice(1, array.length);
```

Examples of **correct** code:

```javascript
const copy = array.slice(1);
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Differences from upstream

Locally shadowed, disabled, or previously reassigned `Infinity` and `Number`
globals are not assumed to be built-ins. When type information proves that the
receiver is not an array, string, or typed array, the call is also left alone.
For `.length` rewrites, the shared slice/splice safety helper also requires a
repeatable receiver and a simple side-effect-free first argument; more complex
first-argument expressions are conservatively left unchanged. These guards avoid
semantics-changing autofixes that Unicorn v75.0.0 can still offer.

## Original Documentation

- [eslint-plugin-unicorn: no-unnecessary-slice-end](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-unnecessary-slice-end.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-unnecessary-slice-end.js)
