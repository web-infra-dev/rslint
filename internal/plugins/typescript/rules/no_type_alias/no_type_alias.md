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

- https://typescript-eslint.io/rules/no-type-alias
