# no-useless-empty-export

## Rule Details

Disallow empty exports (`export {}`) that don't affect the module's exports. An empty `export {}` is used to turn a script file into a module, but if the file already has other imports or exports, this empty export is unnecessary and can be removed.

This rule does not flag empty exports in `.d.ts` definition files, where they may be needed for module encapsulation.

Examples of **incorrect** code for this rule:

```typescript
export const value = 'Hello';
export {};

import { foo } from 'bar';
export {};
```

Examples of **correct** code for this rule:

```typescript
export const value = 'Hello';

export {};
// (when no other imports/exports exist, this is the only module indicator)
```

## Original Documentation

- [typescript-eslint no-useless-empty-export](https://typescript-eslint.io/rules/no-useless-empty-export)
