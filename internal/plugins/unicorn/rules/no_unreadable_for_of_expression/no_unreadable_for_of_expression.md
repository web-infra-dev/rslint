# no-unreadable-for-of-expression

## Rule details

Keep `for…of` and `for await…of` loop headers easy to read by assigning
complex iterable expressions to a variable before the loop.

The rule allows identifiers, literals, arrays, dot-property chains, and calls
with simple arguments, such as `Object.values(object)` and
`Iterator.zip([names, scores])`. It reports lazy iterator helpers such as
`Iterator.from(items).map(transform)`. The `no-duplicate-loops` rule handles
ordinary `.map()` and `.filter()` calls directly in loop headers.

Examples of **incorrect** code for this rule

```javascript
for (const item of getItems(createArgument(seed))) {
  console.log(item);
}

for (const key of Object.keys(Object.fromEntries(entries))) {
  console.log(key);
}
```

Examples of **correct** code for this rule

```javascript
const argument = createArgument(seed);
for (const item of getItems(argument)) {
  console.log(item);
}

const object = Object.fromEntries(entries);
for (const key of Object.keys(object)) {
  console.log(key);
}

for (const item of items) {
  console.log(item);
}

for (const value of Object.values(object)) {
  console.log(value);
}
```

## Options

This rule has no options. It reports diagnostics without fixes or suggestions.

## Differences from upstream

In a TypeScript project, Rslint reports a loop header such as
`for (const row of factories.rows().map(format))` when `factories.rows()`
returns `IteratorObject<Row>`. Unicorn v79.0.0 with syntax-only TypeScript
linting leaves that loop header unreported.

## Original documentation

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unreadable-for-of-expression.md)
- [Upstream rule source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-unreadable-for-of-expression.js)
