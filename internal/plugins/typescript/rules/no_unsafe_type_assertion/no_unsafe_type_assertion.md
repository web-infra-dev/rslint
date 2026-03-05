# no-unsafe-type-assertion

## Rule Details

Disallow type assertions that narrow a type.

Type assertions (`as` or angle-bracket syntax) that narrow the type of an expression are unsafe because they tell TypeScript to trust the developer rather than the type system. This rule flags assertions from `any` types, assertions to `any` types, assertions to unconstrained type parameters, and assertions that narrow a type to a more specific one.

Examples of **incorrect** code for this rule:

```typescript
const x = {} as string;
const y = value as any;
const z = 1 as any as string;

function fn<T>(x: string) {
  return x as T; // T is unconstrained
}
```

Examples of **correct** code for this rule:

```typescript
const x = 'hello' as string; // same type, no narrowing
const y = someValue as unknown; // widening is safe

function fn(x: string | number) {
  if (typeof x === 'string') {
    const str: string = x; // use type guards instead
  }
}
```

## Original Documentation

- [typescript-eslint no-unsafe-type-assertion](https://typescript-eslint.io/rules/no-unsafe-type-assertion)
