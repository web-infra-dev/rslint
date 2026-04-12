# import/no-duplicates

## Rule Details

Reports if a resolved path is imported more than once.

This rule is similar to ESLint core's `no-duplicate-imports`, but differs in two key ways:

1. The paths in the source code don't have to exactly match — they just have to point to the same module on the filesystem (e.g., `./foo` and `./foo.js`).
2. This version distinguishes `type` imports from standard imports.

Examples of **incorrect** code for this rule:

```javascript
import { x } from './foo';
import { y } from './foo';
```

```javascript
import SomeDefaultClass from './mod';
import * as names from './mod';
import { something } from './mod.js';
```

Examples of **correct** code for this rule:

```javascript
import SomeDefaultClass, * as names from './mod';
import type SomeType from './mod';
```

```javascript
import { x } from './foo';
import { y } from './bar';
```

## Options

### `considerQueryString`

When set to `true`, imports with different query strings are treated as different modules.

```json
"import/no-duplicates": ["error", { "considerQueryString": true }]
```

### `prefer-inline`

When set to `true`, supports TypeScript inline type imports, allowing `import type { X }` to be merged into `import { type X }`.

```json
"import/no-duplicates": ["error", { "prefer-inline": true }]
```

## Original Documentation

- [eslint-plugin-import/no-duplicates](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-duplicates.md)
