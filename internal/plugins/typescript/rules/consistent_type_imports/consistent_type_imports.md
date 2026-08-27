# consistent-type-imports

## Rule Details

Enforce consistent use of type-only imports. TypeScript erases imports marked
with `type`, which makes their runtime behavior explicit and works with module
settings that preserve value imports.

Examples of **incorrect** code for this rule with the default options:

```typescript
import { Model, createModel } from './models';

export function create(): Model {
  return createModel();
}

type External = import('./external').External;
```

Examples of **correct** code for this rule with the default options:

```typescript
import type { Model } from './models';
import { createModel } from './models';

export function create(): Model {
  return createModel();
}
```

### `prefer`

The default, `"type-imports"`, requires imports used only in type positions to
be marked as type-only. `"no-type-imports"` instead forbids both top-level and
inline `type` modifiers.

```json
{
  "consistent-type-imports": ["error", { "prefer": "no-type-imports" }]
}
```

```typescript
import { Model } from './models';

type LocalModel = Model;
```

### `fixStyle`

When `prefer` is `"type-imports"`, `"separate-type-imports"` (the default)
moves type-only names into a separate declaration. `"inline-type-imports"`
keeps named type imports in the value declaration when possible.

```json
{
  "consistent-type-imports": [
    "error",
    { "fixStyle": "inline-type-imports" }
  ]
}
```

```typescript
import { type Model, createModel } from './models';
```

### `disallowTypeAnnotations`

`disallowTypeAnnotations` defaults to `true` and reports `import()` type
annotations. Set it to `false` to allow them.

```json
{
  "consistent-type-imports": [
    "error",
    { "disallowTypeAnnotations": false }
  ]
}
```

```typescript
type Model = import('./models').Model;
```

## Differences from ESLint

The pinned upstream implementation stores module sources in a plain JavaScript
object, so imports from `"constructor"`, `"toString"`, `"__proto__"`, or
`"hasOwnProperty"` throw while linting. The Go map intentionally handles these
names normally and still reports and fixes the import.

With `"inline-type-imports"`, rslint also emits a valid inline fix when an
earlier default-only or default-plus-namespace value import from the same module
causes the pinned upstream implementation to suppress its fix.

## Original Documentation

- [typescript-eslint: consistent-type-imports](https://typescript-eslint.io/rules/consistent-type-imports)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/consistent-type-imports.ts)
