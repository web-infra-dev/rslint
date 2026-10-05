# no-unnecessary-await

## Rule Details

Disallow awaiting values whose syntax shows they are not promises, such as
arrays, literals, functions, and values that have already been awaited.

Examples of **incorrect** code:

```javascript
await await promise;
await [promise1, promise2];
```

Examples of **correct** code:

```javascript
await promise;
await Promise.allSettled([promise1, promise2]);
```

This rule uses syntax rather than type information. TypeScript assertions such
as `as`, `satisfies`, and non-null assertions do not hide the awaited value.

Awaiting a non-promise value still pauses execution for a microtask. Automatic
fixes are offered only at the end of a function or module, where removing
`await` does not make subsequent code run earlier. Function and class
expressions are reported without a fix.

## Options

This rule has no options.

## Differences from upstream

In JavaScript modules, a top-level `await !value` produces a syntax error
instead of this rule's diagnostic. Write `await (!value)` to receive the
diagnostic and automatic fix. This limitation does not affect
TypeScript files or `await !value` inside an async function.

## Original Documentation

- [eslint-plugin-unicorn: no-unnecessary-await](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-unnecessary-await.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-unnecessary-await.js)
