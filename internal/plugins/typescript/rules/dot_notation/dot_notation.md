# dot-notation

## Rule Details

Enforce dot notation whenever possible. This is the TypeScript-enhanced version of the ESLint `dot-notation` rule. Dot notation is generally preferred over bracket notation for readability. The TypeScript version adds options to allow bracket notation for private/protected class members and index signature properties.

The rule provides autofixes to convert bracket notation (`obj["prop"]`) to dot notation (`obj.prop`) when safe.

Examples of **incorrect** code for this rule:

```typescript
const x = obj['foo'];
const y = obj['bar'];
```

Examples of **correct** code for this rule:

```typescript
const x = obj.foo;
const y = obj.bar;
const z = obj['some-kebab-case']; // not a valid identifier
const w = obj[dynamicKey]; // computed access
```

## Original Documentation

- [typescript-eslint dot-notation](https://typescript-eslint.io/rules/dot-notation)
