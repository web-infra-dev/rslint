# switch-exhaustiveness-check

## Rule Details

Require switch statements over union types to be exhaustive. When switching over a union type (such as a discriminated union or enum), it is easy to forget to handle all possible cases. This rule ensures that every possible value of the union is handled with a `case` clause, either explicitly or via a `default` clause, preventing runtime errors from unhandled values.

Examples of **incorrect** code for this rule:

```typescript
type Direction = 'north' | 'south' | 'east' | 'west';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    case 'south':
      break;
    // 'east' and 'west' are not handled
  }
}
```

Examples of **correct** code for this rule:

```typescript
type Direction = 'north' | 'south' | 'east' | 'west';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    case 'south':
      break;
    case 'east':
      break;
    case 'west':
      break;
  }
}
```

## Original Documentation

- [typescript-eslint switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
