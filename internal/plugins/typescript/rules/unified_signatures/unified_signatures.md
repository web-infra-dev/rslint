# unified-signatures

## Rule Details

Disallow two overloads that could be unified into one with a union or optional parameter. Function overloads that only differ in one parameter's type can typically be replaced with a single signature using a union type. Similarly, overloads where one signature has an additional optional parameter can be unified into one signature with that parameter marked as optional. Reducing overloads makes the API simpler and easier to understand.

Examples of **incorrect** code for this rule:

```typescript
function foo(x: number): void;
function foo(x: string): void;

interface Bar {
  method(x: number): void;
  method(x: string): void;
}
```

Examples of **correct** code for this rule:

```typescript
function foo(x: number | string): void;

interface Bar {
  method(x: number | string): void;
}

// Different enough signatures are OK
function baz(x: number): number;
function baz(x: string): string;
```

## Original Documentation

- [typescript-eslint unified-signatures](https://typescript-eslint.io/rules/unified-signatures)
