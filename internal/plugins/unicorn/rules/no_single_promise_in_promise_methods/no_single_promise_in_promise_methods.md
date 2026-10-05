# no-single-promise-in-promise-methods

## Rule Details

Disallow passing single-element arrays to `Promise.all()`, `Promise.any()`, or `Promise.race()`.

Promise combinators coordinate multiple inputs. Passing just one value can hide a missing promise or introduce unnecessary array handling.

Examples of **incorrect** code for this rule:

```javascript
const foo = await Promise.all([promise]);
const bar = await Promise.any([promise]);
const baz = await Promise.race([promise]);
const pending = Promise.all([nonPromise]);
```

Examples of **correct** code for this rule:

```javascript
const foo = await promise;
const pending = Promise.resolve(nonPromise);
const results = await Promise.all(promises);
const first = await Promise.any([promise, anotherPromise]);
const [{ value, reason }] = await Promise.allSettled([promise]);
```

## Fixes and Suggestions

Awaited `Promise.any()` and `Promise.race()` calls can be replaced with the single value. Calls without `await` offer suggestions to use the value directly or switch to `Promise.resolve()`.

`Promise.all()` produces an array, so it is only automatically unwrapped when the awaited result is discarded, destructured into one identifier, or accessed at index `0` in a variable initializer or assignment:

```javascript
const [value] = await Promise.all([promise]);
// Fixed to:
const value = await promise;
```

Comments that would be lost prevent unwrapping. When available, the `Promise.resolve()` suggestion preserves comments. Other `Promise.all()` calls are reported without a fix or suggestions.

The rule checks direct, non-optional calls on the identifier `Promise`. Computed methods, such as `Promise['race']()`, and `globalThis.Promise` are not checked. `Promise.allSettled()` is excluded because its result contains settlement details.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-single-promise-in-promise-methods](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-single-promise-in-promise-methods.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-single-promise-in-promise-methods.js)
