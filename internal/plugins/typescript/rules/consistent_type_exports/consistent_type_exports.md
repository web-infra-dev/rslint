# consistent-type-exports

## Rule Details

Enforce consistent usage of type exports. TypeScript allows marking exports as type-only using `export type`, which is erased at compile time and results in no runtime code. This rule enforces that type-only exports use the `export type` syntax.

When all exports in a declaration are types, the entire declaration should use `export type`. When a declaration contains a mix of type and value exports, the rule can suggest using inline `type` specifiers.

Examples of **incorrect** code for this rule:

```typescript
interface Foo {}
type Bar = string;

export { Foo, Bar };

export { SomeType } from './types';
```

Examples of **correct** code for this rule:

```typescript
interface Foo {}
type Bar = string;

export type { Foo, Bar };

export type { SomeType } from './types';

export { value, type MyType } from './mixed';
```

## Original Documentation

- [typescript-eslint consistent-type-exports](https://typescript-eslint.io/rules/consistent-type-exports)
