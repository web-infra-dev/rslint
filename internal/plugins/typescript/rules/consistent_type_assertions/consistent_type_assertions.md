# consistent-type-assertions

## Rule Details

Enforce consistent usage of type assertions. TypeScript provides two syntaxes for type assertions: `as` expressions (`value as Type`) and angle-bracket syntax (`<Type>value`). This rule enforces a consistent style and can also restrict type assertions on object and array literals.

The `assertionStyle` option supports `"as"` (default), `"angle-bracket"`, and `"never"`. Additional options `objectLiteralTypeAssertions` and `arrayLiteralTypeAssertions` control whether assertions on literals are allowed.

Examples of **incorrect** code for this rule (with default `"as"` option):

```typescript
const x = <string>value;
const y = <number>42;
```

Examples of **correct** code for this rule (with default `"as"` option):

```typescript
const x = value as string;
const y = 42 as number;
const z = value as const;
```

## Original Documentation

- [typescript-eslint consistent-type-assertions](https://typescript-eslint.io/rules/consistent-type-assertions)
