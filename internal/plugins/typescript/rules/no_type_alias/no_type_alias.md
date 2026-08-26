# no-type-alias

## Rule Details

Disallow type aliases.

Examples of **incorrect** code for this rule:

```ts
type Name = string;
```

Examples of **correct** code:

```ts
interface Name {
  value: string;
}
```

## Original Documentation

- [typescript-eslint: no-type-alias](https://typescript-eslint.io/rules/no-type-alias)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-type-alias.ts)
