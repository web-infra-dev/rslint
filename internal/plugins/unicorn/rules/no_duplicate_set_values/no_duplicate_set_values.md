# no-duplicate-set-values

## Rule Details

Disallow duplicate values in `Set` constructor array literals.

`Set` values are unique, so repeated static values in a `new Set([...])` array literal are redundant. Values are compared with SameValueZero, the comparison `Set` uses: `NaN` matches `NaN`, `0` matches `-0`, and an array hole matches `undefined`. Constant bindings are resolved, while values that could change through getters or mutation are not treated as static.

The rule also reports repeated reference expressions, like `foo.bar` and `foo["bar"]`. It does not model mutations or side effects between elements.

Examples of **incorrect** code:

```javascript
const set = new Set(['foo', 'bar', 'foo']);

const foo = 2;
const numbers = new Set([foo, 2]);

const references = new Set([foo.bar, foo.bar]);
```

Examples of **correct** code:

```javascript
const set = new Set(['foo', 'bar']);

const foo = 2;
const numbers = new Set([foo, 3]);

const references = new Set([foo.bar, foo.baz]);
```

This rule only checks `new Set([...])` with a single array literal argument, including one wrapped in TypeScript `as`, `satisfies`, `!` or `<Type>` assertions. It does not check arrays, maps, non-array iterables, or chained `Set#add()` calls.

## Differences from upstream

Regular expression literals wrapped in TypeScript assertions are not reported. Each literal evaluates to a new object, so the `Set` keeps both values. Upstream reports the second one:

```typescript
// Not reported by rslint; reported by upstream.
const patterns = new Set([/a/ as RegExp, /a/ as RegExp]);
```

A repeated reference to the same regular expression is still reported:

```javascript
const pattern = /a/;
const patterns = new Set([pattern, pattern]);
```

## Original Documentation

- [eslint-plugin-unicorn: no-duplicate-set-values](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/docs/rules/no-duplicate-set-values.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-set-values.js)
