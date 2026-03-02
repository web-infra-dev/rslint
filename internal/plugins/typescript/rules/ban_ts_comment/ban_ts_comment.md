# ban-ts-comment

## Rule Details

Disallow `@ts-<directive>` comments or require descriptions after directives. TypeScript provides several directive comments (`@ts-expect-error`, `@ts-ignore`, `@ts-nocheck`, `@ts-check`) that alter how the compiler processes code. Overusing these directives can hide real errors and reduce type safety.

By default, `@ts-expect-error`, `@ts-ignore`, and `@ts-nocheck` are banned. Directives can optionally be allowed if accompanied by a description meeting a minimum length requirement.

Examples of **incorrect** code for this rule:

```typescript
// @ts-ignore
const x: number = 'hello';

// @ts-nocheck

/* @ts-ignore */
const y = undefined;
```

Examples of **correct** code for this rule:

```typescript
// @ts-expect-error: this is intentional for testing
const x: number = 'hello';

// @ts-check

// Regular comments that mention @ts-ignore in passing are fine
```

## Original Documentation

- [typescript-eslint ban-ts-comment](https://typescript-eslint.io/rules/ban-ts-comment)
