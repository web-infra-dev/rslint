# no-duplicate-enum-values

## Rule Details

Disallows duplicate enum member values. Enum members that share the same literal value (number, string, or template literal) are almost always a mistake. This rule checks for duplicate initializer values within a single enum declaration.

Examples of **incorrect** code for this rule:

```typescript
enum Direction {
  Up = 0,
  Down = 0,
}

enum Color {
  Red = 'red',
  Blue = 'red',
}

enum Num {
  A = 1,
  B = -1,
  C = 1,
}
```

Examples of **correct** code for this rule:

```typescript
enum Direction {
  Up = 0,
  Down = 1,
  Left = 2,
  Right = 3,
}

enum Color {
  Red = 'red',
  Blue = 'blue',
  Green = 'green',
}
```

## Original Documentation

- [typescript-eslint no-duplicate-enum-values](https://typescript-eslint.io/rules/no-duplicate-enum-values)
