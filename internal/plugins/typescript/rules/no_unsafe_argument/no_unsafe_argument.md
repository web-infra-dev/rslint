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

## Differences from typescript-eslint

Spread handling matches typescript-eslint: this rule checks `any`, `any[]`, and concrete tuple spreads. Other iterable spreads, including `Set<any>`, arrays with generic element types such as `Set<any>[]`, and generic `Parameters<T>` spreads, are ignored without advancing the parameter position used to check later arguments.

## Original Documentation

- [typescript-eslint: no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unsafe-argument.ts)
