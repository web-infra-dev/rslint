# array-type

## Rule Details

Require consistently using either `T[]` or `Array<T>` for arrays. TypeScript provides two equivalent ways to define array types. This rule enforces a consistent style across the codebase.

The rule supports three modes via the `default` option: `"array"` (prefer `T[]`), `"generic"` (prefer `Array<T>`), and `"array-simple"` (prefer `T[]` for simple types, `Array<T>` for complex types). A separate `readonly` option controls readonly array syntax.

Examples of **incorrect** code for this rule (with default `"array"` option):

```typescript
const a: Array<string> = [];
const b: ReadonlyArray<number> = [1, 2];
const c: Array<string | number> = [];
```

Examples of **correct** code for this rule (with default `"array"` option):

```typescript
const a: string[] = [];
const b: readonly number[] = [1, 2];
const c: (string | number)[] = [];
```

## Original Documentation

- [typescript-eslint array-type](https://typescript-eslint.io/rules/array-type)
