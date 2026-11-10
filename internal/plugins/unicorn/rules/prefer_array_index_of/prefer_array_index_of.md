# prefer-array-index-of

## Rule details

Use `indexOf()` or `lastIndexOf()` to find the position of a value. This rule
reports `findIndex()` and `findLastIndex()` callbacks that only compare their
single parameter to a value using `===`.

The replacement evaluates the search value once, even for an empty array.
Search values with explicit side effects, such as function calls and assignments,
receive an editor suggestion. If replacing the callback would remove a comment,
the rule reports the comparison without offering an edit.

When type information is available, the rule skips searches whose value type
is incompatible with the parameter of `indexOf()` or `lastIndexOf()`.

Examples of **incorrect** code for this rule

```javascript
const index = values.findIndex(value => value === 'ready');
const lastIndex = values.findLastIndex(value => 'ready' === value);
```

Examples of **correct** code for this rule

```javascript
const index = values.indexOf('ready');
const lastIndex = values.lastIndexOf('ready');
values.findIndex((value, index) => value === index);
values.findLastIndex(value => value === 'ready' && isValid());
```

## Options

This rule has no options.

## Differences from upstream

Rslint skips sparse array literals. For example,
`[, 1].findIndex(value => value === undefined)` returns `0` because the callback
visits the hole. `[, 1].indexOf(undefined)` skips the hole and returns `-1`.

Rslint keeps the callback when moving the search value would take a TypeScript
type parameter out of scope or change a reference to `new.target`, `this`, or
`arguments`. In this example, `T` is available only inside the callback.

```typescript
declare const values: unknown[], target: unknown;
values.findIndex(<T>(value: T) => value === (target as T));
```

Rslint reports without an edit when a JSDoc type assertion applies to the whole
callback. The replacement would apply the function's type to the search value.

A search value such as `` makeValue`ready` `` receives an editor suggestion.
The tag function can run a different number of times after the replacement.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-array-index-of.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-array-index-of.js)
