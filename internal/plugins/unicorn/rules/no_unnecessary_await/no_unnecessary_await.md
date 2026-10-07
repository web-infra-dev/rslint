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

This rule has no options. It is enabled at `error` severity in
`unicornPlugin.configs.recommended`.

## Differences from upstream

In `.js` files without an `import` or `export`, setting
`languageOptions.sourceType: "module"` alone does not reliably enable this
rule for top-level `await`. For example, `await []` produces a syntax error,
while `await (a + b)` is not reported. Add `export {};` or use a tsconfig with
`compilerOptions.moduleDetection: "force"` selected through
`languageOptions.parserOptions.project` to check these expressions.

Even when the file is recognized as a module, a top-level `await !value` in
JavaScript produces a syntax error instead of this rule's diagnostic. Write
`await (!value)` to receive the diagnostic and automatic fix. This additional
limitation does not affect TypeScript files or `await !value` inside an async
function.

## Original Documentation

- [eslint-plugin-unicorn: no-unnecessary-await](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-unnecessary-await.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-unnecessary-await.js)
