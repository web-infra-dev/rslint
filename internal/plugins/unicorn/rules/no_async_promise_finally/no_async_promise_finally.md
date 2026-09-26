# no-async-promise-finally

## Rule Details

Disallow async functions as `Promise#finally()` callbacks.

An async finalizer can replace the original promise outcome if it rejects. For
example, when the original promise rejects and the async finalizer also rejects,
the resulting promise rejects with the finalizer error rather than the original
one.

The rule reports async callbacks passed directly to `.finally()`, as well as
callbacks referenced through a local async function declaration or a `const`
initializer.

Examples of **incorrect** code:

```js
promise.finally(async () => {
  await cleanup();
});

async function cleanupAgain() {}
promise.finally(cleanupAgain);
```

Examples of **correct** code:

```js
promise.finally(() => {
  cleanupSynchronously();
});

function cleanupAgain() {}
promise.finally(cleanupAgain);
```

Imported functions, member expressions, `let` bindings, and other dynamic
callback references are intentionally ignored because the rule cannot locally
prove that they are async.

## Options

This rule has no options. It is enabled at `error` severity in
`unicornPlugin.configs.recommended`.

## Differences from upstream

Rslint preserves Unicorn v75.0.0's best-effort source-only behavior. When type
information is available, a call is skipped only when the receiver is proven
not to be Promise-like. `any`, `unknown`, TypeScript error types, and mixed
Promise/non-Promise unions remain indeterminate and are therefore checked just
like upstream.

Static computed property names use Rslint's canonical static string evaluator,
so `promise["finally"](...)`, template literals, and locally constant
`"finally"` keys are recognized without a bespoke resolver.

## Original Documentation

- [eslint-plugin-unicorn: no-async-promise-finally](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-async-promise-finally.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-async-promise-finally.js)
