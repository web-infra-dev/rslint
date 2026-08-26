# no-non-null-asserted-nullish-coalescing

## Rule Details

Disallow non-null assertions in the left operand of a nullish coalescing operator.

The `??` nullish coalescing operator is designed to provide a default value when dealing with `null` or `undefined`. Using a non-null assertion `!` in the left operand is contradictory and likely a mistake.

Examples of **incorrect** code for this rule:

```typescript
foo! ?? bar;
foo.bazz! ?? bar;
foo()! ?? bar;
```

Examples of **correct** code for this rule:

```typescript
foo ?? bar;
foo ?? bar!;
foo.bazz ?? bar;
```

## Original Documentation

- [typescript-eslint: no-non-null-asserted-nullish-coalescing](https://typescript-eslint.io/rules/no-non-null-asserted-nullish-coalescing)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-non-null-asserted-nullish-coalescing.ts)
