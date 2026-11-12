# no-duplicate-loops

## Rule Details

Disallow calling `.map()` or `.filter()` directly in a `for…of` or
`for await…of` loop header. On arrays, these methods create an intermediate
array that the loop then traverses. When it preserves the behavior, move the
mapping or filtering into the loop body to traverse the input once.

The rule ignores recognized lazy iterators, such as synchronous generators
and chains starting with `Iterator.from()`, `Iterator.concat()`,
`Iterator.zip()`, `Iterator.zipKeyed()`, `.entries()`, `.keys()`, `.values()`,
or `.matchAll()`.

### Incorrect

```javascript
for (const value of values.map(value => value.property)) {
  console.log(value);
}

for (const value of values.filter(value => value.isEnabled)) {
  console.log(value);
}
```

### Correct

```javascript
for (const value of values) {
  console.log(value.property);
}

for (const value of values) {
  if (value.isEnabled) {
    console.log(value);
  }
}
```

The rule offers no automatic edits. For example, moving an async mapper into
the loop body can change when promises start and whether they run concurrently.

## Options

This rule has no options.

## Differences from upstream

In TypeScript files, rslint can use resolved types to recognize more lazy
iterators. For example, it allows `for (const row of factories.rows().map(fn))`
when `factories.rows()` returns `IteratorObject<Row>`. ESLint needs type
information to recognize that return type.

## Original Documentation

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-duplicate-loops.md)
- [Upstream rule source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-duplicate-loops.js)
