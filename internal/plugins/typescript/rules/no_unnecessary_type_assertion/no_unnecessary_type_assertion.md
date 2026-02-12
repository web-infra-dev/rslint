# no-unnecessary-type-assertion

## Rule Details

Disallow type assertions that do not change the type of an expression.

Type assertions (`as` expressions, angle-bracket syntax, and non-null assertions `!`) that don't actually change the type of an expression are unnecessary and add noise to the code. This includes non-null assertions on values that are already non-nullable.

Examples of **incorrect** code for this rule:

```typescript
const foo = 3;
const bar = foo!; // foo is already non-nullable

const str = 'hello' as string; // already a string

declare const value: number;
const num = value as number; // already a number
```

Examples of **correct** code for this rule:

```typescript
const foo: number | undefined = getValue();
const bar = foo!; // non-null assertion is meaningful

const value = someUnknown as string; // type narrowing

const x = 3 as const; // const assertions are allowed
```

## Original Documentation

- [typescript-eslint no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
