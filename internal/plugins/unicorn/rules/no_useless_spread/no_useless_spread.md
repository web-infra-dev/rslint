# no-useless-spread

## Rule Details

Disallow unnecessary spread when an array, object, or iterable can be used directly.

This rule reports spreading array literals into arrays or argument lists,
spreading object literals into objects, and cloning arrays returned by known
array-producing operations. It also reports array conversions passed to
`Map`, `WeakMap`, `Set`, `WeakSet`, `Array.from`, `TypedArray.from`,
`Promise.all`, `Promise.allSettled`, `Promise.any`, `Promise.race`, or
`Object.fromEntries`, and conversions used by `for…of` or `yield*`.

Examples of **incorrect** code:

```javascript
const values = [first, ...[second], third];
const options = { first, ...{ second } };
consume(...[first, second]);
const set = new Set([...values]);
const keys = [...Object.keys(options)];
for (const value of [...values]) consume(value);
```

Examples of **correct** code:

```javascript
const values = [first, second, third];
const options = { first, second };
consume(first, second);
const set = new Set(values);
const keys = Object.keys(options);
for (const value of values) consume(value);
```

Most reports can be fixed automatically. An `Object.assign` source containing
only spread properties offers an editor suggestion instead:

```javascript
Object.assign(target, { ...source });
// Suggested replacement:
Object.assign(target, source);
```

Applying this suggestion can change when getters and setters run. Comments
inside the source object prevent the suggestion from being offered.

Spreading the only argument to a collection constructor, such as
`new Set(...values)`, is reported without a fix. Pass the intended iterable
directly. Cloning `new Array(...)` or calling `.slice()` on a receiver whose
array type cannot be established also produces a report without a fix.

TypedArray constructors are excluded: `new Uint8Array([...'ab'])` and
`new Uint8Array('ab')` have different lengths. This rule assumes dense arrays;
disable it where sparse arrays are intentional, since spread converts holes
to `undefined` while some array operations preserve them.

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-spread](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-useless-spread.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-useless-spread.js)
