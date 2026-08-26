# unified-signatures

Disallow overloads that can be combined into one signature with a union,
optional parameter, or rest parameter.

## Rule details

Multiple overloads make an API harder to read when they differ only in a
single parameter type or in one parameter that can be omitted. This rule
reports those signatures and recommends expressing the difference directly.

Examples of **incorrect** code for this rule:

```typescript
function parse(value: string): void;
function parse(value: number): void;

interface Factory {
  create(): Result;
  create(options?: Options): Result;
}
```

Examples of **correct** code for this rule:

```typescript
function parse(value: string | number): void;

interface Factory {
  create(options?: Options): Result;
}

function convert(value: string): string;
function convert(value: number): number;
```

The last pair is allowed because its return types differ.

## Options

The rule accepts an optional object:

```json
{
  "ignoreDifferentlyNamedParameters": false,
  "ignoreOverloadsWithDifferentJSDoc": false
}
```

### `ignoreDifferentlyNamedParameters`

When `true`, overloads whose corresponding parameters have different names
are not combined.

### `ignoreOverloadsWithDifferentJSDoc`

When `true`, overloads with different preceding block comments are not
combined.

## Original documentation

- [typescript-eslint: unified-signatures](https://typescript-eslint.io/rules/unified-signatures)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/unified-signatures.ts)
- [Tests](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/tests/rules/unified-signatures.test.ts)
