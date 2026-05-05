# import/order

Enforces a convention in module import order. Imports are sorted into groups
(`builtin`, `external`, `parent`, `sibling`, `index`, plus optional
`internal`, `unknown`, `object`, `type`) and each group must appear in the
configured order.

This rule is autofixable.

## Rule Details

By default, imports are placed in this order:

```
builtin → external → parent → sibling → index
```

Examples of **incorrect** code:

```javascript
var sibling = require('./foo');
var fs = require('fs');
```

Examples of **correct** code:

```javascript
var fs = require('fs');
var sibling = require('./foo');
```

## Options

### `groups`

**Default:** `["builtin", "external", "parent", "sibling", "index"]`

Defines the relative order of import groups. Items can be a single string or
an array (entries in the same array share a rank — they're "interchangeable"
within that group).

```json
{
  "import/order": [
    "error",
    {
      "groups": ["builtin", ["external", "internal"], "parent", "sibling", "index"]
    }
  ]
}
```

### `pathGroups`

**Default:** `[]`

Refines the group ordering by matching specifiers against minimatch patterns.
Each entry has a `pattern`, a target `group`, and an optional `position`
(`"before"` or `"after"`).

```json
{
  "import/order": [
    "error",
    {
      "pathGroups": [
        { "pattern": "@app/**", "group": "external", "position": "after" }
      ]
    }
  ]
}
```

### `pathGroupsExcludedImportTypes`

**Default:** `["builtin", "external", "object"]`

Lists import types that are NOT subject to `pathGroups` matching. If you want
`@scope/*` imports to be re-ranked by a `pathGroup`, remove `"external"` from
this list.

### `distinctGroup`

**Default:** `true`

When `true`, `pathGroups` with a `position` form their own sub-group
(separated by an enforced newline when `newlines-between` is `"always"`).
When `false`, they slot back into the parent group.

### `newlines-between`

**Default:** `"ignore"`

Controls newlines between import groups:

- `"ignore"` — no enforcement
- `"always"` — exactly one empty line between different groups, none within
- `"never"` — no empty lines between any imports
- `"always-and-inside-groups"` — one empty line between groups, allowed within

```json
{ "import/order": ["error", { "newlines-between": "always" }] }
```

```javascript
import fs from 'fs';

import sibling from './foo';
```

### `newlines-between-types`

Identical to `newlines-between` but only applies to type-only imports when
`sortTypesGroup` is `true`. Defaults to the value of `newlines-between`.

### `alphabetize`

**Default:** `{ "order": "ignore", "orderImportKind": "ignore", "caseInsensitive": false }`

Sorts imports alphabetically within each group.

- `order`: `"asc"` | `"desc"` | `"ignore"`
- `orderImportKind`: `"asc"` | `"desc"` | `"ignore"` — secondary sort key
  used when two imports compare equal on path; sorts by kind (`type` vs
  `value`).
- `caseInsensitive`: when `true`, lowercases values before comparison.

```json
{
  "import/order": [
    "error",
    { "alphabetize": { "order": "asc", "caseInsensitive": true } }
  ]
}
```

### `named`

**Default:** `false`

Enforces alphabetical ordering of named import / export specifiers.

Forms accepted:

- `false` — disabled.
- `true` — enable for named imports, exports, requires, and CJS exports.
- Object form:
  - `enabled`: default for the four sub-toggles below.
  - `import`: check `import { ... } from 'mod'`.
  - `export`: check `export { ... } from 'mod'`.
  - `require`: check `var { ... } = require('mod')`.
  - `cjsExports`: check `module.exports = { ... }`. Requires type-info to
    distinguish the global `module` / `exports` from a user binding.
  - `types`: `"mixed"` | `"types-first"` | `"types-last"`. Controls how
    `import { type T, a, b }` interleaves type and value specifiers.

### `sortTypesGroup`

**Default:** `false`

When `true` and `"type"` is in `groups`, type-only imports form a parallel
sub-group hierarchy mirroring the value-import group order.

### `warnOnUnassignedImports`

**Default:** `false`

By default, side-effect imports (`import './styles.css'`) are ignored. Set
this to `true` to treat them like other imports for ordering. Side-effect
imports are never autofixed because their evaluation order can be
load-bearing.

### `consolidateIslands`

**Default:** `"never"`

When `"inside-groups"`, multi-line imports are separated from neighboring
imports with empty lines, while consecutive single-line imports stay
together. Only meaningful with `"always-and-inside-groups"` newline modes.

## Settings

| Setting | Behaviour |
| --- | --- |
| `import/internal-regex` | Specifier matching this regex classifies as `internal`. |
| `import/core-modules` | Extra names treated as `builtin`. |
| `import/external-module-folders` | Resolved paths under any of these folders classify as `external` (default `["node_modules"]`). |

## Differences from ESLint

These are **observable** differences in input → output behaviour. Mechanism
notes live in the source, not here.

- **Module classification uses TypeScript's resolver.** A specifier is
  classified `external` when the TypeScript compiler resolves it to a
  package under any directory listed in
  `settings["import/external-module-folders"]` (default `["node_modules"]`)
  or marks it as a library import. In monorepo layouts where ESLint's
  `eslint-import-resolver-*` walks package boundaries differently, a small
  number of imports may classify as `internal` here while ESLint says
  `external`, or vice versa. Workaround: spell out the boundary with
  `import/internal-regex` or override `import/external-module-folders`.
- **Custom resolvers are not consulted.** ESLint's
  `settings["import/resolver"]` (e.g. `eslint-import-resolver-webpack`,
  `eslint-import-resolver-typescript` configured with non-default options)
  has no effect. Resolution is whatever the TypeScript program already does
  for the file — tsconfig `paths`, `baseUrl`, and conditional exports are
  honoured.
- **TypeScript wrappers around `require()` are recognised.** Forms like
  `(require('fs') as any)`, `require('fs')!`, `require('fs') satisfies T`,
  and combinations of the above are still treated as static requires for
  ordering purposes. ESLint's parser can't represent these, so they are
  rslint-only; the practical effect is that mixed-language codebases get
  consistent ordering across `.ts` files where these wrappers are common.
- **`module.exports = { ... }` ordering** (`named.cjsExports`) only fires
  when type-info is available. On a `.js` file outside any
  `tsconfig.json`, the toggle is silently inert — the rule cannot tell a
  user-declared `module` binding from the global one without the
  type-checker. Add the file to a tsconfig if you want the check.

## Original Documentation

- [`eslint-plugin-import` — `order`](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/order.md)
