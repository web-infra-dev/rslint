# no-new-buffer

Enforce `Buffer.from()` and `Buffer.alloc()` instead of the deprecated `new Buffer()` constructor.

## Examples

```js
// Incorrect
const bytes = new Buffer([0x62, 0x75, 0x66]);
const empty = new Buffer(10);

// Correct
const bytes = Buffer.from([0x62, 0x75, 0x66]);
const empty = Buffer.alloc(10);
```

When the argument does not establish whether it is a size or source data, the rule reports a diagnostic with `Buffer.from()` and `Buffer.alloc()` suggestions.

## Further reading

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-new-buffer.md)
