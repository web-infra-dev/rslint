# import/default

## Rule Details

If a default import is requested, this rule reports when the imported module has no default export.

```javascript
// ./named-exports.ts only has named exports
import foo from "./named-exports";  // ✘ No default export found...
```

```javascript
// ./default-export.ts has a default export
import foo from "./default-export";  // ✓
import bar from "./default-export";  // ✓ renaming is fine
import bar, { baz } from "./default-export";  // ✓ default + named
```

The rule does not apply to imports without a default specifier:

```javascript
import { foo } from "./named-exports";  // ✓
import * as ns from "./named-exports";  // ✓
import "./named-exports";              // ✓
```

`export { default as X } from "./mod"` is a normal named-export form and is not flagged by this rule.

## Configuration

This rule has no per-rule options.

It honours `settings["import/ignore"]` — an array of regex strings tested against each resolved file path. A match marks the module as un-analyzable and skips the check. When unset, the rule skips imports that resolve to `.coffee`, `.eslintrc` (and friends), `.es6`, `.exs`, `.json`, or `.node` files.

## Differences from ESLint

### Imports inside `node_modules`

```javascript
import _ from "lodash";       // ✓ skipped
import crypto from "crypto";  // ✓ skipped
```

rslint silently skips any import resolved into `node_modules`. ESLint skips only Node built-ins. Set `settings["import/ignore"]` to any value (including `[]`) to take over — once set, `node_modules` paths are no longer auto-skipped.

### `import X = require("./mod")`

```typescript
import x = require("./named-exports");  // ✓ never flagged
```

This TypeScript syntax is never reported, even when the target has no default export. Use `import x from "./mod"` if you want the check.

### Babel-only re-export forms

```javascript
export bar from "./mod";              // syntax error — won't parse
export bar, { foo } from "./mod";     // syntax error — won't parse
export bar, * as ns from "./mod";     // syntax error — won't parse
```

These forms don't parse under rslint, so they're never reported.

### Files with parse errors

When you `import foo from "./broken"` and `./broken` has parse errors:

- ESLint reports `Parse errors in imported module './broken': <error> (line:col)` on the import statement.
- rslint stays silent if the parser can recover a default; otherwise reports the regular `No default export found in imported module "./broken".` Parse errors themselves come through separately, not through this rule.

### `import/ignore` pattern matching

```javascript
// settings: { "import/ignore": ["named-exports"] }
import baz from "./default-files/named-exports";  // ✓ skipped
```

Patterns match the **resolved file path**, not the source string as written. `"./named-exports"` won't match — write `"named-exports"` instead.

### `export = X` requires `esModuleInterop` or `allowSyntheticDefaultImports`

```typescript
// some.ts: export = function () {}

// tsconfig has esModuleInterop: true
import fn from "./some";  // ✓

// tsconfig has neither
import fn from "./some";  // ✘ No default export found...
```

### Settings ignored by rslint

The following ESLint settings have no effect under rslint and are silently accepted:

- `settings["import/parsers"]`
- `settings["import/resolver"]`
- `settings["import/extensions"]`
- `settings["import/cache"]`

### Diagnostic message

Identical to ESLint: `No default export found in imported module "<path>".` (`messageId: "default"`).

### Case-sensitive paths

Whether `import Foo from "./named-exports"` (mismatched case) resolves depends on the host filesystem. Same as ESLint.

## Original Documentation

- [import/default](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/default.md)
