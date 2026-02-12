# consistent-type-definitions

## Rule Details

Enforce type definitions to consistently use either `interface` or `type`. TypeScript provides two ways to define object types: `interface` declarations and `type` alias declarations with object literal types. This rule enforces one style for consistency.

The rule supports two modes: `"interface"` (default) prefers interfaces over type literals, and `"type"` prefers type aliases over interfaces.

Examples of **incorrect** code for this rule (with default `"interface"` option):

```typescript
type Foo = {
  name: string;
  age: number;
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
```

## Original Documentation

- [typescript-eslint consistent-type-definitions](https://typescript-eslint.io/rules/consistent-type-definitions)
