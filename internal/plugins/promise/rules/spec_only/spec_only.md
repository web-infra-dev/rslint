# spec-only

Disallow non-standard Promise methods (`promise/spec-only`).

Depending on Promise extensions can make it harder to migrate between implementations. This rule checks member access on `Promise`, including method references and calls, and checks members of `Promise.prototype`.

## Rule Details

The allowed static methods are `all`, `allSettled`, `any`, `race`, `reject`, `resolve`, and `withResolvers`. The allowed prototype methods are `then`, `catch`, and `finally`.

Examples of incorrect code:

```js
const x = Promise.done('bad')
const done = Promise['done']
Promise.prototype.done
```

Examples of correct code:

```js
const x = Promise.resolve('good')
Promise.allSettled(tasks)
Promise.withResolvers()
Promise.prototype.then
Promise[method]
```

Computed identifiers such as `Promise[method]` are allowed. String keys are checked directly. Only members whose object is named `Promise` are checked; aliases and promise instances are not tracked.

## Options

### `allowedMethods`

An array of additional method names to allow on both `Promise` and `Promise.prototype`. Defaults to an empty array.

```js
{
  'promise/spec-only': ['error', { allowedMethods: ['done'] }]
}
```

## Compatibility

Optional access to a private member, such as `Promise?.#done` inside a class declaring `#done`, is rejected by the TypeScript parser before this rule runs.

For computed regular-expression keys, rslint uses the literal's string representation. ESLint running on an older Node.js version can instead report `Promise.null` when that runtime cannot construct a regular expression using newer syntax, such as modifier groups in `Promise[/(?i:a)/]`.

## Original Documentation

- [eslint-plugin-promise: spec-only](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/spec-only.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/rules/spec-only.js)
