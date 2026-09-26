# no-useless-continue

Disallow an unlabeled `continue` statement when the current loop iteration would end at the same point without it. The rule fixes the statement automatically and preserves adjacent comments.

```javascript
// Incorrect
for (const item of items) {
  process(item);
  continue;
}

// Correct
for (const item of items) {
  process(item);
}
```

Labeled `continue` statements, bare loop bodies, and statements inside a `switch` or `try` are left alone.

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Differences from upstream

When removing a standalone statement from a CRLF file, the fix removes the whole preceding CRLF sequence. Unicorn v76.0.0 removes only the newline byte in this case, leaving an extra carriage return.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-continue](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/no-useless-continue.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/no-useless-continue.js)
