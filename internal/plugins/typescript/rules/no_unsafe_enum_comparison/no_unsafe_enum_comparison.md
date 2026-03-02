# no-unsafe-enum-comparison

## Rule Details

Disallow comparing an enum value with a non-enum value.

TypeScript enums are a special type that represent a set of named constants. Comparing enum values against raw literals or values of a different enum type is often a mistake and can lead to subtle bugs, since enums in TypeScript have their own type identity.

Examples of **incorrect** code for this rule:

```typescript
enum Fruit {
  Apple,
  Banana,
}

declare const fruit: Fruit;
fruit === 0;
fruit === 'Apple';

enum Vegetable {
  Carrot,
}
fruit === Vegetable.Carrot;
```

Examples of **correct** code for this rule:

```typescript
enum Fruit {
  Apple,
  Banana,
}

declare const fruit: Fruit;
fruit === Fruit.Apple;
fruit === Fruit.Banana;
```

## Original Documentation

- [typescript-eslint no-unsafe-enum-comparison](https://typescript-eslint.io/rules/no-unsafe-enum-comparison)
