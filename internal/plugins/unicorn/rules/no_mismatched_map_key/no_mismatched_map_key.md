# no-mismatched-map-key

## Rule details

Use the same key when checking a map with `has()` and accessing it with
`get()` or `set()` in the guarded branch. A different key can indicate a
copy-paste error.

The rule checks single-statement `if` and `while` branches, conditional
expressions, and short-circuit expressions with the check on the left.
Reassigning the checked receiver or the index that selects it suppresses
reports in that branch.

Examples of **incorrect** code for this rule

```javascript
const value = map.has(key) ? map.get(anotherKey) : fallback;

if (!map.has(key)) {
  map.set(anotherKey, value);
}
```

Examples of **correct** code for this rule

```javascript
const value = map.has(key) ? map.get(key) : fallback;

if (!map.has(key)) {
  map.set(key, value);
}
```

## Options

This rule has no options or automatic fixes.

## Differences from upstream

Rslint can use generic type aliases and JSDoc annotations such as inline
`@type` casts and array types to recognize additional non-Map receivers.
Rslint skips the cache accesses below. ESLint Unicorn v79.0.0 reports them
when linting without TypeScript project information.

```typescript
type Cache = { has(key: string): boolean; get(key: string): number };
type Identity<T> = T;

function read(cache: Identity<Cache>) {
  if (cache.has('a')) cache.get('b');
}
```

```javascript
const cache = /** @type {{has: (key: string) => boolean, get: (key: string) => number}} */ (
  loadCache()
);
if (cache.has('a')) cache.get('b');
```

Rslint also recognizes `const` aliases of the same object key. For a receiver
such as `cache[index]`, Rslint checks that the same binding selects the map and
skips accesses after index reassignments. Destructuring assignments such as
`[holder.map] = maps` can also invalidate the check. ESLint Unicorn v79.0.0
can report these accesses.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-mismatched-map-key.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-mismatched-map-key.js)
