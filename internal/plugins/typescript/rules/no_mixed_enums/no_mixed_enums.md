# no-mixed-enums

## Rule Details

Disallows enums that mix string and number member values. TypeScript enums can contain either string values or number values, but mixing both types in a single enum declaration can lead to confusing behavior. For example, number enum members have reverse mappings while string enum members do not. This rule ensures all members in an enum use the same type of initializer.

Examples of **incorrect** code for this rule:

```typescript
enum Status {
  Active = 0,
  Inactive = 'inactive',
}

enum Mixed {
  A = 0,
  B = 1,
  C = 'c',
}
```

Examples of **correct** code for this rule:

```typescript
enum Status {
  Active = 0,
  Inactive = 1,
}

enum Color {
  Red = 'red',
  Blue = 'blue',
  Green = 'green',
}

enum Direction {
  Up,
  Down,
  Left,
  Right,
}
```

## Original Documentation

- [typescript-eslint no-mixed-enums](https://typescript-eslint.io/rules/no-mixed-enums)
