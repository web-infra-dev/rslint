# no-non-null-assertion

## Rule Details

Disallows non-null assertions using the `!` postfix operator.

TypeScript's `!` non-null assertion operator asserts to the type system that an expression is non-nullable. Using assertions to tell the type system new information is often a sign that code is not fully type-safe. It's generally better to structure program logic so that TypeScript understands when values may be nullable.

Examples of **incorrect** code for this rule:

```typescript
interface Example {
  property?: string;
}
declare const example: Example;
const includesBaz = example.property!.includes('baz');
```

Examples of **correct** code for this rule:

```typescript
interface Example {
  property?: string;
}
declare const example: Example;
const includesBaz = example.property?.includes('baz') ?? false;
```

## Original Documentation

https://typescript-eslint.io/rules/no-non-null-assertion
