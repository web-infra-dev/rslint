# no-unnecessary-array-splice-count

## Rule Details

Remove a redundant `.length`, `Infinity`, or `Number.POSITIVE_INFINITY` count from a two-argument `splice` or `toSpliced` call.

Calls with additional insertion arguments are left unchanged. Receivers known not to be arrays are excluded; the receiver classification also accepts typed arrays, matching upstream.

Examples of **incorrect** code:

```javascript
array.splice(1, array.length);
```

Examples of **correct** code:

```javascript
array.splice(1);
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Differences from upstream

Locally shadowed, disabled, or previously reassigned `Infinity` and `Number`
globals are not assumed to be built-ins. For `.length` rewrites, known
getter-backed receiver paths or a first argument that can change the receiver
are skipped. These guards avoid semantics-changing autofixes that Unicorn
v75.0.0 can still offer.

## Original Documentation

- [eslint-plugin-unicorn: no-unnecessary-array-splice-count](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-unnecessary-array-splice-count.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-unnecessary-array-splice-count.js)
