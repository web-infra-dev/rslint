# prefer-readonly-parameter-types

Requires function parameters to have deeply readonly types, preventing code
from accidentally mutating values supplied by callers.

## Rule Details

Examples of **incorrect** code for this rule:

```typescript
function consume(items: string[]) {}
function update(options: { enabled: boolean }) {}
function register(value: Set<string>) {}
```

Examples of **correct** code for this rule:

```typescript
function consume(items: readonly string[]) {}
function update(options: Readonly<{ enabled: boolean }>) {}
function register(value: ReadonlySet<string>) {}
function format(value: string) {}
```

The rule checks nested property and element types, so a readonly outer property
whose value is mutable is still reported.

## Options

- `allow` (default `[]`): type specifiers that the rule should accept without
  checking their readonlyness. String names and `file`, `lib`, or `package`
  specifiers are supported.
- `checkParameterProperties` (default `true`): check TypeScript constructor
  parameter properties.
- `ignoreInferredTypes` (default `false`): skip parameters without an explicit
  type annotation.
- `treatMethodsAsReadonly` (default `false`): treat method declarations as
  readonly while checking object types. This is useful for types such as
  `ReadonlySet` and `ReadonlyMap`.

## Original Documentation

- [typescript-eslint: prefer-readonly-parameter-types](https://typescript-eslint.io/rules/prefer-readonly-parameter-types)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/prefer-readonly-parameter-types.ts)
