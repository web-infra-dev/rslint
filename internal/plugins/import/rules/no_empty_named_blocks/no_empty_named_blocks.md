# no-empty-named-blocks

## Rule Details

Disallow empty named import blocks, including type-only imports. Side-effect
imports and imports containing named bindings are allowed.

Examples of **incorrect** code:

```javascript
import {} from 'mod';
import Default, {} from 'mod';
```

```typescript
import type {} from 'mod';
```

Examples of **correct** code:

```javascript
import 'mod';
import Default from 'mod';
import { Named } from 'mod';
import Default, { Named } from 'mod';
```

```typescript
import type { Named } from 'mod';
```

When a default binding remains, the automatic fix removes the empty block,
including comments inside it, and its preceding comma. For example,
`import Default, {} from 'mod'` becomes `import Default from 'mod'`.

When there is no default binding, the rule offers two editor suggestions: remove
the entire import, or convert it to a side-effect import such as `import 'mod'`.
These suggestions also apply to type-only imports, except that type-only imports
with attributes only offer removal. Choose the side-effect form only when the
module should run at runtime.

## Options

This rule has no options.

## Differences from upstream

Compared with eslint-plugin-import v2.32.0:

- **Flow imports are unsupported.** This rule supports JavaScript and TypeScript
  imports, but not Flow forms such as `import typeof {} from 'mod'`.
- **Import attributes are preserved.** Empty attributes, such as
  `import Default from 'mod' with {}`, are allowed. An empty named import with
  attributes, such as `import {} from 'mod' with { mode: 'custom' }`, offers
  suggestions instead of an automatic fix. Upstream can remove the wrong braces
  or produce an invalid import.
- **Default imports named `type` or `from` are kept.** For example,
  `import type, {} from 'mod'` is automatically fixed to `import type from 'mod'`.
  Upstream can instead suggest removing the default import.
- **Suggestions preserve other imports.** For
  `import value from 'first'; import {} from 'second';`, converting the second
  import produces `import value from 'first'; import 'second';`. Upstream can
  duplicate an earlier import and leave the empty import unchanged.
- **Compact imports and following comments stay valid.**
  `import Default,{}from'mod'` becomes `import Default from'mod'`, and
  `import {} from/* keep */ 'mod'` becomes `import /* keep */ 'mod'`. Upstream can
  join the default name with `from` or break the comment.
- **Type-only imports with attributes are not converted to runtime imports.**
  For example, `import type {} from 'mod' with { 'resolution-mode': 'require' }`
  only offers the suggestion to remove the import. Upstream also offers a
  conversion that produces invalid TypeScript. If the module should run at
  runtime, write the intended import explicitly.

## Original Documentation

- [eslint-plugin-import: no-empty-named-blocks](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-empty-named-blocks.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-empty-named-blocks.js)
