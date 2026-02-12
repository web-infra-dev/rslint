# no-explicit-any

## Rule Details

Disallows the `any` type. Using the `any` type defeats the purpose of TypeScript's type system. When `any` is used, all compiler type checks around that value are ignored. This rule reports on explicit uses of the `any` keyword as a type annotation. It suggests using `unknown` for safe type assertions, `never` for generic type parameters that should not be used, and `PropertyKey` instead of `keyof any`.

Examples of **incorrect** code for this rule:

```typescript
const age: any = 'seventeen';

function greet(): any {}

function foo(arg: any): void {}

const key: keyof any = 'name';
```

Examples of **correct** code for this rule:

```typescript
const age: number = 17;

function greet(): string {
  return 'hello';
}

function foo(arg: unknown): void {}

const key: PropertyKey = 'name';
```

## Original Documentation

- [typescript-eslint no-explicit-any](https://typescript-eslint.io/rules/no-explicit-any)
