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

Awaited `Promise.race()` calls can be replaced with the single value. Calls without `await` offer suggestions to use the value directly or switch to `Promise.resolve()`. `Promise.any()` is reported without fixes or suggestions because replacing it changes how rejections are handled.

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

## Differences from upstream

- `Promise.any([promise])` keeps its diagnostic but has no fixes or suggestions. Unlike eslint-plugin-unicorn v77.0.0, rslint preserves the `AggregateError` and its `.errors` array when the input rejects. For example, `await Promise.any([Promise.reject('failure')])` must not become `await Promise.reject('failure')`.
- `Promise.all()` is not automatically unwrapped when JSDoc `@type` or `@satisfies` constrains the original promise or tuple result. This includes annotations on destructuring declarations and casts around the call, awaited result, or destructuring assignment. For example, `/** @satisfies {[number]} */ const [value] = await Promise.all([Promise.resolve(1)])` keeps the tuple result required by its annotation, avoiding a new type error under `checkJs`. Annotations on the scalar result of `(await Promise.all([promise]))[0]` still allow the fix.
- Direct-value suggestions keep parentheses around members whose receiver starts with an object, function, or class expression. For example, `Promise.race([{ value: promise }.value])` becomes `({ value: promise }.value)`, which remains valid as a statement or an arrow function body.

## Original Documentation

- [eslint-plugin-unicorn: no-single-promise-in-promise-methods](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-single-promise-in-promise-methods.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-single-promise-in-promise-methods.js)
