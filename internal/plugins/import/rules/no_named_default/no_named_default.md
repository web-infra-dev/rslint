# no-named-default

## Rule Details

Require default import syntax when importing a default export. Named imports
such as `import { default as foo } from './foo.js'` are reported on the local
binding `foo`. Quoted `'default'` imports are also reported.

Examples of **incorrect** code for this rule:

```javascript
import { default as foo } from './foo.js';
import { default as foo, bar } from './foo.js';
import { 'default' as foo } from './foo.js';
```

```typescript
import type { default as Foo } from './foo.js';
```

Examples of **correct** code for this rule:

```javascript
import foo from './foo.js';
import foo, { bar } from './foo.js';
```

```typescript
import { type default as Foo } from './foo.js';
import type Foo from './foo.js';
```

Inline type imports are ignored. Following upstream behavior, a declaration-level
`import type { default as Foo }` is still reported; use `import type Foo` instead.

## Options

This rule has no options and provides no automatic fixes or suggestions.

## Differences from upstream

Compared with eslint-plugin-import v2.32.0:

- **Flow `typeof` imports are unsupported.** JavaScript and TypeScript imports
  are supported, but Flow code such as
  `import { typeof default as Foo } from './foo.js'` cannot be checked with
  rslint. Upstream accepts and ignores this syntax. Use ESLint for files that
  contain these Flow imports.

## Original Documentation

- [eslint-plugin-import: no-named-default](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-default.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-default.js)
