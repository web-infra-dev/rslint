# no-duplicates

## Rule Details

Reports if a resolved path is imported more than once.

This rule is similar to ESLint core's `no-duplicate-imports`, but differs in two key ways:

1. The paths in the source code don't have to exactly match — they just have to point to the same module on the filesystem (e.g., `./foo` and `./foo.js`).
2. This version distinguishes `type` imports from standard imports.

Import attributes are part of a module request. Imports with different attributes are not duplicates, while equivalent attribute sets compare equally regardless of property order. Autofix only merges imports whose attributes are equivalent.

Explicit local paths that name existing files use those files as module identities. For example, when both `text.js` and `text.ts` exist, imports of `./text.js` and `./text.ts` are not duplicates, even if TypeScript resolves both to `text.ts`. The same applies to `.jsx`/`.tsx`, `.mjs`/`.mts`, and `.cjs`/`.cts` pairs.

Package lookup uses runtime targets rather than shared type declarations, retaining import/require conditions and the project's custom conditions. Entry points with different type declarations remain separate even when they share a runtime file. TypeScript path aliases take precedence when a substitution for the selected pattern resolves; missing substitutions allow package lookup to continue.

Queries and fragments selected by package exports remain part of ESM module identity, independently of `considerQueryString`. CommonJS lookup also resolves relative extensionless requests using Node's runtime extension and directory rules, so adjacent declarations do not hide genuine duplicates.

If the written file does not exist, the rule retains unambiguous TypeScript resolution for aliases and extension substitution. A declaration file alone does not establish runtime identity for a value import. Neither does a result with an existing JavaScript sibling, or one resolved from a different importer directory. In these cases only identical written requests are grouped.

```javascript
import data from './data.json' with { type: 'json' };
import source from './data.json' with { type: 'text' };
```

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

- **Default:**
  - `false` by default
  - `true` if using the `importPlugin.configs.recommended` preset

When set to `true`, imports with different query strings are treated as different modules.

```json
{
  "import/no-duplicates": ["error", { "considerQueryString": false }]
}
```

For example, these imports are reported as duplicates when `considerQueryString` is `false`, but are allowed when it is `true`:

```javascript
import iconUrl from './icon.svg?url';
import iconSource from './icon.svg?raw';
```

### `prefer-inline`

- **Default:** `false`

When set to `true`, supports TypeScript inline type imports, allowing `import type { X }` to be merged into `import { type X }`.

```json
{
  "import/no-duplicates": ["error", { "prefer-inline": true }]
}
```

## Differences from upstream

When a TypeScript alias resolves to a file with an existing JavaScript sibling, the rule conservatively keeps differently written requests separate. For example, if both `text.js` and `text.ts` exist and `@files/*` maps to `./*`, these imports remain separate even if the runtime alias selects the same JavaScript file:

```typescript
import { x } from '@files/text.js';
import { y } from './text.js';
```

## Original Documentation

- [eslint-plugin-import: no-duplicates](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-duplicates.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-duplicates.js)
