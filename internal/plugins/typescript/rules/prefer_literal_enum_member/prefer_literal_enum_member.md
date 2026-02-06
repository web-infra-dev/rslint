# prefer-literal-enum-member

## Rule Details

Require that all enum members are literal values to prevent unintended enum member values.

TypeScript allows the value of an enum member to be many different kinds of valid JavaScript expressions. However, because enums create their own scope whereby each enum member becomes a variable in that scope, developers are often surprised at the result of using non-literal values. Explicit enum values should only be literals (string or number).

Examples of **incorrect** code for this rule:

```typescript
const impliedValueIsString = 'a';

enum Foo {
  ReadWrite = 2 | 4,
  Read = getMask(),
  Write = impliedValueIsString,
  Shift = 1 << 1,
}
```

Examples of **correct** code for this rule:

```typescript
enum Foo {
  Read,
  Write,
  Shift = 1,
  Name = 'hello',
  Combined = 1 | 2, // only with allowBitwiseExpressions
}
```

## Options

### `allowBitwiseExpressions`

When set to `true`, allows using bitwise expressions in enum initializers, which is common in flag-style enums.

```json
{
  "@typescript-eslint/prefer-literal-enum-member": [
    "warn",
    { "allowBitwiseExpressions": true }
  ]
}
```

## Original Documentation

- [typescript-eslint prefer-literal-enum-member](https://typescript-eslint.io/rules/prefer-literal-enum-member)
