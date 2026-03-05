# no-unsafe-assignment

## Rule Details

Disallow assigning a value with type `any` to variables and properties.

Assigning an `any`-typed value to a variable or property circumvents TypeScript's type checking. This rule flags unsafe assignments including direct assignments, variable declarations, array destructuring with `any` elements, object destructuring with `any` properties, and array spreads of `any`-typed values.

Examples of **incorrect** code for this rule:

```typescript
const x = 1 as any;
const [y] = [1] as any;
const [z] = [1 as any];

function fn(arg: any) {
  const val: string = arg;
}
```

Examples of **correct** code for this rule:

```typescript
const x = 1;
const [y] = [1];
const val: unknown = someAnyValue;

function fn(arg: string) {
  const val: string = arg;
}
```

## Original Documentation

- [typescript-eslint no-unsafe-assignment](https://typescript-eslint.io/rules/no-unsafe-assignment)
