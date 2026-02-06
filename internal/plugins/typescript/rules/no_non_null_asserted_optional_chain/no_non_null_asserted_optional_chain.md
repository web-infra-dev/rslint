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

https://typescript-eslint.io/rules/no-non-null-asserted-optional-chain
