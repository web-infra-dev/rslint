# no-duplicate-imports

## Rule Details

Disallow duplicate module imports. Multiple `import` statements (and optionally
`export ... from` re-exports) referencing the same module can usually be merged
into a single statement, which makes dependencies easier to scan and avoids
redundant module entries.

Examples of **incorrect** code for this rule:

```javascript
import { merge } from "lodash-es";
import { find } from "lodash-es";
```

```javascript
import { merge } from "lodash-es";
import _ from "lodash-es";
```

Examples of **correct** code for this rule:

```javascript
import { merge, find } from "lodash-es";
```

```javascript
import _, { merge, find } from "lodash-es";
```

A namespace import together with a named import is allowed because the two
forms cannot be merged into a single statement:

```javascript
import * as bar from "os";
import { baz } from "os";
```

## Options

This rule accepts an options object with the following properties:

- `includeExports` (default: `false`) — when `true`, also flag `export { ... } from`,
  `export * from`, and `export * as ns from` re-exports that duplicate (or could be
  merged with) an earlier import or export of the same module.
- `allowSeparateTypeImports` (default: `false`) — when `true`, a declaration-level
  `import type` / `export type` does NOT collide with a non-type import/export of
  the same module, so `import { foo } from "m"` and `import type { Bar } from "m"`
  are allowed. Specifier-level `type` keywords (`import { type Foo } from "m"`)
  do NOT count for this exemption — they only apply when the entire declaration
  is type-only.

### `includeExports`

Examples of **incorrect** code with `{ "includeExports": true }`:

```json
{ "no-duplicate-imports": ["error", { "includeExports": true }] }
```

```javascript
import os from "os";
export { something } from "os";
```

```json
{ "no-duplicate-imports": ["error", { "includeExports": true }] }
```

```javascript
export * from "os";
export * from "os";
```

Examples of **correct** code with `{ "includeExports": true }`:

```json
{ "no-duplicate-imports": ["error", { "includeExports": true }] }
```

```javascript
import os from "os";
export * from "os";
```

A `import * as` plus an `export { x } from` of the same module is allowed
because the two forms cannot be merged into a single statement:

```json
{ "no-duplicate-imports": ["error", { "includeExports": true }] }
```

```javascript
import * as os from "os";
export { something } from "os";
```

### `allowSeparateTypeImports`

Examples of **correct** code with `{ "allowSeparateTypeImports": true }`:

```json
{ "no-duplicate-imports": ["error", { "allowSeparateTypeImports": true }] }
```

```javascript
import { foo } from "module";
import type { Bar } from "module";
```

Examples of **incorrect** code with `{ "allowSeparateTypeImports": true }`:

```json
{ "no-duplicate-imports": ["error", { "allowSeparateTypeImports": true }] }
```

```javascript
import { type Foo } from "module";
import { type Bar } from "module";
```

```json
{ "no-duplicate-imports": ["error", { "allowSeparateTypeImports": true }] }
```

```javascript
import type { Merge } from "lodash-es";
import type { Find } from "lodash-es";
```

## Original Documentation

- ESLint rule: [no-duplicate-imports](https://eslint.org/docs/latest/rules/no-duplicate-imports)
