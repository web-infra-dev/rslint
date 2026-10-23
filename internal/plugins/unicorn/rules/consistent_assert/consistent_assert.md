# consistent-assert

## Rule Details

Enforce the explicit `node:assert` assertion style by preferring `assert.ok(value)` over calling an imported assert function directly.

Examples of **incorrect** code:

```javascript
import assert from 'node:assert/strict';

assert(value);
```

Examples of **correct** code:

```javascript
import assert from 'node:assert/strict';

assert.ok(value);
```

The rule recognizes default imports from `assert`, `node:assert`, `assert/strict`, and `node:assert/strict`. It also recognizes `default` named imports and `strict` imported from `assert` or `node:assert`.

Only direct calls to the imported binding are changed. Other references, namespace imports, type-only imports, and shadowed bindings are left alone.

## Original Documentation

- [eslint-plugin-unicorn: consistent-assert](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/consistent-assert.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/consistent-assert.js)
