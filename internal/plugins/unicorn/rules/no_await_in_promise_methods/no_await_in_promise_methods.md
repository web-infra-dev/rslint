# no-await-in-promise-methods

## Rule Details

Disallow using `await` on promises passed directly to `Promise.all()`,
`Promise.allSettled()`, `Promise.any()`, or `Promise.race()`. Awaiting an
individual promise before passing it to one of these methods prevents that
promise from running concurrently with the others and is likely a mistake.

Examples of **incorrect** code for this rule:

```javascript
Promise.all([await promise, anotherPromise]);
Promise.allSettled([await promise, anotherPromise]);
Promise.any([await promise, anotherPromise]);
Promise.race([await promise, anotherPromise]);
```

Examples of **correct** code for this rule:

```javascript
Promise.all([promise, anotherPromise]);
Promise.allSettled([promise, anotherPromise]);
Promise.any([promise, anotherPromise]);
Promise.race([promise, anotherPromise]);
```

## Original Documentation

- [eslint-plugin-unicorn: no-await-in-promise-methods](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-await-in-promise-methods.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-await-in-promise-methods.js)
