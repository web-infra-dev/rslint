# consistent-type-exports

## Rule Details

Enforce consistent usage of type exports. TypeScript allows marking exports as type-only using `export type`, which is erased at compile time and results in no runtime code. This rule enforces that type-only exports use the `export type` syntax.

This rule requires type information and supports automatic fixes. When all exports in a declaration are types, the fix uses `export type`. This also applies to `export *` and `export * as ns` from modules with no runtime exports. Values explicitly exported as types are allowed.

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

## Options

`fixMixedExportsWithInlineTypeSpecifier` is a boolean that defaults to `false`. By default, a mixed declaration is split into separate type and value exports:

```typescript
// Before
export { SomeType, value } from './mixed';

// After
export type { SomeType } from './mixed';
export { value } from './mixed';
```

With `{ "fixMixedExportsWithInlineTypeSpecifier": true }`, the fix adds inline `type` specifiers instead:

```typescript
export { type SomeType, value } from './mixed';
```

Existing inline type specifiers do not exempt other type exports from this rule. A declaration containing only types is converted to `export type` with either option value.

## Differences from ESLint

The diagnostics and options follow typescript-eslint v8.70.1. Two fixes intentionally preserve valid code where that version's default mixed-export fix does not:

- Import attributes remain on the value re-export. For example, splitting `export { T, value } from './mixed' with { type: 'json' };` preserves `from './mixed' with { type: 'json' }` on the value declaration. Upstream removes that source and its attributes.
- Module names that need escaping retain their original string literal. For example, `export { T, value } from "./it's-a-module";` produces a type re-export using the same quoted module name. Upstream inserts an unescaped single quote.

## Original Documentation

- [typescript-eslint: consistent-type-exports](https://typescript-eslint.io/rules/consistent-type-exports)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.70.1/packages/eslint-plugin/src/rules/consistent-type-exports.ts)
