# no-import-type-side-effects

## Rule Details

Enforce the use of a top-level `import type` qualifier when an import only has specifiers with inline `type` qualifiers. Under TypeScript's `--verbatimModuleSyntax`, inline `type` specifiers are stripped one by one, which leaves behind an empty `import {} from 'mod'` — a runtime side-effect import. Hoisting the qualifier to the top level removes the whole statement instead. The rule is auto-fixable.

Examples of **incorrect** code for this rule:

```typescript
import { type A } from 'mod';
import { type A as AA } from 'mod';
import { type A, type B } from 'mod';
import { type A as AA, type B as BB } from 'mod';
```

Examples of **correct** code for this rule:

```typescript
import T from 'mod';
import * as T from 'mod';
import { T } from 'mod';
import type { T } from 'mod';
import type { T, U } from 'mod';
import { type T, U } from 'mod';
import { T, type U } from 'mod';
import type T from 'mod';
import T, { type U } from 'mod';
import type * as T from 'mod';
import 'mod';
```

## Original Documentation

- [typescript-eslint no-import-type-side-effects](https://typescript-eslint.io/rules/no-import-type-side-effects)
