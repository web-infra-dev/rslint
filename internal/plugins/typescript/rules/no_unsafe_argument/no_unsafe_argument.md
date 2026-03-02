# no-unsafe-argument

## Rule Details

Disallow calling a function with a value with type `any`.

The `any` type in TypeScript is a dangerous escape hatch from the type system. Passing an `any`-typed value as an argument to a function defeats the purpose of the parameter's type safety. This rule flags cases where `any`-typed values are passed as arguments, including spread arguments.

Examples of **incorrect** code for this rule:

```typescript
declare function foo(arg: string): void;
const anyVal: any = 'hello';
foo(anyVal);

declare function bar(...args: string[]): void;
const anyArray: any[] = [];
bar(...anyArray);
```

Examples of **correct** code for this rule:

```typescript
declare function foo(arg: string): void;
foo('hello');

declare function bar(arg: any): void;
bar(value); // parameter already typed as any

declare function baz(...args: string[]): void;
const strArray: string[] = [];
baz(...strArray);
```

## Original Documentation

- [typescript-eslint no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)
