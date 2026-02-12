# non-nullable-type-assertion-style

## Rule Details

Enforce non-null assertions over explicit type assertions when the asserted type is the same as the original type minus `null` and `undefined`. A non-null assertion (`!`) is a more concise way to remove `null` and `undefined` from a type than using an `as` type assertion.

Examples of **incorrect** code for this rule:

```typescript
const foo = bar as string; // when bar is string | null
const baz = bar as string; // when bar is string | undefined
const qux = bar as string; // when bar is string | null | undefined
```

Examples of **correct** code for this rule:

```typescript
const foo = bar!;
const baz = bar as string | null;
const qux = bar as SomeDifferentType;
```

## Original Documentation

- [typescript-eslint non-nullable-type-assertion-style](https://typescript-eslint.io/rules/non-nullable-type-assertion-style)
