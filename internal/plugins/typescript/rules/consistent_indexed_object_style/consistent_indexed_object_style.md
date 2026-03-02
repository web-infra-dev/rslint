# consistent-indexed-object-style

## Rule Details

Require or disallow the `Record` type. TypeScript supports defining object types with an index signature or using the built-in `Record` utility type. This rule enforces a consistent style.

The rule supports two modes: `"record"` (default) prefers `Record<string, T>` over index signatures, and `"index-signature"` prefers index signatures over `Record`.

Examples of **incorrect** code for this rule (with default `"record"` option):

```typescript
interface Foo {
  [key: string]: unknown;
}

type Bar = {
  [key: number]: string;
};
```

Examples of **correct** code for this rule (with default `"record"` option):

```typescript
type Foo = Record<string, unknown>;
type Bar = Record<number, string>;

interface Baz {
  [key: string]: unknown;
  name: string; // has other members, so it's fine
}
```

## Original Documentation

- [typescript-eslint consistent-indexed-object-style](https://typescript-eslint.io/rules/consistent-indexed-object-style)
