# no-unsafe-return

## Rule Details

Disallow returning a value with type `any` from a function.

Returning an `any`-typed value from a function that has a typed return type undermines type safety, since the caller will trust the return type but the actual value has no type guarantees. This rule also flags returning `any[]` and `Promise<any>` where the function expects more specific types.

Examples of **incorrect** code for this rule:

```typescript
function foo(): string {
  return 1 as any;
}

function bar(): string[] {
  return [] as any[];
}

const fn = (): Set<string> => new Set<any>();
```

Examples of **correct** code for this rule:

```typescript
function foo(): string {
  return 'hello';
}

function bar(): unknown {
  return 1 as any; // returning any to unknown is allowed
}

function baz(): any {
  return 1 as any; // explicit any return type is allowed
}
```

## Original Documentation

- [typescript-eslint no-unsafe-return](https://typescript-eslint.io/rules/no-unsafe-return)
