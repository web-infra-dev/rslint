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

## Compatibility Notes

The rule is aligned with `typescript-eslint` v8.69.0 for diagnostic selection,
message IDs, ranges, suggestions, and autofixes. Literal-assertion diagnostics
use context-aware descriptions: an untyped variable initializer recommends a
type annotation or the `satisfies` operator, while other expressions recommend
only `satisfies`. Upstream always recommends `const x: T = ...`, including in
return expressions, class fields, and other positions where that replacement
cannot be applied.

## Original Documentation

- [typescript-eslint: consistent-type-assertions](https://typescript-eslint.io/rules/consistent-type-assertions)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.69.0/packages/eslint-plugin/src/rules/consistent-type-assertions.ts)
