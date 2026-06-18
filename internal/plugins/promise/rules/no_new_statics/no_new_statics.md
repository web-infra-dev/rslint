# no-new-statics

Disallow calling `new` on a Promise static method.

## Rule Details

Promise static methods (`resolve`, `reject`, `all`, `allSettled`, `any`, `race`, `withResolvers`) return Promises directly — they are not constructors. Calling `new` on them is almost always a mistake and produces unexpected behavior.

Examples of **incorrect** code for this rule:

```javascript
new Promise.resolve(value);
new Promise.reject(reason);
new Promise.all([p1, p2]);
new Promise.allSettled([p1, p2]);
new Promise.any([p1, p2]);
new Promise.race([p1, p2]);
new Promise.withResolvers();
```

Examples of **correct** code for this rule:

```javascript
Promise.resolve(value);
Promise.reject(reason);
Promise.all([p1, p2]);
new Promise(function (resolve, reject) {});
new SomeClass.resolve();
```

## Autofix

This rule provides an autofix that removes the `new` keyword.

## Differences from ESLint

- `new (Promise as any).resolve()` and similar TS type-assertion or non-null-assertion wrappers around `Promise` are **not** flagged. Under `@typescript-eslint/parser`, type wrappers are stripped before the rule sees the AST, so ESLint would flag these. rslint preserves the TS node structure and only unwraps parentheses.

## Original Documentation

- [eslint-plugin-promise: no-new-statics](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-new-statics.md)
