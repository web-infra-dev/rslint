# prefer-enum-initializers

## Rule Details

Require each enum member to have an explicitly initialized value.

TypeScript enum members default to a sequential numeric value when no initializer is provided. Relying on that implicit numbering makes the enum fragile — inserting, deleting, or reordering members silently shifts the numeric value of every later member, which can break code that persisted those numbers (in storage, in a network protocol, in serialized output). Explicit initializers make the value of each member a documented, stable choice. For each uninitialized member, this rule reports a diagnostic and offers three suggestions: initialize to the member's current index, to the index plus one, or to a string matching the member's name.

Examples of **incorrect** code for this rule:

```typescript
enum Direction {
  Up,
  Down,
}

enum Status {
  Open = 1,
  Close,
}
```

Examples of **correct** code for this rule:

```typescript
enum Direction {
  Up = 1,
  Down = 2,
}

enum Status {
  Open = 'Open',
  Close = 'Close',
}

enum Empty {}
```

## Original Documentation

- [typescript-eslint prefer-enum-initializers](https://typescript-eslint.io/rules/prefer-enum-initializers)
