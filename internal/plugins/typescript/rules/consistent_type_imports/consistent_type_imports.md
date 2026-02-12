# consistent-type-imports

## Rule Details

Enforce consistent usage of type imports. TypeScript allows marking imports as type-only using `import type`, which ensures the import is erased at compile time. This rule enforces that type-only imports use the `import type` syntax, and can optionally disallow `import()` type annotations.

The `prefer` option supports `"type-imports"` (default) and `"no-type-imports"`. The `disallowTypeAnnotations` option (default `true`) forbids using `import()` in type annotations.

Examples of **incorrect** code for this rule:

```typescript
import { MyType } from './types'; // MyType is only used as a type

type Foo = import('./types').Bar; // disallowTypeAnnotations
```

Examples of **correct** code for this rule:

```typescript
import type { MyType } from './types';

import { value } from './values';

import { value, type MyType } from './mixed';
```

## Original Documentation

- [typescript-eslint consistent-type-imports](https://typescript-eslint.io/rules/consistent-type-imports)
