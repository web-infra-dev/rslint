# consistent-type-definitions

## Rule Details

Enforce type definitions to consistently use either `interface` or `type`. TypeScript provides two ways to define object types: `interface` declarations and `type` alias declarations with object literal types. This rule enforces one style for consistency.

The rule supports two modes: `"interface"` (default) prefers interfaces over type literals, and `"type"` prefers type aliases over interfaces.

This rule is auto-fixable.

Examples of **incorrect** code for this rule (with default `"interface"` option):

```typescript
// type alias with object literal -> should be interface
type Foo = {
  name: string;
  age: number;
};

// including index signatures
type Bar = {
  [key: string]: number;
};

// including parenthesized types
type Baz = {
  x: number;
};
```

Examples of **correct** code for this rule (with default `"interface"` option):

```typescript
interface Foo {
  name: string;
  age: number;
}

// Type aliases for non-object types are always allowed
type ID = string | number;
type Callback = () => void;
type Union = { x: number } | { y: string };
type Intersection = { x: number } & { y: string };
type Mapped<T, U> = { [K in T]: U };
```

Examples of **incorrect** code for this rule (with `"type"` option):

```typescript
interface Foo {
  name: string;
}
```

Examples of **correct** code for this rule (with `"type"` option):

```typescript
type Foo = {
  name: string;
};
```

## Autofix

The rule provides automatic fixes:

- **`interface` mode**: Converts `type T = { ... }` to `interface T { ... }`, handling `export`, `declare`, type parameters, parenthesized types, and trailing semicolons.
- **`type` mode**: Converts `interface T { ... }` to `type T = { ... }`, converting `extends` clauses to intersection types (`& B & C`). Handles `export default interface` by splitting into a type declaration and a separate default export.

Note: Interfaces inside `declare global` blocks report an error but are not auto-fixed to avoid breaking global type augmentation patterns.

## Original Documentation

- [typescript-eslint consistent-type-definitions](https://typescript-eslint.io/rules/consistent-type-definitions)
