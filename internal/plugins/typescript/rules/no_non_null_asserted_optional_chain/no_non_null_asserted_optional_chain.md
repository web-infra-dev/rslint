# no-non-null-asserted-optional-chain

## Rule Details

Disallow non-null assertions after an optional chain expression.

Optional chain expressions (`?.`) are designed to return `undefined` if the value is nullish. Using a non-null assertion (`!`) after an optional chain expression is unsafe, as it defeats the purpose of the optional chain.

Examples of **incorrect** code for this rule:

```typescript
foo?.bar!;
foo?.bar()!;
```

Examples of **correct** code for this rule:

```typescript
foo?.bar;
foo?.bar();
```

## Original Documentation

- [typescript-eslint: no-non-null-asserted-optional-chain](https://typescript-eslint.io/rules/no-non-null-asserted-optional-chain)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-non-null-asserted-optional-chain.ts)
