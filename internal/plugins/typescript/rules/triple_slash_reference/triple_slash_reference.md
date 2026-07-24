# triple-slash-reference

## Rule Details

Disallow certain triple slash directives in favor of ES6-style import declarations. TypeScript's `/// <reference ... />` triple-slash directives are an older mechanism for including type information. In most modern codebases, ES6-style `import` statements are preferred. This rule can ban `/// <reference path="..." />`, `/// <reference types="..." />`, and `/// <reference lib="..." />` directives.

The default options are:

- `lib: "always"`: allow `lib` references.
- `path: "never"`: disallow `path` references.
- `types: "prefer-import"`: disallow a `types` reference only when the same module is also imported.

Examples of **incorrect** code for this rule:

```typescript
/// <reference types="jest" />
import 'jest';

/// <reference path="./types.d.ts" />
```

Examples of **correct** code for this rule:

```typescript
/// <reference types="jest" />
import { foo } from 'bar';
// The "jest" reference is allowed because only "bar" is imported.
```

## Original Documentation

- [typescript-eslint triple-slash-reference](https://typescript-eslint.io/rules/triple-slash-reference)
