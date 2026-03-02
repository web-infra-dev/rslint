# no-empty-interface

## Rule Details

Disallows empty interface declarations. An empty interface with no members is equivalent to the empty object type `{}`. An empty interface that extends a single interface is equivalent to a type alias of that interface. In both cases, the interface declaration adds unnecessary indirection and can be replaced with a simpler construct.

Examples of **incorrect** code for this rule:

```typescript
// Empty interface is equivalent to {}
interface Foo {}

// Equivalent to: type Bar = Baz
interface Bar extends Baz {}
```

Examples of **correct** code for this rule:

```typescript
// Interface with members
interface Foo {
  name: string;
}

// Interface extending multiple interfaces
interface Bar extends Baz, Qux {}

// Type alias instead of empty extending interface
type Bar = Baz;
```

## Original Documentation

- [typescript-eslint no-empty-interface](https://typescript-eslint.io/rules/no-empty-interface)
