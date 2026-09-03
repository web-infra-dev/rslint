# no-array-front-mutation

## Rule Details

Disallow mutating the front of an array with `Array#shift()` or
`Array#unshift()`. Re-indexing the remaining elements can make these methods
unexpectedly expensive for large arrays.

Examples of **incorrect** code for this rule:

```javascript
const first = items.shift();
items.unshift(newItem);
```

Examples of **correct** code for this rule:

```javascript
for (const item of items) {
  consume(item);
}

queue.enqueue(newItem);
```

The rule allows `unshift()` on common Node.js stream objects, where it has
stream-specific semantics rather than array semantics:

```javascript
stream.unshift(chunk);
process.stdin.unshift(chunk);
```

## Original Documentation

- [eslint-plugin-unicorn: no-array-front-mutation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-array-front-mutation.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-array-front-mutation.js)
