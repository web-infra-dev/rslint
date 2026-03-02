# adjacent-overload-signatures

## Rule Details

Require that function overload signatures be consecutive. When function overload signatures are not adjacent to each other, it can be difficult to read and understand the complete set of overloads for a given function.

Examples of **incorrect** code for this rule:

```typescript
declare function foo(s: string): void;
declare function bar(): void;
declare function foo(n: number): void;

class MyClass {
  foo(s: string): void;
  bar(): void;
  foo(n: number): void;
}
```

Examples of **correct** code for this rule:

```typescript
declare function foo(s: string): void;
declare function foo(n: number): void;
declare function bar(): void;

class MyClass {
  foo(s: string): void;
  foo(n: number): void;
  bar(): void;
}
```

## Original Documentation

- [typescript-eslint adjacent-overload-signatures](https://typescript-eslint.io/rules/adjacent-overload-signatures)
