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

`patternOptions` accepts minimatch's string-matching controls, including
`dot`, `matchBase`, `nobrace`, `nocase`, `nocomment`, `noext`, `noglobstar`,
`nonegate`, `partial`, `flipNegate`, and `allowWindowsEscape`.

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
- `"always"` — at least one empty line between different groups, none within
- `"never"` — no empty lines between any imports
- `"always-and-inside-groups"` — at least one empty line between groups, allowed within

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
  - `cjsExports`: check `module.exports = { ... }` and named CommonJS export
    assignments while excluding locally shadowed `module` / `exports` names.
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
  `settings["import/external-module-folders"]` (default `["node_modules"]`).
  In monorepo layouts where ESLint's
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
- **Flow `import typeof` is not parsed.** The TypeScript parser rejects this
  Flow-only syntax before rules run. Ordinary Flow-compatible JavaScript
  covered by the upstream suite behaves normally.
- **CommonJS shadowing follows the full lexical chain.** A `module` or
  `exports` binding in any enclosing scope suppresses CommonJS export sorting.
  Upstream only checks variables declared by the immediate ESLint scope, which
  can mistake an ancestor binding for the CommonJS global.
- **String-named specifiers sort by their literal value.** Modern JavaScript
  permits forms such as `import { "name" as local } from "pkg"`. Upstream reads
  the Identifier-only `name` field and therefore does not order these
  consistently; rslint uses the shared static-property-name helper.
- **Reverse autofixes are EOF-safe.** When moving the first import after a
  later import in a file without a final newline, rslint inserts the missing
  line break. This avoids concatenating the two statements, which the upstream
  fixer can do in that edge case.
- **Autofixes preserve every adjacent same-line comment.** Upstream inspects at
  most 100 tokens or comments on either side of a statement when choosing its
  movable range. Rslint uses the file's shared comment index without that cap,
  so only lines with more than 100 adjacent comments differ.

## Original Documentation

- [`eslint-plugin-import` — `order`](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/order.md)
