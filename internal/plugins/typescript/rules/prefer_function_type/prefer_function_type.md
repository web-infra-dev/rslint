# prefer-function-type

## Rule Details

Enforce using function types instead of interfaces with call signatures.

TypeScript allows declaring a function type in two ways: as a function type (`() => string`) or as an object type with a single call signature (`{ (): string }`). The former is more concise, so this rule reports the latter on interfaces and type literals whose only member is a single call signature (or construct signature) and offers an autofix to rewrite them.

Examples of **incorrect** code for this rule:

```typescript
interface Example {
  (): string;
}

function foo(example: { (): number }): number {
  return example();
}

interface ReturnsSelf {
  // returns `this` directly, not a `this` type parameter
  (arg: string): this;
}
```

Examples of **correct** code for this rule:

```typescript
type Example = () => string;

function foo(example: () => number): number {
  return example();
}

// `this` parameter via a generic to avoid the
// `unexpectedThisOnFunctionOnlyInterface` warning:
type ReturnsSelf<Self> = (this: Self, arg: string) => Self;

// Has additional properties besides the call signature:
function foo(bar: { (): string; baz: number }): string {
  return bar();
}

// Multiple call signatures (overloads):
interface Overloaded {
  (data: string): number;
  (id: number): string;
}

// `extends` something other than `Function`:
interface Foo {
  bar: string;
}
interface Bar extends Foo {
  (): void;
}
```

## Options

This rule has no options.

## When Not To Use It

Disable this rule if you prefer interfaces or object type literals for stylistic consistency, or if you rely on declaration merging (e.g. augmenting the global `Function` interface) — those cases occasionally produce false positives that can be silenced with an inline disable comment.

## Original Documentation

- [typescript-eslint prefer-function-type](https://typescript-eslint.io/rules/prefer-function-type)
