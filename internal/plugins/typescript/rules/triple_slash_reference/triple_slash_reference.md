# triple-slash-reference

## Rule Details

Disallow certain triple slash directives in favor of ES6-style import declarations. TypeScript's `/// <reference ... />` triple-slash directives are an older mechanism for including type information. In most modern codebases, ES6-style `import` statements are preferred. This rule can ban `/// <reference path="..." />`, `/// <reference types="..." />`, and `/// <reference lib="..." />` directives.

By default, `types` references are banned when the file already uses `import` statements (the `prefer-import` setting).

Examples of **incorrect** code for this rule:

```typescript
/// <reference types="jest" />
import { foo } from 'bar';
```

Examples of **correct** code for this rule:

```typescript
import { foo } from 'bar';

/// <reference types="jest" />
// (OK when there are no import statements in the file)
```

## Original Documentation

- [typescript-eslint triple-slash-reference](https://typescript-eslint.io/rules/triple-slash-reference)
